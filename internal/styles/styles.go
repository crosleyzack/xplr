package styles

import "github.com/charmbracelet/lipgloss"

const (
	white       = lipgloss.Color("#ffffff")
	black       = lipgloss.Color("#000000")
	blue        = lipgloss.Color("#7db8f2")
	orange      = lipgloss.Color("#d99c63")
	dark_orange = lipgloss.Color("#cc8e55")
)

type Style struct {
	LeafShapes   lipgloss.Style
	ExpandShapes lipgloss.Style
	Selected     lipgloss.Style
	Unselected   lipgloss.Style
	Help         lipgloss.Style
}

func NewStyle(c *StyleConfig) Style {
	style := DefaultStyles()
	if c.LeafShapeColor != "" {
		style.LeafShapes = style.LeafShapes.Foreground(lipgloss.Color(c.LeafShapeColor))
	}
	if c.ExpandShapeColor != "" {
		style.ExpandShapes = style.ExpandShapes.Foreground(lipgloss.Color(c.ExpandShapeColor))
	}
	if c.SelectedForegroundColor != "" {
		style.Selected = style.Selected.Foreground(lipgloss.Color(c.SelectedForegroundColor))
	}
	if c.SelectedBackgroundColor != "" {
		style.Selected = style.Selected.Background(lipgloss.Color(c.SelectedBackgroundColor))
	}
	if c.UnselectedForegroundColor != "" {
		style.Unselected = style.Unselected.Background(lipgloss.Color(c.UnselectedForegroundColor))
	}
	if c.HelpColor != "" {
		style.Help = style.Help.Foreground(lipgloss.Color(c.HelpColor))
	}
	return style
}

func DefaultStyles() Style {
	return Style{
		LeafShapes:   lipgloss.NewStyle().Margin(0, 0, 0, 0).Foreground(orange),
		ExpandShapes: lipgloss.NewStyle().Margin(0, 0, 0, 0).Foreground(dark_orange),
		Selected:     lipgloss.NewStyle().Margin(0, 0, 0, 0).Background(blue).Foreground(white),
		Unselected:   lipgloss.NewStyle().Margin(0, 0, 0, 0).Foreground(white).Faint(true),
		Help:         lipgloss.NewStyle().Margin(0, 0, 0, 0).Foreground(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#ffffff"}),
	}
}
