package format

import (
	"fmt"
)

type Format func(data []byte) (map[string]any, error)

type FormatType int

const (
	FormatJson FormatType = iota
	FormatYaml
	FormatToml
)

// Parse tries to parse the data using the provided formats and
// returns the first successful result as a map[string]any
func Parse(data []byte) (m map[string]any, err error) {
	for _, fmt := range []Format{ParseJson, ParseYaml, ParseToml} {
		m, err = fmt(data)
		if err == nil {
			break
		}
	}
	if len(m) == 0 {
		return nil, fmt.Errorf("no data")
	}
	return m, nil
}

func As(m map[string]any, f FormatType) ([]byte, error) {
	switch f {
	case FormatJson:
		return AsJson(m)
	case FormatYaml:
		return AsYaml(m)
	case FormatToml:
		return AsToml(m)
	default:
		return nil, fmt.Errorf("unsupported format type: %v", f)
	}
}
