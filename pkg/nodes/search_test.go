package nodes

import (
	"errors"
	"iter"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// testSearchTree builds the standard test tree used across search tests.
// Children iterate in key order (so "bad" sorts before "bar"), but note the
// recursive DFS and the stack-based DFSIter disagree on sibling order: DFS
// visits children in ascending key order, while DFSIter pops from the end of
// its stack and so visits siblings in reverse ("bar" before "bad").
//
//	foo (expand)
//	  bad (collapsed)
//	    unreached
//	  bar (expand)
//	    baz
func testSearchTree() *Node {
	return &Node{
		Key:    "foo",
		Expand: true,
		Children: childMap(
			&Node{
				Key:      "bar",
				Expand:   true,
				Children: childMap(&Node{Key: "baz"}),
			},
			&Node{
				Key:      "bad",
				Expand:   false,
				Children: childMap(&Node{Key: "unreached"}),
			},
		),
	}
}

// rootOf wraps top-level nodes under a sentinel root (Parent == nil), mirroring
// how New builds a tree. DFS/DFSIter skip this sentinel and start at its
// children, so the wrapped nodes are visited at layer 0.
func rootOf(children ...*Node) *Node {
	root := &Node{Children: childMap(children...)}
	for _, c := range children {
		c.Parent = root
	}
	return root
}

func TestDFS(t *testing.T) {
	tests := []struct {
		name       string
		opts       []DFSOption
		wantKeys   []string
		wantLayers []int
	}{
		{
			name:       "default ObeyExpand skips collapsed children",
			wantKeys:   []string{"foo", "bad", "bar", "baz"},
			wantLayers: []int{0, 1, 1, 2},
		},
		{
			name:       "AllChildren visits all nodes",
			opts:       []DFSOption{WithNextNodes(AllChildren)},
			wantKeys:   []string{"foo", "bad", "unreached", "bar", "baz"},
			wantLayers: []int{0, 1, 2, 1, 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotKeys []string
			var gotLayers []int
			err := DFS(rootOf(testSearchTree()), func(n *Node, layer int) error {
				gotKeys = append(gotKeys, n.Key)
				gotLayers = append(gotLayers, layer)
				return nil
			}, tt.opts...)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantKeys, gotKeys)
			assert.Equal(t, tt.wantLayers, gotLayers)
		})
	}
}

func TestDFSError(t *testing.T) {
	sentinel := errors.New("stop here")
	var visited []string
	err := DFS(rootOf(testSearchTree()), func(n *Node, _ int) error {
		visited = append(visited, n.Key)
		if n.Key == "bar" {
			return sentinel
		}
		return nil
	})
	assert.ErrorIs(t, err, sentinel)
	assert.Equal(t, []string{"foo", "bad", "bar"}, visited)
}

func TestDFSIter(t *testing.T) {
	tests := []struct {
		name     string
		filter   func(*Node) bool
		opts     []DFSOption
		wantKeys []string
	}{
		{
			name:     "ObeyExpand",
			filter:   func(*Node) bool { return true },
			wantKeys: []string{"foo", "bar", "baz", "bad"},
		},
		{
			name:     "match all with AllChildren",
			filter:   func(*Node) bool { return true },
			opts:     []DFSOption{WithNextNodes(AllChildren)},
			wantKeys: []string{"foo", "bar", "baz", "bad", "unreached"},
		},
		{
			name:     "filter by key substring",
			filter:   func(n *Node) bool { return strings.Contains(n.Key, "ba") },
			wantKeys: []string{"bar", "baz", "bad"},
		},
		{
			name:     "filter matching nothing",
			filter:   func(*Node) bool { return false },
			wantKeys: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []string
			for n := range DFSIter(rootOf(testSearchTree()), tt.filter, tt.opts...) {
				got = append(got, n.Key)
			}
			assert.Equal(t, tt.wantKeys, got)
		})
	}
}

