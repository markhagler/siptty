#!/usr/bin/env python3
"""Wait for Asterisk to be ready by sending SIP OPTIONS probes.

Sends a SIP OPTIONS request via raw UDP socket and waits for any SIP
response. Retries every second until a response is received or the
timeout is reached.
"""

import argparse
import socket
import sys
import time
import uuid


def build_options_request(host: str, port: int) -> bytes:
    """Build a minimal SIP OPTIONS request."""
    branch = f"z9hG4bK-{uuid.uuid4().hex[:12]}"
    call_id = f"{uuid.uuid4().hex[:16]}@siptty-probe"
    tag = uuid.uuid4().hex[:8]

    request = (
        f"OPTIONS sip:{host}:{port} SIP/2.0\r\n"
        f"Via: SIP/2.0/UDP 127.0.0.1:15060;branch={branch};rport\r\n"
        f"From: <sip:probe@siptty>;tag={tag}\r\n"
        f"To: <sip:probe@{host}:{port}>\r\n"
        f"Call-ID: {call_id}\r\n"
        f"CSeq: 1 OPTIONS\r\n"
        f"Max-Forwards: 70\r\n"
        f"Content-Length: 0\r\n"
        f"\r\n"
    )
    return request.encode("utf-8")


def probe_sip(host: str, port: int, recv_timeout: float = 2.0) -> bool:
    """Send a SIP OPTIONS probe and return True if we get any SIP response."""
    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    sock.settimeout(recv_timeout)
    try:
        request = build_options_request(host, port)
        sock.sendto(request, (host, port))
        data, _ = sock.recvfrom(4096)
        response = data.decode("utf-8", errors="replace")
        return response.startswith("SIP/2.0")
    except (TimeoutError, OSError):
        return False
    finally:
        sock.close()


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Wait for Asterisk SIP to be ready"
    )
    parser.add_argument(
        "--host", default="127.0.0.1", help="Asterisk host (default: 127.0.0.1)"
    )
    parser.add_argument(
        "--port", type=int, default=5060, help="SIP port (default: 5060)"
    )
    parser.add_argument(
        "--timeout",
        type=int,
        default=30,
        help="Max seconds to wait (default: 30)",
    )
    args = parser.parse_args()

    deadline = time.monotonic() + args.timeout
    attempt = 0

    print(
        f"Waiting for Asterisk SIP at {args.host}:{args.port} "
        f"(timeout {args.timeout}s)..."
    )

    while time.monotonic() < deadline:
        attempt += 1
        if probe_sip(args.host, args.port):
            print(f"Asterisk SIP ready after {attempt} attempt(s).")
            return 0
        time.sleep(1)

    print(f"Timeout: Asterisk SIP not responding after {args.timeout}s.")
    return 1


if __name__ == "__main__":
    sys.exit(main())
