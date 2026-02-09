"""SipEngine — pjsua2 lifecycle wrapper."""

from __future__ import annotations

import logging
from collections.abc import Callable
from typing import Any

from siptty.config.models import AppConfig
from siptty.engine import PJSUA2_AVAILABLE

if PJSUA2_AVAILABLE:
    import pjsua2 as pj

log = logging.getLogger(__name__)


class SipEngine:
    """Thin wrapper around pjsua2 :class:`pj.Endpoint` with lifecycle management."""

    def __init__(self, event_callback: Callable[[Any], None]) -> None:
        if not PJSUA2_AVAILABLE:
            raise RuntimeError("pjsua2 is not available – cannot create SipEngine")
        self._callback = event_callback
        self._ep: pj.Endpoint | None = None
        self._accounts: dict[str, Any] = {}  # PhoneAccount – placeholder for now
        self._calls: dict[int, Any] = {}  # PhoneCall – placeholder for now
        self._started = False
        self._transport_id: int = -1

    # ------------------------------------------------------------------
    # Lifecycle
    # ------------------------------------------------------------------

    def start(self, config: AppConfig) -> None:
        """Create the pjsua2 Endpoint, initialise and start the library."""
        if self._started:
            raise RuntimeError("SipEngine is already started")

        ep = pj.Endpoint()
        ep.libCreate()

        try:
            # ---- Endpoint configuration ----
            ep_cfg = pj.EpConfig()
            ep_cfg.uaConfig.userAgent = config.general.user_agent

            # Library-level log: honour config but suppress console spew
            ep_cfg.logConfig.level = config.general.log_level
            ep_cfg.logConfig.consoleLevel = 0

            ep.libInit(ep_cfg)

            # ---- UDP transport (ephemeral port) ----
            tp_cfg = pj.TransportConfig()
            tp_cfg.port = 0
            self._transport_id = ep.transportCreate(pj.PJSIP_TRANSPORT_UDP, tp_cfg)

            # ---- Audio device ----
            if config.audio.mode == "null":
                ep.audDevManager().setNullDev()

            ep.libStart()
        except Exception:
            # If anything above fails, tear down what we created.
            try:
                ep.libDestroy()
            except Exception:  # noqa: S110
                pass
            raise

        self._ep = ep
        self._started = True
        log.info("SipEngine started (ua=%s)", config.general.user_agent)

    def stop(self) -> None:
        """Destroy the pjsua2 Endpoint.  Safe to call multiple times."""
        if not self._started or self._ep is None:
            return

        try:
            self._ep.libDestroy()
        except Exception:
            log.exception("Error during libDestroy")
        finally:
            self._ep = None
            self._started = False
            self._accounts.clear()
            self._calls.clear()
            log.info("SipEngine stopped")
