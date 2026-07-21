package omap

import (
	"cmp"
	"testing"

	"github.com/stretchr/testify/assert"
)

// vals collects the values yielded by Iter, in iteration order.
func vals[K cmp.Ordered, V any](o *OMap[K, V]) []V {
	var out []V
	for _, v := range o.Iter() {
		out = append(out, v)
	}
	return out
}

// keysOf collects the keys yielded by Keys, in iteration order.
func keysOf[K cmp.Ordered, V any](o *OMap[K, V]) []K {
	var out []K
	for k := range o.Keys() {
		out = append(out, k)
	}
	return out
}

// seed builds an OMap in the given order mode and fills it from m via PutAll.
func seed[K cmp.Ordered, V any](order bool, m map[K]V) OMap[K, V] {
	o := New[K, V](WithOrder(order))
	o.PutAll(m)
	return o
}

// orderModes drives tests that must hold in both ordered and unordered mode.
var orderModes = []struct {
	name  string
	order bool
}{
	{name: "ordered", order: true},
	{name: "unordered", order: false},
}

func TestNew(t *testing.T) {
	t.Run("empty and ordered by default", func(t *testing.T) {
		o := New[string, int]()
		assert.Equal(t, 0, o.Len())
		assert.Nil(t, o.Arr())
		o.Put("b", 2)
		o.Put("a", 1)
		assert.Equal(t, []int{1, 2}, vals(&o)) // ordered
	})

	t.Run("unordered skips sorting", func(t *testing.T) {
		o := New[string, int](WithOrder(false))
		o.Put("b", 2)
		o.Put("a", 1)
		assert.ElementsMatch(t, []int{1, 2}, vals(&o))
	})
}

func TestPutAll(t *testing.T) {
	for _, mode := range orderModes {
		t.Run(mode.name, func(t *testing.T) {
			o := New[string, int](WithOrder(mode.order))
			o.PutAll(map[string]int{"a": 1, "b": 2})
			o.PutAll(map[string]int{"b": 20, "c": 3}) // overwrites b, adds c
			o.PutAll(nil)                             // no-op
			assert.Equal(t, 3, o.Len())

			if mode.order {
				assert.Equal(t, []int{1, 20, 3}, vals(&o)) // a, b, c ascending
			} else {
				assert.ElementsMatch(t, []int{1, 20, 3}, vals(&o))
			}
		})
	}
}

func TestLen(t *testing.T) {
	for _, mode := range orderModes {
		t.Run(mode.name, func(t *testing.T) {
			empty := New[string, int](WithOrder(mode.order))
			assert.Equal(t, 0, empty.Len())
			three := seed(mode.order, map[string]int{"a": 1, "b": 2, "c": 3})
			assert.Equal(t, 3, three.Len())
		})
	}
}

func TestGet(t *testing.T) {
	for _, mode := range orderModes {
		t.Run(mode.name, func(t *testing.T) {
			o := seed(mode.order, map[string]int{"a": 1, "b": 2})
			got, ok := o.Get("a")
			assert.True(t, ok)
			assert.Equal(t, 1, got)
			_, ok = o.Get("z")
			assert.False(t, ok)
		})
	}
}

func TestPut(t *testing.T) {
	for _, mode := range orderModes {
		t.Run(mode.name, func(t *testing.T) {
			o := New[string, int](WithOrder(mode.order))
			o.Put("b", 2)
			o.Put("a", 1)
			o.Put("a", 9) // overwrite, must not grow
			assert.Equal(t, 2, o.Len())
			got, _ := o.Get("a")
			assert.Equal(t, 9, got)
			assert.ElementsMatch(t, []int{9, 2}, vals(&o))
		})
	}
}

func TestDelete(t *testing.T) {
	for _, mode := range orderModes {
		t.Run(mode.name, func(t *testing.T) {
			o := seed(mode.order, map[string]int{"a": 1, "b": 2, "c": 3})
			o.Delete("b")
			o.Delete("z") // absent: no-op
			assert.Equal(t, 2, o.Len())
			_, ok := o.Get("b")
			assert.False(t, ok)
			assert.ElementsMatch(t, []int{1, 3}, vals(&o))
		})
	}
}

func TestKeys(t *testing.T) {
	in := map[string]int{"c": 3, "a": 1, "b": 2}
	t.Run("ordered ascending", func(t *testing.T) {
		o := seed(true, in)
		assert.Equal(t, []string{"a", "b", "c"}, keysOf(&o))
	})
	t.Run("unordered has all keys", func(t *testing.T) {
		o := seed(false, in)
		assert.ElementsMatch(t, []string{"a", "b", "c"}, keysOf(&o))
	})
	t.Run("empty", func(t *testing.T) {
		o := New[string, int]()
		assert.Empty(t, keysOf(&o))
	})
}

func TestArr(t *testing.T) {
	in := map[string]int{"c": 3, "a": 1, "b": 2}
	t.Run("ordered ascending", func(t *testing.T) {
		o := seed(true, in)
		assert.Equal(t, []int{1, 2, 3}, o.Arr())
	})
	t.Run("unordered has all values", func(t *testing.T) {
		o := seed(false, in)
		assert.ElementsMatch(t, []int{1, 2, 3}, o.Arr())
	})
	t.Run("empty returns nil", func(t *testing.T) {
		o := New[string, int]()
		assert.Nil(t, o.Arr())
	})
}

func TestIter(t *testing.T) {
	in := map[string]int{"c": 3, "a": 1, "b": 2}
	t.Run("ordered yields pairs in key order", func(t *testing.T) {
		o := seed(true, in)
		var ks []string
		var vs []int
		for k, v := range o.Iter() {
			ks = append(ks, k)
			vs = append(vs, v)
		}
		assert.Equal(t, []string{"a", "b", "c"}, ks)
		assert.Equal(t, []int{1, 2, 3}, vs)
	})
	t.Run("unordered yields all pairs", func(t *testing.T) {
		o := seed(false, in)
		got := map[string]int{}
		for k, v := range o.Iter() {
			got[k] = v
		}
		assert.Equal(t, in, got)
	})
}

func TestIterEarlyStop(t *testing.T) {
	o := seed(true, map[string]int{"a": 1, "b": 2, "c": 3})
	var got []int
	for _, v := range o.Iter() {
		got = append(got, v)
		if len(got) == 2 {
			break // exercises the yield-returns-false path
		}
	}
	assert.Equal(t, []int{1, 2}, got)
}

// TestZeroValueUsableUnordered documents that a zero-value OMap works but is
// unordered (the btree is only built when ordered).
func TestZeroValueUsableUnordered(t *testing.T) {
	var o OMap[string, int]
	assert.NotPanics(t, func() {
		o.Put("b", 2)
		o.Put("a", 1)
	})
	got, ok := o.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 1, got)
	assert.ElementsMatch(t, []int{1, 2}, o.Arr())
}

// TestOrderingIntKeys documents that integer keys iterate in numeric order.
func TestOrderingIntKeys(t *testing.T) {
	o := seed(true, map[int]string{10: "ten", 2: "two", 1: "one", 3: "three"})
	assert.Equal(t, []string{"one", "two", "three", "ten"}, vals(&o))
}

// TestOrderingStringKeysLexicographic documents that string keys iterate in
// lexicographic order, so numeric-looking keys sort as text ("10" before "2").
func TestOrderingStringKeysLexicographic(t *testing.T) {
	o := seed(true, map[string]int{"0": 0, "1": 1, "2": 2, "10": 10})
	assert.Equal(t, []int{0, 1, 10, 2}, vals(&o))
}
