from sqlite3 import Connection

from pandas import DataFrame

from cs.analytics.courseSchedule import CourseSchedule
from cs.utils.analytic import Analytic
from cs.utils.result import AnalyticResult, ChartData


class AssignmentsPerFaculty(Analytic):
    def __init__(self, conn: Connection) -> None:
        self.conn = conn

    def compute(self) -> DataFrame:
        df: DataFrame = CourseSchedule(conn=self.conn).compute()

        # nunique on COMBINED ID means combined sections count as one assignment
        summary = df.groupby("INSTRUCTOR")["COMBINED ID"].nunique().reset_index()
        summary.columns = ["INSTRUCTOR", "ASSIGNMENTS"]
        summary = summary[summary["INSTRUCTOR"] != "UNKNOWN"]
        summary.sort_values("ASSIGNMENTS", ascending=False, inplace=True)
        summary.reset_index(drop=True, inplace=True)
        return summary

    def plot(self, data=None) -> list:
        return []

    def display(self, filter_zero: bool = False) -> AnalyticResult:
        df = self.compute()

        names = df["INSTRUCTOR"].tolist()
        raw_counts = df["ASSIGNMENTS"].tolist()

        # Cap overloaded faculty at 5 for display; color their bar red
        counts = [min(c, 5) for c in raw_counts]
        colors = ["red" if c > 4 else "cyan" for c in raw_counts]
        x_ticks = list(range(0, 6))

        chart = ChartData(
            chart_type="bar",
            title="Assignments per Faculty  (red bar = 5+ assignments)",
            x=names,
            y=counts,
            x_label="Assignments",
            y_label="Instructor",
            colors=colors,
            x_ticks=x_ticks,
        )

        return AnalyticResult(
            title="Assignments Per Faculty",
            subtitle=(
                "Each distinct COMBINED ID counts as one assignment — "
                "combined sections are not double-counted"
            ),
            charts=[chart],
        )
