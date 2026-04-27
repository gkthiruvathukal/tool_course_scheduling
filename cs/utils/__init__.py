from datetime import datetime


def datetimeToMinutes(dt: datetime) -> int:
    return dt.hour * 60 + dt.minute
