from sqlite3 import Connection

import pandas
from pandas import DataFrame

from cs.utils.result import AnalyticResult


class CourseSchedule:
    def __init__(self, conn: Connection) -> None:
        self.conn: Connection = conn
        self.departmentFilters: dict[str, str] = {
            "COMP": """SUBJECT = 'COMP' AND "CATALOG NUMBER" NOT IN ('391', '398', '490', '499', '605') AND "CATALOG NUMBER" NOT IN ('215', '231', '331', '431', '381', '386', '383', '483') AND SECTION NOT IN ('01L', '02L', '03L', '04L', '05L', '06L', '700N')"""  # noqa: E501
        }

    def compute(self, filterZeroEnrollment: bool = False) -> DataFrame:
        whereClauses: str = "WHERE " + " or ".join(
            ["(" + self.departmentFilters[f] + ")" for f in self.departmentFilters]
        )

        query: str = (
            """SELECT SUBJECT, "WEIGHTED ENROLL TOTAL", "CATALOG NUMBER", """
            """"FQ CATALOG NUMBER", "FQ CLASS SECTION", "CLASS TITLE", INSTRUCTOR, """
            """"ENROLL TOTAL", "TRAD MEETING PATTERN", "CLASS START TIME", """
            """"CLASS END TIME", "UNIT CLASS DURATION", "INSTRUCTIONAL TIME", """
            """FACILITY, "COMBINED ID" FROM schedule """ + whereClauses + ";"
        )

        df: DataFrame = pandas.read_sql_query(sql=query, con=self.conn)  # nosec
        df.reset_index(drop=True, inplace=True)

        if filterZeroEnrollment:
            df = df[df["ENROLL TOTAL"] > 0]

        return df

    def plot(self, data=None) -> list:
        return []

    def display(self, filter_zero: bool = False) -> AnalyticResult:
        df = self.compute(filterZeroEnrollment=filter_zero)
        return AnalyticResult(
            title="Course Schedule",
            subtitle="Current course schedule for the COMP department",
            dataframes=[df],
            display_columns=[
                "FQ CATALOG NUMBER",
                "CLASS TITLE",
                "INSTRUCTOR",
                "ENROLL TOTAL",
                "TRAD MEETING PATTERN",
                "CLASS START TIME",
                "CLASS END TIME",
                "FACILITY",
            ],
            filter_zero=filter_zero,
        )
