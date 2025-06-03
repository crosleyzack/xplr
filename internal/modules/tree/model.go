package tree

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/crosleyzack/xplr/internal/keys"
	"github.com/crosleyzack/xplr/internal/nodes"
	"github.com/crosleyzack/xplr/internal/styles"
)

// inspired by https://github.com/savannahostrowski/tree-bubble/blob/main/tree.go

type TreeConfig struct {
	Width          int
	Height         int
	ExpandShape    string
	LeafShape      string
	SpacesPerLayer int
	Style          styles.Style
	Keys           keys.KeyMap
}

// Model for the JSON tree
type Model struct {
	KeyMap         keys.KeyMap
	Styles         styles.Style
	Nodes          []*nodes.Node
	Height         int
	Width          int
	ExpandShape    string
	LeafShape      string
	SpacesPerLayer int
	cursor         int
	searchResults  []*nodes.Node
	currentNode    *nodes.Node
}

var _ tea.Model = &Model{}

// New creates a new Model for the tree
func New(conf *TreeConfig, nodes []*nodes.Node) *Model {
	return &Model{
		KeyMap:         conf.Keys,
		Styles:         conf.Style,
		Nodes:          nodes,
		Height:         conf.Height,
		Width:          conf.Width,
		ExpandShape:    conf.ExpandShape,
		LeafShape:      conf.LeafShape,
		SpacesPerLayer: conf.SpacesPerLayer,
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
