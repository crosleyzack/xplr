package styles

import "github.com/charmbracelet/lipgloss"

const (
	white  = lipgloss.Color("#ffffff")
	black  = lipgloss.Color("#000000")
	blue   = lipgloss.Color("#7DB8F2")
	orange = lipgloss.Color("#D99C63")
)

type Style struct {
	Shapes     lipgloss.Style
	Selected   lipgloss.Style
	Unselected lipgloss.Style
	Help       lipgloss.Style
}

func NewStyle(c *StyleConfig) Style {
	style := DefaultStyles()
	if c.ShapeColor != "" {
		style.Shapes = style.Shapes.Foreground(lipgloss.Color(c.ShapeColor))
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
		Shapes:     lipgloss.NewStyle().Margin(0, 0, 0, 0).Foreground(orange),
		Selected:   lipgloss.NewStyle().Margin(0, 0, 0, 0).Background(blue).Foreground(white),
		Unselected: lipgloss.NewStyle().Margin(0, 0, 0, 0).Foreground(white).Faint(true),
		Help:       lipgloss.NewStyle().Margin(0, 0, 0, 0).Foreground(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#ffffff"}),
	}
}
