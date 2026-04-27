from pathlib import Path
from sqlite3 import Connection

from textual.app import ComposeResult
from textual.binding import Binding
from textual.containers import Horizontal, Vertical
from textual.screen import Screen
from textual.widgets import (
    DataTable,
    Footer,
    Header,
    Label,
    ListItem,
    ListView,
)

from cs.analytics.assignmentsPerFaculty import AssignmentsPerFaculty
from cs.analytics.courseEnrollmentHealth import CourseEnrollmentHealth
from cs.analytics.courseSchedule import CourseSchedule
from cs.analytics.instructorAssignments import InstructorAssignments
from cs.tui.widgets.chart import ChartWidget
from cs.utils.result import AnalyticResult

# Ordered list of (label, key) for the sidebar
ANALYTICS = [
    ("Course Schedule", "course_schedule"),
    ("Online Only Courses", "online_courses"),
    ("Schedule Density", "schedule_density"),
    ("Enrollment Health", "enrollment_health"),
    ("Instructor Assignments", "instructor_assignments"),
    ("Zero Enrollment", "zero_enrollment"),
    ("Assignments Per Faculty", "assignments_per_faculty"),
    ("Courses by Number", "courses_by_number"),
    ("Teaching Distribution", "teaching_distribution"),
    ("Enrollment by Level", "enrollment_by_level"),
    ("In Trouble Courses", "in_trouble_courses"),
    ("Filter Schedule", "filter_schedule"),
    ("School Credit Hours", "school_credit_hours"),
]


class MainScreen(Screen):
    DEFAULT_CSS = """
    MainScreen {
        layout: vertical;
    }
    #body {
        layout: horizontal;
        height: 1fr;
    }
    #sidebar {
        width: 28;
        height: 100%;
        border-right: solid $primary;
        padding: 0 1;
    }
    #sidebar-title {
        text-style: bold;
        color: $accent;
        padding: 1 0;
        width: 1fr;
        text-align: center;
    }
    #nav-list {
        height: 1fr;
    }
    #content {
        width: 1fr;
        height: 100%;
        padding: 1 2;
        layout: vertical;
    }
    #content-title {
        text-style: bold;
        color: $accent;
    }
    #content-subtitle {
        color: $text-muted;
        margin-bottom: 1;
    }
    #schedule-table {
        height: 1fr;
    }
    #placeholder {
        color: $text-muted;
        padding: 2;
    }
    """

    BINDINGS = [
        Binding("q", "app.quit", "Quit"),
        Binding("f", "app.load_file", "Load File"),
        Binding("r", "reload", "Reload"),
    ]

    def __init__(self, conn: Connection, file_path: Path) -> None:
        super().__init__()
        self._conn = conn
        self._file_path = file_path
        self._active_key = "course_schedule"

    def compose(self) -> ComposeResult:
        yield Header()
        with Horizontal(id="body"):
            with Vertical(id="sidebar"):
                yield Label("Analytics", id="sidebar-title")
                yield ListView(
                    *[
                        ListItem(Label(label), id=f"nav-{key}")
                        for label, key in ANALYTICS
                    ],
                    id="nav-list",
                )
            with Vertical(id="content"):
                yield Label("", id="content-title")
                yield Label("", id="content-subtitle")
                yield DataTable(
                    id="schedule-table", zebra_stripes=True, cursor_type="row"
                )
                yield ChartWidget(id="chart-widget")
                yield Label("", id="placeholder")
        yield Footer()

    def on_mount(self) -> None:
        self.sub_title = self._file_path.name
        self.query_one("#chart-widget", ChartWidget).display = False
        self._show_course_schedule()

    def on_list_view_selected(self, event: ListView.Selected) -> None:
        if event.list_view.id != "nav-list":
            return
        item_id = event.item.id or ""
        key = item_id.removeprefix("nav-")
        self._activate(key)

    def _activate(self, key: str) -> None:
        self._active_key = key
        if key == "course_schedule":
            self._show_course_schedule()
        elif key == "enrollment_health":
            self._show_enrollment_health()
        elif key == "instructor_assignments":
            self._show_instructor_assignments()
        elif key == "assignments_per_faculty":
            self._show_assignments_per_faculty()
        else:
            self._show_placeholder(key)

    def _show_course_schedule(self) -> None:
        self._render_result(CourseSchedule(conn=self._conn).display())

    def _show_enrollment_health(self) -> None:
        self._render_result(CourseEnrollmentHealth(conn=self._conn).display())

    def _show_instructor_assignments(self) -> None:
        self._render_result(InstructorAssignments(conn=self._conn).display())

    def _show_assignments_per_faculty(self) -> None:
        self._render_result(AssignmentsPerFaculty(conn=self._conn).display())

    def _render_result(self, result: AnalyticResult) -> None:
        self.query_one("#content-title", Label).update(result.title)
        self.query_one("#content-subtitle", Label).update(result.subtitle)
        self.query_one("#placeholder", Label).update("")

        table = self.query_one("#schedule-table", DataTable)
        table.clear(columns=True)
        table.display = bool(result.dataframes)

        chart = self.query_one("#chart-widget", ChartWidget)
        if result.charts:
            chart.display = True
            chart.update_chart(result.charts[0])
        else:
            chart.display = False

        if not result.dataframes:
            return

        df = result.dataframes[0]
        cols = (
            [c for c in result.display_columns if c in df.columns]
            if result.display_columns
            else list(df.columns)
        )
        table.add_columns(*cols)

        for _, row in df[cols].iterrows():
            table.add_row(*[str(v) for v in row])

    def _show_placeholder(self, key: str) -> None:
        label = next((lbl for lbl, k in ANALYTICS if k == key), key)
        self.query_one("#content-title", Label).update(label)
        self.query_one("#content-subtitle", Label).update("")
        table = self.query_one("#schedule-table", DataTable)
        table.clear(columns=True)
        table.display = False
        self.query_one("#placeholder", Label).update(
            f"'{label}' is not yet implemented in the TUI."
        )

    def action_reload(self) -> None:
        self._activate(self._active_key)
