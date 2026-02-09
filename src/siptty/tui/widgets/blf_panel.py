"""BLF / Presence panel widget."""

from textual.widgets import Static


class BlfPanel(Static):
    """Displays BLF (Busy Lamp Field) / presence subscription status."""

    DEFAULT_CSS = """
    BlfPanel {
        height: 100%;
    }
    """

    def __init__(self) -> None:
        super().__init__("(no subscriptions)", id="blf-panel")
        self.border_title = "BLF / PRESENCE"