func TestDFSIterPull(t *testing.T) {
	tests := []struct {
		name     string
		filter   func(*Node) bool
		opts     []DFSOption
		wantKeys []string
	}{
		{
			name:     "default ObeyExpand skips collapsed children",
			filter:   func(*Node) bool { return true },
			wantKeys: []string{"foo", "bar", "baz", "bad"},
		},
		{
			name:     "match all with AllChildren",
			filter:   func(*Node) bool { return true },
			opts:     []DFSOption{WithNextNodes(AllChildren)},
			wantKeys: []string{"foo", "bar", "baz", "bad", "unreached"},
		},
		{
			name:     "filter by key substring",
			filter:   func(n *Node) bool { return strings.Contains(n.Key, "ba") },
			wantKeys: []string{"bar", "baz", "bad"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []string
			next, stop := iter.Pull(DFSIter(rootOf(testSearchTree()), tt.filter, tt.opts...))
			defer stop()
			for {
				n, ok := next()
				if !ok {
					break
				}
				got = append(got, n.Key)
			}
			assert.Equal(t, tt.wantKeys, got)
		})
	}
}

// nodeAt records what a single tree resolved to at a visited path. Key is the
// Key of the resolved node; Rem holds the unresolved path segments when the
// path is absent from that tree (empty when the node exists).
type nodeAt struct {
	Key string
	Rem []string
}

// multiVisit records one DFSMulti callback invocation. The nodes slice DFSMulti
// passes is reused across calls, so we snapshot the fields we care about here
// rather than retaining the ChildNode values.
type multiVisit struct {
	Path  []string
	Nodes []nodeAt
}

// collectMulti runs DFSMulti over trees and returns one multiVisit per callback
// invocation. Visit order is not asserted because getNewPaths derives the next
// paths from a map, so callers compare with assert.ElementsMatch.
func collectMulti(t *testing.T, trees ...*Node) []multiVisit {
	t.Helper()
	var got []multiVisit
	err := DFSMulti(func(path []string, ns []ChildNode) error {
		snap := make([]nodeAt, len(ns))
		for i, n := range ns {
			snap[i] = nodeAt{Rem: n.Rem}
			if n.Node != nil {
				snap[i].Key = n.Node.Key
			}
		}
		got = append(got, multiVisit{Path: path, Nodes: snap})
		return nil
	}, trees...)
	assert.NoError(t, err)
	return got
}

