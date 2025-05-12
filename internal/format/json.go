package format

import (
	"encoding/json"
	"fmt"
)

func ParseJson(data []byte) (map[string]any, error) {
	var j map[string]any
	err := json.Unmarshal(data, &j)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshall json: %w", err)
	}
	return j, nil
}
