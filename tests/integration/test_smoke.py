"""Smoke test: verify Asterisk is reachable via SIP."""

import socket
import uuid

import pytest


@pytest.mark.timeout(15)
def test_asterisk_responds_to_options(asterisk: None) -> None:
    """Send SIP OPTIONS to Asterisk and verify a SIP/2.0 response."""
    host, port = "127.0.0.1", 5060

    branch = f"z9hG4bK-{uuid.uuid4().hex[:12]}"
    call_id = f"{uuid.uuid4().hex[:16]}@smoke"
    tag = uuid.uuid4().hex[:8]

    request = (
        f"OPTIONS sip:{host}:{port} SIP/2.0\r\n"
        f"Via: SIP/2.0/UDP 127.0.0.1:15060;branch={branch};rport\r\n"
        f"From: <sip:probe@smoke>;tag={tag}\r\n"
        f"To: <sip:probe@{host}:{port}>\r\n"
        f"Call-ID: {call_id}\r\n"
        f"CSeq: 1 OPTIONS\r\n"
        f"Max-Forwards: 70\r\n"
        f"Content-Length: 0\r\n"
        f"\r\n"
    ).encode()

    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    sock.settimeout(5.0)
    try:
        sock.sendto(request, (host, port))
        data, _ = sock.recvfrom(4096)
        response = data.decode("utf-8", errors="replace")
        assert response.startswith("SIP/2.0"), f"Unexpected response: {response[:80]}"
        # Should be a 200 or 401 — either proves Asterisk is alive
        status_line = response.split("\r\n")[0]
        assert "SIP/2.0" in status_line
    finally:
        sock.close()
