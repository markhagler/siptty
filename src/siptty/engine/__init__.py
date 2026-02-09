"""SIP engine layer — wraps pjsua2."""

from siptty.engine.events import CallStateEvent, RegStateEvent, SipTraceEvent

try:
    import pjsua2 as pj

    PJSUA2_AVAILABLE = True
except ImportError:
    pj = None  # type: ignore[assignment]
    PJSUA2_AVAILABLE = False

# Deferred import to avoid referencing pj at module level when unavailable.
from siptty.engine.core import SipEngine  # noqa: E402

__all__ = [
    "PJSUA2_AVAILABLE",
    "CallStateEvent",
    "RegStateEvent",
    "SipEngine",
    "SipTraceEvent",
]
