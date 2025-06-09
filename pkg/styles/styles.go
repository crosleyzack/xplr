package styles

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

const (
	// default colors
	white       = lipgloss.Color("#ffffff")
	black       = lipgloss.Color("#000000")
	blue        = lipgloss.Color("#7db8f2")
	orange      = lipgloss.Color("#d99c63")
	dark_orange = lipgloss.Color("#cc8e55")
)

type Style struct {
	LeafStyle       lipgloss.Style
	ExpandedStyle   lipgloss.Style
	ExpandableStyle lipgloss.Style
	Selected        lipgloss.Style
	Unselected      lipgloss.Style
	Help            lipgloss.Style
}

func NewStyle(c *StyleConfig) Style {
	style := DefaultStyles()
	if c.LeafShapeColor != "" {
		style.LeafStyle = style.LeafStyle.Foreground(lipgloss.Color(c.LeafShapeColor))
	}
	if c.ExpandedShapeColor != "" {
		style.ExpandedStyle = style.ExpandedStyle.Foreground(lipgloss.Color(c.ExpandedShapeColor))
	}
	if c.ExpandableShapeColor != "" {
		style.ExpandableStyle = style.ExpandableStyle.Foreground(lipgloss.Color(c.ExpandableShapeColor))
	}
	if c.SelectedForegroundColor != "" {
		fmt.Printf("SelectedForegroundColor: %s\n", c.SelectedForegroundColor)
		style.Selected = style.Selected.Foreground(lipgloss.Color(c.SelectedForegroundColor))
	}
	if c.SelectedBackgroundColor != "" {
		fmt.Printf("SelectedBackgroundColor: %s\n", c.SelectedBackgroundColor)
		style.Selected = style.Selected.Background(lipgloss.Color(c.SelectedBackgroundColor))
	}
	if c.UnselectedForegroundColor != "" {
		fmt.Printf("UnselectedForegroundColor: %s\n", c.UnselectedForegroundColor)
		style.Unselected = style.Unselected.Foreground(lipgloss.Color(c.UnselectedForegroundColor))
	}
	if c.HelpColor != "" {
		style.Help = style.Help.Foreground(lipgloss.Color(c.HelpColor))
	}
	return style
}

func DefaultStyles() Style {
	return Style{
		LeafStyle:       lipgloss.NewStyle().Margin(0, 0, 0, 0).Foreground(orange),
		ExpandedStyle:   lipgloss.NewStyle().Margin(0, 0, 0, 0).Foreground(orange),
		ExpandableStyle: lipgloss.NewStyle().Margin(0, 0, 0, 0).Foreground(dark_orange),
		Selected:        lipgloss.NewStyle().Margin(0, 0, 0, 0).Background(blue).Foreground(white),
		Unselected:      lipgloss.NewStyle().Margin(0, 0, 0, 0).Foreground(white).Faint(true),
		Help:            lipgloss.NewStyle().Margin(0, 0, 0, 0).Foreground(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#ffffff"}),
	}
}
