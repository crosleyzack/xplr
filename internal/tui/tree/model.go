package tree

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/crosleyzack/xplr/internal/keys"
	"github.com/crosleyzack/xplr/internal/nodes"
	"github.com/crosleyzack/xplr/internal/styles"
)

// inspired by https://github.com/savannahostrowski/tree-bubble/blob/main/tree.go

type TreeConfig struct {
	Width  int
	Height int
	Style  styles.Style
	Keys   keys.KeyMap
}

// Model for the JSON tree
type Model struct {
	KeyMap keys.KeyMap
	Styles styles.Style
	Nodes  []*nodes.Node

	width  int
	height int
	cursor int

	currentNode *nodes.Node

	Help     help.Model
	showHelp bool

	AdditionalShortHelpKeys func() []key.Binding
}

var _ tea.Model = &Model{}

func New(conf *TreeConfig, nodes []*nodes.Node) *Model {
	return &Model{
		KeyMap: conf.Keys,
		Styles: conf.Style,
		Nodes:  nodes,

		width:  conf.Width,
		height: conf.Height,

		showHelp: true,
		Help:     help.New(),
	}
}

func (m *Model) NumberOfNodes() int {
	count := 0
	var countNodes func([]*nodes.Node)
	countNodes = func(nodes []*nodes.Node) {
		for _, node := range nodes {
			count++
			if node.Children != nil && node.Expand {
				// Recursively count the children, if expanded
				countNodes(node.Children)
			}
		}
	}
	countNodes(m.Nodes)
	return count
}

func (m *Model) ShortHelp() []key.Binding {
	kb := []key.Binding{
		m.KeyMap.Up,
		m.KeyMap.Down,
		m.KeyMap.CollapseToggle,
		m.KeyMap.CollapseAll,
	}
	if m.AdditionalShortHelpKeys != nil {
		kb = append(kb, m.AdditionalShortHelpKeys()...)
	}
	return append(kb,
		m.KeyMap.Quit,
	)
}

func (m *Model) FullHelp() [][]key.Binding {
	return [][]key.Binding{{
		m.KeyMap.Up,
		m.KeyMap.Down,
		m.KeyMap.CollapseToggle,
		m.KeyMap.CollapseAll,
		m.KeyMap.ExpandAll,
		m.KeyMap.Quit,
		m.KeyMap.Help,
	}}
}

// Init Initialize the dashboard
func (m *Model) Init() tea.Cmd {
	return tea.ClearScreen
}
