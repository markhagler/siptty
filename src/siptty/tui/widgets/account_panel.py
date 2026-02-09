"""Account panel widget."""

from textual.widgets import Static


class AccountPanel(Static):
    """Displays configured SIP accounts and their registration status."""

    DEFAULT_CSS = """
    AccountPanel {
        height: 100%;
    }
    """

    def __init__(self) -> None:
        super().__init__("(no accounts configured)", id="account-panel")
        self.border_title = "ACCOUNTS"
