"""Tests for siptty.config (models + loader)."""

from __future__ import annotations

from pathlib import Path
from textwrap import dedent

import pytest

from siptty.config import AccountConfig, AppConfig, ConfigError, load_config


def _write_toml(tmp_path: Path, content: str) -> Path:
    """Write TOML content to a temp file and return the path."""
    p = tmp_path / "test.toml"
    p.write_text(dedent(content), encoding="utf-8")
    return p


# --- Minimal config -----------------------------------------------------------


def test_minimal_config(tmp_path: Path) -> None:
    """A config with only one account (just sip_uri) should load with defaults."""
    path = _write_toml(
        tmp_path,
        """\
        [[accounts]]
        sip_uri = "sip:alice@pbx.example.com"
        """,
    )
    cfg = load_config(path)
    assert isinstance(cfg, AppConfig)
    assert len(cfg.accounts) == 1
    acct = cfg.accounts[0]
    assert acct.sip_uri == "sip:alice@pbx.example.com"
    # auth_user derived from sip_uri
    assert acct.auth_user == "alice"
    assert acct.transport == "udp"
    assert acct.register is True
    assert acct.reg_expiry == 300


# --- Multi-account ------------------------------------------------------------


def test_multi_account(tmp_path: Path) -> None:
    path = _write_toml(
        tmp_path,
        """\
        [[accounts]]
        name = "alice"
        sip_uri = "sip:alice@pbx.example.com"

        [[accounts]]
        name = "bob"
        sip_uri = "sip:bob@other.example.com"
        transport = "tcp"
        """,
    )
    cfg = load_config(path)
    assert len(cfg.accounts) == 2
    assert cfg.accounts[0].name == "alice"
    assert cfg.accounts[1].name == "bob"
    assert cfg.accounts[1].transport == "tcp"


# --- Defaults applied ---------------------------------------------------------


def test_defaults_applied(tmp_path: Path) -> None:
    """Sections not present in the file should get sensible defaults."""
    path = _write_toml(
        tmp_path,
        """\
        [general]
        log_level = 5
        """,
    )
    cfg = load_config(path)
    assert cfg.general.log_level == 5
    assert cfg.general.log_file == "siptty.log"
    assert cfg.general.user_agent == "siptty/0.1"
    assert cfg.audio.mode == "null"
    assert cfg.history.enabled is True
    assert cfg.history.db_file == "~/.siptty/history.db"
    assert cfg.accounts == []


# --- Invalid transport --------------------------------------------------------


def test_invalid_transport_rejected(tmp_path: Path) -> None:
    path = _write_toml(
        tmp_path,
        """\
        [[accounts]]
        sip_uri = "sip:alice@pbx.example.com"
        transport = "sctp"
        """,
    )
    with pytest.raises(ConfigError, match="Invalid transport"):
        load_config(path)


# --- Missing sip_uri ----------------------------------------------------------


def test_missing_sip_uri_raises(tmp_path: Path) -> None:
    path = _write_toml(
        tmp_path,
        """\
        [[accounts]]
        name = "broken"
        """,
    )
    with pytest.raises(ConfigError, match="sip_uri"):
        load_config(path)


# --- Invalid sip_uri ----------------------------------------------------------


def test_invalid_sip_uri_raises(tmp_path: Path) -> None:
    path = _write_toml(
        tmp_path,
        """\
        [[accounts]]
        sip_uri = "http://not-sip"
        """,
    )
    with pytest.raises(ConfigError, match="must start with 'sip:'"):
        load_config(path)


# --- Empty file ---------------------------------------------------------------


def test_empty_file_raises(tmp_path: Path) -> None:
    path = _write_toml(tmp_path, "")
    with pytest.raises(ConfigError, match="empty"):
        load_config(path)


# --- File not found -----------------------------------------------------------


def test_missing_file_raises(tmp_path: Path) -> None:
    with pytest.raises(FileNotFoundError):
        load_config(tmp_path / "nonexistent.toml")


# --- BLF + header overrides ---------------------------------------------------


def test_blf_and_headers_parsed(tmp_path: Path) -> None:
    path = _write_toml(
        tmp_path,
        """\
        [[accounts]]
        sip_uri = "sip:alice@pbx.example.com"

          [[accounts.blf]]
          extension = "201"
          label = "Bob"

          [[accounts.blf]]
          extension = "202"
          label = "Carol"

          [accounts.headers]
          User-Agent = "Yealink SIP-T54W 96.86.0.100"
          X-Custom = "test"
        """,
    )
    cfg = load_config(path)
    acct = cfg.accounts[0]
    assert len(acct.blf) == 2
    assert acct.blf[0].extension == "201"
    assert acct.blf[0].label == "Bob"
    assert acct.blf[1].extension == "202"
    assert acct.headers["User-Agent"] == "Yealink SIP-T54W 96.86.0.100"
    assert acct.headers["X-Custom"] == "test"


# --- auth_user explicit override ----------------------------------------------


def test_explicit_auth_user(tmp_path: Path) -> None:
    path = _write_toml(
        tmp_path,
        """\
        [[accounts]]
        sip_uri = "sip:alice@pbx.example.com"
        auth_user = "alice-override"
        """,
    )
    cfg = load_config(path)
    assert cfg.accounts[0].auth_user == "alice-override"


# --- NAT / TLS / SRTP / codecs sub-sections -----------------------------------


def test_subsections_parsed(tmp_path: Path) -> None:
    path = _write_toml(
        tmp_path,
        """\
        [[accounts]]
        sip_uri = "sip:alice@pbx.example.com"
        transport = "tls"

          [accounts.nat]
          stun_server = "stun:stun.l.google.com:19302"
          ice_enabled = true

          [accounts.tls]
          verify_server = false

          [accounts.srtp]
          enabled = true
          require = true

          [accounts.codecs]
          priority = ["pcmu/8000"]
        """,
    )
    cfg = load_config(path)
    acct = cfg.accounts[0]
    assert acct.nat.stun_server == "stun:stun.l.google.com:19302"
    assert acct.nat.ice_enabled is True
    assert acct.tls.verify_server is False
    assert acct.srtp.enabled is True
    assert acct.srtp.require is True
    assert acct.codecs.priority == ["pcmu/8000"]


# --- AccountConfig is importable at package level ---


def test_account_config_importable() -> None:
    assert AccountConfig is not None
    acct = AccountConfig(sip_uri="sip:test@example.com")
    assert acct.sip_uri == "sip:test@example.com"
