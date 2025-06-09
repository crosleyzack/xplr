package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// View renders the TUI view for the tree model
func (m *Model) View() string {
	if m == nil {
		return "no data"
	}

	availableHeight := m.height
	if availableHeight <= 0 {
		availableHeight = 80 // Default height if not set
	}

	sections := []string{}

	// add help
	sections = append(sections, m.Styles.Help.Render(m.HelpView.View(m)))
	if m.HelpView.ShowAll {
		availableHeight -= m.KeyMap.Len()
	} else {
		availableHeight -= 1
	}

	var search string
	if m.SearchView.Focused() {
		search = m.Styles.Help.Render(m.SearchView.View())
		sections = append(sections, search)
		availableHeight -= 1
	}

	m.TreeView.Height = availableHeight - 1 // add a line of padding
	tree := m.TreeView.View()

	sections = append([]string{tree}, sections...)
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}
