package tree

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/crosleyzack/xplr/pkg/nodes"
	"github.com/tiagomelo/go-clipboard/clipboard"
)

// Update the JSON component
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m == nil {
		return nil, nil
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Bottom):
			m.cursor = (m.NumberOfNodes() - 1)
		case key.Matches(msg, m.KeyMap.Top):
			m.cursor = 0
		case key.Matches(msg, m.KeyMap.Down):
			m.NavDown()
		case key.Matches(msg, m.KeyMap.Up):
			m.NavUp()
		case key.Matches(msg, m.KeyMap.CollapseToggle):
			m.InvertCollaped()
		case key.Matches(msg, m.KeyMap.CollapseAll):
			m.ExpandCollapseAll(m.currentNode, false)
		case key.Matches(msg, m.KeyMap.ExpandAll):
			m.ExpandCollapseAll(m.currentNode, true)
		case key.Matches(msg, m.KeyMap.Next):
			m.NextMatchingNode()
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
	if m.currentNode != nil {
		// TODO do we want to keep children expanded so
		// they expand automatically when toggled again
		m.currentNode.Expand = !m.currentNode.Expand
	}
}

// ExpandCollapseAll set the expand flag on every node
func (m *Model) ExpandCollapseAll(n *nodes.Node, expand bool) {
	err := nodes.DFS(
		[]*nodes.Node{n},
		func(n *nodes.Node, _ int) error {
			n.Expand = expand
			return nil
		},
		&nodes.SearchConfig{SearchAll: true},
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to expand-collapse all: %v", err))
	}
}

// GetMatchingNodes find nodes which match request
func (m *Model) GetMatchingNodes(searchTerm string) error {
	m.searchResults = []*nodes.Node{}
	f := func(node *nodes.Node, layer int) error {
		if len(node.Children) == 0 {
			if out, err := regexp.Match(searchTerm, []byte(node.Value)); err == nil && out {
				m.searchResults = append(m.searchResults, node)
			}
		}
		return nil
	}
	err := nodes.DFS(m.Nodes, f, &nodes.SearchConfig{SearchAll: true})
	if err != nil {
		return err
	}
	return nil
}

// NextMatchingNode sets the current node to the next matching node
func (m *Model) NextMatchingNode() {
	if len(m.searchResults) > 0 {
		m.currentNode, m.searchResults = m.searchResults[0], append(m.searchResults[1:], m.searchResults[0])
		// set all parents of this node to be expanded
		for n := m.currentNode; n != nil; n = n.Parent {
			n.Expand = true
		}
		// get cursor position of current node
		count := 0
		nodes.DFS(m.Nodes, func(node *nodes.Node, layer int) error {
			if node.Equal(m.currentNode) {
				m.cursor = count
				return errors.New("break out")
			}
			count++
			return nil
		}, nil)
	}
}

// CopyNodePath find path to node and copies it to clipboard
func (m *Model) CopyNodePath() error {
	// TODO: if value is empty and has children, get string json
	s := nodes.GetPathToNode(m.currentNode) + " = " + m.currentNode.Value
	c := clipboard.New()
	if err := c.CopyText(s); err != nil {
		return fmt.Errorf("failed to copy %s to clipboard: %w", s, err)
	}
	return nil
}
