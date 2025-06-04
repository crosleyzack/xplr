package nodes

import (
	"strings"
	"testing"
)

func TestLeafValuesWithBrackets(t *testing.T) {
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
			result := LeafValuesWithBrackets(tt.node)
			if result != tt.expected {
				t.Errorf("String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestLeafValuesOnly(t *testing.T) {
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
			result := LeafValuesOnly(tt.node)
			if result != tt.expected {
				t.Errorf("ShortString() = %v, want %v", result, tt.expected)
			}
		})
	}
}
