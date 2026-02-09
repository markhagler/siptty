"""Textual TUI application for siptty."""

from __future__ import annotations

import logging
import sys
from pathlib import Path
from typing import Any

from textual.app import App, ComposeResult
from textual.binding import Binding
from textual.containers import Horizontal
from textual.widgets import (
    Footer,
    Header,
    RichLog,
    Static,
    TabbedContent,
    TabPane,
)

from siptty.config import ConfigError, load_config
from siptty.config.models import AppConfig
from siptty.engine import PJSUA2_AVAILABLE
from siptty.engine.events import CallStateEvent, RegStateEvent, SipTraceEvent
from siptty.tui.widgets import AccountPanel, BlfPanel, CallControlPanel

log = logging.getLogger(__name__)

_DEFAULT_CONFIG_PATHS = [
    Path("siptty.toml"),
    Path.home() / ".config" / "siptty" / "config.toml",
]


class SipttyApp(App):
    """The main siptty terminal UI application."""

    TITLE = "siptty"
    SUB_TITLE = "SIP Softphone"
    CSS_PATH = "siptty.tcss"

    BINDINGS = [
        Binding("q", "quit", "Quit", show=False),
        Binding("f10", "quit", "Quit"),
        Binding("f1", "help", "Help"),
        Binding("d", "focus_dial", "Dial", show=False),
        Binding("a", "answer_call", "Answer", show=False),
        Binding("h", "hangup_call", "Hangup", show=False),
        Binding("o", "toggle_hold", "Hold", show=False),
        Binding("m", "toggle_mute", "Mute", show=False),
        Binding("p", "toggle_dtmf", "DTMF", show=False),
    ]

    def __init__(
        self,
        config_path: Path | None = None,
        **kwargs: Any,
    ) -> None:
        super().__init__(**kwargs)
        self._config_path = config_path
        self._config: AppConfig | None = None
        self._engine: Any = None  # SipEngine, set in on_mount if available

    def compose(self) -> ComposeResult:
        yield Header()
        with Horizontal(id="top-panels"):
            yield AccountPanel()
            yield CallControlPanel()
            yield BlfPanel()
        with TabbedContent():
            with TabPane("Calls", id="tab-calls"):
                yield Static("No active calls")
            with TabPane("SIP Trace", id="tab-sip-trace"):
                yield RichLog(id="sip-trace", highlight=True, markup=True)
            with TabPane("Dialogs", id="tab-dialogs"):
                yield Static("No SIP dialogs captured")
            with TabPane("Config", id="tab-config"):
                yield Static("Configuration viewer (not yet implemented)")
        yield Footer()

    def on_mount(self) -> None:
        """Load config and start the SIP engine."""
        # Load config
        config_path = self._config_path
        if config_path is None:
            for p in _DEFAULT_CONFIG_PATHS:
                if p.exists():
                    config_path = p
                    break

        if config_path and config_path.exists():
            try:
                self._config = load_config(config_path)
                self.notify(f"Config loaded: {config_path}")
            except ConfigError as exc:
                self.notify(f"Config error: {exc}", severity="error")
                return
        else:
            self._config = AppConfig()  # defaults
            self.notify("No config file found — using defaults")

        # Start engine
        if PJSUA2_AVAILABLE:
            from siptty.engine.core import SipEngine

            try:
                self._engine = SipEngine(event_callback=self._on_engine_event)
                self._engine.start(self._config)

                # Register accounts
                panel = self.query_one(AccountPanel)
                for acc_cfg in self._config.accounts:
                    if acc_cfg.enabled:
                        panel.add_account(acc_cfg.name, acc_cfg.sip_uri)
                        self._engine.add_account(acc_cfg)

            except Exception as exc:
                self.notify(f"Engine error: {exc}", severity="error")
                log.exception("Failed to start SipEngine")
        else:
            self.notify("pjsua2 not available — running in UI-only mode")

    def on_unmount(self) -> None:
        """Stop the engine on app exit."""
        if self._engine:
            try:
                self._engine.stop()
            except Exception:
                log.exception("Error stopping engine")

    # ------------------------------------------------------------------
    # Engine event callback (called from pjsua2 threads)
    # ------------------------------------------------------------------

    def _on_engine_event(self, event: Any) -> None:
        """Dispatch engine events to the TUI (thread-safe via call_from_thread)."""
        self.call_from_thread(self._handle_event, event)

    def _handle_event(self, event: Any) -> None:
        """Handle engine events on the Textual main thread."""
        if isinstance(event, RegStateEvent):
            self._on_reg_state(event)
        elif isinstance(event, CallStateEvent):
            self._on_call_state(event)
        elif isinstance(event, SipTraceEvent):
            self._on_sip_trace(event)

    def _on_reg_state(self, event: RegStateEvent) -> None:
        panel = self.query_one(AccountPanel)
        panel.update_state(event.account_id, event.state, event.reason)

    def _on_call_state(self, event: CallStateEvent) -> None:
        panel = self.query_one(CallControlPanel)
        if event.state == "disconnected":
            panel.clear_call()
        else:
            panel.set_call(event.call_id, event.state, event.remote_uri)

    def _on_sip_trace(self, event: SipTraceEvent) -> None:
        try:
            trace_log = self.query_one("#sip-trace", RichLog)
            arrow = "→" if event.direction == "send" else "←"
            # Get first line for summary
            first_line = event.message.split("\n", 1)[0] if event.message else ""
            trace_log.write(f"[dim]{arrow}[/] {first_line}")
        except Exception:
            pass

    # ------------------------------------------------------------------
    # Actions
    # ------------------------------------------------------------------

    def action_help(self) -> None:
        """Show help (placeholder)."""
        self.notify("Help is not yet implemented.", title="Help")

    def action_focus_dial(self) -> None:
        """Focus the dial input."""
        try:
            self.query_one("#dial-input").focus()
        except Exception:
            pass

    def action_answer_call(self) -> None:
        """Answer an incoming call."""
        if not self._engine:
            return
        panel = self.query_one(CallControlPanel)
        if panel.call_state == "incoming" and panel.call_id >= 0:
            try:
                self._engine.answer(panel.call_id)
            except Exception as exc:
                self.notify(f"Answer failed: {exc}", severity="error")

    def action_hangup_call(self) -> None:
        """Hang up or reject the current call."""
        if not self._engine:
            return
        panel = self.query_one(CallControlPanel)
        if panel.call_id >= 0:
            try:
                if panel.call_state == "incoming":
                    self._engine.reject(panel.call_id)
                else:
                    self._engine.hangup(panel.call_id)
            except Exception as exc:
                self.notify(f"Hangup failed: {exc}", severity="error")

    def action_toggle_hold(self) -> None:
        """Toggle hold on the current call."""
        if not self._engine:
            return
        panel = self.query_one(CallControlPanel)
        if panel.call_id >= 0:
            try:
                if panel.call_state == "hold":
                    self._engine.resume(panel.call_id)
                else:
                    self._engine.hold(panel.call_id)
            except Exception as exc:
                self.notify(f"Hold failed: {exc}", severity="error")

    def action_toggle_mute(self) -> None:
        """Toggle mute (placeholder)."""
        self.notify("Mute not yet implemented")

    def action_toggle_dtmf(self) -> None:
        """Toggle DTMF input mode."""
        panel = self.query_one(CallControlPanel)
        if panel.call_id < 0:
            self.notify("No active call for DTMF")
            return
        panel.dtmf_mode = not panel.dtmf_mode
        if panel.dtmf_mode:
            self.notify("DTMF mode ON — press digits, Escape to exit")
        else:
            self.notify("DTMF mode OFF")

    def on_input_submitted(self, event: Any) -> None:
        """Handle dial input submission."""
        if not self._engine or not self._config:
            return
        if hasattr(event, "input") and event.input.id == "dial-input":
            uri = event.value.strip()
            if not uri:
                return
            # Use first enabled account
            account_id = None
            for acc in self._config.accounts:
                if acc.enabled:
                    account_id = acc.name
                    break
            if account_id is None:
                self.notify("No accounts configured", severity="error")
                return
            try:
                self._engine.dial(account_id, uri)
                event.input.value = ""
            except Exception as exc:
                self.notify(f"Dial failed: {exc}", severity="error")

    def on_key(self, event: Any) -> None:
        """Handle keypress — send DTMF digits in DTMF mode."""
        panel = self.query_one(CallControlPanel)
        if not panel.dtmf_mode or panel.call_id < 0:
            return
        if event.key == "escape":
            panel.dtmf_mode = False
            panel.dtmf_digits = ""
            self.notify("DTMF mode OFF")
            event.prevent_default()
            return
        ch = event.character
        if ch and ch in "0123456789*#ABCDabcd":
            try:
                if self._engine:
                    self._engine.send_dtmf(panel.call_id, ch)
                    panel.dtmf_digits += ch
            except Exception as exc:
                self.notify(f"DTMF error: {exc}", severity="error")
            event.prevent_default()


def main() -> None:
    """Entry point for the siptty console script."""
    config_path = None
    if len(sys.argv) > 1:
        config_path = Path(sys.argv[1])

    app = SipttyApp(config_path=config_path)
    try:
        app.run()
    except KeyboardInterrupt:
        pass


if __name__ == "__main__":
    main()
