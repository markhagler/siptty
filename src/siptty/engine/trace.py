"""SIP trace log writer — captures raw SIP messages from pjsua2 logs."""

from __future__ import annotations

import logging
import re
import time
from collections.abc import Callable

from siptty.engine import PJSUA2_AVAILABLE
from siptty.engine.events import SipTraceEvent

if PJSUA2_AVAILABLE:
    import pjsua2 as pj

log = logging.getLogger(__name__)

# pjsip logs transmitted packets with a header like:
#   "TX 1234 bytes Request msg INVITE/cseq=1 (tdta0x...) to UDP 10.0.0.1:5060:\n"
# and received packets with:
#   "RX 567 bytes Request msg REGISTER/cseq=2 (rdata0x...) from UDP 10.0.0.1:5060:\n"
# The full SIP message follows on subsequent lines until "--end msg--".
_TX_HEADER_RE = re.compile(r"TX\s+\d+\s+bytes\s+.+\s+to\s+")
_RX_HEADER_RE = re.compile(r"RX\s+\d+\s+bytes\s+.+\s+from\s+")
_END_MSG_RE = re.compile(r"--end msg--")


class SipTraceWriter(pj.LogWriter if PJSUA2_AVAILABLE else object):  # type: ignore[misc]
    """Custom :class:`pj.LogWriter` that captures SIP messages from pjsua2 log output.

    pjsua2 emits full SIP messages at log level 5.  Transmitted messages are
    preceded by a line matching ``TX <n> bytes … to …`` and received messages
    by ``RX <n> bytes … from …``.  The message body follows over multiple
    subsequent log lines and is terminated by a line containing ``--end msg--``.

    This writer accumulates those lines and, on seeing the end-of-message
    sentinel, constructs a :class:`SipTraceEvent` and delivers it via the
    supplied *callback*.
    """

    def __init__(self, callback: Callable[[SipTraceEvent], None]) -> None:
        super().__init__()
        self._callback = callback
        # Accumulation state – accessed only from pjsip's log-writer thread.
        self._direction: str | None = None
        self._lines: list[str] = []

    # ------------------------------------------------------------------
    # pj.LogWriter interface
    # ------------------------------------------------------------------

    def write(self, entry: pj.LogEntry) -> None:  # type: ignore[override]
        """Called by pjsua2 for every log line.

        We inspect each line to detect the start / end of a SIP message block
        and accumulate the intermediate lines.
        """
        msg: str = entry.msg

        # Check for the start of a new SIP message block.
        if _TX_HEADER_RE.search(msg):
            # If we were already accumulating (shouldn't happen, but be safe),
            # discard the incomplete block.
            self._direction = "send"
            self._lines = []
            return

        if _RX_HEADER_RE.search(msg):
            self._direction = "recv"
            self._lines = []
            return

        # If we're not inside a message block, nothing to do.
        if self._direction is None:
            return

        # Check for the end-of-message sentinel.
        if _END_MSG_RE.search(msg):
            self._emit()
            return

        # Accumulate the line (strip the trailing newline pjsip usually adds).
        self._lines.append(msg.rstrip("\n"))

    # ------------------------------------------------------------------
    # Internal helpers
    # ------------------------------------------------------------------

    def _emit(self) -> None:
        """Construct a :class:`SipTraceEvent` from the accumulated lines and fire callback."""
        direction = self._direction
        lines = self._lines

        # Reset accumulation state *before* invoking the callback so that any
        # re-entrant log writes don't corrupt state.
        self._direction = None
        self._lines = []

        if not lines or direction is None:
            return

        full_message = "\n".join(lines)
        event = SipTraceEvent(
            direction=direction,
            message=full_message,
            timestamp=time.time(),
        )

        try:
            self._callback(event)
        except Exception:
            # Never let a callback error propagate into pjsip's log thread.
            log.exception("SipTraceWriter callback failed")
