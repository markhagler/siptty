"""SIP engine layer \u2014 wraps pjsua2."""

try:
    import pjsua2 as pj

    PJSUA2_AVAILABLE = True
except ImportError:
    pj = None  # type: ignore[assignment]
    PJSUA2_AVAILABLE = False
