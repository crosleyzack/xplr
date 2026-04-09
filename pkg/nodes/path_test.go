package nodes

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// buildTestTree builds:
//
//	top (foo)
//	├── sibling1 (bar)
//	│   └── parent1 (final)
//	│       └── leaf (target)
//	└── sibling2 (bad)
func buildTestTree() (top, sibling1, sibling2, parent1, leaf *Node) {
	leaf = &Node{ID: uuid.New(), Key: "target", Value: "hello"}
	parent1 = &Node{ID: uuid.New(), Key: "final", Children: []*Node{leaf}}
	leaf.Parent = parent1
	sibling1 = &Node{ID: uuid.New(), Key: "bar", Children: []*Node{parent1}}
	parent1.Parent = sibling1
	sibling2 = &Node{ID: uuid.New(), Key: "bad"}
	top = &Node{ID: uuid.New(), Key: "foo", Children: []*Node{sibling1, sibling2}}
	sibling1.Parent = top
	sibling2.Parent = top
	return
}

func TestGetPathToNode(t *testing.T) {
	top, sibling1, _, parent1, leaf := buildTestTree()

	tests := []struct {
		name     string
		node     *Node
		expected []string
	}{
		{
			name:     "multi-level path",
			node:     leaf,
			expected: []string{"foo", "bar", "final", "target"},
		},
		{
			name:     "root node only",
			node:     top,
			expected: []string{"foo"},
		},
		{
			name:     "mid-level node",
			node:     sibling1,
			expected: []string{"foo", "bar"},
		},
		{
			name:     "two levels deep",
			node:     parent1,
			expected: []string{"foo", "bar", "final"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetPathToNode(tt.node)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestGetNodeFromPath(t *testing.T) {
	top, sibling1, sibling2, parent1, leaf := buildTestTree()

	tests := []struct {
		name     string
		path     []string
		expected *Node
	}{
		{
			name:     "empty path returns root",
			path:     []string{},
			expected: top,
		},
		{
			name:     "first child",
			path:     []string{"bar"},
			expected: sibling1,
		},
		{
			name:     "second child",
			path:     []string{"bad"},
			expected: sibling2,
		},
		{
			name:     "multi-level path",
			path:     []string{"bar", "final", "target"},
			expected: leaf,
		},
		{
			name:     "partial path stops at matching node",
			path:     []string{"bar", "final"},
			expected: parent1,
		},
		{
			name:     "non-matching key returns nil",
			path:     []string{"nonexistent"},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetNodeFromPath(top, tt.path)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestGetNodeFromTree(t *testing.T) {
	top, sibling1, sibling2, parent1, leaf := buildTestTree()

	tests := []struct {
		name     string
		tree     []*Node
		path     []string
		expected *Node
	}{
		{
			name:     "empty tree returns nil",
			tree:     []*Node{},
			path:     []string{"bar"},
			expected: nil,
		},
		{
			name:     "empty path returns first root",
			tree:     []*Node{top},
			path:     []string{},
			expected: top,
		},
		{
			name:     "match in first root",
			tree:     []*Node{top},
			path:     []string{"bar"},
			expected: sibling1,
		},
		{
			name:     "multi-level path through first root",
			tree:     []*Node{top},
			path:     []string{"bar", "final", "target"},
			expected: leaf,
		},
		{
			name:     "match in second root when first has no match",
			tree:     []*Node{sibling2, sibling1},
			path:     []string{"final"},
			expected: parent1,
		},
		{
			name:     "no match returns nil",
			tree:     []*Node{top},
			path:     []string{"nonexistent"},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetNodeFromTree(tt.tree, tt.path)
			assert.Equal(t, tt.expected, got)
		})
	}
}
