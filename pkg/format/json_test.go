package format

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseJson(t *testing.T) {
	m := map[string]any{
		"foo": 1,
		"bar": "two",
		"baz": map[string]any{
			"bad": []any{1, 2},
		},
	}
	b, err := json.Marshal(m)
	assert.NoError(t, err)
	m2, err := ParseJson(b)
	assert.NoError(t, err)
	reflect.DeepEqual(m, m2)
}
