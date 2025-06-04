package tui

import (
	"os"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/term"
	"github.com/crosleyzack/xplr/internal/keys"
	"github.com/crosleyzack/xplr/internal/modules/tree"
	"github.com/crosleyzack/xplr/internal/nodes"
	"github.com/crosleyzack/xplr/internal/styles"
)

// Model for the JSON tree
type Model struct {
	KeyMap     keys.KeyMap
	Styles     styles.Style
	TreeView   *tree.Model
	HelpView   help.Model
	SearchView textinput.Model

	width  int
	height int
}

var _ tea.Model = &Model{}

// New creates a new Model for the TUI
func New(format *tree.TreeFormat, keymap keys.KeyMap, style styles.Style, nodes []*nodes.Node) (*Model, error) {
	w, h, err := term.GetSize(os.Stdout.Fd())
	if err != nil {
		return nil, err
	}
	format.Width = w
	format.Height = h
	treeView := tree.New(format, keymap, style, nodes)
	helpView := help.New()
	searchView := textinput.New()
	return &Model{
		KeyMap:     keymap,
		Styles:     style,
		TreeView:   treeView,
		HelpView:   helpView,
		SearchView: searchView,
		width:      w,
		height:     h,
	}, nil
}

// ShortHelp returns a short help view for the TUI
func (m *Model) ShortHelp() []key.Binding {
	kb := []key.Binding{
		m.KeyMap.Up,
		m.KeyMap.Down,
		m.KeyMap.CollapseToggle,
		m.KeyMap.CollapseAll,
		m.KeyMap.ExpandAll,
		m.KeyMap.Quit,
	}
	return kb
}

// FullHelp returns a full help view for the TUI
func (m *Model) FullHelp() [][]key.Binding {
	return [][]key.Binding{{
		m.KeyMap.Up,
		m.KeyMap.Down,
		m.KeyMap.Top,
		m.KeyMap.Bottom,
		m.KeyMap.CollapseToggle,
		m.KeyMap.CollapseAll,
		m.KeyMap.ExpandAll,
		m.KeyMap.Search,
		m.KeyMap.Submit,
		m.KeyMap.Next,
		m.KeyMap.Quit,
		m.KeyMap.Help,
	}}
}

// Init Initialize the dashboard
func (m *Model) Init() tea.Cmd {
	return tea.ClearScreen
}
