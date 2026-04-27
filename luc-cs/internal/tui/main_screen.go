package tui

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gkthiruvathukal/luc-cs/internal/analytics"
)

// analyticsLoadedMsg carries the result of running an analytic.
type analyticsLoadedMsg struct {
	result *analytics.Result
	err    error
}

type focus int

const (
	focusSidebar      focus = iota
	focusContent
	focusColumnPicker
)

// sidebarItem implements list.Item.
type sidebarItem struct {
	title    string
	analytic analytics.Analytic
}

func (s sidebarItem) Title() string       { return s.title }
func (s sidebarItem) Description() string { return "" }
func (s sidebarItem) FilterValue() string { return s.title }

// mainModel is the main screen after a file is loaded.
type mainModel struct {
	db       *sql.DB
	filePath string
	width    int
	height   int

	sidebar  list.Model
	tbl      table.Model
	focus    focus
	loading  bool
	errMsg   string
	result   *analytics.Result

	colOffset    int
	colVisible   []bool // parallel to result.Columns
	pickerCursor int
}

func newMainModel(db *sql.DB, path string, w, h int) mainModel {
	items := []list.Item{
		sidebarItem{title: "Course Schedule", analytic: &analytics.CourseSchedule{DB: db}},
	}

	l := list.New(items, list.NewDefaultDelegate(), sidebarWidth, h-4)
	l.Title = "Analytics"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle

	tbl := table.New(
		table.WithFocused(false),
		table.WithHeight(h-6),
	)
	ts := table.DefaultStyles()
	ts.Header = ts.Header.Bold(true).Foreground(lipgloss.Color("212"))
	ts.Selected = ts.Selected.Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57"))
	tbl.SetStyles(ts)

	return mainModel{
		db:       db,
		filePath: path,
		width:    w,
		height:   h,
		sidebar:  l,
		tbl:      tbl,
		focus:    focusSidebar,
	}
}

func (m mainModel) Init() tea.Cmd {
	return m.loadSelected()
}

func (m mainModel) loadSelected() tea.Cmd {
	item, ok := m.sidebar.SelectedItem().(sidebarItem)
	if !ok {
		return nil
	}
	a := item.analytic
	return func() tea.Msg {
		r, err := a.Display()
		return analyticsLoadedMsg{result: r, err: err}
	}
}

func (m mainModel) Update(msg tea.Msg) (mainModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.sidebar.SetHeight(m.height - 4)
		m.tbl.SetHeight(m.height - 6)
		m.applyTableData()

	case analyticsLoadedMsg:
		m.loading = false
		m.colOffset = 0
		m.pickerCursor = 0
		if msg.err != nil {
			m.errMsg = msg.err.Error()
		} else {
			m.result = msg.result
			m.errMsg = ""
			m.colVisible = make([]bool, len(m.result.Columns))
			for i := range m.colVisible {
				m.colVisible[i] = true
			}
			m.applyTableData()
		}
		return m, nil

	case tea.KeyMsg:
		switch m.focus {
		case focusColumnPicker:
			return m.updatePicker(msg)
		case focusSidebar:
			return m.updateSidebar(msg, &cmds)
		case focusContent:
			return m.updateContent(msg, &cmds)
		}
	}

	// Non-key messages: pass to active pane.
	switch m.focus {
	case focusSidebar:
		var cmd tea.Cmd
		m.sidebar, cmd = m.sidebar.Update(msg)
		cmds = append(cmds, cmd)
	case focusContent:
		var cmd tea.Cmd
		m.tbl, cmd = m.tbl.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m mainModel) updatePicker(msg tea.KeyMsg) (mainModel, tea.Cmd) {
	n := len(m.result.Columns)
	switch msg.String() {
	case "up", "k":
		if m.pickerCursor > 0 {
			m.pickerCursor--
		}
	case "down", "j":
		if m.pickerCursor < n-1 {
			m.pickerCursor++
		}
	case " ":
		m.colVisible[m.pickerCursor] = !m.colVisible[m.pickerCursor]
		m.colOffset = 0
		m.applyTableData()
	case "a":
		for i := range m.colVisible {
			m.colVisible[i] = true
		}
		m.colOffset = 0
		m.applyTableData()
	case "n":
		for i := range m.colVisible {
			m.colVisible[i] = false
		}
		m.applyTableData()
	case "esc", "c", "enter":
		m.focus = focusContent
		m.tbl.Focus()
	}
	return m, nil
}

func (m mainModel) updateSidebar(msg tea.KeyMsg, cmds *[]tea.Cmd) (mainModel, tea.Cmd) {
	switch msg.String() {
	case "tab":
		m.focus = focusContent
		m.tbl.Focus()
		return m, nil
	case "c":
		if m.result != nil {
			m.focus = focusColumnPicker
			m.tbl.Blur()
		}
		return m, nil
	case "enter":
		m.loading = true
		m.result = nil
		m.colOffset = 0
		return m, m.loadSelected()
	}
	var cmd tea.Cmd
	m.sidebar, cmd = m.sidebar.Update(msg)
	*cmds = append(*cmds, cmd)
	// Reload when arrow keys change selection.
	m.loading = true
	m.result = nil
	m.colOffset = 0
	*cmds = append(*cmds, m.loadSelected())
	return m, tea.Batch(*cmds...)
}

