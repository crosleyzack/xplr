package keys

import "github.com/charmbracelet/bubbles/key"

// KeyMap holds the key bindings for the table.
type KeyMap struct {
	Bottom   key.Binding
	Top      key.Binding
	Down     key.Binding
	Up       key.Binding
	Collapse key.Binding
	Help     key.Binding
	Quit     key.Binding
}

func NewKeyMap(c *KeyConfig) KeyMap {
	keys := DefaultKeyMap()
	if len(c.BottomKeys) != 0 {
		keys.Bottom.SetKeys(c.BottomKeys...)
	}
	if len(c.TopKeys) != 0 {
		keys.Top.SetKeys(c.TopKeys...)
	}
	if len(c.DownKeys) != 0 {
		keys.Down.SetKeys(c.DownKeys...)
	}
	if len(c.UpKeys) != 0 {
		keys.Down.SetKeys(c.UpKeys...)
	}
	if len(c.CollapseKeys) != 0 {
		keys.Collapse.SetKeys(c.CollapseKeys...)
	}
	if len(c.HelpKeys) != 0 {
		keys.Help.SetKeys(c.HelpKeys...)
	}
	if len(c.QuitKeys) != 0 {
		keys.Quit.SetKeys(c.QuitKeys...)
	}
	return keys
}

// DefaultKeyMap is the default key bindings for the table.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Bottom: key.NewBinding(
			key.WithKeys("bottom"),
			key.WithHelp("end", "bottom"),
		),
		Top: key.NewBinding(
			key.WithKeys("top"),
			key.WithHelp("home", "top"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓", "down"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑", "up"),
		),
		Collapse: key.NewBinding(
			key.WithKeys("tab", "enter"),
			key.WithHelp("tab/enter", "collapse/expand"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("esc", "return"),
		),
	}
}
