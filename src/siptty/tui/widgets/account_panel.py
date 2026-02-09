"""Account panel widget."""

from __future__ import annotations

from textual.app import ComposeResult
from textual.reactive import reactive
from textual.widget import Widget
from textual.widgets import Static


class AccountEntry(Static):
    """A single account with registration state."""

    ICONS = {
        "registered": "●",     # green
        "failed": "○",         # red
        "unregistered": "◌",   # grey
    }
    COLORS = {
        "registered": "green",
        "failed": "red",
        "unregistered": "dim",
    }

    state: reactive[str] = reactive("unregistered")
    reason: reactive[str] = reactive("")

    def __init__(self, name: str, uri: str) -> None:
        super().__init__(id=f"account-{name}")
        self._name = name
        self._uri = uri

    def render(self) -> str:
        icon = self.ICONS.get(self.state, "◌")
        color = self.COLORS.get(self.state, "dim")
        return f"[{color}]{icon}[/] {self._name}\n  [{color}]{self.state.title()}[/]"


class AccountPanel(Widget):
    """Displays configured SIP accounts and their registration status."""

    DEFAULT_CSS = """
    AccountPanel {
        height: 100%;
        layout: vertical;
    }
    """

    def __init__(self) -> None:
        super().__init__(id="account-panel")
        self.border_title = "ACCOUNTS"
        self._entries: dict[str, AccountEntry] = {}

    def compose(self) -> ComposeResult:
        yield Static("(no accounts configured)", id="no-accounts-msg")

    def add_account(self, name: str, uri: str) -> None:
        """Add an account entry to the panel."""
        # Hide the placeholder
        try:
            self.query_one("#no-accounts-msg").display = False
        except Exception:
            pass
        entry = AccountEntry(name, uri)
        self._entries[name] = entry
        self.mount(entry)

    def update_state(self, account_id: str, state: str, reason: str = "") -> None:
        """Update the registration state of an account."""
        entry = self._entries.get(account_id)
        if entry:
            entry.state = state
            entry.reason = reason
