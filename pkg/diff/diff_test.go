package diff

import (
	"strings"
	"testing"

	"github.com/crosleyzack/xplr/pkg/nodes"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNodesEquivalent(t *testing.T) {
	leaf := func(key, value string) *nodes.Node {
		return &nodes.Node{ID: uuid.New(), Key: key, Value: value}
	}
	nonLeaf := func(key string, children ...*nodes.Node) *nodes.Node {
		n := &nodes.Node{ID: uuid.New(), Key: key}
		for _, c := range children {
			n.Children.Put(c.Key, c)
		}
		return n
	}
	leafArray := func(key string) *nodes.Node {
		return nonLeaf(key, leaf("0", "a"), leaf("1", "b"))
	}

	tests := []struct {
		name     string
		nodes    []*nodes.Node
		expected bool
	}{
		// pairwise cases
		{
			name:     "empty slice is equivalent",
			nodes:    nil,
			expected: true,
		},
		{
			name:     "single node is equivalent",
			nodes:    []*nodes.Node{leaf("x", "1")},
			expected: true,
		},
		{
			name:     "equal leaf nodes",
			nodes:    []*nodes.Node{leaf("x", "v"), leaf("x", "v")},
			expected: true,
		},
		{
			name:     "both nil are equivalent",
			nodes:    []*nodes.Node{nil, nil},
			expected: true,
		},
		{
			name:     "n1 nil n2 non-nil",
			nodes:    []*nodes.Node{nil, leaf("x", "v")},
			expected: false,
		},
		{
			name:     "n1 non-nil n2 nil",
			nodes:    []*nodes.Node{leaf("x", "v"), nil},
			expected: false,
		},
		{
			name:     "different keys",
			nodes:    []*nodes.Node{leaf("a", "v"), leaf("b", "v")},
			expected: false,
		},
		{
			name:     "different values",
			nodes:    []*nodes.Node{leaf("x", "1"), leaf("x", "2")},
			expected: false,
		},
		{
			name:     "one leaf one non-leaf",
			nodes:    []*nodes.Node{leaf("x", "v"), nonLeaf("x", leaf("c", "v"))},
			expected: false,
		},
		{
			name:     "leaf array vs non-array non-leaf",
			nodes:    []*nodes.Node{leafArray("x"), nonLeaf("x", leaf("a", "v"), leaf("b", "v"))},
			expected: false,
		},
		{
			name:     "equal non-leaf nodes",
			nodes:    []*nodes.Node{nonLeaf("x", leaf("c", "v")), nonLeaf("x", leaf("c", "v"))},
			expected: true,
		},
		{
			name:     "all equal leaves",
			nodes:    []*nodes.Node{leaf("x", "1"), leaf("x", "1"), leaf("x", "1")},
			expected: true,
		},
		{
			name:     "one differing value among three",
			nodes:    []*nodes.Node{leaf("x", "1"), leaf("x", "1"), leaf("x", "2")},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, nodesEquivalent(tt.nodes...))
		})
	}
}

func TestAddNode(t *testing.T) {
	newNode := func(key, value string) *nodes.Node {
		return &nodes.Node{ID: uuid.New(), Key: key, Value: value}
	}
	// newTree builds a sentinel root holding the given top-level keys as leaf
	// children, matching how New represents a tree.
	newTree := func(keys ...string) *nodes.Node {
		root := nodes.New(map[string]any{}, 0, nodes.EmptyRepr)
		for _, k := range keys {
			child := &nodes.Node{ID: uuid.New(), Key: k}
			root.Children.Put(k, child)
			child.Parent = root
		}
		return root
	}

	tests := []struct {
		name        string
		tree        *nodes.Node
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
			// got is the sentinel root; its children are the top-level nodes.
			assert.Equal(t, tt.expectedMap, nodes.ToMap(got.Children.Arr()...))
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
			// DFSMulti walks the union of paths, so "2" (present only in m2)
			// now appears in the diff.
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
					"2": map[string]any{
						"f1": "nil",
						"f2": map[string]any{"0": "3"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree1 := nodes.New(tt.m1, 0, nodes.EmptyRepr)
			tree2 := nodes.New(tt.m2, 0, nodes.EmptyRepr)
			conf := defaultDiffConf()
			conf.Keys = []string{"f1", "f2"}
			diff, err := createDiffTree([]*nodes.Node{tree1, tree2}, conf)
			require.NoError(t, err)
			require.Equal(t, tt.expected, nodes.ToMap(diff.Children.Arr()...))
		})
	}
}

