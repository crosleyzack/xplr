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
		name              string
		path              []string
		expectedNode      *Node
		expectedRemaining []string
	}{
		{
			name:              "empty path returns root with nil remaining",
			path:              []string{},
			expectedNode:      top,
			expectedRemaining: nil,
		},
		{
			name:              "first child fully matched",
			path:              []string{"bar"},
			expectedNode:      sibling1,
			expectedRemaining: nil,
		},
		{
			name:              "second child fully matched",
			path:              []string{"bad"},
			expectedNode:      sibling2,
			expectedRemaining: nil,
		},
		{
			name:              "multi-level path fully matched",
			path:              []string{"bar", "final", "target"},
			expectedNode:      leaf,
			expectedRemaining: nil,
		},
		{
			name:              "partial path fully matched",
			path:              []string{"bar", "final"},
			expectedNode:      parent1,
			expectedRemaining: nil,
		},
		{
			name:              "no match returns root with full path as remaining",
			path:              []string{"nonexistent"},
			expectedNode:      top,
			expectedRemaining: []string{"nonexistent"},
		},
		{
			name:              "partial match returns deepest matched node with remaining path",
			path:              []string{"bar", "nonexistent"},
			expectedNode:      sibling1,
			expectedRemaining: []string{"nonexistent"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNode, gotRemaining := GetNodeFromPath(top, tt.path)
			assert.Equal(t, tt.expectedNode, gotNode)
			assert.Equal(t, tt.expectedRemaining, gotRemaining)
		})
	}
}

func TestGetNodeFromTree(t *testing.T) {
	top, sibling1, sibling2, parent1, leaf := buildTestTree()

	tests := []struct {
		name              string
		tree              []*Node
		path              []string
		expectedNode      *Node
		expectedRemaining []string
	}{
		{
			name:              "empty tree returns nil with path as remaining",
			tree:              []*Node{},
			path:              []string{"foo"},
			expectedNode:      nil,
			expectedRemaining: []string{"foo"},
		},
		{
			name:              "empty path returns nil",
			tree:              []*Node{top},
			path:              []string{},
			expectedNode:      nil,
			expectedRemaining: nil,
		},
		{
			name:              "path matching root key returns root",
			tree:              []*Node{top},
			path:              []string{"foo"},
			expectedNode:      top,
			expectedRemaining: nil,
		},
		{
			name:              "path matching child of first root",
			tree:              []*Node{top},
			path:              []string{"foo", "bar"},
			expectedNode:      sibling1,
			expectedRemaining: nil,
		},
		{
			name:              "multi-level path through first root",
			tree:              []*Node{top},
			path:              []string{"foo", "bar", "final", "target"},
			expectedNode:      leaf,
			expectedRemaining: nil,
		},
		{
			name:              "match in second root when first key differs",
			tree:              []*Node{sibling2, sibling1},
			path:              []string{"bar", "final"},
			expectedNode:      parent1,
			expectedRemaining: nil,
		},
		{
			name:              "no root key match returns nil with full path as remaining",
			tree:              []*Node{top},
			path:              []string{"nonexistent"},
			expectedNode:      nil,
			expectedRemaining: []string{"nonexistent"},
		},
		{
			name:              "root matched but child missing returns deepest node with remaining",
			tree:              []*Node{top},
			path:              []string{"foo", "nonexistent"},
			expectedNode:      top,
			expectedRemaining: []string{"nonexistent"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNode, gotRemaining := GetNodeFromTree(tt.tree, tt.path)
			assert.Equal(t, tt.expectedNode, gotNode)
			assert.Equal(t, tt.expectedRemaining, gotRemaining)
		})
	}
}
