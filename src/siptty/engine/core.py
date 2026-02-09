"""SipEngine — pjsua2 lifecycle wrapper."""

from __future__ import annotations

import logging
from collections.abc import Callable
from typing import Any

from siptty.config.models import AccountConfig, AppConfig
from siptty.engine import PJSUA2_AVAILABLE
from siptty.engine.events import CallStateEvent

if PJSUA2_AVAILABLE:
    import pjsua2 as pj

    from siptty.engine.account import PhoneAccount
    from siptty.engine.call import PhoneCall
    from siptty.engine.trace import SipTraceWriter

log = logging.getLogger(__name__)


class SipEngine:
    """Thin wrapper around pjsua2 :class:`pj.Endpoint` with lifecycle management."""

    def __init__(self, event_callback: Callable[[Any], None]) -> None:
        if not PJSUA2_AVAILABLE:
            raise RuntimeError("pjsua2 is not available – cannot create SipEngine")
        self._callback = event_callback
        self._ep: pj.Endpoint | None = None
        self._accounts: dict[str, PhoneAccount] = {}
        self._calls: dict[int, PhoneCall] = {}
        self._started = False
        self._transport_id: int = -1
        self._trace_writer: SipTraceWriter | None = None

    # ------------------------------------------------------------------
    # Lifecycle
    # ------------------------------------------------------------------

    def start(self, config: AppConfig) -> None:
        """Create the pjsua2 Endpoint, initialise and start the library."""
        if self._started:
            raise RuntimeError("SipEngine is already started")

        ep = pj.Endpoint()
        ep.libCreate()

        try:
            # ---- Endpoint configuration ----
            ep_cfg = pj.EpConfig()
            ep_cfg.uaConfig.userAgent = config.general.user_agent

            # SIP trace writer
            self._trace_writer = SipTraceWriter(self._callback)
            ep_cfg.logConfig.writer = self._trace_writer
            ep_cfg.logConfig.level = max(config.general.log_level, 5)
            ep_cfg.logConfig.consoleLevel = 0

            ep.libInit(ep_cfg)

            # ---- UDP transport (ephemeral port) ----
            tp_cfg = pj.TransportConfig()
            tp_cfg.port = 0
            self._transport_id = ep.transportCreate(pj.PJSIP_TRANSPORT_UDP, tp_cfg)

            # ---- Audio device ----
            if config.audio.mode == "null":
                ep.audDevManager().setNullDev()

            ep.libStart()
        except Exception:
            try:
                ep.libDestroy()
            except Exception:  # noqa: S110
                pass
            raise

        self._ep = ep
        self._started = True
        log.info("SipEngine started (ua=%s)", config.general.user_agent)

    def stop(self) -> None:
        """Destroy the pjsua2 Endpoint.  Safe to call multiple times."""
        if not self._started or self._ep is None:
            return

        # Remove accounts first
        for name in list(self._accounts):
            try:
                self.remove_account(name)
            except Exception:
                log.exception("Error removing account %s during stop", name)

        try:
            self._ep.libDestroy()
        except Exception:
            log.exception("Error during libDestroy")
        finally:
            self._ep = None
            self._started = False
            self._accounts.clear()
            self._calls.clear()
            self._trace_writer = None
            log.info("SipEngine stopped")

    # ------------------------------------------------------------------
    # Account management
    # ------------------------------------------------------------------

    def add_account(self, cfg: AccountConfig) -> str:
        """Create and register a SIP account.  Returns the account name."""
        if not self._started:
            raise RuntimeError("SipEngine not started")
        if cfg.name in self._accounts:
            raise ValueError(f"Account '{cfg.name}' already exists")

        account = PhoneAccount(
            cfg,
            event_callback=self._callback,
            incoming_call_factory=self._on_incoming_call,
        )
        account.create_and_register()
        self._accounts[cfg.name] = account
        log.info("Added account '%s' (%s)", cfg.name, cfg.sip_uri)
        return cfg.name

    def remove_account(self, account_id: str) -> None:
        """Unregister and remove a SIP account."""
        account = self._accounts.pop(account_id, None)
        if account is None:
            return
        try:
            account.setRegistration(False)
        except Exception:
            log.exception("Error unregistering %s", account_id)
        try:
            account.shutdown()
        except Exception:
            log.exception("Error shutting down account %s", account_id)
        log.info("Removed account '%s'", account_id)

    # ------------------------------------------------------------------
    # Outbound calls
    # ------------------------------------------------------------------

    def dial(
        self,
        account_id: str,
        uri: str,
        headers: dict[str, str] | None = None,
    ) -> int:
        """Place an outbound call.  Returns the call ID."""
        if not self._started:
            raise RuntimeError("SipEngine not started")

        account = self._accounts.get(account_id)
        if account is None:
            raise ValueError(f"Account '{account_id}' not found")

        call = PhoneCall(
            account,
            direction="outbound",
            event_callback=self._callback,
            on_disconnected=self._on_call_disconnected,
        )

        prm = pj.CallOpParam(True)
        prm.opt.audioCount = 1
        prm.opt.videoCount = 0

        # Custom SIP headers
        if headers:
            for name, value in headers.items():
                hdr = pj.SipHeader()
                hdr.hName = name
                hdr.hValue = value
                prm.txOption.headers.append(hdr)

        call.makeCall(uri, prm)
        call_id = call.call_id_key
        self._calls[call_id] = call
        log.info("Outbound call %d to %s via %s", call_id, uri, account_id)
        return call_id

    # ------------------------------------------------------------------
    # Inbound call handling
    # ------------------------------------------------------------------

    def _on_incoming_call(
        self, account: PhoneAccount, call_id: int
    ) -> None:
        """Factory method called by PhoneAccount.onIncomingCall."""
        call = PhoneCall(
            account,
            call_id=call_id,
            direction="inbound",
            event_callback=self._callback,
            on_disconnected=self._on_call_disconnected,
        )
        self._calls[call_id] = call
        # Fire incoming event
        try:
            ci = call.getInfo()
            event = CallStateEvent(
                call_id=ci.id,
                state="incoming",
                remote_uri=ci.remoteUri,
                duration=0.0,
                direction="inbound",
            )
            self._callback(event)
        except Exception:
            log.exception("Error handling incoming call %d", call_id)

    def answer(self, call_id: int, code: int = 200) -> None:
        """Answer an incoming call."""
        call = self._calls.get(call_id)
        if call is None:
            raise ValueError(f"Call {call_id} not found")
        call.do_answer(code)

    def reject(self, call_id: int, code: int = 486) -> None:
        """Reject an incoming call."""
        call = self._calls.get(call_id)
        if call is None:
            raise ValueError(f"Call {call_id} not found")
        call.do_answer(code)

    # ------------------------------------------------------------------
    # Call control
    # ------------------------------------------------------------------

    def hangup(self, call_id: int) -> None:
        """Hang up a call."""
        call = self._calls.get(call_id)
        if call is None:
            raise ValueError(f"Call {call_id} not found")
        call.do_hangup()

    def hold(self, call_id: int) -> None:
        """Put a call on hold."""
        call = self._calls.get(call_id)
        if call is None:
            raise ValueError(f"Call {call_id} not found")
        call.do_hold()

    def resume(self, call_id: int) -> None:
        """Resume a held call."""
        call = self._calls.get(call_id)
        if call is None:
            raise ValueError(f"Call {call_id} not found")
        call.do_unhold()

    def send_dtmf(self, call_id: int, digits: str) -> None:
        """Send DTMF digits on an active call."""
        call = self._calls.get(call_id)
        if call is None:
            raise ValueError(f"Call {call_id} not found")
        call.do_send_dtmf(digits)

    # ------------------------------------------------------------------
    # Internal helpers
    # ------------------------------------------------------------------

    def _on_call_disconnected(self, call_id: int) -> None:
        """Remove a call from tracking when it disconnects."""
        self._calls.pop(call_id, None)
        log.info("Call %d removed (disconnected)", call_id)
