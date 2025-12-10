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
	expected := map[string]any{
		"foo": 1.0,
		"bar": "two",
		"baz": map[string]any{
			"bad": []any{1.0, 2.0},
		},
	}
	assert.True(t, reflect.DeepEqual(m2, expected))
}

func TestParseJsonMapArray(t *testing.T) {
	m := []map[string]any{
		{
			"foo": 1,
		},
		{
			"bar": "two",
			"baz": map[string]any{
				"bad": []any{1, 2},
			},
		},
	}
	b, err := json.Marshal(m)
	assert.NoError(t, err)
	m2, err := ParseJson(b)
	assert.NoError(t, err)
	expected := map[string]any{
		"0": map[string]any{
			"foo": 1.0,
		},
		"1": map[string]any{
			"bar": "two",
			"baz": map[string]any{
				"bad": []any{1.0, 2.0},
			},
		},
	}
	assert.True(t, reflect.DeepEqual(m2, expected))
}

func TestParseJsonBasicArray(t *testing.T) {
	m := []any{
		"foo",
		"bar",
		"baz",
		2,
	}
	b, err := json.Marshal(m)
	assert.NoError(t, err)
	m2, err := ParseJson(b)
	assert.NoError(t, err)
	expected := map[string]any{
		"0": "foo",
		"1": "bar",
		"2": "baz",
		"3": 2.0,
	}
	assert.True(t, reflect.DeepEqual(m2, expected))
}
