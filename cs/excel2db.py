from datetime import datetime
from pathlib import Path
from sqlite3 import Connection, connect
from typing import Union

import pandas
from pandas import DataFrame, Series, read_excel

from cs.utils import datetimeToMinutes


def _computeInstructionalTime(row: Series):
    """
    Compute the instructional time for a given row.

    This function calculates the instructional time based on the provided row
    data. Currently, it returns 0 as a placeholder.

    :param row: The row of data for which to compute the instructional time.
    :type row: Series
    :return: The computed instructional time.
    :rtype: int
    """
    return 0


def _computeTotalTime(row: Series) -> int:
    """
    Compute the total class duration in minutes for a given row.

    This function calculates the total duration of a class session in minutes
    based on the start and end times provided in the row data.

    :param row: The row of data containing class start and end times.
    :type row: Series
    :return: The total class duration in minutes.
    :rtype: int
    """
    startTime: datetime = pandas.to_datetime(
        row["CLASS START TIME"],
    )
    endTime: datetime = pandas.to_datetime(
        row["CLASS END TIME"],
    )

    if pandas.isnull(obj=startTime):
        return 0

    if pandas.isnull(endTime):
        return 0

    startMinutes: int = datetimeToMinutes(dt=startTime)
    endMinutes: int = datetimeToMinutes(dt=endTime)

    return endMinutes - startMinutes


def _computeWeightedEnrollment(row: Series):
    """
    Compute the weighted enrollment for a given row.

    This function calculates the weighted enrollment based on the course
    level. Upper-level courses (400 and above) have a higher weight.

    :param row: The row of data for which to compute the weighted enrollment.
    :type row: Series
    :return: The computed weighted enrollment.
    :rtype: float
    """
    if "300" <= row["CATALOG NUMBER"] < "400":
        return row["ENROLL TOTAL"] * 1.0
    elif "400" <= row["CATALOG NUMBER"]:
        return row["ENROLL TOTAL"] * 5 / 3
    else:
        return row["ENROLL TOTAL"]


def _computeWeightedSchedule(row: Series):
    """
    Compute the weighted schedule for a given row.

    This function calculates the weighted schedule based on the course credits
    and the weighted enrollment total.

    :param row: The row of data for which to compute the weighted schedule.
    :type row: Series
    :return: The computed weighted schedule.
    :rtype: int
    """
    credits = 3
    if row["CATALOG NUMBER"] in set(["395"]):
        credits = 1
    return int(credits * row["WEIGHTED ENROLL TOTAL"])


def _createCombinedID(row: Series) -> str:
    instructor: str = row["INSTRUCTOR"]
    facility: str = row["FACILITY"]
    meetingPattern: str = row["TRAD MEETING PATTERN"]

    startTime: str = row["CLASS START TIME"]
    endTime: str = row["CLASS END TIME"]

    return f"({instructor},{facility},{meetingPattern},{startTime},{endTime})"


def readExcelToDB(uf: Union[str, Path], dbPath: str = ":memory:") -> Connection:
    """
    Read an Excel file and populate the database with the data.

    This function reads course schedule data from an uploaded Excel file,
    processes it, and stores it in an SQLite database.

    :param uf: The uploaded Excel file.
    :type uf: UploadedFile
    :param dbPath: The path to the SQLite database file, defaults to
        ":memory:".
    :type dbPath: str, optional
    :return: The SQLite database connection.
    :rtype: Connection
    """
    conn: Connection = connect(database=dbPath, check_same_thread=False)

    df: DataFrame = read_excel(io=uf, engine="openpyxl")

    df.columns = df.columns.str.upper()
    df["ENROLL TOTAL"] = df["ENROLLMENT TOTAL"]

    df["INSTRUCTOR"] = df["INSTRUCTOR"].fillna(value="Turing,Alan")
    df["FACILITY"] = df["FACILITY"].fillna(value="Doyole Hall")

    df["FQ CATALOG NUMBER"] = df["SUBJECT"] + "-" + df["CATALOG NUMBER"]
    df["FQ CLASS SECTION"] = df["CATALOG NUMBER"] + "-" + df["SECTION"]

    df.drop_duplicates(
        subset=["FQ CLASS SECTION"],
        inplace=True,
        ignore_index=True,
    )

    df["CLASS START TIME"] = pandas.to_datetime(
        df["CLASS START TIME"],
        format="%I:%M %p",
    ).dt.strftime("%H:%M:%S")

    df["CLASS END TIME"] = pandas.to_datetime(
        df["CLASS END TIME"],
        format="%I:%M %p",
    ).dt.strftime("%H:%M:%S")

    df["TRAD MEETING PATTERN"] = (
        df["MEETING PATTERN"]
        # .replace("TR", "R")
        .replace("Th", "R")
        .replace("TTh", "TR")
        # .replace("TTR", "TR")
        .replace("SA", "S")
        .replace("SU", "X")
        .fillna("No Meeting Pattern")
    )

    df["UNIT CLASS DURATION"] = df.apply(_computeTotalTime, axis=1)

    df["COMBINED ID"] = df.apply(_createCombinedID, axis=1)

    df["INSTRUCTIONAL TIME"] = df.apply(_computeInstructionalTime, axis=1)

    df["WEIGHTED ENROLL TOTAL"] = df.apply(
        _computeWeightedEnrollment,
        axis=1,
    )

    df["WEIGHTED SCH TOTAL"] = df.apply(_computeWeightedSchedule, axis=1)

    df.to_sql(
        name="schedule",
        con=conn,
        if_exists="replace",
        index=False,
    )

    return conn
