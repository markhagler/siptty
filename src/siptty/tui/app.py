"""Textual TUI application for siptty."""

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

from siptty.tui.widgets import AccountPanel, BlfPanel, CallControlPanel


class SipttyApp(App):
    """The main siptty terminal UI application."""

    TITLE = "siptty"
    SUB_TITLE = "SIP Softphone"
    CSS_PATH = "siptty.tcss"

    BINDINGS = [
        Binding("q", "quit", "Quit", show=False),
        Binding("f10", "quit", "Quit"),
        Binding("f1", "help", "Help"),
    ]

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

    def action_help(self) -> None:
        """Show help (placeholder)."""
        self.notify("Help is not yet implemented.", title="Help")


def main() -> None:
    """Entry point for the siptty console script."""
    app = SipttyApp()
    try:
        app.run()
    except KeyboardInterrupt:
        pass


if __name__ == "__main__":
    main()
