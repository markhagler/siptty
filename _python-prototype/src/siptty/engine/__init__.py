"""SIP engine layer — wraps pjsua2."""

try:
    import pjsua2 as pj

    PJSUA2_AVAILABLE = True
except ImportError:
    pj = None  # type: ignore[assignment]
    PJSUA2_AVAILABLE = False

from siptty.engine.events import CallStateEvent, RegStateEvent, SipTraceEvent

__all__ = [
    "PJSUA2_AVAILABLE",
    "CallStateEvent",
    "RegStateEvent",
    "SipTraceEvent",
]

# Conditionally export engine classes that need pjsua2
if PJSUA2_AVAILABLE:
    from siptty.engine.account import PhoneAccount
    from siptty.engine.call import PhoneCall
    from siptty.engine.core import SipEngine
    from siptty.engine.trace import SipTraceWriter

    __all__ += ["SipEngine", "PhoneAccount", "PhoneCall", "SipTraceWriter"]
