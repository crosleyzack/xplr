package tree

import (
	"testing"

	"github.com/crosleyzack/wndr/pkg/keys"
	"github.com/crosleyzack/wndr/pkg/nodes"
	"github.com/crosleyzack/wndr/pkg/styles"
	"github.com/stretchr/testify/assert"
)

func TestDefaultFormat(t *testing.T) {
	f := DefaultFormat()
	assert.Equal(t, 80, f.Width)
	assert.Equal(t, 20, f.Height)
	assert.Equal(t, "└─", f.LeafShape)
	assert.Equal(t, "❭ ", f.ExpandableShape)
	assert.Equal(t, "╰─", f.ExpandedShape)
	assert.Equal(t, 2, f.SpacesPerLayer)
	assert.Equal(t, 8, f.SpacesAfterKey)
	assert.False(t, f.HideSummaryWhenExpanded)
}

func TestNewFormatDefaults(t *testing.T) {
	f := NewFormat(&TreeConfig{})
	assert.Equal(t, DefaultFormat(), f)
}

func TestNewFormatOverrides(t *testing.T) {
	f := NewFormat(&TreeConfig{
		ExpandedShape:           "-+",
		ExpandableShape:         "+>",
		LeafShape:               "--",
		SpacesPerLayer:          3,
		HideSummaryWhenExpanded: true,
		SpacesAfterKey:          2,
	})
	assert.Equal(t, "+>", f.ExpandableShape)
	assert.Equal(t, "--", f.LeafShape)
	assert.Equal(t, "-+", f.ExpandedShape)
	assert.Equal(t, 3, f.SpacesPerLayer)
	assert.True(t, f.HideSummaryWhenExpanded)
	assert.Equal(t, 2, f.SpacesAfterKey)
	// non-overridden fields keep defaults
	assert.Equal(t, 80, f.Width)
	assert.Equal(t, 20, f.Height)
}

func TestNewFormatIgnoresZeroValues(t *testing.T) {
	f := NewFormat(&TreeConfig{})
	assert.Equal(t, 80, f.Width)
	assert.Equal(t, 20, f.Height)
}

func TestNew(t *testing.T) {
	format := DefaultFormat()
	km := keys.DefaultKeyMap()
	st := styles.DefaultStyles()
	root := nodes.New(map[string]any{
		"alpha": "1",
		"beta":  "2",
	}, 2, nodes.LeafValuesOnly)

	m := New(format, km, st, root)
	assert.Equal(t, format.Height, m.Height)
	assert.Equal(t, format.Width, m.Width)
	assert.Equal(t, format.ExpandedShape, m.ExpandedShape)
	assert.Equal(t, format.ExpandableShape, m.ExpandableShape)
	assert.Equal(t, format.LeafShape, m.LeafShape)
	assert.Equal(t, format.SpacesPerLayer, m.SpacesPerLayer)
	assert.Equal(t, format.SpacesAfterKey, m.spacesAfterKey)
	assert.Equal(t, format.HideSummaryWhenExpanded, m.hideSummaryWhenExpanded)
	assert.Same(t, root, m.Root)
	assert.Nil(t, m.searchResults)
	assert.Nil(t, m.searchNext)
	assert.Nil(t, m.searchStop)
	assert.Nil(t, m.currentNode)
}

func TestNumberOfNodes(t *testing.T) {
	root := nodes.New(map[string]any{
		"a": map[string]any{
			"a1": "1",
			"a2": "2",
		},
		"b": "3",
	}, 5, nodes.LeafValuesOnly)

	m := &Model{Root: root}
	// descendants = a, a1, a2, b = 4 (the sentinel root is not counted)
	assert.Equal(t, 4, m.NumberOfNodes())
}

func TestNumberOfNodesEmptyRoot(t *testing.T) {
	root := &nodes.Node{}
	m := &Model{Root: root}
	assert.Equal(t, 0, m.NumberOfNodes())
}

func TestInit(t *testing.T) {
	m := &Model{}
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("Init() returned nil command")
	}
}
