package nodes

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeNode(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    any
		expected Node
	}{
		{
			name:     "string value",
			key:      "test",
			value:    "hello",
			expected: Node{Key: "test", Value: "hello", Expand: true},
		},
		{
			name:     "int value",
			key:      "number",
			value:    42,
			expected: Node{Key: "number", Value: "42", Expand: true},
		},
		{
			name:     "float value",
			key:      "float",
			value:    3.14,
			expected: Node{Key: "float", Value: "3.14", Expand: true},
		},
		{
			name:     "bool value",
			key:      "flag",
			value:    true,
			expected: Node{Key: "flag", Value: "true", Expand: true},
		},
		{
			name:  "array value",
			key:   "array",
			value: []any{"a", "b", "c"},
			expected: Node{Key: "array", Value: "a b c", Expand: true, Children: []*Node{
				{Key: "0", Value: "a", Expand: true},
				{Key: "1", Value: "b", Expand: true},
				{Key: "2", Value: "c", Expand: true},
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := makeNode(tt.key, tt.value, 0, 2, LeafValuesOnly)
			sort.Slice(node.Children, sortNodes(node.Children))
			assert.True(t, compareNodes(node, &tt.expected))
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected []*Node
	}{
		{
			name: "simple key-value pairs",
			input: map[string]any{
				"string": "value",
				"int":    42,
				"float":  3.14,
				"bool":   true,
			},
			expected: []*Node{
				{Key: "bool", Value: "true", Expand: true},
				{Key: "float", Value: "3.14", Expand: true},
				{Key: "int", Value: "42", Expand: true},
				{Key: "string", Value: "value", Expand: true},
			},
		},
		{
			name: "arrays",
			input: map[string]any{
				"numbers": []any{1, 2, 3},
				"mixed":   []any{"a", 1, true},
			},
			expected: []*Node{
				{
					Key:    "mixed",
					Value:  "a 1 true",
					Expand: true,
					Children: []*Node{
						{Key: "0", Value: "a", Expand: true},
						{Key: "1", Value: "1", Expand: true},
						{Key: "2", Value: "true", Expand: true},
					},
				},
				{
					Key:    "numbers",
					Value:  "1 2 3",
					Expand: true,
					Children: []*Node{
						{Key: "0", Value: "1", Expand: true},
						{Key: "1", Value: "2", Expand: true},
						{Key: "2", Value: "3", Expand: true},
					},
				},
			},
		},
		{
			name:     "empty object",
			input:    map[string]any{},
			expected: []*Node{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := New(tt.input, 2, LeafValuesOnly)
			sort.Slice(result, sortNodes(result))
			if len(result) != len(tt.expected) {
				t.Errorf("New() returned %d nodes, want %d", len(result), len(tt.expected))
				return
			}
			for i, node := range result {
				if !compareNodes(node, tt.expected[i]) {
					t.Errorf("New()[%d] = %v, want %v", i, node, tt.expected[i])
				}
			}
		})
	}
}

func sortNodes(nodes []*Node) func(i, j int) bool {
	return func(i, j int) bool { return nodes[i].Key < nodes[j].Key }
}

// compareNodes compares two nodes for equality, including their children.
// ignores keys
func compareNodes(a, b *Node) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.Key != b.Key || a.Value != b.Value || a.Expand != b.Expand {
		return false
	}
	if len(a.Children) != len(b.Children) {
		return false
	}
	for i := range a.Children {
		if !compareNodes(a.Children[i], b.Children[i]) {
			return false
		}
	}
	return true
}
