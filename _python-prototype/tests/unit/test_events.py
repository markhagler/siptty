"""Tests for engine event dataclasses."""

from __future__ import annotations

import pytest

from siptty.engine.events import CallStateEvent, RegStateEvent, SipTraceEvent

# ---- RegStateEvent ---------------------------------------------------------


class TestRegStateEvent:
    def test_constructable(self) -> None:
        evt = RegStateEvent(account_id="acc1", state="registered", reason="OK")
        assert evt.account_id == "acc1"
        assert evt.state == "registered"
        assert evt.reason == "OK"

    def test_frozen(self) -> None:
        evt = RegStateEvent(account_id="acc1", state="registered", reason="OK")
        with pytest.raises(AttributeError):
            evt.state = "failed"  # type: ignore[misc]

    def test_field_types(self) -> None:
        evt = RegStateEvent(account_id="acc1", state="registered", reason="OK")
        assert isinstance(evt.account_id, str)
        assert isinstance(evt.state, str)
        assert isinstance(evt.reason, str)


# ---- CallStateEvent --------------------------------------------------------


class TestCallStateEvent:
    def test_constructable(self) -> None:
        evt = CallStateEvent(
            call_id=1,
            state="confirmed",
            remote_uri="sip:bob@example.com",
            duration=12.5,
            direction="outbound",
        )
        assert evt.call_id == 1
        assert evt.state == "confirmed"
        assert evt.remote_uri == "sip:bob@example.com"
        assert evt.duration == 12.5
        assert evt.direction == "outbound"

    def test_frozen(self) -> None:
        evt = CallStateEvent(
            call_id=1,
            state="confirmed",
            remote_uri="sip:bob@example.com",
            duration=12.5,
            direction="outbound",
        )
        with pytest.raises(AttributeError):
            evt.call_id = 2  # type: ignore[misc]

    def test_field_types(self) -> None:
        evt = CallStateEvent(
            call_id=1,
            state="confirmed",
            remote_uri="sip:bob@example.com",
            duration=12.5,
            direction="outbound",
        )
        assert isinstance(evt.call_id, int)
        assert isinstance(evt.state, str)
        assert isinstance(evt.remote_uri, str)
        assert isinstance(evt.duration, float)
        assert isinstance(evt.direction, str)


# ---- SipTraceEvent ---------------------------------------------------------


class TestSipTraceEvent:
    def test_constructable(self) -> None:
        evt = SipTraceEvent(direction="send", message="INVITE sip:bob@x", timestamp=1.0)
        assert evt.direction == "send"
        assert evt.message == "INVITE sip:bob@x"
        assert evt.timestamp == 1.0

    def test_frozen(self) -> None:
        evt = SipTraceEvent(direction="send", message="INVITE sip:bob@x", timestamp=1.0)
        with pytest.raises(AttributeError):
            evt.direction = "recv"  # type: ignore[misc]

    def test_field_types(self) -> None:
        evt = SipTraceEvent(direction="recv", message="200 OK", timestamp=2.5)
        assert isinstance(evt.direction, str)
        assert isinstance(evt.message, str)
        assert isinstance(evt.timestamp, float)


# ---- Re-exports from engine package ----------------------------------------


def test_importable_from_engine_package() -> None:
    from siptty.engine import CallStateEvent, RegStateEvent, SipTraceEvent

    assert RegStateEvent is not None
    assert CallStateEvent is not None
    assert SipTraceEvent is not None
