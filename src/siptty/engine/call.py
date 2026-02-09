"""PhoneCall — pjsua2 Call wrapper with state events and media plumbing."""

from __future__ import annotations

import logging
import time
from collections.abc import Callable
from typing import Any

from siptty.engine import PJSUA2_AVAILABLE
from siptty.engine.events import CallStateEvent

if PJSUA2_AVAILABLE:
    import pjsua2 as pj

log = logging.getLogger(__name__)

# Map pjsua2 call state ints to our state strings.
_CALL_STATE_MAP: dict[int, str] = {}
if PJSUA2_AVAILABLE:
    _CALL_STATE_MAP = {
        pj.PJSIP_INV_STATE_NULL: "disconnected",
        pj.PJSIP_INV_STATE_CALLING: "calling",
        pj.PJSIP_INV_STATE_INCOMING: "incoming",
        pj.PJSIP_INV_STATE_EARLY: "early",
        pj.PJSIP_INV_STATE_CONNECTING: "connecting",
        pj.PJSIP_INV_STATE_CONFIRMED: "confirmed",
        pj.PJSIP_INV_STATE_DISCONNECTED: "disconnected",
    }


class PhoneCall(pj.Call if PJSUA2_AVAILABLE else object):  # type: ignore[misc]
    """Wraps a pjsua2 Call.  Fires CallStateEvent on state changes."""

    def __init__(
        self,
        account: Any,  # PhoneAccount
        call_id: int = -1,
        *,
        direction: str = "outbound",
        event_callback: Callable[[Any], None] | None = None,
        on_disconnected: Callable[[int], None] | None = None,
    ) -> None:
        if call_id >= 0:
            super().__init__(account, call_id)
        else:
            super().__init__(account)
        self._event_callback = event_callback
        self._on_disconnected = on_disconnected
        self._direction = direction
        self._start_time: float = time.monotonic()
        self._remote_uri: str = ""
        self._is_hold = False
        # Audio plumbing state
        self._audio_player: Any | None = None
        self._audio_recorder: Any | None = None

    @property
    def call_id_key(self) -> int:
        """Return the pjsua2 call ID for use as dict key."""
        try:
            return self.getId()
        except Exception:
            return -1

    @property
    def duration(self) -> float:
        """Elapsed seconds since call start."""
        return time.monotonic() - self._start_time

    # ------------------------------------------------------------------
    # pjsua2 callbacks
    # ------------------------------------------------------------------

    def onCallState(self, prm: pj.OnCallStateParam) -> None:  # noqa: N802
        """Called by pjsua2 when the call state changes."""
        try:
            ci = self.getInfo()
            state_str = _CALL_STATE_MAP.get(ci.state, "unknown")
            self._remote_uri = ci.remoteUri

            # Track hold state locally
            if self._is_hold and state_str == "confirmed":
                self._is_hold = False

            event = CallStateEvent(
                call_id=ci.id,
                state=state_str,
                remote_uri=ci.remoteUri,
                duration=self.duration,
                direction=self._direction,
            )
            if self._event_callback:
                self._event_callback(event)

            # Cleanup on disconnect
            if ci.state == pj.PJSIP_INV_STATE_DISCONNECTED:
                self._cleanup_audio()
                if self._on_disconnected:
                    self._on_disconnected(ci.id)
        except Exception:
            log.exception("Error in onCallState")

    def onCallMediaState(self, prm: pj.OnCallMediaStateParam) -> None:  # noqa: N802
        """Called by pjsua2 when media state changes — wire up audio."""
        try:
            ci = self.getInfo()
            for mi_idx in range(len(ci.media)):
                if ci.media[mi_idx].type == pj.PJMEDIA_TYPE_AUDIO:
                    media = self.getAudioMedia(mi_idx)
                    if ci.media[mi_idx].status == pj.PJSUA_CALL_MEDIA_ACTIVE:
                        # Connect to sound device (or null dev)
                        aud_mgr = pj.Endpoint.instance().audDevManager()
                        media.startTransmit(aud_mgr.getPlaybackDevMedia())
                        aud_mgr.getCaptureDevMedia().startTransmit(media)
                        log.info("Audio connected for call %d", ci.id)
        except Exception:
            log.exception("Error in onCallMediaState")

    # ------------------------------------------------------------------
    # Call operations
    # ------------------------------------------------------------------

    def do_hold(self) -> None:
        """Put the call on hold."""
        prm = pj.CallOpParam(True)
        self.setHold(prm)
        self._is_hold = True

    def do_unhold(self) -> None:
        """Resume a held call."""
        prm = pj.CallOpParam(True)
        prm.opt.flag = pj.PJSUA_CALL_UNHOLD
        self.reinvite(prm)

    def do_hangup(self, code: int = 200) -> None:
        """Hang up the call."""
        prm = pj.CallOpParam(True)
        prm.statusCode = code
        self.hangup(prm)

    def do_answer(self, code: int = 200) -> None:
        """Answer the call with given status code."""
        prm = pj.CallOpParam(True)
        prm.statusCode = code
        self.answer(prm)

    def do_send_dtmf(self, digits: str) -> None:
        """Send DTMF digits via RFC 4733 telephone-event."""
        valid = set("0123456789*#ABCDabcd")
        for ch in digits:
            if ch not in valid:
                raise ValueError(f"Invalid DTMF digit: {ch!r}")
        self.dialDtmf(digits)

    # ------------------------------------------------------------------
    # Audio helpers
    # ------------------------------------------------------------------

    def _cleanup_audio(self) -> None:
        """Stop and cleanup any active audio player/recorder."""
        self._audio_player = None
        self._audio_recorder = None
