// Package db handles loading an xlsx course schedule export into SQLite.
package db

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	_ "modernc.org/sqlite"
)

const createTableSQL = `
CREATE TABLE IF NOT EXISTS schedule (
	"SUBJECT"              TEXT,
	"CATALOG NUMBER"       TEXT,
	"SECTION"              TEXT,
	"CLASS TITLE"          TEXT,
	"INSTRUCTOR"           TEXT,
	"FACILITY"             TEXT,
	"MEETING PATTERN"      TEXT,
	"CLASS START TIME"     TEXT,
	"CLASS END TIME"       TEXT,
	"ENROLLMENT TOTAL"     INTEGER,
	"ENROLL TOTAL"         INTEGER,
	"FQ CATALOG NUMBER"    TEXT,
	"FQ CLASS SECTION"     TEXT,
	"TRAD MEETING PATTERN" TEXT,
	"UNIT CLASS DURATION"  INTEGER,
	"COMBINED ID"          TEXT,
	"INSTRUCTIONAL TIME"   INTEGER,
	"WEIGHTED ENROLL TOTAL" REAL,
	"WEIGHTED SCH TOTAL"   INTEGER
)`

const insertSQL = `
INSERT INTO schedule VALUES (
	?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?
)`

// LoadExcel reads the xlsx file at path and returns a populated in-memory DB.
func LoadExcel(path string) (*sql.DB, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

	sheet := f.GetSheetName(0)
	rawRows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("reading sheet: %w", err)
	}
	if len(rawRows) < 2 {
		return nil, fmt.Errorf("spreadsheet has no data rows")
	}

	// Map uppercase header name -> column index
	idx := make(map[string]int)
	for i, h := range rawRows[0] {
		idx[strings.ToUpper(strings.TrimSpace(h))] = i
	}

	col := func(row []string, name string) string {
		i, ok := idx[name]
		if !ok || i >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[i])
	}

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, fmt.Errorf("opening sqlite: %w", err)
	}

	if _, err = db.Exec(createTableSQL); err != nil {
		return nil, fmt.Errorf("creating table: %w", err)
	}

	seen := make(map[string]bool)

	for _, row := range rawRows[1:] {
		subject := col(row, "SUBJECT")
		catalogNum := cleanString(col(row, "CATALOG NUMBER"))
		section := cleanString(col(row, "SECTION"))
		classTitle := col(row, "CLASS TITLE")
		instructor := col(row, "INSTRUCTOR")
		facility := col(row, "FACILITY")
		meetingPattern := col(row, "MEETING PATTERN")
		startRaw := col(row, "CLASS START TIME")
		endRaw := col(row, "CLASS END TIME")
		enrollTotalStr := col(row, "ENROLLMENT TOTAL")

		if instructor == "" {
			instructor = "Turing,Alan"
		}
		if facility == "" {
			facility = "Doyle Hall"
		}

		fqCatalog := subject + "-" + catalogNum
		fqSection := catalogNum + "-" + section

		// Deduplicate on FQ CLASS SECTION
		if seen[fqSection] {
			continue
		}
		seen[fqSection] = true

		startTime := parseTime(startRaw)
		endTime := parseTime(endRaw)
		tradPattern := normalizeMeetingPattern(meetingPattern)
		duration := timeToMinutes(endTime) - timeToMinutes(startTime)
		if duration < 0 {
			duration = 0
		}

		combinedID := fmt.Sprintf("(%s,%s,%s,%s,%s)",
			instructor, facility, tradPattern, startTime, endTime)

		enrollTotal := cleanInt(enrollTotalStr)
		weightedEnroll := computeWeightedEnrollment(catalogNum, float64(enrollTotal))
		weightedSCH := computeWeightedSCH(catalogNum, weightedEnroll)

		_, err = db.Exec(insertSQL,
			subject, catalogNum, section, classTitle, instructor, facility,
			meetingPattern, startTime, endTime,
			enrollTotal, enrollTotal,
			fqCatalog, fqSection, tradPattern, duration, combinedID,
			0, // INSTRUCTIONAL TIME — placeholder
			weightedEnroll, weightedSCH,
		)
		if err != nil {
			return nil, fmt.Errorf("inserting row %q: %w", fqSection, err)
		}
	}

	return db, nil
}

// --- helpers -----------------------------------------------------------------

var timeFormats = []string{"3:04 PM", "3:04:05 PM", "15:04", "15:04:05"}

func parseTime(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	for _, layout := range timeFormats {
		if t, err := time.Parse(layout, s); err == nil {
			return t.Format("15:04:05")
		}
	}
	return ""
}

func timeToMinutes(s string) int {
	parts := strings.SplitN(s, ":", 3)
	if len(parts) < 2 {
		return 0
	}
	h, _ := strconv.Atoi(parts[0])
	m, _ := strconv.Atoi(parts[1])
	return h*60 + m
}

func normalizeMeetingPattern(p string) string {
	if p == "" {
		return "No Meeting Pattern"
	}
	// Order matters: replace longer patterns before shorter substrings
	p = strings.ReplaceAll(p, "TTh", "TR")
	p = strings.ReplaceAll(p, "Th", "R")
	p = strings.ReplaceAll(p, "SA", "S")
	p = strings.ReplaceAll(p, "SU", "X")
	return p
}

func computeWeightedEnrollment(catalogNum string, enroll float64) float64 {
	switch {
	case catalogNum >= "400":
		return enroll * 5.0 / 3.0
	case catalogNum >= "300":
		return enroll * 1.0
	default:
		return enroll
	}
}

func computeWeightedSCH(catalogNum string, weightedEnroll float64) int {
	credits := 3.0
	if catalogNum == "395" {
		credits = 1.0
	}
	return int(credits * weightedEnroll)
}

// cleanString removes trailing ".0" that Excel sometimes adds to numeric cells.
func cleanString(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, ".0") {
		return s[:len(s)-2]
	}
	return s
}

func cleanInt(s string) int {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, "."); i >= 0 {
		s = s[:i]
	}
	n, _ := strconv.Atoi(s)
	return n
}
