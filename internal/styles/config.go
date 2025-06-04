package styles

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type StyleConfig struct {
	ExpandShape                       string
	LeafShape                         string
	SpacesPerLayer                    int
	MergedObjectExpandOverride        string
	MergedObjectShowMetadata          bool
	MergedObjectShowKeyCount          bool
	MergedObjectShowKeyNamesWithTypes bool
	MergedObjectMetadataPrefix        string
	// colors
	ExpandShapeColor          string
	LeafShapeColor            string
	SelectedForegroundColor   string
	SelectedBackgroundColor   string
	UnselectedForegroundColor string
	HelpColor                 string
}

func NewConfig(data []byte) (*StyleConfig, error) {
	var c StyleConfig
	err := toml.Unmarshal(data, &c)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshall style config: %w", err)
	}
	return &c, nil
}
