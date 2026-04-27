package tui

import (
	"database/sql"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gkthiruvathukal/luc-cs/internal/db"
)

// Run starts the Bubble Tea application.
func Run() error {
	p := tea.NewProgram(newRootModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// loadExcelDB is the bridge between the tui package and db.LoadExcel.
func loadExcelDB(path string) (*sql.DB, error) {
	return db.LoadExcel(path)
}
