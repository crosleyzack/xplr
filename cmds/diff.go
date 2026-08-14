package cmds

import (
	"fmt"
	"os"
	"strings"

	"github.com/crosleyzack/xplr/pkg/format"
	"github.com/crosleyzack/xplr/pkg/nodes"
	"github.com/crosleyzack/xplr/pkg/tui"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// NewDiffCmd builds the `diff` command. It compares N trees drawn from any mix
// of files (-f), positional arguments, and a piped stdin, walking them together
// with DFSMulti so nodes present in only some trees still appear in the diff.
func NewDiffCmd() *cobra.Command {
	var files, keys []string
	var output string
	var nilValue string
	cmd := &cobra.Command{
		Use:     "diff []",
		Aliases: []string{"d"},
		Version: "0.3.1",
		Short:   "Diff two or more tree data files with a TUI graphical interface",
		Long:    "Takes in two or more tree data sources (JSON, YAML, TOML) via file flags, positional arguments, or a piped stdin and compares them.",
		Example: "xplr diff -f foo.json -f bar.json",
		Args:    cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			// get config
			c, err := tui.NewConfig()
			if err != nil {
				return fmt.Errorf("failed to parse config: %w", err)
			}
			// gather every operand from files, arguments and a piped stdin.
			inputs, err := gatherInputs(args, files, os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to get data: %w", err)
			}
			if len(inputs) < 2 {
				return fmt.Errorf("diff needs at least two inputs, got %d", len(inputs))
			}

			// keys default to _f1.._fN; when supplied, one is required per input.
			if len(keys) == 0 {
				keys = defaultKeys(len(inputs))
			}
			if len(keys) != len(inputs) {
				return fmt.Errorf("need one key per input: got %d keys for %d inputs", len(keys), len(inputs))
			}

			// parse each input into its own tree.
			trees := make([]*nodes.Node, len(inputs))
			for i, in := range inputs {
				m, err := format.Parse(in)
				if err != nil {
					return fmt.Errorf("failed to parse input %d: %w", i+1, err)
				}
				trees[i] = nodes.New(m, 0, nodes.EmptyRepr)
			}

			diffTree, err := createDiffTree(
				trees,
				WithKeys(keys...),
				WithNilValue(nilValue),
			)
			if err != nil {
				return fmt.Errorf("failed to create diff tree: %w", err)
			}
			// set the tree repr values
			if err := updateRepr(diffTree, nodes.LeafKeyAndValues); err != nil {
				return fmt.Errorf("failed to update tree repr: %w", err)
			}
			// add meta
			diffTree, err = addMeta(diffTree, c, keys...)
			if err != nil {
				return fmt.Errorf("failed to add tree meta: %w", err)
			}

			// TODO output diff tree to output
			switch output {
			case "json":
				err := printOutput(diffTree, format.AsJson)
				if err != nil {
					return fmt.Errorf("failed to print output: %w", err)
				}
			case "yaml":
				err := printOutput(diffTree, format.AsYaml)
				if err != nil {
					return fmt.Errorf("failed to print output: %w", err)
				}
			case "toml":
				err := printOutput(diffTree, format.AsToml)
				if err != nil {
					return fmt.Errorf("failed to print output: %w", err)
				}
			default:
				if err := renderTree(c, diffTree); err != nil {
					return fmt.Errorf("failed to render tree: %w", err)
				}
			}
			return nil
		},
	}
	cmd.Flags().StringSliceVarP(&files, "file", "f", nil, "files to read data from")
	cmd.Flags().StringVarP(&output, "out", "o", "", "what to output the diff to (defaults to tree display)")

	cmd.Flags().StringSliceVarP(&keys, "key", "k", nil, "key to label each input in the diff (one per input, defaults to _f1.._fN)")
	cmd.Flags().StringVar(&nilValue, "nilValue", "nil", "what to use as value for missing nodes in one tree")
	return cmd
}

// defaultKeys returns the default label for each of n inputs: _f1, _f2, ... _fn.
func defaultKeys(n int) []string {
	keys := make([]string, n)
	for i := range keys {
		keys[i] = fmt.Sprintf("_f%d", i+1)
	}
	return keys
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

type diffConf struct {
	// Keys labels each tree in the diff output; Keys[i] is used for trees[i].
	Keys     []string
	NilValue string
}

func defaultDiffConf() *diffConf {
	return &diffConf{
		Keys:     []string{"f1", "f2"},
		NilValue: "nil",
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

func createDiffTree(trees []*nodes.Node, opts ...DiffTreeOption) (*nodes.Node, error) {
	conf := defaultDiffConf()
	for _, opt := range opts {
		opt(conf)
	}
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

func printOutput(diffTree *nodes.Node, formatter func(map[string]any) ([]byte, error)) error {
	// diffTree is a sentinel root with an empty key; passing it to ToMap would
	// nest the whole output under a "" key. Map its children directly instead so
	// the top-level entries sit at the document root.
	b, err := formatter(nodes.ToMap(diffTree.Children.Arr()...))
	if err != nil {
		return fmt.Errorf("failed to convert diff tree to json: %w", err)
	}
	fmt.Println(string(b))
	return nil
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

func copyNode(n *nodes.Node) *nodes.Node {
	if n == nil {
		return nil
	}
	return &nodes.Node{
		ID:       uuid.New(),
		Key:      n.Key,
		Value:    n.Value,
		Expand:   n.Expand,
		Parent:   n.Parent,
		Children: n.Children,
	}
}

var defaultDiffColors = []string{"#ad0116", "#006222", "#38DB89", "#61707D", "#9D69A3"}

func addMeta(tree *nodes.Node, conf *tui.Config, keys ...string) (*nodes.Node, error) {
	colors := conf.DiffColors
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
