"""Call control panel widget."""

from __future__ import annotations

from textual.app import ComposeResult
from textual.reactive import reactive
from textual.widget import Widget
from textual.widgets import Input, Static


class CallControlPanel(Widget):
    """Central call-control panel with state display and dial input."""

    DEFAULT_CSS = """
    CallControlPanel {
        height: 100%;
        layout: vertical;
    }
    #call-state {
        height: auto;
        margin: 0 0 1 0;
    }
    #call-remote {
        height: auto;
    }
    #call-dtmf-display {
        height: auto;
        color: $accent;
    }
    #call-hints {
        height: auto;
        margin: 1 0 0 0;
        color: $text-muted;
    }
    """

    call_state: reactive[str] = reactive("IDLE")
    remote_uri: reactive[str] = reactive("")
    call_id: reactive[int] = reactive(-1)
    dtmf_mode: reactive[bool] = reactive(False)
    dtmf_digits: reactive[str] = reactive("")

    def __init__(self) -> None:
        super().__init__(id="call-panel")
        self.border_title = "CALL CONTROL"

    def compose(self) -> ComposeResult:
        yield Static("State: IDLE", id="call-state")
        yield Static("", id="call-remote")
        yield Input(placeholder="Enter SIP URI...", id="dial-input")
        yield Static("", id="call-dtmf-display")
        yield Static(
            "[d]Dial [a]Answer [h]Hangup [p]DTMF [o]Hold",
            id="call-hints",
        )

    def watch_call_state(self, value: str) -> None:
        try:
            label = self.query_one("#call-state", Static)
            if value == "IDLE":
                label.update("State: IDLE")
            elif value == "incoming":
                label.update("[bold red]INCOMING CALL[/]")
            elif value == "confirmed":
                label.update("[bold green]CONNECTED[/]")
            elif value == "hold":
                label.update("[bold yellow]ON HOLD[/]")
            else:
                label.update(f"State: {value.upper()}")
        except Exception:
            pass

    def watch_remote_uri(self, value: str) -> None:
        try:
            self.query_one("#call-remote", Static).update(value)
        except Exception:
            pass

    def watch_dtmf_digits(self, value: str) -> None:
        try:
            if value:
                self.query_one("#call-dtmf-display", Static).update(
                    f"DTMF: {value}"
                )
            else:
                self.query_one("#call-dtmf-display", Static).update("")
        except Exception:
            pass

    def set_call(self, call_id: int, state: str, remote: str) -> None:
        """Update the panel with call information."""
        self.call_id = call_id
        self.call_state = state
        self.remote_uri = remote

    def clear_call(self) -> None:
        """Reset the panel to idle state."""
        self.call_id = -1
        self.call_state = "IDLE"
        self.remote_uri = ""
        self.dtmf_mode = False
        self.dtmf_digits = ""
