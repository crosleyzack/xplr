package tree

import (
	"errors"
	"fmt"
	"iter"
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
	// stop previous search if it exists
	if m.searchStop != nil {
		m.searchStop()
	}
	m.searchResults = make([]*nodes.Node, 0)
	m.searchNext, m.searchStop = iter.Pull(nodes.DFSIter(m.Nodes, func(node *nodes.Node) bool {
		// match on leaf ndes which match search term
		if len(node.Children) == 0 {
			if out, err := regexp.Match(searchTerm, []byte(node.Value)); err == nil && out {
				return true
			}
			if out, err := regexp.Match(searchTerm, []byte(node.Key)); err == nil && out {
				return true
			}
		}
		return false
	}))
	return nil
}

// nextNodeFromResults get next item from stored results
func (m *Model) nextNodeFromResults() bool {
	// rotate through built up results, if any
	if len(m.searchResults) == 0 {
		return false
	}
	m.currentNode = m.searchResults[0]
	m.searchResults = append(m.searchResults[1:], m.searchResults[0])
	return true
}

// NextMatchingNode sets the current node to the next matching node
func (m *Model) NextMatchingNode() {
	if m.searchNext != nil {
		node, ok := m.searchNext()
		if ok && node != nil {
			// set n as current node and add to results
			m.currentNode, m.searchResults = node, append(m.searchResults, node)
		} else {
			// stop search and rotate through built up results
			m.searchStop()
			m.searchStop = nil
			m.searchNext = nil
			if ok := m.nextNodeFromResults(); !ok {
				// we couldn't get another node, just do nothing
				return
			}
		}
	} else if ok := m.nextNodeFromResults(); !ok {
		// we couldn't get another node, just do nothing
		return
	}
	// expand parents
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
