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
	ExpandShape                string
	LeafShape                  string
	SpacesPerLayer             int
	MergedObjectExpandOverride string
	MergedObjectShowMetadata          bool
	MergedObjectShowKeyCount          bool
	MergedObjectShowKeyNamesWithTypes bool
	MergedObjectMetadataPrefix        string
	LeafShapes                        lipgloss.Style
	ExpandShapes                      lipgloss.Style
	Selected                          lipgloss.Style
	Unselected                        lipgloss.Style
	Help                              lipgloss.Style
}

func NewStyle(c *StyleConfig) Style {
	style := DefaultStyles()
	if c.LeafShape != "" {
		style.LeafShape = c.LeafShape
	}
	if c.ExpandShape != "" {
		style.ExpandShape = c.ExpandShape
	}
	if c.SpacesPerLayer > 0 {
		style.SpacesPerLayer = c.SpacesPerLayer
	}
	if c.MergedObjectExpandOverride != "" {
		style.MergedObjectExpandOverride = c.MergedObjectExpandOverride
	}
	style.MergedObjectShowMetadata = c.MergedObjectShowMetadata
	style.MergedObjectShowKeyCount = c.MergedObjectShowKeyCount
	style.MergedObjectShowKeyNamesWithTypes = c.MergedObjectShowKeyNamesWithTypes
	if c.MergedObjectMetadataPrefix != "" {
		style.MergedObjectMetadataPrefix = c.MergedObjectMetadataPrefix
	}

	if c.LeafShapeColor != "" {
		fmt.Printf("LeafShapeColor: %s\n", c.LeafShapeColor)
		style.LeafShapes = style.LeafShapes.Foreground(lipgloss.Color(c.LeafShapeColor))
	}
	if c.ExpandShapeColor != "" {
		fmt.Printf("ExpandShapeColor: %s\n", c.ExpandShapeColor)
		style.ExpandShapes = style.ExpandShapes.Foreground(lipgloss.Color(c.ExpandShapeColor))
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
		LeafShape:                  "╰─",
		ExpandShape:                "🯒🯑",
		SpacesPerLayer:             2,
		MergedObjectMetadataPrefix: "ⓘ ",
		LeafShapes:                 lipgloss.NewStyle().Margin(0, 0, 0, 0).Foreground(orange),
		ExpandShapes:               lipgloss.NewStyle().Margin(0, 0, 0, 0).Foreground(dark_orange),
		Selected:                   lipgloss.NewStyle().Margin(0, 0, 0, 0).Background(blue).Foreground(white),
		Unselected:                 lipgloss.NewStyle().Margin(0, 0, 0, 0).Foreground(white).Faint(true),
		Help:                       lipgloss.NewStyle().Margin(0, 0, 0, 0).Foreground(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#ffffff"}),
	}
}
