package tree

type TreeConfig struct {
	ExpandedShape           string
	ExpandableShape         string
	LeafShape               string
	SpacesPerLayer          int
	HideSummaryWhenExpanded bool
}

type TreeFormat struct {
	Width                   int
	Height                  int
	ExpandedShape           string
	ExpandableShape         string
	LeafShape               string
	SpacesPerLayer          int
	HideSummaryWhenExpanded bool
}

func NewFormat(c *TreeConfig) *TreeFormat {
	format := DefaultFormat()
	if c.ExpandedShape != "" {
		format.ExpandedShape = c.ExpandedShape
	}
	if c.ExpandableShape != "" {
		format.ExpandableShape = c.ExpandableShape
	}
	if c.LeafShape != "" {
		format.LeafShape = c.LeafShape
	}
	if c.SpacesPerLayer > 0 {
		format.SpacesPerLayer = c.SpacesPerLayer
	}
	if c.HideSummaryWhenExpanded {
		format.HideSummaryWhenExpanded = c.HideSummaryWhenExpanded
	}
	return format
}

func DefaultFormat() *TreeFormat {
	return &TreeFormat{
		Width:                   80,
		Height:                  20,
		LeafShape:               "└─",
		ExpandableShape:         "❭",
		ExpandedShape:           "╰─",
		SpacesPerLayer:          2,
		HideSummaryWhenExpanded: false,
	}
}
