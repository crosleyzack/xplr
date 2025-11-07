package nodes

import (
	"iter"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDFS(t *testing.T) {
	n := &Node{
		Key:    "foo",
		Expand: true,
		Children: []*Node{
			{
				Key:    "bar",
				Expand: true,
				Children: []*Node{
					{
						Key: "baz",
					},
				},
			},
			{
				Key:    "bad",
				Expand: false,
				Children: []*Node{
					{
						Key: "unreached",
					},
				},
			},
		},
	}
	keys := make([]string, 0)
	f := func(n *Node, _ int) error {
		keys = append(keys, n.Key)
		return nil
	}
	err := DFS([]*Node{n}, f, nil)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"foo", "bar", "baz", "bad"}, keys)
	keys = make([]string, 0)
	err = DFS([]*Node{n}, f, &SearchConfig{SearchAll: true})
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"foo", "bar", "baz", "bad", "unreached"}, keys)
}

func TestDFSIter(t *testing.T) {
	n := &Node{
		Key:    "foo",
		Expand: true,
		Children: []*Node{
			{
				Key:    "bar",
				Expand: true,
				Children: []*Node{
					{
						Key: "baz",
					},
				},
			},
			{
				Key:    "bad",
				Expand: false,
				Children: []*Node{
					{
						Key: "unreached",
					},
				},
			},
		},
	}

	// test matching all nodes to check dfs ordering
	f := func(n *Node) bool {
		return true
	}
	out := make([]string, 0)
	for n := range DFSIter([]*Node{n}, f) {
		out = append(out, n.Key)
	}
	assert.Equal(t, []string{"foo", "bar", "baz", "bad", "unreached"}, out)

	// test matching only nodes which contain "ba"
	f = func(n *Node) bool {
		return strings.Contains(n.Key, "ba")
	}
	out = make([]string, 0)
	for n := range DFSIter([]*Node{n}, f) {
		out = append(out, n.Key)
	}
	assert.Equal(t, []string{"bar", "baz", "bad"}, out)
}

func TestDFSIterPull(t *testing.T) {
	n := &Node{
		Key:    "foo",
		Expand: true,
		Children: []*Node{
			{
				Key:    "bar",
				Expand: true,
				Children: []*Node{
					{
						Key: "baz",
					},
				},
			},
			{
				Key:    "bad",
				Expand: false,
				Children: []*Node{
					{
						Key: "unreached",
					},
				},
			},
		},
	}

	// test matching all nodes to check dfs ordering
	f := func(n *Node) bool {
		return true
	}
	out := make([]string, 0)
	next, stop := iter.Pull(DFSIter([]*Node{n}, f))
	for {
		n, ok := next()
		if !ok {
			break
		}
		out = append(out, n.Key)
	}
	stop()
	assert.Equal(t, []string{"foo", "bar", "baz", "bad", "unreached"}, out)

	// test matching only nodes which contain "ba"
	f = func(n *Node) bool {
		return strings.Contains(n.Key, "ba")
	}
	out = make([]string, 0)
	next, stop = iter.Pull(DFSIter([]*Node{n}, f))
	for {
		n, ok := next()
		if !ok {
			break
		}
		out = append(out, n.Key)
	}
	stop()
	assert.Equal(t, []string{"bar", "baz", "bad"}, out)
}