func TestDFSMulti(t *testing.T) {
	tests := []struct {
		name  string
		trees []*Node
		want  []multiVisit
	}{
		{
			name:  "single tree visits every path",
			trees: []*Node{New(map[string]any{"a": map[string]any{"b": "1"}, "c": "2"}, 5, EmptyRepr)},
			want: []multiVisit{
				{Path: []string{}, Nodes: []nodeAt{{Key: ""}}},
				{Path: []string{"a"}, Nodes: []nodeAt{{Key: "a"}}},
				{Path: []string{"c"}, Nodes: []nodeAt{{Key: "c"}}},
				{Path: []string{"a", "b"}, Nodes: []nodeAt{{Key: "b"}}},
			},
		},
		{
			name: "identical trees resolve on both sides",
			trees: []*Node{
				New(map[string]any{"a": "1"}, 5, EmptyRepr),
				New(map[string]any{"a": "1"}, 5, EmptyRepr),
			},
			want: []multiVisit{
				{Path: []string{}, Nodes: []nodeAt{{Key: ""}, {Key: ""}}},
				{Path: []string{"a"}, Nodes: []nodeAt{{Key: "a"}, {Key: "a"}}},
			},
		},
		{
			name: "divergent trees visit the union of paths",
			trees: []*Node{
				New(map[string]any{"a": map[string]any{"b": "1"}, "shared": "x"}, 5, EmptyRepr),
				New(map[string]any{"a": map[string]any{"c": "2"}, "shared": "y"}, 5, EmptyRepr),
			},
			want: []multiVisit{
				{Path: []string{}, Nodes: []nodeAt{{Key: ""}, {Key: ""}}},
				{Path: []string{"a"}, Nodes: []nodeAt{{Key: "a"}, {Key: "a"}}},
				{Path: []string{"shared"}, Nodes: []nodeAt{{Key: "shared"}, {Key: "shared"}}},
				// b exists only in tree 1; tree 2 stops at "a" with "b" unresolved.
				{Path: []string{"a", "b"}, Nodes: []nodeAt{{Key: "b"}, {Key: "a", Rem: []string{"b"}}}},
				// c exists only in tree 2; tree 1 stops at "a" with "c" unresolved.
				{Path: []string{"a", "c"}, Nodes: []nodeAt{{Key: "a", Rem: []string{"c"}}, {Key: "c"}}},
			},
		},
		{
			name: "remainder holds every unresolved segment",
			trees: []*Node{
				New(map[string]any{"a": map[string]any{"b": map[string]any{"c": "1"}}}, 5, EmptyRepr),
				New(map[string]any{"a": "2"}, 5, EmptyRepr),
			},
			want: []multiVisit{
				{Path: []string{}, Nodes: []nodeAt{{Key: ""}, {Key: ""}}},
				{Path: []string{"a"}, Nodes: []nodeAt{{Key: "a"}, {Key: "a"}}},
				{Path: []string{"a", "b"}, Nodes: []nodeAt{{Key: "b"}, {Key: "a", Rem: []string{"b"}}}},
				{Path: []string{"a", "b", "c"}, Nodes: []nodeAt{{Key: "c"}, {Key: "a", Rem: []string{"b", "c"}}}},
			},
		},
		{
			name: "nil tree contributes no children and never panics",
			trees: []*Node{
				New(map[string]any{"a": "1"}, 5, EmptyRepr),
				nil,
			},
			want: []multiVisit{
				// GetNodeFromPath on a nil tree returns a nil node with the
				// full path unresolved.
				{Path: []string{}, Nodes: []nodeAt{{Key: ""}, {Rem: []string{}}}},
				{Path: []string{"a"}, Nodes: []nodeAt{{Key: "a"}, {Rem: []string{"a"}}}},
			},
		},
		{
			name:  "no trees yields a single root visit",
			trees: nil,
			want: []multiVisit{
				{Path: []string{}, Nodes: []nodeAt{}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := collectMulti(t, tt.trees...)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestDFSMultiError(t *testing.T) {
	sentinel := errors.New("stop here")
	var visited int
	err := DFSMulti(func(path []string, _ []ChildNode) error {
		visited++
		if len(path) == 1 && path[0] == "a" {
			return sentinel
		}
		return nil
	}, New(map[string]any{"a": map[string]any{"deep": "1"}}, 5, EmptyRepr))
	assert.ErrorIs(t, err, sentinel)
	// traversal stops at the erroring node: only the root and "a" are visited,
	// never "a/deep".
	assert.Equal(t, 2, visited)
}

func TestAllChildren(t *testing.T) {
	child := &Node{Key: "child"}
	tests := []struct {
		name string
		node *Node
		want []*Node
	}{
		{
			name: "returns children when expanded",
			node: &Node{Expand: true, Children: childMap(child)},
			want: []*Node{child},
		},
		{
			name: "returns children even when collapsed",
			node: &Node{Expand: false, Children: childMap(child)},
			want: []*Node{child},
		},
		{
			name: "returns nil for no children",
			node: &Node{},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AllChildren(tt.node)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestObeyExpand(t *testing.T) {
	child := &Node{Key: "child"}
	tests := []struct {
		name string
		node *Node
		want []*Node
	}{
		{
			name: "expanded node returns children",
			node: &Node{Expand: true, Children: childMap(child)},
			want: []*Node{child},
		},
		{
			name: "collapsed node returns nil",
			node: &Node{Expand: false, Children: childMap(child)},
			want: nil,
		},
		{
			name: "expanded node with no children returns nil",
			node: &Node{Expand: true},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ObeyExpand(tt.node)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWithNextNodes(t *testing.T) {
	// WithNextNodes should override the default ObeyExpand with AllChildren,
	// causing collapsed nodes' children to be visited.
	var got []string
	err := DFS(rootOf(testSearchTree()), func(n *Node, _ int) error {
		got = append(got, n.Key)
		return nil
	}, WithNextNodes(AllChildren))
	assert.NoError(t, err)
	assert.Equal(t, []string{"foo", "bad", "unreached", "bar", "baz"}, got)
}
