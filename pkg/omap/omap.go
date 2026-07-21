// Package omap provides OMap, an ordered map that keeps its keys sorted.
package omap

import (
	"cmp"
	"iter"

	"github.com/tidwall/btree"
)

// OMap is a map that keeps its keys in sorted order. It pairs a hash map (for
// constant-time lookup) with a btree.Set over the keys (for ordered iteration),
// so lookups stay fast without giving up a stable iteration order.
//
// A zero-value OMap is usable but unordered; use New (ordered by default) when
// you need sorted iteration.
//
// Time complexity (n = number of entries):
//
//	New     O(1)
//	Len     O(1)
//	Get     O(1)
//	Put     O(log n)
//	PutAll  O(n log n)
//	Delete  O(log n)
//	Keys    O(n) time, O(1) space
//	Iter    O(n) time, O(1) space
//	Arr     O(n) time, O(n) space
//
// Keys order by the built-in "<": numeric for integer keys, lexicographic for
// strings.
type OMap[K cmp.Ordered, V any] struct {
	mp    map[K]V
	keys  btree.Set[K]
	order bool
}

// options holds the settings applied by New's Option arguments. It is
// non-generic so WithOrder needs no type arguments; New's type parameters are
// supplied explicitly at the call site.
type options struct {
	order bool
}

// Option configures an OMap built by New.
type Option func(*options)

// WithOrder sets whether the OMap maintains sorted key order. Defaults to true;
// pass false to skip the key btree and use less memory at the cost of ordered
// iteration.
func WithOrder(order bool) Option {
	return func(o *options) { o.order = order }
}

// New returns an empty OMap. It is ordered by default; pass WithOrder(false) to
// skip the key btree for lower memory use. Seed it from a map with PutAll.
//
// Complexity: O(1).
func New[K cmp.Ordered, V any](opts ...Option) OMap[K, V] {
	cfg := options{order: true}
	for _, opt := range opts {
		opt(&cfg)
	}
	ret := OMap[K, V]{
		mp:    make(map[K]V, 0),
		order: cfg.order,
	}
	return ret
}

// Len returns the number of entries.
//
// Complexity: O(1).
func (o *OMap[K, V]) Len() int {
	return len(o.mp)
}

// Get returns the value stored for key and whether it was present.
//
// Complexity: O(1).
func (o *OMap[K, V]) Get(key K) (V, bool) {
	v, ok := o.mp[key]
	return v, ok
}

// Put stores val under key, overwriting any existing value.
//
// Complexity: O(log n) — a constant-time map write plus a logarithmic btree
// insert.
func (o *OMap[K, V]) Put(key K, val V) {
	if o.mp == nil {
		o.mp = make(map[K]V)
	}
	o.mp[key] = val
	if o.order {
		o.keys.Insert(key)
	}
}

// PutAll stores every entry of m, overwriting any existing keys. A nil map is
// a no-op.
//
// Complexity: O(n log n) for n entries (O(n) when unordered).
func (o *OMap[K, V]) PutAll(m map[K]V) {
	if m == nil {
		return
	}
	for k, v := range m {
		o.Put(k, v)
	}
}

// Keys returns an iterator over the keys in ascending order.
//
// Complexity: O(n) time, O(1) space.
func (o *OMap[K, V]) Keys() iter.Seq[K] {
	if o.order {
		return func(yield func(K) bool) {
			o.keys.Scan(yield) // Scan's callback IS func(K) bool
		}
	}
	return func(yield func(K) bool) {
		for key := range o.mp {
			if !yield(key) {
				return
			}
		}
	}
}

// Iter returns an iterator over the key/value pairs in ascending key order.
//
// Complexity: O(1) to obtain the iterator; consuming it is O(n) time and O(1)
// space (it streams over the key btree without allocating a snapshot).
func (o *OMap[K, V]) Iter() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for key := range o.Keys() {
			if !yield(key, o.mp[key]) {
				return
			}
		}
	}
}

// Arr returns all values as a slice in ascending key order, or nil if empty.
//
// Complexity: O(n) time and space.
func (o *OMap[K, V]) Arr() []V {
	if o.Len() == 0 {
		return nil
	}
	ret := make([]V, 0, o.Len())
	for _, v := range o.Iter() {
		ret = append(ret, v)
	}
	return ret
}

// Delete removes key and its value. Deleting an absent key is a no-op.
//
// Complexity: O(log n).
func (o *OMap[K, V]) Delete(key K) {
	delete(o.mp, key)
	if o.order {
		o.keys.Delete(key)
	}
}
