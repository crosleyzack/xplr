// Package diff compares two or more tree-based data structures and produces a
// single tree describing where they differ. Every tree is walked together so a
// node present in only some trees still appears in the diff, and each tree's
// contribution is labeled under its own key (default "f1", "f2", ...).
//
// The Diff function accepts any number of trees and is configured through
// functional options: WithKeys renames each tree in the output, WithNilValue
// customizes the placeholder for a path absent from a tree, WithReprMethod
// selects how node values are rendered, and WithColors sets the color assigned
// to each tree. The result is a *nodes.Node tree ready to render or serialize.
package diff

import (
	"fmt"
	"strings"

	"github.com/crosleyzack/wndr/pkg/nodes"
	"github.com/google/uuid"
)

type diffConf struct {
	// Keys labels each tree in the diff output; Keys[i] is used for trees[i].
	Keys     []string
	NilValue string
	Repr     nodes.ReprNode
	Colors   []string
}

var defaultDiffColors = []string{"#ad0116", "#006222", "#38DB89", "#61707D", "#9D69A3"}

func defaultDiffConf() *diffConf {
	return &diffConf{
		Keys:     []string{"f1", "f2"},
		NilValue: "nil",
		Repr:     nodes.LeafKeyAndValues,
		Colors:   defaultDiffColors,
	}
}

type DiffTreeOption func(*diffConf)

// WithKeys sets the label used for each tree in the diff output. One key must be
// provided per tree passed to createDiffTree.
func WithKeys(keys ...string) DiffTreeOption {
	return func(c *diffConf) {
		c.Keys = keys
	}
}

func WithNilValue(val string) DiffTreeOption {
	return func(c *diffConf) {
		c.NilValue = val
	}
}

func WithReprMethod(repr nodes.ReprNode) DiffTreeOption {
	return func(c *diffConf) {
		c.Repr = repr
	}
}

func WithColors(colors []string) DiffTreeOption {
	return func(c *diffConf) {
		c.Colors = colors
	}
}

func Diff(trees []*nodes.Node, opts ...DiffTreeOption) (*nodes.Node, error) {
	conf := defaultDiffConf()
	for _, opt := range opts {
		opt(conf)
	}
	tree, err := createDiffTree(trees, conf)
	if err != nil {
		return nil, fmt.Errorf("failed to created diff tree: %w", err)
	}
	// set the tree repr values
	if err := updateRepr(tree, conf.Repr); err != nil {
		return nil, fmt.Errorf("failed to update tree repr: %w", err)
	}
	tree, err = addMeta(tree, conf, conf.Keys...)
	if err != nil {
		return nil, fmt.Errorf("failed to add tree meta: %w", err)
	}
	return tree, nil
}

func createDiffTree(trees []*nodes.Node, conf *diffConf) (*nodes.Node, error) {
	// keys[i] labels the node contributed by trees[i] in the diff tree, so one
	// key is required per tree.
	keys := conf.Keys
	if len(keys) != len(trees) {
		return nil, fmt.Errorf("need one key per tree: got %d keys for %d trees", len(keys), len(trees))
	}
	diffTree := nodes.New(map[string]any{}, 0, nodes.EmptyRepr)
	// pruned holds diff-tree paths whose whole subtree is already recorded as a
	// difference. DFSMulti still walks the descendants of any tree that has
	// them, so we skip every path under a pruned prefix to avoid recording the
	// same difference twice.
	pruned := make(map[string]bool)
	isPruned := func(path []string) bool {
		for i := 1; i <= len(path); i++ {
			if pruned[strings.Join(path[:i], "\x00")] {
				return true
			}
		}
		return false
	}
	err := nodes.DFSMulti(
		func(path []string, cnodes []nodes.ChildNode) error {
			// the sentinel root holds no value to compare; always descend.
			if len(path) == 0 {
				return nil
			}
			// skip anything under a difference we have already recorded.
			if isPruned(path) {
				return nil
			}
			// resolve each tree's node at this path (nil when the path is
			// absent from that tree).
			resolved := make([]*nodes.Node, len(cnodes))
			for i, cn := range cnodes {
				if cn.Node != nil && len(cn.Rem) == 0 {
					resolved[i] = cn.Node
				}
			}
			// when every tree has an equivalent node there is no difference at
			// this level; let DFSMulti descend to compare their children.
			if nodesEquivalent(resolved...) {
				return nil
			}
			// when at least two trees still share an equivalent non-leaf node,
			// keep descending so their differing children are compared one level
			// at a time. Any tree that is absent here stays absent (nil) on
			// those deeper paths and surfaces at the leaves where a difference
			// is recorded.
			if canDescend(resolved) {
				return nil
			}
			// record the difference: each tree contributes its node (or the nil
			// value when the path is absent) under its key, then the whole
			// subtree is pruned so it is not expanded further.
			var err error
			for i, node := range resolved {
				add := copyNode(node)
				if add == nil {
					add = &nodes.Node{ID: uuid.New(), Value: conf.NilValue}
				}
				add.Key = keys[i]
				diffTree, err = addNode(diffTree, path, add)
				if err != nil {
					return fmt.Errorf("failed to add node to diff tree: %w", err)
				}
			}
			// the whole subtree at this path is now recorded; do not expand it.
			pruned[strings.Join(path, "\x00")] = true
			return nil
		},
		trees...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to traverse trees: %w", err)
	}
	return diffTree, nil
}

