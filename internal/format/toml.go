package format

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

func ParseToml(data []byte) (map[string]any, error) {
	var t map[string]any
	err := toml.Unmarshal(data, &t)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshall yaml: %w", err)
	}
	return t, nil
}
