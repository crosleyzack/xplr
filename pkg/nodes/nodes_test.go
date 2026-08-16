package nodes

import (
	"testing"

	"github.com/crosleyzack/xplr/pkg/omap"
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
			expected: Node{Key: "array", Value: "a b c", Expand: true, Children: childMap(
				&Node{Key: "0", Value: "a", Expand: true},
				&Node{Key: "1", Value: "b", Expand: true},
				&Node{Key: "2", Value: "c", Expand: true},
			)},
		},
		{
			name:  "map value sets Value via repr",
			key:   "map",
			value: map[string]any{"b": "2", "a": "1"},
			expected: Node{Key: "map", Value: "1 2", Expand: true, Children: childMap(
				&Node{Key: "a", Value: "1", Expand: true},
				&Node{Key: "b", Value: "2", Expand: true},
			)},
		},
		{
			name:     "empty map value",
			key:      "map",
			value:    map[string]any{},
			expected: Node{Key: "map", Value: "{}", Expand: true, Children: childMap()},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewNode(tt.key, tt.value, 0, 2, LeafValuesOnly)
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
					Children: childMap(
						&Node{Key: "0", Value: "a", Expand: true},
						&Node{Key: "1", Value: "1", Expand: true},
						&Node{Key: "2", Value: "true", Expand: true},
					),
				},
				{
					Key:    "numbers",
					Value:  "1 2 3",
					Expand: true,
					Children: childMap(
						&Node{Key: "0", Value: "1", Expand: true},
						&Node{Key: "1", Value: "2", Expand: true},
						&Node{Key: "2", Value: "3", Expand: true},
					),
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
			// New returns a single sentinel root; the top-level nodes are its
			// children, held in sorted key order.
			result := New(tt.input, 2, LeafValuesOnly).Children.Arr()
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

func TestToMap(t *testing.T) {
	tests := []struct {
		name     string
		nodes    []*Node
		expected map[string]any
	}{
		{
			name:     "empty slice",
			nodes:    []*Node{},
			expected: map[string]any{},
		},
		{
			name:     "single leaf node",
			nodes:    []*Node{{Key: "name", Value: "alice"}},
			expected: map[string]any{"name": "alice"},
		},
		{
			name:     "multiple leaf nodes",
			nodes:    []*Node{{Key: "a", Value: "1"}, {Key: "b", Value: "2"}},
			expected: map[string]any{"a": "1", "b": "2"},
		},
		{
			name: "nested node",
			nodes: []*Node{
				{
					Key: "person",
					Children: childMap(
						&Node{Key: "name", Value: "alice"},
						&Node{Key: "age", Value: "30"},
					),
				},
			},
			expected: map[string]any{
				"person": map[string]any{"name": "alice", "age": "30"},
			},
		},
		{
			name:     "leaf node with empty value",
			nodes:    []*Node{{Key: "empty"}},
			expected: map[string]any{"empty": ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToMap(tt.nodes...)
			assert.Equal(t, tt.expected, got)
		})
	}
}

// childMap builds a Children map keyed by each node's Key, matching how the
// package stores children internally.
func childMap(children ...*Node) omap.OMap[string, *Node] {
	m := omap.New[string, *Node]()
	for _, c := range children {
		m.Put(c.Key, c)
	}
	return m
}

// compareNodes compares two nodes for equality, including their children.
// ignores ID and Parent so nodes built by NewNode (which assigns random IDs)
// can be compared against literal expectations.
func compareNodes(a, b *Node) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.Key != b.Key || a.Value != b.Value || a.Expand != b.Expand {
		return false
	}
	if a.Children.Len() != b.Children.Len() {
		return false
	}
	for _, child := range a.Children.Arr() {
		other, ok := b.Children.Get(child.Key)
		if !ok || !compareNodes(child, other) {
			return false
		}
	}
	return true
}

func TestIsArray(t *testing.T) {
	tests := []struct {
		name     string
		node     *Node
		expected bool
	}{
		{
			name: "numeric keys array",
			node: &Node{
				Children: childMap(
					&Node{Key: "0", Value: "first"},
					&Node{Key: "1", Value: "second"},
					&Node{Key: "2", Value: "third"},
				),
			},
			expected: true,
		},
		{
			name: "mixed keys not array",
			node: &Node{
				Children: childMap(
					&Node{Key: "name", Value: "John"},
					&Node{Key: "age", Value: "30"},
				),
			},
			expected: false,
		},
		{
			name: "non-sequential numeric keys still array",
			node: &Node{
				Children: childMap(
					&Node{Key: "0", Value: "first"},
					&Node{Key: "2", Value: "third"},
				),
			},
			expected: true,
		},
		{
			name: "empty children not array",
			node: &Node{
				Children: childMap(),
			},
			expected: false,
		},
		{
			name: "no children not array",
			node: &Node{
				Value: "simple",
			},
			expected: false,
		},
		{
			name: "mixed numeric and string keys not array",
			node: &Node{
				Children: childMap(
					&Node{Key: "0", Value: "first"},
					&Node{Key: "name", Value: "test"},
				),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsArray(tt.node)
			if result != tt.expected {
				t.Errorf("isArray() = %v, want %v", result, tt.expected)
			}
		})
	}
}
