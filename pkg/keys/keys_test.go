package keys

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultKeyMap(t *testing.T) {
	km := DefaultKeyMap()
	assert.Equal(t, 12, km.Len())
	assert.Equal(t, []string{"bottom", "G"}, km.Bottom.Keys())
	assert.Equal(t, []string{"top", "g"}, km.Top.Keys())
	assert.Equal(t, []string{"down", "j"}, km.Down.Keys())
	assert.Equal(t, []string{"up", "k"}, km.Up.Keys())
	assert.Equal(t, []string{"tab", "h", "l"}, km.CollapseToggle.Keys())
	assert.Equal(t, []string{"<", "H"}, km.CollapseAll.Keys())
	assert.Equal(t, []string{">", "L"}, km.ExpandAll.Keys())
	assert.Equal(t, []string{"?"}, km.Help.Keys())
	assert.Equal(t, []string{"q", "esc"}, km.Quit.Keys())
	assert.Equal(t, []string{"/"}, km.Search.Keys())
	assert.Equal(t, []string{"enter"}, km.Submit.Keys())
	assert.Equal(t, []string{"n"}, km.Next.Keys())
	assert.Equal(t, []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"}, km.Num.Keys())
}

func TestLen(t *testing.T) {
	assert.Equal(t, 12, (KeyMap{}).Len())
}

func TestNewKeyMapDefaults(t *testing.T) {
	km := NewKeyMap(&KeyConfig{})
	assert.Equal(t, km, DefaultKeyMap())
}

func TestNewKeyMapCustom(t *testing.T) {
	c := &KeyConfig{
		BottomKeys:         []string{"ctrl+e"},
		TopKeys:            []string{"ctrl+a"},
		DownKeys:           []string{"d"},
		UpKeys:             []string{"u"},
		CollapseToggleKeys: []string{"c"},
		CollapseAllKeys:    []string{"["},
		ExpandAllKeys:      []string{"]"},
		HelpKeys:           []string{"h"},
		QuitKeys:           []string{"x"},
		SearchKeys:         []string{"s"},
		SubmitKeys:         []string{"return"},
		NextKeys:           []string{"m"},
	}
	km := NewKeyMap(c)
	assert.Equal(t, []string{"ctrl+e"}, km.Bottom.Keys())
	assert.Equal(t, []string{"ctrl+a"}, km.Top.Keys())
	assert.Equal(t, []string{"d"}, km.Down.Keys())
	assert.Equal(t, []string{"u"}, km.Up.Keys())
	assert.Equal(t, []string{"c"}, km.CollapseToggle.Keys())
	assert.Equal(t, []string{"["}, km.CollapseAll.Keys())
	assert.Equal(t, []string{"]"}, km.ExpandAll.Keys())
	assert.Equal(t, []string{"h"}, km.Help.Keys())
	assert.Equal(t, []string{"x"}, km.Quit.Keys())
	assert.Equal(t, []string{"s"}, km.Search.Keys())
	assert.Equal(t, []string{"return"}, km.Submit.Keys())
	assert.Equal(t, []string{"m"}, km.Next.Keys())
	// fields not overridden fall back to defaults
	assert.Equal(t, []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"}, km.Num.Keys())
}

func TestNewConfig(t *testing.T) {
	data := []byte(`
BottomKeys = ["ctrl+b"]
TopKeys = ["ctrl+t"]
DownKeys = ["d"]
`)
	c, err := NewConfig(data)
	assert.NoError(t, err)
	assert.Equal(t, []string{"ctrl+b"}, c.BottomKeys)
	assert.Equal(t, []string{"ctrl+t"}, c.TopKeys)
	assert.Equal(t, []string{"d"}, c.DownKeys)
}

func TestNewConfigInvalid(t *testing.T) {
	_, err := NewConfig([]byte("not valid toml =="))
	assert.Error(t, err)
}