func (m mainModel) updateContent(msg tea.KeyMsg, cmds *[]tea.Cmd) (mainModel, tea.Cmd) {
	switch msg.String() {
	case "tab":
		m.focus = focusSidebar
		m.tbl.Blur()
		return m, nil
	case "c":
		if m.result != nil {
			m.focus = focusColumnPicker
			m.tbl.Blur()
		}
		return m, nil
	case "]", "l":
		if m.result != nil && m.colOffset < len(m.result.Columns)-1 {
			m.colOffset++
			m.applyTableData()
		}
		return m, nil
	case "[":
		if m.colOffset > 0 {
			m.colOffset--
			m.applyTableData()
		}
		return m, nil
	}
	var cmd tea.Cmd
	m.tbl, cmd = m.tbl.Update(msg)
	return m, cmd
}

func (m *mainModel) contentWidth() int {
	w := m.width - sidebarWidth - 2
	if w < 10 {
		return 10
	}
	return w
}

func (m *mainModel) applyTableData() {
	if m.result == nil {
		return
	}

	allCols := m.result.Columns
	allRows := m.result.Rows
	cw := m.contentWidth()

	// Natural per-column width: max(header+2, 12), capped at 24.
	naturalWidths := make([]int, len(allCols))
	for i, c := range allCols {
		w := len(c) + 2
		if w < 12 {
			w = 12
		}
		if w > 24 {
			w = 24
		}
		naturalWidths[i] = w
	}

	// Collect indices of visible columns.
	visible := make([]int, 0, len(allCols))
	for i, on := range m.colVisible {
		if on {
			visible = append(visible, i)
		}
	}

	// Clamp offset to visible range.
	if m.colOffset >= len(visible) && len(visible) > 0 {
		m.colOffset = len(visible) - 1
	}

	// Select visible columns from offset that fit in contentWidth.
	var cols []table.Column
	used := 0
	for _, idx := range visible[m.colOffset:] {
		w := naturalWidths[idx]
		if used+w > cw && len(cols) > 0 {
			break
		}
		cols = append(cols, table.Column{Title: allCols[idx], Width: w})
		used += w
	}

	m.tbl.SetColumns(cols)
	m.tbl.SetWidth(cw)

	// Project rows to only the shown columns.
	shownIndices := make([]int, len(cols))
	for j := range cols {
		shownIndices[j] = visible[m.colOffset+j]
	}

	rows := make([]table.Row, len(allRows))
	for i, r := range allRows {
		proj := make([]string, len(cols))
		for j, idx := range shownIndices {
			if idx < len(r) {
				proj[j] = r[idx]
			}
		}
		rows[i] = table.Row(proj)
	}
	m.tbl.SetRows(rows)
}

func (m mainModel) View() string {
	header := headerStyle.Render(fmt.Sprintf("luc-cs  ·  %s", m.filePath))
	footer := m.footerView()
	sidebar := sidebarStyle.Height(m.height - 4).Render(m.sidebar.View())

	var content string
	cw := m.contentWidth()

	switch {
	case m.loading:
		content = placeholderStyle.Render("Loading…")
	case m.errMsg != "":
		content = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("Error: " + m.errMsg)
	case m.result == nil:
		content = placeholderStyle.Render("Select an analytic from the sidebar.")
	case m.focus == focusColumnPicker:
		content = m.pickerView(cw)
	case m.result.Chart != nil:
		content = renderChart(m.result.Chart, cw)
	default:
		title := titleStyle.Render(m.result.Title)
		sub := subtitleStyle.Render(m.result.Subtitle)
		content = lipgloss.JoinVertical(lipgloss.Left, title, sub, "", m.tbl.View())
	}

	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, content)
	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

func (m mainModel) footerView() string {
	switch m.focus {
	case focusColumnPicker:
		return footerStyle.Render("↑/↓ move  ·  space toggle  ·  a all  ·  n none  ·  esc close")
	case focusContent:
		hint := ""
		if m.result != nil {
			on := 0
			for _, v := range m.colVisible {
				if v {
					on++
				}
			}
			hint = fmt.Sprintf("  ·  cols %d/%d shown  ·  [/] scroll  ·  c pick cols", on, len(m.result.Columns))
		}
		return footerStyle.Render("tab focus  ·  ↑/↓ rows" + hint + "  ·  q quit")
	default:
		return footerStyle.Render("tab focus  ·  ↑/↓ navigate  ·  c pick cols  ·  q quit")
	}
}

func (m mainModel) pickerView(width int) string {
	var sb strings.Builder
	sb.WriteString(titleStyle.Render("Column Visibility"))
	sb.WriteString("\n")
	sb.WriteString(subtitleStyle.Render("space toggle  ·  a all  ·  n none  ·  esc close"))
	sb.WriteString("\n\n")

	for i, col := range m.result.Columns {
		check := "[ ]"
		if m.colVisible[i] {
			check = lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Render("[x]")
		}

		line := fmt.Sprintf("%s %s", check, col)
		if i == m.pickerCursor {
			line = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("229")).
				Background(lipgloss.Color("57")).
				Width(width - 2).
				Render(line)
		}
		sb.WriteString(line + "\n")
	}

	return sb.String()
}
