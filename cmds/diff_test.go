package cmds

import (
	"strings"
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
					"1": map[string]any{
						"f1": "2",
						"f2": "3",
					},
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
					"f2": "nil",
				},
			},
		},
		{
			// m1: {"foo": [{"bar": [1, 2]}}]
			// m2: {"foo": [{"bar": [1, 3]}}]
			name: "nested arrays",
			m1:   map[string]any{"foo": []any{map[string]any{"bar": map[string]any{"baz": []any{1, 2}}}}},
			m2:   map[string]any{"foo": []any{map[string]any{"bar": map[string]any{"baz": []any{1, 3}}}}},
			expected: map[string]any{
				"foo": map[string]any{
					"0": map[string]any{
						"bar": map[string]any{
							"baz": map[string]any{
								"1": map[string]any{
									"f1": "2",
									"f2": "3",
								},
							},
						},
					},
				},
			},
		},
		{
			// m1: {"foo": {"0": [1, 2], "1": [2, 3]}}
			// m2: {"foo": {"0": [1], "2": [3]}}
			name: "complex",
			m1:   map[string]any{"foo": map[string]any{"0": []any{1, 2}, "1": []any{2, 3}}},
			m2:   map[string]any{"foo": map[string]any{"0": []any{1}, "2": []any{3}}},
			expected: map[string]any{
				"foo": map[string]any{
					"0": map[string]any{
						"1": map[string]any{
							"f1": "2",
							"f2": "nil",
						},
					},
					"1": map[string]any{
						"f1": map[string]any{"0": "2", "1": "3"},
						"f2": "nil",
					},
					// TODO this is impossible with current impl
					// "2": map[string]any{
					// 	"f1": "nil",
					// 	"f2": map[string]any{"0": "3"},
					// },
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

func TestUpdateRepr(t *testing.T) {
	// collectValues walks the entire tree (ignoring Expand) and returns a map of
	// dot-separated path -> node Value for every node.
	collectValues := func(tree []*nodes.Node) map[string]string {
		result := map[string]string{}
		_ = nodes.DFS(tree, func(n *nodes.Node, _ int) error {
			path := strings.Join(nodes.GetPathToNode(n), ".")
			result[path] = n.Value
			return nil
		}, nodes.WithNextNodes(nodes.AllChildren))
		return result
	}

	tests := []struct {
		name          string
		input         map[string]any
		displayLayers uint
		repr          nodes.ReprNode
		expected      map[string]string
	}{
		{
			name:          "empty tree returns no error",
			input:         map[string]any{},
			displayLayers: 0,
			repr:          nodes.EmptyRepr,
			expected:      map[string]string{},
		},
		{
			// leaf nodes have no children so IsLeaf=true; updateRepr skips them
			name:          "leaf nodes are not modified",
			input:         map[string]any{"x": "1", "y": "2"},
			displayLayers: 0,
			repr:          nodes.LeafKeyAndValues,
			expected:      map[string]string{"x": "1", "y": "2"},
		},
		{
			name:          "root non-leaf value is updated",
			input:         map[string]any{"foo": map[string]any{"bar": "v"}},
			displayLayers: 0,
			repr:          nodes.LeafKeyAndValues,
			expected:      map[string]string{"foo": "bar:v", "foo.bar": "v"},
		},
		{
			// updateRepr uses AllChildren so it visits every node regardless of Expand.
			// All non-leaf Values are updated even with displayLayers=0.
			name:          "all non-leaf nodes updated regardless of expand state",
			input:         map[string]any{"root": map[string]any{"a": map[string]any{"x": "1"}}},
			displayLayers: 0,
			repr:          nodes.LeafKeyAndValues,
			expected: map[string]string{
				"root":     "x:1",
				"root.a":   "x:1",
				"root.a.x": "1",
			},
		},
		{
			// building with LeafKeyAndValues sets non-leaf Values; updateRepr with EmptyRepr
			// clears them back to "" while leaving leaf Values intact.
			name:          "EmptyRepr clears all non-leaf values",
			input:         map[string]any{"root": map[string]any{"a": map[string]any{"x": "1"}, "b": map[string]any{"y": "2"}}},
			displayLayers: 0,
			repr:          nodes.EmptyRepr,
			expected: map[string]string{
				"root":     "",
				"root.a":   "",
				"root.a.x": "1",
				"root.b":   "",
				"root.b.y": "2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := nodes.New(tt.input, tt.displayLayers, nodes.EmptyRepr)
			err := updateRepr(tree, tt.repr)
			require.NoError(t, err)
			require.Equal(t, tt.expected, collectValues(tree))
		})
	}
}
