"""Tests for SipEngine lifecycle (start / stop)."""

from __future__ import annotations

import pytest

from siptty.config.models import AppConfig, AudioConfig, GeneralConfig
from siptty.engine import SipEngine

pytestmark = pytest.mark.requires_pjsua2


def _noop_callback(event: object) -> None:  # noqa: ARG001
    """No-op event callback for tests."""


def _null_audio_config() -> AppConfig:
    """Return a minimal AppConfig with null audio."""
    return AppConfig(
        general=GeneralConfig(log_level=0),
        audio=AudioConfig(mode="null"),
    )


class TestStartStop:
    """Basic lifecycle tests."""

    def test_start_stop(self) -> None:
        """Engine starts and stops cleanly."""
        engine = SipEngine(event_callback=_noop_callback)
        engine.start(_null_audio_config())
        try:
            assert engine._started is True
            assert engine._ep is not None
        finally:
            engine.stop()
        assert engine._started is False
        assert engine._ep is None

    def test_double_stop_is_safe(self) -> None:
        """Calling stop() twice must not raise."""
        engine = SipEngine(event_callback=_noop_callback)
        engine.start(_null_audio_config())
        engine.stop()
        engine.stop()  # second call — should be a no-op

    def test_stop_without_start_is_safe(self) -> None:
        """Calling stop() on a never-started engine must not raise."""
        engine = SipEngine(event_callback=_noop_callback)
        engine.stop()

    def test_double_start_raises(self) -> None:
        """Starting an already-started engine must raise RuntimeError."""
        engine = SipEngine(event_callback=_noop_callback)
        engine.start(_null_audio_config())
        try:
            with pytest.raises(RuntimeError, match="already started"):
                engine.start(_null_audio_config())
        finally:
            engine.stop()

    def test_restart_after_stop(self) -> None:
        """An engine can be started again after being stopped."""
        engine = SipEngine(event_callback=_noop_callback)
        engine.start(_null_audio_config())
        engine.stop()
        # Second cycle
        engine.start(_null_audio_config())
        engine.stop()
