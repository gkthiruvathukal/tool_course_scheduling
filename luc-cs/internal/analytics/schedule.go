package analytics

import "database/sql"

const departmentFilter = `SUBJECT = 'COMP'
	AND "CATALOG NUMBER" NOT IN ('391','398','490','499','605')
	AND "CATALOG NUMBER" NOT IN ('215','231','331','431','381','386','383','483')
	AND SECTION NOT IN ('01L','02L','03L','04L','05L','06L','700N')`

var scheduleDisplayColumns = []string{
	"FQ CATALOG NUMBER",
	"CLASS TITLE",
	"INSTRUCTOR",
	"ENROLL TOTAL",
	"TRAD MEETING PATTERN",
	"CLASS START TIME",
	"CLASS END TIME",
	"FACILITY",
}

const scheduleQuery = `
SELECT
	"SUBJECT", "WEIGHTED ENROLL TOTAL", "CATALOG NUMBER",
	"FQ CATALOG NUMBER", "FQ CLASS SECTION", "CLASS TITLE",
	"INSTRUCTOR", "ENROLL TOTAL", "TRAD MEETING PATTERN",
	"CLASS START TIME", "CLASS END TIME", "UNIT CLASS DURATION",
	"INSTRUCTIONAL TIME", "FACILITY", "COMBINED ID"
FROM schedule
WHERE ` + departmentFilter

// CourseSchedule returns the filtered COMP department schedule.
type CourseSchedule struct {
	DB            *sql.DB
	FilterZero    bool
}

func (cs *CourseSchedule) Display() (*Result, error) {
	query := scheduleQuery
	if cs.FilterZero {
		query += ` AND "ENROLL TOTAL" > 0`
	}

	allCols, allRows, err := QueryRows(cs.DB, query)
	if err != nil {
		return nil, err
	}

	// Build a column-index lookup so we can project only display columns.
	colIdx := make(map[string]int, len(allCols))
	for i, c := range allCols {
		colIdx[c] = i
	}

	rows := make([][]string, len(allRows))
	for i, r := range allRows {
		proj := make([]string, len(scheduleDisplayColumns))
		for j, name := range scheduleDisplayColumns {
			if idx, ok := colIdx[name]; ok {
				proj[j] = r[idx]
			}
		}
		rows[i] = proj
	}

	return &Result{
		Title:    "Course Schedule",
		Subtitle: "Current course schedule for the COMP department",
		Columns:  scheduleDisplayColumns,
		Rows:     rows,
	}, nil
}
