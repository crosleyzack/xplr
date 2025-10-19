package nodes

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetPathToNode(t *testing.T) {
	// setup path foo.bar[0].baz.final
	parent1 := &Node{
		ID:  uuid.New(),
		Key: "final",
	}
	leaf := &Node{
		ID:     uuid.New(),
		Key:    "target",
		Value:  "hello",
		Parent: parent1,
	}
	parent1.Children = []*Node{leaf}
	sibling1 := &Node{
		ID:  uuid.New(),
		Key: "bar",
		Children: []*Node{
			parent1,
		},
	}
	sibling2 := &Node{
		ID:  uuid.New(),
		Key: "bad",
	}
	parent1.Parent = sibling1
	top := &Node{
		ID:  uuid.New(),
		Key: "foo",
		Children: []*Node{
			sibling1, sibling2,
		},
	}
	sibling1.Parent = top
	sibling2.Parent = top
	// get path
	path := GetPathToNode(leaf)
	assert.Equal(t, "foo.bar.final.target", path)
}
