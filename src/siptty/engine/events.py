"""Frozen dataclasses for engine-level events."""

from __future__ import annotations

from dataclasses import dataclass


@dataclass(frozen=True)
class RegStateEvent:
    """Registration state change."""

    account_id: str
    state: str  # "registered", "unregistered", "failed"
    reason: str


@dataclass(frozen=True)
class CallStateEvent:
    """Call state change."""

    call_id: int
    state: str  # "calling", "incoming", "early", "connecting",
    #              "confirmed", "disconnected", "hold"
    remote_uri: str
    duration: float
    direction: str  # "inbound" | "outbound"


@dataclass(frozen=True)
class SipTraceEvent:
    """Raw SIP message trace."""

    direction: str  # "send" | "recv"
    message: str  # Full SIP message text
    timestamp: float
