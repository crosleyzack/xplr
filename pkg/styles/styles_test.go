package styles

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

func TestDefaultStyles(t *testing.T) {
	s := DefaultStyles()
	assert.NotNil(t, s.KeyBasedStyles)
	assert.Empty(t, s.KeyBasedStyles)
	assert.Equal(t, lipgloss.Color("#d99c63"), s.LeafStyle.GetForeground())
	assert.Equal(t, lipgloss.Color("#d99c63"), s.ExpandedStyle.GetForeground())
	assert.Equal(t, lipgloss.Color("#cc8e55"), s.ExpandableStyle.GetForeground())
	assert.Equal(t, lipgloss.Color("#7db8f2"), s.Selected.GetBackground())
}

func TestNewStyleDefaults(t *testing.T) {
	s := NewStyle(&StyleConfig{})
	expected := DefaultStyles()
	assert.Equal(t, expected, s)
}

func TestNewStyleOverrides(t *testing.T) {
	s := NewStyle(&StyleConfig{
		LeafShapeColor:            "#ff0000",
		ExpandedShapeColor:        "#00ff00",
		ExpandableShapeColor:      "#0000ff",
		SelectedForegroundColor:   "#ffffff",
		SelectedBackgroundColor:   "#000000",
		UnselectedForegroundColor: "#aaaaaa",
		HelpColor:                 "#bbbbbb",
	})
	assert.Equal(t, lipgloss.Color("#ff0000"), s.LeafStyle.GetForeground())
	assert.Equal(t, lipgloss.Color("#00ff00"), s.ExpandedStyle.GetForeground())
	assert.Equal(t, lipgloss.Color("#0000ff"), s.ExpandableStyle.GetForeground())
	assert.Equal(t, lipgloss.Color("#ffffff"), s.Selected.GetForeground())
	assert.Equal(t, lipgloss.Color("#000000"), s.Selected.GetBackground())
	assert.Equal(t, lipgloss.Color("#aaaaaa"), s.Unselected.GetForeground())
	assert.Equal(t, lipgloss.Color("#bbbbbb"), s.Help.GetForeground())
}

func TestAddConditionalStyle(t *testing.T) {
	s := DefaultStyles()
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("#123456"))
	s.AddConditionalStyle("custom", style)
	assert.Equal(t, style, s.KeyBasedStyles["custom"])
	assert.Equal(t, 1, len(s.KeyBasedStyles))
}

func TestNewConfig(t *testing.T) {
	data := []byte(`
LeafShapeColor = "#aa0000"
SelectedForegroundColor = "#ffffff"
DiffColors = ["#111111", "#ffffff"]
`)
	c, err := NewConfig(data)
	assert.NoError(t, err)
	assert.Equal(t, "#aa0000", c.LeafShapeColor)
	assert.Equal(t, "#ffffff", c.SelectedForegroundColor)
	assert.Equal(t, []string{"#111111", "#ffffff"}, c.DiffColors)
}

func TestNewConfigInvalid(t *testing.T) {
	_, err := NewConfig([]byte("not valid toml =="))
	assert.Error(t, err)
}
