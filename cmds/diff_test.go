package cmds

import (
	"testing"

	"github.com/crosleyzack/xplr/pkg/nodes"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompareNodes(t *testing.T) {
	leaf := func(key, value string) *nodes.Node {
		return &nodes.Node{ID: uuid.New(), Key: key, Value: value}
	}
	nonLeaf := func(key string, children ...*nodes.Node) *nodes.Node {
		return &nodes.Node{ID: uuid.New(), Key: key, Children: children}
	}
	leafArray := func(key string) *nodes.Node {
		return nonLeaf(key, leaf("0", "a"), leaf("1", "b"))
	}

	tests := []struct {
		name     string
		n1, n2   *nodes.Node
		expected bool
	}{
		{
			name:     "equal leaf nodes",
			n1:       leaf("x", "v"),
			n2:       leaf("x", "v"),
			expected: true,
		},
		{
			name:     "n1 nil n2 non-nil",
			n1:       nil,
			n2:       leaf("x", "v"),
			expected: false,
		},
		{
			name:     "n1 non-nil n2 nil",
			n1:       leaf("x", "v"),
			n2:       nil,
			expected: false,
		},
		{
			name:     "different keys",
			n1:       leaf("a", "v"),
			n2:       leaf("b", "v"),
			expected: false,
		},
		{
			name:     "different values",
			n1:       leaf("x", "1"),
			n2:       leaf("x", "2"),
			expected: false,
		},
		{
			name:     "one leaf one non-leaf",
			n1:       leaf("x", "v"),
			n2:       nonLeaf("x", leaf("c", "v")),
			expected: false,
		},
		{
			name:     "leaf array vs non-array non-leaf",
			n1:       leafArray("x"),
			n2:       nonLeaf("x", leaf("a", "v"), leaf("b", "v")),
			expected: false,
		},
		{
			name:     "equal non-leaf nodes",
			n1:       nonLeaf("x", leaf("c", "v")),
			n2:       nonLeaf("x", leaf("c", "v")),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, nodesEquivalent(tt.n1, tt.n2))
		})
	}
}

func TestAddNode(t *testing.T) {
	newNode := func(key, value string) *nodes.Node {
		return &nodes.Node{ID: uuid.New(), Key: key, Value: value}
	}
	newTree := func(keys ...string) []*nodes.Node {
		tree := make([]*nodes.Node, len(keys))
		for i, k := range keys {
			tree[i] = &nodes.Node{ID: uuid.New(), Key: k}
		}
		return tree
	}

	tests := []struct {
		name        string
		tree        []*nodes.Node
		path        []string
		n           *nodes.Node
		expectedMap map[string]any
		wantErr     bool
	}{
		{
			name:        "nil node returns tree unchanged",
			tree:        newTree("foo"),
			path:        []string{"foo", "f1"},
			n:           nil,
			expectedMap: map[string]any{"foo": ""},
		},
		{
			name:        "empty path appends node at root",
			tree:        newTree(),
			path:        []string{},
			n:           newNode("x", "v"),
			expectedMap: map[string]any{"x": "v"},
		},
		{
			name:        "root not in tree adds new subtree at root",
			tree:        newTree("foo"),
			path:        []string{"bar"},
			n:           newNode("f1", "v"),
			expectedMap: map[string]any{"foo": "", "bar": map[string]any{"f1": "v"}},
		},
		{
			name:        "multi-level missing root attaches node at correct level",
			tree:        newTree(),
			path:        []string{"foo", "f1"},
			n:           newNode("leaf", "v"),
			expectedMap: map[string]any{"foo": map[string]any{"f1": map[string]any{"leaf": "v"}}},
		},
		{
			name:        "node attached under existing root",
			tree:        newTree("foo"),
			path:        []string{"foo", "f1"},
			n:           newNode("bar", "v"),
			expectedMap: map[string]any{"foo": map[string]any{"f1": map[string]any{"bar": "v"}}},
		},
		{
			name:        "intermediate nodes created for deep path",
			tree:        newTree("foo"),
			path:        []string{"foo", "bar", "f1"},
			n:           newNode("baz", "v"),
			expectedMap: map[string]any{"foo": map[string]any{"bar": map[string]any{"f1": map[string]any{"baz": "v"}}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := addNode(tt.tree, tt.path, tt.n)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMap, nodes.ToMap(got))
		})
	}
}

func TestCreateDiffTree(t *testing.T) {
	tests := []struct {
		name     string
		m1       map[string]any
		m2       map[string]any
		expected map[string]any
	}{
		{
			name:     "identical trees produce empty diff",
			m1:       map[string]any{"foo": map[string]any{"bar": 1}},
			m2:       map[string]any{"foo": map[string]any{"bar": 1}},
			expected: map[string]any{},
		},
		{
			// m1: {"foo": {"bar": 1, "baz": 2}}
			// m2: {"foo": {"bar": 1, "baz": 3}}
			// only baz differs; bar is the same so it should not appear in the diff
			name: "leaf value difference in nested structure",
			m1:   map[string]any{"foo": map[string]any{"bar": 1, "baz": 2}},
			m2:   map[string]any{"foo": map[string]any{"bar": 1, "baz": 3}},
			expected: map[string]any{
				"foo": map[string]any{
					"baz": map[string]any{
						"f1": "2",
						"f2": "3",
					},
				},
			},
		},
		{
			// m1: {"foo": {"bar": [1, 2]}}
			// m2: {"foo": {"bar": [1, 3]}}
			name: "leaf value different array",
			m1:   map[string]any{"foo": []any{1, 2}},
			m2:   map[string]any{"foo": []any{1, 3}},
			expected: map[string]any{
				"foo": map[string]any{
					"f1": map[string]any{"0": "1", "1": "2"},
					"f2": map[string]any{"0": "1", "1": "3"},
				},
			},
		},
		{
			// m1: {"foo": {"bar": [1, 3]}}
			name: "one map is empty",
			m1:   map[string]any{"foo": []any{1, 2}},
			m2:   map[string]any{},
			expected: map[string]any{
				"foo": map[string]any{
					"f1": map[string]any{"0": "1", "1": "2"},
					"f2": "",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree1 := nodes.New(tt.m1, 0, nodes.EmptyRepr)
			tree2 := nodes.New(tt.m2, 0, nodes.EmptyRepr)
			diff, err := createDiffTree(tree1, tree2)
			require.NoError(t, err)
			require.Equal(t, tt.expected, nodes.ToMap(diff))
		})
	}
}
