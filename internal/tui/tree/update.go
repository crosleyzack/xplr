package tree

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Update the JSON component
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m == nil {
		return nil, nil
	}
	switch msg := msg.(type) {
	case tea.QuitMsg:
		return m, tea.Batch(tea.ClearScreen, tea.Quit)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Bottom):
			m.cursor = m.NumberOfNodes()
		case key.Matches(msg, m.KeyMap.Top):
			m.cursor = 0
		case key.Matches(msg, m.KeyMap.Down):
			m.NavDown()
		case key.Matches(msg, m.KeyMap.Up):
			m.NavUp()
		case key.Matches(msg, m.KeyMap.Collapse):
			m.InvertCollaped()
		case key.Matches(msg, m.KeyMap.Help):
			m.Help.ShowAll = !m.Help.ShowAll
		case key.Matches(msg, m.KeyMap.Quit):
			return m, tea.Batch(tea.Quit, tea.ClearScreen)
		}
	}
	return m, nil
}

// NavUp moves the cursor up in component
func (m *Model) NavUp() {
	m.cursor--
	if m.cursor < 0 {
		m.cursor = 0
		return
	}
}

// NavDown moves the cursor down in component
func (m *Model) NavDown() {
	m.cursor++
	if m.cursor >= m.NumberOfNodes() {
		m.cursor = m.NumberOfNodes() - 1
		return
	}
}

// InvertCollaped inverts the collapsed state of the current node
func (m *Model) InvertCollaped() {
	if m.currentNode != nil && m.currentNode.Children != nil {
		m.currentNode.Expand = !m.currentNode.Expand
	}
}
