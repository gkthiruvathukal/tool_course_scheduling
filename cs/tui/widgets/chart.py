from rich.markup import escape
from textual.widgets import Static

from cs.utils.result import ChartData


class ChartWidget(Static):
    DEFAULT_CSS = """
    ChartWidget {
        padding: 1 2;
        border: solid $panel;
        margin-top: 1;
    }
    """

    def __init__(self, *args, **kwargs) -> None:
        super().__init__(*args, **kwargs)
        self._data: ChartData | None = None

    def update_chart(self, data: ChartData) -> None:
        self._data = data
        # 1 line per bar + title(2) + axis(2) + ticks(1) + label(1) + border+padding(4)
        self.styles.height = len(data.y) + 10
        self.update(self._render())

    def _render(self) -> str:
        if self._data is None:
            return ""
        if self._data.chart_type == "bar":
            return self._render_bar(self._data)
        return f"Unsupported chart type: {self._data.chart_type}"

    def _render_bar(self, data: ChartData) -> str:
        names = [str(n) for n in data.x]
        values = data.y
        colors = data.colors if data.colors else ["cyan"] * len(names)
        ticks = data.x_ticks

        max_val = max(ticks) if ticks else (max(values) if values else 1)
        max_label = max((len(n) for n in names), default=10)
        bar_width = 50

        out = f"[bold]{escape(data.title)}[/bold]\n\n"

        for name, value, color in zip(names, values, colors):
            bar_chars = round(value / max_val * bar_width) if max_val else 0
            label = escape(name.rjust(max_label))
            out += f"{label} │[{color}]{'█' * bar_chars} {value}[/{color}]\n"

        out += f"{' ' * (max_label + 2)}└{'─' * bar_width}\n"

        if ticks:
            gap = bar_width - len(str(ticks[-1]))
            out += f"{' ' * (max_label + 3)}{ticks[0]}{' ' * gap}{ticks[-1]}\n"

        if data.x_label:
            center = max_label + 2 + bar_width // 2 - len(data.x_label) // 2
            out += f"{' ' * center}{escape(data.x_label)}\n"

        return out
