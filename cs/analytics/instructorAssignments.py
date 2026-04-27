from collections import defaultdict
from sqlite3 import Connection
from typing import List

from pandas import DataFrame
from pandas.core.groupby import DataFrameGroupBy

from cs.analytics.courseSchedule import CourseSchedule
from cs.utils.analytic import Analytic
from cs.utils.result import AnalyticResult


class InstructorAssignments(Analytic):
    def __init__(self, conn: Connection) -> None:
        self.conn: Connection = conn

    def compute(self, filterZeroEnrollment: bool = False) -> DataFrameGroupBy:
        df: DataFrame = CourseSchedule(conn=self.conn).compute()

        if filterZeroEnrollment:
            df = df[df["ENROLL TOTAL"] > 0]

        return df.groupby(by="INSTRUCTOR")

    def plot(self, data=None) -> list:
        return []

    def display(self, filter_zero: bool = False) -> AnalyticResult:
        by_instructor: DataFrameGroupBy = self.compute(filterZeroEnrollment=filter_zero)

        records = []

        instructor: str
        instructor_df: DataFrame
        for instructor, instructor_df in by_instructor:
            # Each distinct COMBINED ID is one assignment for this instructor.
            combined_groups = list(instructor_df.groupby(by="COMBINED ID"))
            total = len(combined_groups)

            for assign_num, (_, group_df) in enumerate(combined_groups, start=1):
                assign_label = f"{assign_num}/{total}"

                for _, row in group_df.iterrows():
                    records.append(
                        {
                            "INSTRUCTOR": instructor,
                            "ASSIGN": assign_label,
                            "SECTION": str(row["FQ CLASS SECTION"]),
                            "TITLE": str(row["CLASS TITLE"]),
                            "ENROLL": int(row["ENROLL TOTAL"]),
                            "WEIGHTED": int(row["WEIGHTED ENROLL TOTAL"]),
                            "PATTERN": str(row["TRAD MEETING PATTERN"]),
                            "START": str(row["CLASS START TIME"]),
                            "END": str(row["CLASS END TIME"]),
                        }
                    )

        return AnalyticResult(
            title="Instructor Assignments",
            subtitle=(
                "Each distinct teaching slot is one assignment — "
                "combined sections share the same assignment number"
            ),
            dataframes=[DataFrame(records)],
        )
