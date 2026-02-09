"""Configuration dataclasses for siptty."""

from __future__ import annotations

from dataclasses import dataclass, field


@dataclass
class NatConfig:
    """NAT traversal settings for an account."""

    stun_server: str = ""
    ice_enabled: bool = False
    turn_enabled: bool = False
    turn_server: str = ""
    turn_user: str = ""
    turn_pass: str = ""


@dataclass
class TlsConfig:
    """TLS settings for an account."""

    cert_file: str = ""
    key_file: str = ""
    ca_file: str = ""
    verify_server: bool = True


@dataclass
class SrtpConfig:
    """SRTP settings for an account."""

    enabled: bool = False
    require: bool = False


@dataclass
class CodecConfig:
    """Codec priority settings for an account."""

    priority: list[str] = field(
        default_factory=lambda: ["opus/48000", "g722/16000", "pcmu/8000", "pcma/8000"]
    )


@dataclass
class BlfConfig:
    """Busy Lamp Field entry."""

    extension: str = ""
    label: str = ""


@dataclass
class AccountConfig:
    """Configuration for a single SIP account."""

    name: str = ""
    enabled: bool = True
    sip_uri: str = ""
    auth_user: str = ""
    auth_password: str = ""
    registrar: str = ""
    outbound_proxy: str = ""
    transport: str = "udp"
    register: bool = True
    reg_expiry: int = 300
    nat: NatConfig = field(default_factory=NatConfig)
    tls: TlsConfig = field(default_factory=TlsConfig)
    srtp: SrtpConfig = field(default_factory=SrtpConfig)
    codecs: CodecConfig = field(default_factory=CodecConfig)
    blf: list[BlfConfig] = field(default_factory=list)
    headers: dict[str, str] = field(default_factory=dict)


@dataclass
class GeneralConfig:
    """General application settings."""

    log_level: int = 3
    log_file: str = "siptty.log"
    null_audio: bool = False
    user_agent: str = "siptty/0.1"


@dataclass
class AudioConfig:
    """Audio subsystem settings."""

    mode: str = "null"
    play_file: str = ""
    record_dir: str = ""
    record_format: str = "wav"
    browser_port: int = 8080


@dataclass
class HistoryConfig:
    """Call history settings."""

    enabled: bool = True
    db_file: str = "~/.siptty/history.db"
    max_entries: int = 1000


@dataclass
class AppConfig:
    """Top-level application configuration."""

    general: GeneralConfig = field(default_factory=GeneralConfig)
    accounts: list[AccountConfig] = field(default_factory=list)
    audio: AudioConfig = field(default_factory=AudioConfig)
    history: HistoryConfig = field(default_factory=HistoryConfig)
