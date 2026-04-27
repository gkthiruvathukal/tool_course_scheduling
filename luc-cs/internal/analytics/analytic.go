// Package analytics defines the Analytic interface and shared result types.
package analytics

import (
	"database/sql"
	"fmt"
)

// Result is everything a TUI view needs to render.
type Result struct {
	Title    string
	Subtitle string
	Columns  []string    // ordered column headers for the table view
	Rows     [][]string  // one entry per column, per row
	Chart    *ChartData  // nil if this analytic has no chart
}

// ChartData describes a bar chart renderable in the terminal.
type ChartData struct {
	Title  string
	Labels []string  // y-axis labels (one per bar)
	Values []float64 // bar lengths
	Colors []string  // per-bar color name ("red", "cyan", …)
	XTicks []int     // explicit x-axis tick values; nil = auto
	XLabel string
}

// Analytic is implemented by every report.
type Analytic interface {
	Display() (*Result, error)
}

// QueryRows runs a SELECT and returns column names + string rows.
// It is a shared helper used by analytics that do simple table queries.
func QueryRows(db *sql.DB, query string, args ...any) ([]string, [][]string, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}

	var result [][]string
	vals := make([]any, len(cols))
	ptrs := make([]any, len(cols))
	for i := range vals {
		ptrs[i] = &vals[i]
	}

	for rows.Next() {
		if err := rows.Scan(ptrs...); err != nil {
			return nil, nil, err
		}
		row := make([]string, len(cols))
		for i, v := range vals {
			switch t := v.(type) {
			case []byte:
				row[i] = string(t)
			case nil:
				row[i] = ""
			default:
				row[i] = fmt.Sprintf("%v", t)
			}
		}
		result = append(result, row)
	}
	return cols, result, rows.Err()
}
