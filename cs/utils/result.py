from dataclasses import dataclass, field

from pandas import DataFrame


@dataclass
class ChartData:
    chart_type: str  # "bar" | "scatter"
    title: str
    x: list
    y: list
    x_label: str = ""
    y_label: str = ""
    colors: list = field(
        default_factory=list
    )  # per-bar colors; empty = plotext default
    x_ticks: list = field(
        default_factory=list
    )  # explicit x-axis tick values; empty = auto


@dataclass
class AnalyticResult:
    title: str
    subtitle: str
    dataframes: list[DataFrame] = field(default_factory=list)
    df_titles: list[str] = field(default_factory=list)
    df_subtitles: list[str] = field(default_factory=list)
    charts: list[ChartData] = field(default_factory=list)
    display_columns: list[str] = field(default_factory=list)  # empty = all columns
    filter_zero: bool = False
