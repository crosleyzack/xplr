package nodes

import (
	"errors"
	"iter"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// testSearchTree builds the standard test tree used across search tests:
//
//	foo (expand)
//	  bar (expand)
//	    baz
//	  bad (collapsed)
//	    unreached
func testSearchTree() *Node {
	return &Node{
		Key:    "foo",
		Expand: true,
		Children: []*Node{
			{
				Key:    "bar",
				Expand: true,
				Children: []*Node{
					{Key: "baz"},
				},
			},
			{
				Key:    "bad",
				Expand: false,
				Children: []*Node{
					{Key: "unreached"},
				},
			},
		},
	}
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
			wantKeys:   []string{"foo", "bar", "baz", "bad"},
			wantLayers: []int{0, 1, 2, 1},
		},
		{
			name:       "AllChildren visits all nodes",
			opts:       []DFSOption{WithNextNodes(AllChildren)},
			wantKeys:   []string{"foo", "bar", "baz", "bad", "unreached"},
			wantLayers: []int{0, 1, 2, 1, 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotKeys []string
			var gotLayers []int
			err := DFS([]*Node{testSearchTree()}, func(n *Node, layer int) error {
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
	err := DFS([]*Node{testSearchTree()}, func(n *Node, _ int) error {
		visited = append(visited, n.Key)
		if n.Key == "bar" {
			return sentinel
		}
		return nil
	})
	assert.ErrorIs(t, err, sentinel)
	assert.Equal(t, []string{"foo", "bar"}, visited)
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
			for n := range DFSIter([]*Node{testSearchTree()}, tt.filter, tt.opts...) {
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
			next, stop := iter.Pull(DFSIter([]*Node{testSearchTree()}, tt.filter, tt.opts...))
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
			node: &Node{Expand: true, Children: []*Node{child}},
			want: []*Node{child},
		},
		{
			name: "returns children even when collapsed",
			node: &Node{Expand: false, Children: []*Node{child}},
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
			node: &Node{Expand: true, Children: []*Node{child}},
			want: []*Node{child},
		},
		{
			name: "collapsed node returns nil",
			node: &Node{Expand: false, Children: []*Node{child}},
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
	err := DFS([]*Node{testSearchTree()}, func(n *Node, _ int) error {
		got = append(got, n.Key)
		return nil
	}, WithNextNodes(AllChildren))
	assert.NoError(t, err)
	assert.Equal(t, []string{"foo", "bar", "baz", "bad", "unreached"}, got)
}
