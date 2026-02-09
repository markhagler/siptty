"""Shared pytest configuration and fixtures."""

import subprocess
import time

import pytest


def pytest_addoption(parser: pytest.Parser) -> None:
    parser.addoption(
        "--asterisk-up",
        action="store_true",
        default=False,
        help="Skip Asterisk container start/stop (assume already running)",
    )


def _sip_probe(host: str = "127.0.0.1", port: int = 5060, timeout: float = 2.0) -> bool:
    """Send a SIP OPTIONS probe and return True if we get a response."""
    import socket
    import uuid

    branch = f"z9hG4bK-{uuid.uuid4().hex[:12]}"
    call_id = f"{uuid.uuid4().hex[:16]}@siptty-test"
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
    ).encode()

    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    sock.settimeout(timeout)
    try:
        sock.sendto(request, (host, port))
        data, _ = sock.recvfrom(4096)
        return data.decode("utf-8", errors="replace").startswith("SIP/2.0")
    except (TimeoutError, OSError):
        return False
    finally:
        sock.close()


def _wait_for_asterisk(
    host: str = "127.0.0.1", port: int = 5060, timeout: float = 30.0
) -> None:
    """Block until Asterisk responds to SIP OPTIONS."""
    deadline = time.monotonic() + timeout
    while time.monotonic() < deadline:
        if _sip_probe(host, port):
            return
        time.sleep(1)
    raise TimeoutError(f"Asterisk not responding on {host}:{port} after {timeout}s")


@pytest.fixture(scope="session")
def asterisk(request: pytest.FixtureRequest) -> None:  # type: ignore[misc]
    """Ensure Asterisk Docker container is running."""
    if request.config.getoption("--asterisk-up"):
        # Assume container is already running; just verify.
        _wait_for_asterisk(timeout=5)
        yield
        return

    compose_file = "docker-compose.test.yml"
    subprocess.run(
        ["docker", "compose", "-f", compose_file, "up", "-d"],
        check=True,
        capture_output=True,
    )
    try:
        _wait_for_asterisk(timeout=45)
        yield
    finally:
        subprocess.run(
            ["docker", "compose", "-f", compose_file, "down"],
            check=True,
            capture_output=True,
        )


# Marker for tests that need pjsua2
def pytest_configure(config: pytest.Config) -> None:
    config.addinivalue_line(
        "markers", "requires_pjsua2: skip if pjsua2 is not installed"
    )


def pytest_collection_modifyitems(
    config: pytest.Config, items: list[pytest.Item]
) -> None:
    try:
        import pjsua2  # noqa: F401
    except ImportError:
        skip_pjsua2 = pytest.mark.skip(reason="pjsua2 not installed")
        for item in items:
            if "requires_pjsua2" in item.keywords:
                item.add_marker(skip_pjsua2)
