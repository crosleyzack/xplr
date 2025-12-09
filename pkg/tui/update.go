package tui

import (
	"github.com/sirupsen/logrus"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/crosleyzack/xplr/pkg/modules/tree"
)

// Update the tree view component
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m == nil {
		return nil, nil
	}
	log := logrus.New()
	switch msg := msg.(type) {
	case tea.QuitMsg:
		return m, tea.Batch(tea.ClearScreen, tea.Quit)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Help):
			m.HelpView.ShowAll = !m.HelpView.ShowAll
		case key.Matches(msg, m.KeyMap.Quit):
			return m, tea.Batch(tea.ClearScreen, tea.Quit)
		case key.Matches(msg, m.KeyMap.Search):
			m.SearchView.Reset()
			m.SearchView.Focus()
		case key.Matches(msg, m.KeyMap.Submit):
			if !m.SearchView.Focused() {
				// copy the path do this node
				err := m.TreeView.CopyNodePath()
				if err != nil {
					log.Errorf("Failed to copy node path: %v", err)
				}
				return m, nil
			}
			err := m.TreeView.GetMatchingNodes(m.SearchView.Value())
			if err != nil {
				log.Errorf("Failed to get matching nodes: %v", err)
			}
			m.TreeView.NextMatchingNode()
			m.SearchView.Blur()
			m.SearchView.Reset()
		default:
			if m.SearchView.Focused() {
				// If the search view is focused, update it with any key
				m.SearchView, _ = m.SearchView.Update(msg)
			} else {
				model, _ := m.TreeView.Update(msg)
				var ok bool
				if m.TreeView, ok = model.(*tree.Model); !ok {
					log.Errorf("Failed to update tree model")
				}
			}
		}
	}
	return m, nil
}
