package cmds

import (
	"fmt"

	"github.com/crosleyzack/xplr/pkg/format"
	"github.com/crosleyzack/xplr/pkg/nodes"
	"github.com/crosleyzack/xplr/pkg/tui"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// NOTE: current limitation of this is it will not show items in
// tree2 that are not in tree1, as it only does a single dfs over tree1
func NewDiffCmd() *cobra.Command {
	var files []string
	var output string
	var key1, key2 string
	var nilValue string
	cmd := &cobra.Command{
		Use:     "diff []",
		Aliases: []string{"d"},
		Version: "0.2.5",
		Short:   "Diff two tree data files with a TUI graphical interface",
		Long:    "Takes in two tree data files (JSON, YAML, TOML) either via flag parameter to compare the two files.",
		Example: "xplr diff -f foo.json -f bar.json",
		Args:    cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			// get config
			c, err := tui.NewConfig()
			if err != nil {
				return fmt.Errorf("failed to parse config: %w", err)
			}

			d1, err := getData(getSafe(args, 0), getSafe(files, 0))
			if err != nil {
				return fmt.Errorf("failed to get data: %w", err)
			}
			d2, err := getData(getSafe(args, 1), getSafe(files, 1))
			if err != nil {
				return fmt.Errorf("failed to get data: %w", err)
			}

			// get data as map[string]any
			m1, err := format.Parse(d1)
			if err != nil {
				return fmt.Errorf("failed to parse data: %w", err)
			}
			m2, err := format.Parse(d2)
			if err != nil {
				return fmt.Errorf("failed to parse data: %w", err)
			}

			tree1 := nodes.New(m1, 0, nodes.EmptyRepr)
			tree2 := nodes.New(m2, 0, nodes.EmptyRepr)
			diffTree, err := createDiffTree(
				tree1, tree2,
				WithKeyOne(key1),
				WithKeyTwo(key2),
			)
			if err != nil {
				return fmt.Errorf("failed to create diff tree: %w", err)
			}
			// set the tree repr values
			if err := updateRepr(diffTree, nodes.LeafKeyAndValues); err != nil {
				return fmt.Errorf("failed to update tree repr: %v", err)
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
	cmd.Flags().StringVar(&key1, "key1", "f1", "what to put as key in tree to denote this is from tree1")
	cmd.Flags().StringVar(&key2, "key2", "f2", "what to put as key in tree to denote this is from tree2")
	cmd.Flags().StringVar(&nilValue, "nilValue", "nil", "what to use as value for missing nodes in one tree")
	return cmd
}

// nodesEquivalent compares two nodes and returns true if they are equal, false otherwise
// in this case, we only compare the key and value of the nodes, not their children.
// we will compare their children when we traverse the tree
func nodesEquivalent(n1, n2 *nodes.Node) bool {
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

func createDiffTree(tree1, tree2 []*nodes.Node, opts ...DiffTreeOption) ([]*nodes.Node, error) {
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
			other, remaining := nodes.GetNodeFromTree(tree2, path)
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
				n, _ = nodes.GetNodeFromTree(tree1, path)
			case len(remaining) > 0:
				// this path is not in tree2, calculate where they
				// diverge and get that node. Where they diverge
				// is the first item in remaining
				path = nodes.TrimPath(path, remaining[1:])
				// set n to this divergent node
				n, _ = nodes.GetNodeFromTree(tree1, path)
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
				return n.Children
			}
			return nil
		}))
	if err != nil {
		return nil, fmt.Errorf("failed to traverse tree: %w", err)
	}
	// TODO we need to apply nodes.LeafKeyAndValue repr to tree
	return diffTree, nil
}

func addNode(tree []*nodes.Node, path []string, n *nodes.Node) ([]*nodes.Node, error) {
	if n == nil {
		return tree, nil
	}

	if len(path) == 0 {
		tree = append(tree, n)
		return tree, nil
	}

	// find or create root-level node matching path[0]
	current, remaining := nodes.GetNodeFromTree(tree, path)

	// create tree for remaining path and n at the end
	m := make(map[string]any)
	mapPtr := m
	for _, part := range remaining {
		mapPtr[part] = map[string]any{}
		mapPtr = mapPtr[part].(map[string]any)
	}
	subtree := nodes.New(m, 0, nodes.LeafKeyAndValues)

	// attach new tree to end of the current tree
	// if there is no tree as current is nil, add to root and set current
	if current == nil {
		current.Children = append(current.Children, subtree...)
	} else {
		tree = append(tree, subtree...)
		current = subtree[0]
	}

	// get node from end of subtree
	node, remaining := nodes.GetNodeFromTree(subtree, remaining)
	if node == nil || len(remaining) != 0 {
		fmt.Printf("remaining %s does not match any node in subtree for path: %v, subtree: %v\n", remaining, path, subtree)
	}

	node.Children = append(node.Children, n)
	n.Parent = node
	return tree, nil
}

func printOutput(diffTree []*nodes.Node, formatter func(map[string]any) ([]byte, error)) error {
	b, err := formatter(nodes.ToMap(diffTree))
	if err != nil {
		return fmt.Errorf("failed to convert diff tree to json: %w", err)
	}
	fmt.Println(string(b))
	return nil
}

func updateRepr(tree []*nodes.Node, f nodes.ReprNode) error {
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
