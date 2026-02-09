"""PhoneAccount — pjsua2 Account wrapper with registration state events."""

from __future__ import annotations

import logging
from collections.abc import Callable
from typing import TYPE_CHECKING, Any

from siptty.engine import PJSUA2_AVAILABLE
from siptty.engine.events import RegStateEvent

if PJSUA2_AVAILABLE:
    import pjsua2 as pj

if TYPE_CHECKING:
    from siptty.config.models import AccountConfig

log = logging.getLogger(__name__)


def _build_account_config(cfg: AccountConfig) -> pj.AccountConfig:
    """Translate our AccountConfig dataclass into a pjsua2 AccountConfig."""
    acc_cfg = pj.AccountConfig()

    # Identity
    acc_cfg.idUri = cfg.sip_uri

    # Registration
    if cfg.register and cfg.registrar:
        acc_cfg.regConfig.registrarUri = cfg.registrar
        acc_cfg.regConfig.timeoutSec = cfg.reg_expiry

    # Auth credentials
    if cfg.auth_user and cfg.auth_password:
        cred = pj.AuthCredInfo()
        cred.scheme = "digest"
        cred.realm = "*"
        cred.username = cfg.auth_user
        cred.data = cfg.auth_password
        cred.dataType = 0  # plain text
        acc_cfg.sipConfig.authCreds.append(cred)

    # Outbound proxy
    if cfg.outbound_proxy:
        proxy = cfg.outbound_proxy
        if not proxy.startswith("sip:"):
            proxy = f"sip:{proxy}"
        acc_cfg.sipConfig.proxies.append(proxy)

    # Media — prefer audio only
    acc_cfg.mediaConfig.transportConfig.port = 0  # ephemeral RTP port

    return acc_cfg


class PhoneAccount(pj.Account if PJSUA2_AVAILABLE else object):  # type: ignore[misc]
    """Wraps a pjsua2 Account.  Fires RegStateEvent on registration changes."""

    def __init__(
        self,
        cfg: AccountConfig,
        event_callback: Callable[[Any], None],
        incoming_call_factory: Callable[..., Any] | None = None,
    ) -> None:
        super().__init__()
        self.cfg = cfg
        self._callback = event_callback
        self._incoming_call_factory = incoming_call_factory
        self.account_name = cfg.name

    def create_and_register(self) -> None:
        """Build the pjsua2 AccountConfig and create the account."""
        acc_cfg = _build_account_config(self.cfg)
        self.create(acc_cfg)

    # ------------------------------------------------------------------
    # pjsua2 callbacks
    # ------------------------------------------------------------------

    def onRegState(self, prm: pj.OnRegStateParam) -> None:  # noqa: N802
        """Called by pjsua2 when registration state changes."""
        try:
            info = self.getInfo()
            code = info.regStatus
            if info.regIsActive:
                state = "registered"
            elif code == 0:
                state = "unregistered"
            else:
                state = "failed"

            reason = f"{code} {info.regStatusText}"
            event = RegStateEvent(
                account_id=self.account_name,
                state=state,
                reason=reason,
            )
            self._callback(event)
        except Exception:
            log.exception("Error in onRegState for %s", self.account_name)

    def onIncomingCall(self, prm: pj.OnIncomingCallParam) -> None:  # noqa: N802
        """Called by pjsua2 when an incoming call arrives."""
        try:
            if self._incoming_call_factory:
                self._incoming_call_factory(self, prm.callId)
        except Exception:
            log.exception("Error in onIncomingCall for %s", self.account_name)
