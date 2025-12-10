package format

import (
	"encoding/json"
	"errors"
	"strconv"
)

// ParseJson convert json byte array to map[string]any
// if not JSON type, returns err
func ParseJson(data []byte) (map[string]any, error) {
	var mp map[string]any
	if err := json.Unmarshal(data, &mp); err == nil {
		return mp, nil
	}
	var arr []any
	if err := json.Unmarshal(data, &arr); err == nil {
		mp := map[string]any{}
		for i, item := range arr {
			mp[strconv.Itoa(i)] = item
		}
		return mp, nil
	}
	return nil, errors.New("data is not json type")
}
