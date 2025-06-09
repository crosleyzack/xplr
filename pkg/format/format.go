package format

type Format func(data []byte) (map[string]any, error)
