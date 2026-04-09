package cmds

import (
	"fmt"
	"slices"

	"github.com/crosleyzack/xplr/pkg/format"
	"github.com/crosleyzack/xplr/pkg/nodes"
	"github.com/crosleyzack/xplr/pkg/tui"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func NewDiffCmd() *cobra.Command {
	var files []string
	var output string
	var key1, key2 string
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
			diffTree, err := createDiffTree(tree1, tree2, &diffConf{
				KeyOne: key1,
				KeyTwo: key2,
			})
			if err != nil {
				return fmt.Errorf("failed to create diff tree: %w", err)
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
	return cmd
}

// nodesEquivalent compares two nodes and returns true if they are equal, false otherwise
// in this case, we only compare the key and value of the nodes, not their children.
// we will compare their children when we traverse the tree
func nodesEquivalent(n1, n2 *nodes.Node) bool {
	if (n1 == nil) != (n2 == nil) {
		return false
	}
	if nodes.IsLeaf(n1) != nodes.IsLeaf(n2) {
		return false
	}
	if n1.Key != n2.Key {
		return false
	}
	if n1.Value != n2.Value {
		return false
	}
	leafArray := nodes.IsLeafArray(n1)
	if leafArray != nodes.IsLeafArray(n2) {
		return false
	}
	if leafArray {
		// check if arrays match
		array1 := make([]string, 0, len(n1.Children))
		for _, child := range n1.Children {
			array1 = append(array1, child.Value)
		}
		array2 := make([]string, 0, len(n2.Children))
		for _, child := range n2.Children {
			array2 = append(array2, child.Value)
		}
		return slices.Compare(array1, array2) == 0
	}
	return true
}

type diffConf struct {
	KeyOne string
	KeyTwo string
}

func createDiffTree(tree1, tree2 []*nodes.Node, conf *diffConf) ([]*nodes.Node, error) {
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
					Expand: false,
				}
				path = path[:1]
			case len(remaining) > 0:
				// this path is not in tree2, calculate where they
				// diverge and get that node.
				path = nodes.TrimPath(path, remaining)
				// Since this is a subset of path1, remainder
				// will always be empty
				n, _ = nodes.GetNodeFromTree(tree1, path)
				//
				if n == nil {
					return fmt.Errorf("idk how this happens: %v", path)
				}
			}
			// if these nodes aren't equal, add to diff tree
			if len(remaining) != 0 || !nodesEquivalent(n, other) {
				// add to diff tree at path with value of n and other
				nCopy := *n
				nCopy.Key = conf.KeyOne
				diffTree, err = addNode(diffTree, path, &nCopy)
				if err != nil {
					return fmt.Errorf("failed to add node to diff tree: %w", err)
				}
				oCopy := *other
				oCopy.Key = conf.KeyTwo
				diffTree, err = addNode(diffTree, path, &oCopy)
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
	rootFound := current != nil

	// create tree for remaining path and n at the end
	m := make(map[string]any)
	mapPtr := m
	for _, part := range remaining {
		mapPtr[part] = map[string]any{}
		mapPtr = mapPtr[part].(map[string]any)
	}
	fmt.Printf("creating subtree for remaining path: %v, m: %v\n", remaining, m)
	subtree := nodes.New(m, 0, nodes.EmptyRepr)
	if len(subtree) != 1 {
		fmt.Printf("unexpected subtree length for path: %v, subtree: %v\n", remaining, subtree)
	}

	// attach new tree to end of the current tree
	for _, n := range subtree {
		if current != nil {
			n.Parent = current
			current.Children = append(current.Children, n)
		} else {
			//
			tree = append(tree, n)
			current = n
		}
	}

	// get leaf to attach value to; if root was absent, current is now remaining[0]'s node — skip it
	navPath := remaining
	if !rootFound && len(remaining) > 0 {
		navPath = remaining[1:]
	}
	node, remaining := nodes.GetNodeFromPath(current, navPath)
	if node == nil || len(remaining) != 0 {
		fmt.Printf("remaining %s does not match any node in subtree for path: %v, subtree: %v\n", remaining, path, subtree)
	}

	node.Children = append(node.Children, n)
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
