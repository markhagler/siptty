"""TOML configuration loader for siptty."""

from __future__ import annotations

import tomllib
from pathlib import Path
from typing import Any

from .models import (
    AccountConfig,
    AppConfig,
    AudioConfig,
    BlfConfig,
    CodecConfig,
    GeneralConfig,
    HistoryConfig,
    NatConfig,
    SrtpConfig,
    TlsConfig,
)

_VALID_TRANSPORTS = {"udp", "tcp", "tls"}


class ConfigError(Exception):
    """Raised when configuration is invalid or incomplete."""


def _parse_general(data: dict[str, Any]) -> GeneralConfig:
    """Parse the [general] section."""
    return GeneralConfig(
        log_level=data.get("log_level", 3),
        log_file=data.get("log_file", "siptty.log"),
        null_audio=data.get("null_audio", False),
        user_agent=data.get("user_agent", "siptty/0.1"),
    )


def _parse_nat(data: dict[str, Any]) -> NatConfig:
    """Parse an [accounts.nat] section."""
    return NatConfig(
        stun_server=data.get("stun_server", ""),
        ice_enabled=data.get("ice_enabled", False),
        turn_enabled=data.get("turn_enabled", False),
        turn_server=data.get("turn_server", ""),
        turn_user=data.get("turn_user", ""),
        turn_pass=data.get("turn_pass", ""),
    )


def _parse_tls(data: dict[str, Any]) -> TlsConfig:
    """Parse an [accounts.tls] section."""
    return TlsConfig(
        cert_file=data.get("cert_file", ""),
        key_file=data.get("key_file", ""),
        ca_file=data.get("ca_file", ""),
        verify_server=data.get("verify_server", True),
    )


def _parse_srtp(data: dict[str, Any]) -> SrtpConfig:
    """Parse an [accounts.srtp] section."""
    return SrtpConfig(
        enabled=data.get("enabled", False),
        require=data.get("require", False),
    )


def _parse_codecs(data: dict[str, Any]) -> CodecConfig:
    """Parse an [accounts.codecs] section."""
    default_priority = ["opus/48000", "g722/16000", "pcmu/8000", "pcma/8000"]
    return CodecConfig(
        priority=data.get("priority", default_priority),
    )


def _parse_blf_list(data: list[dict[str, Any]]) -> list[BlfConfig]:
    """Parse [[accounts.blf]] entries."""
    return [
        BlfConfig(
            extension=entry.get("extension", ""),
            label=entry.get("label", ""),
        )
        for entry in data
    ]


def _user_from_sip_uri(sip_uri: str) -> str:
    """Extract the user part from a sip: URI.

    Example: 'sip:alice@pbx.example.com' -> 'alice'
    """
    # Strip 'sip:' prefix, then take everything before '@'
    without_scheme = sip_uri[4:]  # remove 'sip:'
    if "@" in without_scheme:
        return without_scheme.split("@", 1)[0]
    return without_scheme


def _parse_account(data: dict[str, Any]) -> AccountConfig:
    """Parse a single [[accounts]] entry."""
    sip_uri = data.get("sip_uri", "")
    if not sip_uri:
        raise ConfigError("Account is missing required field 'sip_uri'")
    if not sip_uri.startswith("sip:"):
        raise ConfigError(
            f"Invalid sip_uri '{sip_uri}': must start with 'sip:'"
        )

    transport = data.get("transport", "udp")
    if transport not in _VALID_TRANSPORTS:
        raise ConfigError(
            f"Invalid transport '{transport}': must be one of {sorted(_VALID_TRANSPORTS)}"
        )

    auth_user = data.get("auth_user", "") or _user_from_sip_uri(sip_uri)

    return AccountConfig(
        name=data.get("name", ""),
        enabled=data.get("enabled", True),
        sip_uri=sip_uri,
        auth_user=auth_user,
        auth_password=data.get("auth_password", ""),
        registrar=data.get("registrar", ""),
        outbound_proxy=data.get("outbound_proxy", ""),
        transport=transport,
        register=data.get("register", True),
        reg_expiry=data.get("reg_expiry", 300),
        nat=_parse_nat(data.get("nat", {})),
        tls=_parse_tls(data.get("tls", {})),
        srtp=_parse_srtp(data.get("srtp", {})),
        codecs=_parse_codecs(data.get("codecs", {})),
        blf=_parse_blf_list(data.get("blf", [])),
        headers=data.get("headers", {}),
    )


def _parse_audio(data: dict[str, Any]) -> AudioConfig:
    """Parse the [audio] section."""
    return AudioConfig(
        mode=data.get("mode", "null"),
        play_file=data.get("play_file", ""),
        record_dir=data.get("record_dir", ""),
        record_format=data.get("record_format", "wav"),
        browser_port=data.get("browser_port", 8080),
    )


def _parse_history(data: dict[str, Any]) -> HistoryConfig:
    """Parse the [history] section."""
    return HistoryConfig(
        enabled=data.get("enabled", True),
        db_file=data.get("db_file", "~/.siptty/history.db"),
        max_entries=data.get("max_entries", 1000),
    )


def load_config(path: Path) -> AppConfig:
    """Load and validate configuration from a TOML file.

    Args:
        path: Path to the TOML configuration file.

    Returns:
        Fully populated AppConfig instance.

    Raises:
        ConfigError: If the file is missing, unparseable, or contains
            invalid values.
        FileNotFoundError: If the config file does not exist.
    """
    if not path.exists():
        raise FileNotFoundError(f"Config file not found: {path}")

    text = path.read_text(encoding="utf-8")
    try:
        data = tomllib.loads(text)
    except tomllib.TOMLDecodeError as exc:
        raise ConfigError(f"Failed to parse TOML: {exc}") from exc

    if not data:
        raise ConfigError("Configuration file is empty")

    accounts_data: list[dict[str, Any]] = data.get("accounts", [])
    accounts = [_parse_account(acct) for acct in accounts_data]

    return AppConfig(
        general=_parse_general(data.get("general", {})),
        accounts=accounts,
        audio=_parse_audio(data.get("audio", {})),
        history=_parse_history(data.get("history", {})),
    )
