package nodes

import (
	"sort"
	"strings"
	"testing"
)

func TestString(t *testing.T) {
	tests := []struct {
		name     string
		node     *Node
		expected string
	}{
		{
			name: "simple node",
			node: &Node{
				Key:   "test",
				Value: "value",
			},
			expected: "test: value",
		},
		{
			name: "node with children",
			node: &Node{
				Key: "parent",
				Children: []*Node{
					{Key: "child1", Value: "value1"},
					{Key: "child2", Value: "value2"},
				},
			},
			expected: "parent: {child1: value1 child2: value2}",
		},
		{
			name: "nested children",
			node: &Node{
				Key: "root",
				Children: []*Node{
					{
						Key: "level1",
						Children: []*Node{
							{Key: "level2", Value: "value"},
						},
					},
				},
			},
			expected: "root: {level1: {level2: value}}",
		},
		{
			name: "empty children",
			node: &Node{
				Key:      "empty",
				Children: []*Node{},
			},
			expected: "empty",
		},
		{
			name: "long string truncation",
			node: &Node{
				Key:   "long",
				Value: strings.Repeat("a", MaxStringLength+10),
			},
			expected: "long: " + strings.Repeat("a", MaxStringLength-6) + "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.node.String()
			if result != tt.expected {
				t.Errorf("String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestShortString(t *testing.T) {
	tests := []struct {
		name     string
		node     *Node
		expected string
	}{
		{
			name: "simple leaf node",
			node: &Node{
				Key:   "test",
				Value: "value",
			},
			expected: "value",
		},
		{
			name: "node with leaf children",
			node: &Node{
				Key: "parent",
				Children: []*Node{
					{Key: "child1", Value: "value1"},
					{Key: "child2", Value: "value2"},
				},
			},
			expected: "value1 value2",
		},
		{
			name: "mixed leaf and non-leaf children",
			node: &Node{
				Key: "root",
				Children: []*Node{
					{
						Key: "nonleaf",
						Children: []*Node{
							{Key: "nested", Value: "value"},
						},
					},
					{Key: "leaf", Value: "leafvalue"},
				},
			},
			expected: "value leafvalue",
		},
		{
			name: "no leaf children",
			node: &Node{
				Key: "root",
				Children: []*Node{
					{
						Key: "level1",
						Children: []*Node{
							{Key: "level2", Children: []*Node{}},
						},
					},
				},
			},
			expected: "",
		},
		{
			name: "long string truncation",
			node: &Node{
				Key:   "long",
				Value: strings.Repeat("a", MaxStringLength+10),
			},
			expected: strings.Repeat("a", MaxStringLength) + "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.node.ShortString()
			if result != tt.expected {
				t.Errorf("ShortString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

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
			node := makeNode(tt.key, tt.value, 0, 2)
			sort.Slice(node.Children, sortNodes(node.Children))
			if !node.Equal(&tt.expected) {
				t.Errorf("makeNode() = %v, want %v", node.String(), tt.expected)
			}
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
			result := New(tt.input, 2)
			sort.Slice(result, sortNodes(result))
			if len(result) != len(tt.expected) {
				t.Errorf("New() returned %d nodes, want %d", len(result), len(tt.expected))
				return
			}
			for i, node := range result {
				if !node.Equal(tt.expected[i]) {
					t.Errorf("New()[%d] = %v, want %v", i, node, tt.expected[i])
				}
			}
		})
	}
}

func sortNodes(nodes []*Node) func(i, j int) bool {
	return func(i, j int) bool { return nodes[i].Key < nodes[j].Key }
}
