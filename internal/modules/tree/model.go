package tree

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/crosleyzack/xplr/internal/keys"
	"github.com/crosleyzack/xplr/internal/nodes"
	"github.com/crosleyzack/xplr/internal/styles"
)

// inspired by https://github.com/savannahostrowski/tree-bubble/blob/main/tree.go

// Model for the JSON tree
type Model struct {
	KeyMap                  keys.KeyMap
	Styles                  styles.Style
	Nodes                   []*nodes.Node
	Height                  int
	Width                   int
	ExpandedShape           string
	ExpandableShape         string
	LeafShape               string
	SpacesPerLayer          int
	cursor                  int
	searchResults           []*nodes.Node
	currentNode             *nodes.Node
	hideSummaryWhenExpanded bool
}

var _ tea.Model = &Model{}

// New creates a new Model for the tree
func New(format *TreeFormat, keys keys.KeyMap, style styles.Style, nodes []*nodes.Node) *Model {
	return &Model{
		KeyMap:                  keys,
		Styles:                  style,
		Nodes:                   nodes,
		Height:                  format.Height,
		Width:                   format.Width,
		ExpandedShape:           format.ExpandedShape,
		ExpandableShape:         format.ExpandableShape,
		LeafShape:               format.LeafShape,
		SpacesPerLayer:          format.SpacesPerLayer,
		hideSummaryWhenExpanded: format.HideSummaryWhenExpanded,
	}
}

// NumberOfNodes returns the number of nodes in the tree
func (m *Model) NumberOfNodes() int {
	count := 0
	err := nodes.DFS(m.Nodes, func(node *nodes.Node, _ int) error {
		count++
		return nil
	}, nil)
	if err != nil {
		panic(fmt.Sprintf("error counting nodes: %v", err))
	}
	return count
}

// Init Initialize the dashboard
func (m *Model) Init() tea.Cmd {
	return tea.ClearScreen
}
