from sqlite3 import Connection
from typing import List, Tuple

from pandas import DataFrame
from pandas.core.groupby import DataFrameGroupBy

from cs.analytics.courseSchedule import CourseSchedule
from cs.utils.analytic import Analytic
from cs.utils.result import AnalyticResult

_FILTER_FIELDS: List[str] = [
    "FQ CLASS SECTION",
    "CLASS TITLE",
    "INSTRUCTOR",
    "ENROLL TOTAL",
    "TRAD MEETING PATTERN",
    "CLASS START TIME",
    "CLASS END TIME",
]


class CourseEnrollmentHealth(Analytic):
    def __init__(self, conn: Connection) -> None:
        self.conn = conn

    def compute(
        self, filterZeroEnrollment: bool = False
    ) -> List[Tuple[str, DataFrame, str, int]]:
        data: List[Tuple[str, DataFrame, str, int]] = []

        df: DataFrame = CourseSchedule(conn=self.conn).compute()

        if filterZeroEnrollment:
            df = df[df["ENROLL TOTAL"] > 0]

        groups: DataFrameGroupBy = df.groupby(by="COMBINED ID")
        report: List[Tuple[int, str, DataFrame]] = []

        for name, group in groups:
            report.append((group["WEIGHTED ENROLL TOTAL"].sum(), name, group))

        report = sorted(report, key=lambda t: t[0])

        for group_sum, name, group in report:
            filtered = group[_FILTER_FIELDS]
            color = "blue"
            if group_sum < 12:
                color = "red"
            elif group_sum > 32:
                color = "green"
            data.append((name, filtered, color, group_sum))

        return data

    def plot(self, data=None) -> list:
        return []

    def display(self, filter_zero: bool = False) -> AnalyticResult:
        data = self.compute(filterZeroEnrollment=filter_zero)

        records = []
        for _, group_df, color, group_sum in data:
            if color == "red":
                status = "⚠ Low"
            elif color == "green":
                status = "✓ Healthy"
            else:
                status = "· Normal"

            # Emit one row per section so combined sections are all visible.
            # Sections sharing the same COMBINED ID appear consecutively with
            # the same STATUS and GRP WEIGHTED, making the grouping apparent.
            for _, row in group_df.iterrows():
                records.append(
                    {
                        "STATUS": status,
                        "GRP WEIGHTED": int(group_sum),
                        "SECTION": str(row["FQ CLASS SECTION"]),
                        "TITLE": str(row["CLASS TITLE"]),
                        "INSTRUCTOR": str(row["INSTRUCTOR"]),
                        "ENROLL": int(row["ENROLL TOTAL"]),
                        "PATTERN": str(row["TRAD MEETING PATTERN"]),
                        "START": str(row["CLASS START TIME"]),
                        "END": str(row["CLASS END TIME"]),
                    }
                )

        summary = DataFrame(records)

        return AnalyticResult(
            title="Course Enrollment Health",
            subtitle=(
                "Sorted by group weighted enrollment — "
                "⚠ < 12 low   · 12–32 normal   ✓ > 32 healthy"
            ),
            dataframes=[summary],
        )
