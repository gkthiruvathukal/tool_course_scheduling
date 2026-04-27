package tui

import (
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
)

type filePickedMsg string // carries the chosen path

type filePickerModel struct {
	fp       filepicker.Model
	selected string
	err      error
}

func newFilePickerModel() filePickerModel {
	fp := filepicker.New()
	fp.AllowedTypes = []string{".xlsx"}
	fp.ShowHidden = false

	start := filepath.Join(os.Getenv("HOME"), "Downloads")
	if _, err := os.Stat(start); err != nil {
		start, _ = os.Getwd()
	}
	fp.CurrentDirectory = start

	return filePickerModel{fp: fp}
}

func (m filePickerModel) Init() tea.Cmd {
	return m.fp.Init()
}

func (m filePickerModel) Update(msg tea.Msg) (filePickerModel, tea.Cmd) {
	var cmd tea.Cmd
	m.fp, cmd = m.fp.Update(msg)

	if didSelect, path := m.fp.DidSelectFile(msg); didSelect {
		m.selected = path
	}

	return m, cmd
}

func (m filePickerModel) View() string {
	header := headerStyle.Render("Select a schedule file (.xlsx)")
	footer := footerStyle.Render("↑/↓ navigate  ·  enter select  ·  q quit")
	return header + "\n" + m.fp.View() + "\n" + footer
}
