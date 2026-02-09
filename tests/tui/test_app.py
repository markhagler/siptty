"""Smoke tests for the TUI application."""
import pytest

from siptty.tui.app import SipttyApp
from siptty.tui.widgets import AccountPanel, BlfPanel, CallControlPanel


@pytest.fixture
def app():
    return SipttyApp()


async def test_app_launches_and_has_panels(app):
    """App starts and all main panels exist."""
    async with app.run_test():
        assert app.query_one("#account-panel", AccountPanel)
        assert app.query_one("#call-panel", CallControlPanel)
        assert app.query_one("#blf-panel", BlfPanel)


async def test_app_has_tabbed_content(app):
    """App has the tabbed bottom section."""
    from textual.widgets import TabbedContent
    async with app.run_test():
        assert app.query_one(TabbedContent)


async def test_quit_key(app):
    """Pressing q exits the app."""
    async with app.run_test() as pilot:
        await pilot.press("q")
