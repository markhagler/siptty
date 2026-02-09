#!/usr/bin/env python3
"""Verify pjsua2 Python bindings are installed and working."""
import sys

try:
    import pjsua2 as pj

    ep = pj.Endpoint()
    ep.libCreate()
    print(f"pjsua2 OK \u2014 PJSIP {ep.libVersion().full}")
    ep.libDestroy()
except ImportError:
    print("pjsua2 not installed", file=sys.stderr)
    sys.exit(1)
except Exception as e:
    print(f"pjsua2 error: {e}", file=sys.stderr)
    sys.exit(1)
