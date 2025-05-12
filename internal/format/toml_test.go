package format

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseToml(t *testing.T) {
	tml := `Age = 25
Cats = [ "Cauchy", "Plato" ]
Perfection = [ 6, 28, 496, 8128 ]
[children]
	alpha = 10
	bravo = 20`
	m, err := ParseToml([]byte(tml))
	assert.NoError(t, err)
	assert.Len(t, m, 4)
	assert.EqualValues(t, 25, m["Age"])
	assert.ElementsMatch(t, []any{"Cauchy", "Plato"}, m["Cats"])
	assert.ElementsMatch(t, []any{int64(6), int64(28), int64(496), int64(8128)}, m["Perfection"])
	assert.Len(t, m["children"], 2)
	assert.EqualValues(t, 10, m["children"].(map[string]any)["alpha"])
	assert.EqualValues(t, 20, m["children"].(map[string]any)["bravo"])
}
