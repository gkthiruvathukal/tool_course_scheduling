package tui

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/gkthiruvathukal/luc-cs/internal/analytics"
)

// renderChart returns a string rendering of a horizontal bar chart.
func renderChart(cd *analytics.ChartData, width int) string {
	if cd == nil || len(cd.Labels) == 0 {
		return placeholderStyle.Render("No chart data.")
	}

	maxLabel := 0
	for _, l := range cd.Labels {
		if len(l) > maxLabel {
			maxLabel = len(l)
		}
	}

	// Reserve space for: label + " │ " + bar + " " + value
	barWidth := width - maxLabel - 4
	if barWidth < 4 {
		barWidth = 4
	}

	maxVal := 0.0
	for _, v := range cd.Values {
		if v > maxVal {
			maxVal = v
		}
	}

	var sb strings.Builder
	sb.WriteString(titleStyle.Render(cd.Title))
	sb.WriteString("\n\n")

	for i, label := range cd.Labels {
		val := cd.Values[i]
		color := "cyan"
		if i < len(cd.Colors) && cd.Colors[i] != "" {
			color = cd.Colors[i]
		}

		barLen := 0
		if maxVal > 0 {
			barLen = int(math.Round(val / maxVal * float64(barWidth)))
		}

		bar := strings.Repeat("█", barLen)
		styled := lipgloss.NewStyle().Foreground(lipgloss.Color(colorCode(color))).Render(bar)

		paddedLabel := fmt.Sprintf("%*s", maxLabel, label)
		sb.WriteString(fmt.Sprintf("%s │%s %.0f\n", paddedLabel, styled, val))
	}

	// x-axis
	axis := strings.Repeat("─", barWidth+2)
	sb.WriteString(fmt.Sprintf("%s┴%s\n", strings.Repeat(" ", maxLabel), axis))

	// x-axis ticks
	if len(cd.XTicks) > 0 && maxVal > 0 {
		tickLine := strings.Repeat(" ", maxLabel+1)
		prev := 0
		for _, t := range cd.XTicks {
			pos := int(math.Round(float64(t) / maxVal * float64(barWidth)))
			gap := pos - prev
			if gap > 0 {
				tickLine += strings.Repeat(" ", gap-len(fmt.Sprintf("%d", t))) + fmt.Sprintf("%d", t)
			}
			prev = pos
		}
		sb.WriteString(tickLine + "\n")
	}

	if cd.XLabel != "" {
		indent := strings.Repeat(" ", maxLabel+1+barWidth/2-len(cd.XLabel)/2)
		sb.WriteString(indent + subtitleStyle.Render(cd.XLabel) + "\n")
	}

	return sb.String()
}

// colorCode maps friendly color names to lipgloss terminal color codes.
func colorCode(name string) string {
	switch strings.ToLower(name) {
	case "red":
		return "196"
	case "cyan":
		return "51"
	case "green":
		return "46"
	case "yellow":
		return "226"
	case "blue":
		return "33"
	default:
		return "51" // default cyan
	}
}
