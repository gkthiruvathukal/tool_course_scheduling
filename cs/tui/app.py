from pathlib import Path
from sqlite3 import Connection

from textual.app import App, ComposeResult
from textual.binding import Binding

from cs.excel2db import readExcelToDB
from cs.tui.screens.file_picker import FilePickerScreen
from cs.tui.screens.main import MainScreen


class LucCSApp(App):
    TITLE = "LUC CS Course Scheduling Utility"
    BINDINGS = [
        Binding("f", "load_file", "Load File"),
        Binding("q", "quit", "Quit"),
    ]

    def on_mount(self) -> None:
        self.push_screen(FilePickerScreen(), callback=self._on_file_selected)

    def action_load_file(self) -> None:
        self.push_screen(FilePickerScreen(), callback=self._on_file_selected)

    def _on_file_selected(self, path: Path | None) -> None:
        if path is None:
            return
        conn: Connection = readExcelToDB(uf=path)
        if self.screen_stack and isinstance(self.screen, MainScreen):
            self.pop_screen()
        self.push_screen(MainScreen(conn=conn, file_path=path))


def main() -> None:
    LucCSApp().run()
