"""Tests for DTMF digit validation."""

from __future__ import annotations

import pytest


def _validate_dtmf(digits: str) -> None:
    """Standalone validation matching PhoneCall.do_send_dtmf logic."""
    valid = set("0123456789*#ABCDabcd")
    for ch in digits:
        if ch not in valid:
            raise ValueError(f"Invalid DTMF digit: {ch!r}")


class TestDtmfValidation:
    def test_valid_numeric(self) -> None:
        _validate_dtmf("0123456789")

    def test_valid_star_hash(self) -> None:
        _validate_dtmf("*#")

    def test_valid_letters(self) -> None:
        _validate_dtmf("ABCDabcd")

    def test_empty_is_valid(self) -> None:
        _validate_dtmf("")

    def test_invalid_letter_e(self) -> None:
        with pytest.raises(ValueError, match="Invalid DTMF digit"):
            _validate_dtmf("E")

    def test_invalid_space(self) -> None:
        with pytest.raises(ValueError, match="Invalid DTMF digit"):
            _validate_dtmf("1 2")

    def test_invalid_mixed(self) -> None:
        with pytest.raises(ValueError, match="Invalid DTMF digit"):
            _validate_dtmf("12X4")
