package tui

import (
	"database/sql"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type appState int

const (
	stateFilePicker appState = iota
	stateMain
)

type dbLoadedMsg struct {
	db   *sql.DB
	path string
}

// rootModel is the top-level Bubble Tea model.
type rootModel struct {
	state      appState
	filePicker filePickerModel
	main       mainModel
	width      int
	height     int
}

func newRootModel() rootModel {
	return rootModel{
		state:      stateFilePicker,
		filePicker: newFilePickerModel(),
	}
}

func (m rootModel) Init() tea.Cmd {
	return m.filePicker.Init()
}

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.state == stateMain {
			var cmd tea.Cmd
			m.main, cmd = m.main.Update(msg)
			return m, cmd
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.state == stateFilePicker {
				return m, tea.Quit
			}
		}

	case dbLoadedMsg:
		m.main = newMainModel(msg.db, msg.path, m.width, m.height)
		m.state = stateMain
		return m, m.main.Init()
	}

	switch m.state {
	case stateFilePicker:
		var cmd tea.Cmd
		m.filePicker, cmd = m.filePicker.Update(msg)

		if m.filePicker.selected != "" {
			path := m.filePicker.selected
			m.filePicker.selected = ""
			return m, loadDB(path)
		}
		return m, cmd

	case stateMain:
		var cmd tea.Cmd
		if kMsg, ok := msg.(tea.KeyMsg); ok && (kMsg.String() == "ctrl+c" || kMsg.String() == "q") {
			return m, tea.Quit
		}
		m.main, cmd = m.main.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m rootModel) View() string {
	switch m.state {
	case stateFilePicker:
		return m.filePicker.View()
	case stateMain:
		return m.main.View()
	}
	return ""
}

// loadDB opens the Excel file in a background command and returns a dbLoadedMsg.
func loadDB(path string) tea.Cmd {
	return func() tea.Msg {
		db, err := loadExcelDB(path)
		if err != nil {
			return fmt.Errorf("loading %s: %w", path, err)
		}
		return dbLoadedMsg{db: db, path: path}
	}
}
