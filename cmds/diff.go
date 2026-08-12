package cmds

import (
	"fmt"
	"os"

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
		Version: "0.2.5",
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
			if len(inputs) != 2 {
				// TODO: relax once n diff implemented
				return fmt.Errorf("diff needs exactly two inputs, got %d", len(inputs))
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
				trees[0],
				trees[1],
				WithKeyOne(keys[0]),
				WithKeyTwo(keys[1]),
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

// nodesEquivalent compares two nodes and returns true if they are equal, false otherwise
// in this case, we only compare the key and value of the nodes, not their children.
// we will compare their children when we traverse the tree
func nodesEquivalent(n1, n2 *nodes.Node) bool {
	if (n1 == nil) != (n2 == nil) {
		return false
	}
	if n1 == nil && n2 == nil {
		return true
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
	KeyOne   string
	KeyTwo   string
	NilValue string
}

func defaultDiffConf() *diffConf {
	return &diffConf{
		KeyOne:   "f1",
		KeyTwo:   "f2",
		NilValue: "nil",
	}
}

type DiffTreeOption func(*diffConf)

func WithKeyOne(key string) func(*diffConf) {
	return func(c *diffConf) {
		c.KeyOne = key
	}
}

func WithKeyTwo(key string) func(*diffConf) {
	return func(c *diffConf) {
		c.KeyTwo = key
	}
}

func WithNilValue(val string) func(*diffConf) {
	return func(c *diffConf) {
		c.NilValue = val
	}
}

func createDiffTree(tree1, tree2 *nodes.Node, opts ...DiffTreeOption) (*nodes.Node, error) {
	conf := defaultDiffConf()
	for _, opt := range opts {
		opt(conf)
	}
	// use dfs to hit every node, stopping the downward recusion
	// when we find a difference
	diffTree := nodes.New(map[string]any{}, 0, nodes.EmptyRepr)
	shouldRecurse := true
	err := nodes.DFS(
		tree1,
		func(n *nodes.Node, _ int) (err error) {
			shouldRecurse = true
			// path will never be empty, as we are traversing tree 1
			// so we know it is in tree 1
			path := nodes.GetPathToNode(n)
			other, remaining := nodes.GetNodeFromPath(tree2, path)
			switch {
			case other == nil:
				// this path is not in tree2 diverging from the root,
				// add a new root node and set path to that key.
				other = &nodes.Node{
					ID:     uuid.New(),
					Key:    path[0],
					Value:  conf.NilValue,
					Expand: false,
				}
				path = path[:1]
				// get node from tree1 at this path
				n, _ = nodes.GetNodeFromPath(tree1, path)
			case len(remaining) > 0:
				// this path is not in tree2, calculate where they
				// diverge and get that node. Where they diverge
				// is the first item in remaining
				path = nodes.TrimPath(path, remaining[1:])
				// set n to this divergent node
				n, _ = nodes.GetNodeFromPath(tree1, path)
				if n == nil {
					return fmt.Errorf("this should never happen!")
				}
				// other is now a new node under the existing
				// node other with no value
				newLeaf := &nodes.Node{
					ID:     uuid.New(),
					Key:    n.Key,
					Value:  conf.NilValue,
					Parent: other,
				}
				other = newLeaf
			}
			// if these nodes aren't equal, add to diff tree
			if len(remaining) != 0 || !nodesEquivalent(n, other) {
				// add to diff tree at path with value of n and other
				nCopy := copyNode(n)
				nCopy.Key = conf.KeyOne
				diffTree, err = addNode(diffTree, path, nCopy)
				if err != nil {
					return fmt.Errorf("failed to add node to diff tree: %w", err)
				}
				oCopy := copyNode(other)
				oCopy.Key = conf.KeyTwo
				diffTree, err = addNode(diffTree, path, oCopy)
				if err != nil {
					return fmt.Errorf("failed to add node to diff tree: %w", err)
				}
				shouldRecurse = false
			}
			return nil
		},
		nodes.WithNextNodes(func(n *nodes.Node) []*nodes.Node {
			if shouldRecurse {
				return n.Children.Arr()
			}
			return nil
		}))
	if err != nil {
		return nil, fmt.Errorf("failed to traverse tree: %w", err)
	}
	// TODO we need to apply nodes.LeafKeyAndValue repr to tree
	return diffTree, nil
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
	b, err := formatter(nodes.ToMap(diffTree))
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

var defaultDiffColors = []string{"#ad0116", "#006222"}

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