func TestAddMeta(t *testing.T) {
	// metaColors returns the color value stored for each key under the meta node.
	metaColors := func(tree *nodes.Node) map[string]string {
		m := map[string]string{}
		meta := nodes.Child(tree, nodes.MetaKey)
		if meta == nil {
			return m
		}
		for k, child := range meta.Children.Iter() {
			m[k] = child.Value
		}
		return m
	}

	tests := []struct {
		name       string
		diffColors []string
		keys       []string
		want       map[string]string
	}{
		{
			name:       "two keys use configured colors",
			diffColors: []string{"#111", "#222"},
			keys:       []string{"_f1", "_f2"},
			want:       map[string]string{"_f1": "#111", "_f2": "#222"},
		},
		{
			name:       "three keys with enough configured colors",
			diffColors: []string{"#111", "#222", "#333"},
			keys:       []string{"_f1", "_f2", "_f3"},
			want:       map[string]string{"_f1": "#111", "_f2": "#222", "_f3": "#333"},
		},
		{
			name:       "too few colors fall back to defaults and cycle",
			diffColors: []string{"#111"},
			keys:       []string{"_f1", "_f2", "_f3"},
			want: map[string]string{
				"_f1": defaultDiffColors[0],
				"_f2": defaultDiffColors[1],
				"_f3": defaultDiffColors[2],
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := &diffConf{}
			conf.Colors = tt.diffColors
			tree := nodes.New(map[string]any{}, 0, nodes.EmptyRepr)
			tree, err := addMeta(tree, conf, tt.keys...)
			require.NoError(t, err)
			assert.Equal(t, tt.want, metaColors(tree))
		})
	}
}

func TestCreateDiffTreeThreeTrees(t *testing.T) {
	tests := []struct {
		name     string
		maps     []map[string]any
		expected map[string]any
	}{
		{
			name:     "identical trees produce empty diff",
			maps:     []map[string]any{{"a": 1}, {"a": 1}, {"a": 1}},
			expected: map[string]any{},
		},
		{
			// only the second tree differs at leaf "a".
			name: "single leaf differs across trees",
			maps: []map[string]any{{"a": 1}, {"a": 2}, {"a": 1}},
			expected: map[string]any{
				"a": map[string]any{"f1": "1", "f2": "2", "f3": "1"},
			},
		},
		{
			// two trees agree, the third differs at a nested leaf.
			name: "two trees agree third differs",
			maps: []map[string]any{
				{"a": map[string]any{"b": 1}},
				{"a": map[string]any{"b": 1}},
				{"a": map[string]any{"b": 2}},
			},
			expected: map[string]any{
				"a": map[string]any{
					"b": map[string]any{"f1": "1", "f2": "1", "f3": "2"},
				},
			},
		},
		{
			// t3 lacks "a" entirely, so t1 and t2 keep descending and their
			// difference at a.y is pinpointed while t3 shows nil at a's leaves.
			// "b" exists only in t3.
			name: "missing key keeps the other two descending",
			maps: []map[string]any{
				{"a": map[string]any{"x": 1, "y": 2}},
				{"a": map[string]any{"x": 1, "y": 9}},
				{"b": 5},
			},
			expected: map[string]any{
				"a": map[string]any{
					"x": map[string]any{"f1": "1", "f2": "1", "f3": "nil"},
					"y": map[string]any{"f1": "2", "f2": "9", "f3": "nil"},
				},
				"b": map[string]any{"f1": "nil", "f2": "nil", "f3": "5"},
			},
		},
		{
			// the third tree has a leaf where the others have a map, so the
			// whole subtree is dumped for each tree at "a".
			name: "structural divergence dumps subtrees",
			maps: []map[string]any{
				{"a": map[string]any{"b": 1}},
				{"a": map[string]any{"b": 1}},
				{"a": 5},
			},
			expected: map[string]any{
				"a": map[string]any{
					"f1": map[string]any{"b": "1"},
					"f2": map[string]any{"b": "1"},
					"f3": "5",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trees := make([]*nodes.Node, len(tt.maps))
			for i, m := range tt.maps {
				trees[i] = nodes.New(m, 0, nodes.EmptyRepr)
			}
			conf := defaultDiffConf()
			conf.Keys = []string{"f1", "f2", "f3"}
			diff, err := createDiffTree(trees, conf)
			require.NoError(t, err)
			require.Equal(t, tt.expected, nodes.ToMap(diff.Children.Arr()...))
		})
	}
}

func TestCreateDiffTreeKeyCountMismatch(t *testing.T) {
	// passing fewer keys than trees must fail rather than index past the keys
	// slice.
	one := []*nodes.Node{nodes.New(map[string]any{"a": 1}, 0, nodes.EmptyRepr)}
	twoConf := defaultDiffConf()
	_, err := createDiffTree(one, twoConf)
	require.Error(t, err)

	three := []*nodes.Node{
		nodes.New(map[string]any{"a": 1}, 0, nodes.EmptyRepr),
		nodes.New(map[string]any{"a": 1}, 0, nodes.EmptyRepr),
		nodes.New(map[string]any{"a": 1}, 0, nodes.EmptyRepr),
	}
	_, err = createDiffTree(three, twoConf)
	require.Error(t, err)
}

func TestUpdateRepr(t *testing.T) {
	// collectValues walks the entire tree (ignoring Expand) and returns a map of
	// dot-separated path -> node Value for every node.
	collectValues := func(tree *nodes.Node) map[string]string {
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
