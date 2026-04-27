from pathlib import Path

from textual.app import ComposeResult
from textual.binding import Binding
from textual.containers import Vertical
from textual.screen import Screen
from textual.widgets import DirectoryTree, Footer, Header, Label


class _XlsxTree(DirectoryTree):
    """DirectoryTree that shows only directories and .xlsx files."""

    def filter_paths(self, paths: list[Path]) -> list[Path]:
        return [p for p in paths if p.is_dir() or p.suffix.lower() == ".xlsx"]


class FilePickerScreen(Screen[Path | None]):
    DEFAULT_CSS = """
    FilePickerScreen {
        layout: vertical;
    }
    #picker-title {
        text-style: bold;
        color: $accent;
        padding: 1 2;
    }
    #file-tree {
        height: 1fr;
        margin: 0 1;
        border: solid $panel;
    }
    #hint {
        color: $text-muted;
        padding: 0 2 1 2;
    }
    """

    BINDINGS = [
        Binding("escape", "dismiss_none", "Cancel"),
    ]

    def __init__(self) -> None:
        super().__init__()
        downloads = Path.home() / "Downloads"
        self._start_path = downloads if downloads.exists() else Path.home()

    def compose(self) -> ComposeResult:
        yield Header()
        with Vertical():
            yield Label("Select a Course Schedule (.xlsx) file", id="picker-title")
            yield _XlsxTree(self._start_path, id="file-tree")
            yield Label("↑↓ navigate · Enter expand/select · Esc cancel", id="hint")
        yield Footer()

    def on_directory_tree_file_selected(
        self, event: DirectoryTree.FileSelected
    ) -> None:
        if event.path.suffix.lower() == ".xlsx":
            self.dismiss(event.path)

    def action_dismiss_none(self) -> None:
        self.dismiss(None)
