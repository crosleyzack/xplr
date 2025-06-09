package nodes

import (
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
