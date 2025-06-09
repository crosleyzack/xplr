package format

import (
	"fmt"

	yaml "github.com/goccy/go-yaml"
)

func ParseYaml(data []byte) (map[string]any, error) {
	var y map[string]any
	err := yaml.Unmarshal(data, &y)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshall yaml: %w", err)
	}
	return y, nil
}
