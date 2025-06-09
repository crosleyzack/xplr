package format

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseYaml(t *testing.T) {
	// TODO
	yml := `---
foo: 1
bar: c
baz: [3, 4]
bad:
  guy: moriarty
`
	m, err := ParseYaml([]byte(yml))
	assert.NoError(t, err)
	assert.Len(t, m, 4)
	assert.EqualValues(t, 1, m["foo"])
	assert.Equal(t, "c", m["bar"])
	assert.ElementsMatch(t, []any{uint64(3), uint64(4)}, m["baz"])
	assert.Len(t, m["bad"], 1)
	assert.Equal(t, "moriarty", m["bad"].(map[string]any)["guy"])
}
