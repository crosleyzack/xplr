package styles

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

// StyleConfig defines the style information for the xplr tui
type StyleConfig struct {
	ExpandedShapeColor        string
	ExpandableShapeColor      string
	LeafShapeColor            string
	SelectedForegroundColor   string
	SelectedBackgroundColor   string
	UnselectedForegroundColor string
	HelpColor                 string
	DiffColors                []string
}

// NewConfig creates a style config by unmarshalling data
func NewConfig(data []byte) (*StyleConfig, error) {
	var c StyleConfig
	err := toml.Unmarshal(data, &c)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshall style config: %w", err)
	}
	return &c, nil
}
