"""Call control panel widget."""

from textual.app import ComposeResult
from textual.widget import Widget
from textual.widgets import Input, Static


class CallControlPanel(Widget):
    """Central call-control panel with state display and dial input."""

    DEFAULT_CSS = """
    CallControlPanel {
        height: 100%;
        layout: vertical;
    }
    """

    def __init__(self) -> None:
        super().__init__(id="call-panel")
        self.border_title = "CALL CONTROL"

    def compose(self) -> ComposeResult:
        yield Static("State: IDLE", id="call-state")
        yield Input(placeholder="Enter SIP URI...", id="dial-input")