// nodesEquivalent reports whether every given node is equivalent to the others.
// We only compare the key and value of the nodes, not their children; the
// children are compared separately while traversing the tree. Fewer than two
// nodes are trivially equivalent.
func nodesEquivalent(ns ...*nodes.Node) bool {
	for i := 1; i < len(ns); i++ {
		if !nodesEqual(ns[0], ns[i]) {
			return false
		}
	}
	return true
}

// nodesEqual reports whether two nodes are equivalent. A nil node means the path
// was absent from that tree and is only equivalent to another absent node.
func nodesEqual(n1, n2 *nodes.Node) bool {
	if n1 == nil && n2 == nil {
		// both absent from their trees: nothing differs between them.
		return true
	}
	if (n1 == nil) != (n2 == nil) {
		return false
	}
	n1IsLeaf := nodes.IsLeaf(n1)
	n2IsLeaf := nodes.IsLeaf(n2)
	if n1IsLeaf != n2IsLeaf {
		return false
	}
	if n1.Key != n2.Key {
		return false
	}
	if n1IsLeaf && n2IsLeaf && n1.Value != n2.Value {
		// only compare value for leafs
		return false
	}
	if nodes.IsLeafArray(n1) != nodes.IsLeafArray(n2) {
		return false
	}
	return true
}

// canDescend reports whether at least two trees still hold an equivalent
// non-leaf node at this path. When they do, the diff continues one level deeper
// to compare their children rather than dumping whole subtrees. Leaf nodes and
// paths held by fewer than two trees cannot be descended, so their difference is
// recorded in place.
func canDescend(ns []*nodes.Node) bool {
	present := make([]*nodes.Node, 0, len(ns))
	for _, n := range ns {
		if n != nil {
			present = append(present, n)
		}
	}
	if len(present) < 2 {
		return false
	}
	if !nodesEquivalent(present...) {
		return false
	}
	return !nodes.IsLeaf(present[0])
}

func addNode(tree *nodes.Node, path []string, n *nodes.Node) (*nodes.Node, error) {
	if n == nil {
		return tree, nil
	}

	if len(path) == 0 {
		// we will just append to root
		tree.Children.Put(n.Key, n)
		n.Parent = tree
		return tree, nil
	}

	// find or create root-level node matching path[0]
	current, remaining := nodes.GetNodeFromPath(tree, path)

	// If remaining is empty, current is the exact target node - attach n directly.
	if len(remaining) == 0 {
		current.Children.Put(n.Key, n)
		n.Parent = current
		return tree, nil
	}

	// create tree for remaining path
	m := make(map[string]any)
	mapPtr := m
	for _, part := range remaining {
		mapPtr[part] = map[string]any{}
		mapPtr = mapPtr[part].(map[string]any)
	}
	subtree := nodes.New(m, 0, nodes.LeafKeyAndValues)

	if current == nil {
		// no matching root node found - add subtree to root tree
		current = tree
	}
	// found a node partway through path - attach subtree to it and wire parents
	for key, child := range subtree.Children.Iter() {
		current.Children.Put(key, child)
		child.Parent = current
	}

	// get node from end of subtree and attach n
	node, remaining := nodes.GetNodeFromPath(subtree, remaining)
	if node == nil || len(remaining) != 0 {

		return nil, fmt.Errorf("failed to find node in subtree for path: %v", path)
	}

	node.Children.Put(n.Key, n)
	n.Parent = node
	return tree, nil
}

func updateRepr(tree *nodes.Node, f nodes.ReprNode) error {
	return nodes.DFS(
		tree,
		func(n *nodes.Node, _ int) error {
			if !nodes.IsLeaf(n) {
				n.Value = f(n)
			}
			return nil
		},
		nodes.WithNextNodes(nodes.AllChildren),
	)
}

// copyNode returns a shallow copy of n. This will corrupt the original tree,
// but this is fine for our use case and keeps the code efficient.
func copyNode(n *nodes.Node) *nodes.Node {
	if n == nil {
		return nil
	}
	c := &nodes.Node{
		ID:       uuid.New(),
		Key:      n.Key,
		Value:    n.Value,
		Expand:   n.Expand,
		Parent:   n.Parent,
		Children: n.Children,
	}
	for _, child := range c.Children.Iter() {
		child.Parent = c
	}
	return c
}

// addMeta
func addMeta(tree *nodes.Node, conf *diffConf, keys ...string) (*nodes.Node, error) {
	colors := conf.Colors
	if len(colors) < len(keys) {
		// not enough configured colors; fall back to the defaults and cycle
		// through them for any extra inputs.
		colors = defaultDiffColors
	}
	var err error
	for i, key := range keys {
		color := colors[i%len(colors)]
		tree, err = addNode(
			tree, []string{nodes.MetaKey},
			nodes.NewNode(key, color, 1, 0, nodes.LeafKeyAndValues),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to add meta for %s: %w", key, err)
		}
	}
	return tree, nil
}
