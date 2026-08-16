// Package tree implements the bubbletea model for rendering a tree view.
package tree

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/crosleyzack/wndr/pkg/keys"
	"github.com/crosleyzack/wndr/pkg/nodes"
	"github.com/crosleyzack/wndr/pkg/styles"
)

// inspired by https://github.com/savannahostrowski/tree-bubble/blob/main/tree.go

// Model for the JSON tree
type Model struct {
	KeyMap                  keys.KeyMap
	Styles                  styles.Style
	Root                    *nodes.Node
	Height                  int
	Width                   int
	ExpandedShape           string
	ExpandableShape         string
	LeafShape               string
	SpacesPerLayer          int
	cursor                  int
	searchResults           []*nodes.Node
	searchNext              func() (*nodes.Node, bool)
	searchStop              func()
	currentNode             *nodes.Node
	spacesAfterKey          int
	hideSummaryWhenExpanded bool
}

var _ tea.Model = &Model{}

// New creates a new Model for the tree
func New(format *TreeFormat, keys keys.KeyMap, style styles.Style, root *nodes.Node) *Model {
	return &Model{
		KeyMap:                  keys,
		Styles:                  style,
		Root:                    root,
		Height:                  format.Height,
		Width:                   format.Width,
		ExpandedShape:           format.ExpandedShape,
		ExpandableShape:         format.ExpandableShape,
		LeafShape:               format.LeafShape,
		SpacesPerLayer:          format.SpacesPerLayer,
		hideSummaryWhenExpanded: format.HideSummaryWhenExpanded,
		spacesAfterKey:          format.SpacesAfterKey,
		searchResults:           nil,
		searchNext:              nil,
		searchStop:              nil,
		currentNode:             nil,
	}
}

// NumberOfNodes returns the number of nodes in the tree
func (m *Model) NumberOfNodes() int {
	count := 0
	err := nodes.DFS(m.Root, func(node *nodes.Node, _ int) error {
		count++
		return nil
	})
	if err != nil {
		panic(fmt.Sprintf("error counting nodes: %v", err))
	}
	return count
}

// Init Initialize the dashboard
func (m *Model) Init() tea.Cmd {
	return tea.ClearScreen
}
