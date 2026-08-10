package nodes

import (
	"errors"
	"iter"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// testSearchTree builds the standard test tree used across search tests.
// Children iterate in key order, so "bad" is visited before "bar".
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
			wantKeys: []string{"foo", "bad", "bar", "baz"},
		},
		{
			name:     "match all with AllChildren",
			filter:   func(*Node) bool { return true },
			opts:     []DFSOption{WithNextNodes(AllChildren)},
			wantKeys: []string{"foo", "bad", "unreached", "bar", "baz"},
		},
		{
			name:     "filter by key substring",
			filter:   func(n *Node) bool { return strings.Contains(n.Key, "ba") },
			wantKeys: []string{"bad", "bar", "baz"},
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
			wantKeys: []string{"foo", "bad", "bar", "baz"},
		},
		{
			name:     "match all with AllChildren",
			filter:   func(*Node) bool { return true },
			opts:     []DFSOption{WithNextNodes(AllChildren)},
			wantKeys: []string{"foo", "bad", "unreached", "bar", "baz"},
		},
		{
			name:     "filter by key substring",
			filter:   func(n *Node) bool { return strings.Contains(n.Key, "ba") },
			wantKeys: []string{"bad", "bar", "baz"},
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
