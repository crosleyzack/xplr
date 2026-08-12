package nodes

import (
	"bytes"
	"crypto/sha256"
	"fmt"
)

// SearchConfig configuration for performing a search on the tree
type searchConfig struct {
	NextNodes NextNodes
}

func defaultSearchConfig() *searchConfig {
	return &searchConfig{
		NextNodes: ObeyExpand,
	}
}

type DFSOption func(*searchConfig)
type NextNodes func(n *Node) []*Node

func WithNextNodes(f NextNodes) DFSOption {
	return func(c *searchConfig) {
		c.NextNodes = f
	}
}

// DFS perform depth first search on tree and run f on nodes
func DFS(node *Node, f func(*Node, int) error, opts ...DFSOption) error {
	conf := defaultSearchConfig()
	for _, opt := range opts {
		opt(conf)
	}
	if node == nil {
		return fmt.Errorf("received nil node")
	}
	start := []*Node{node}
	if IsRoot(node) {
		start = node.Children.Arr()
	}
	return dfs(start, f, conf, 0)
}

// dfs implementation of dfs
func dfs(nodes []*Node, f func(*Node, int) error, conf *searchConfig, layer int) error {
	for _, node := range nodes {
		if err := f(node, layer); err != nil {
			return err
		}
		next := conf.NextNodes(node)
		if len(next) > 0 {
			if err := dfs(next, f, conf, layer+1); err != nil {
				return err
			}
		}
	}
	return nil
}

// DFSIter a DFS implementation as an iterator for efficient searches
func DFSIter(node *Node, f func(*Node) bool, opts ...DFSOption) func(func(*Node) bool) {
	// get config
	conf := defaultSearchConfig()
	for _, opt := range opts {
		opt(conf)
	}
	stack := []*Node{node}
	if IsRoot(node) {
		stack = node.Children.Arr()
	}
	var n *Node
	return func(yield func(*Node) bool) {
		for len(stack) > 0 {
			n, stack = pop(stack)
			if next := conf.NextNodes(n); len(next) > 0 {
				// add to front of stack
				stack = append(stack, next...)
			}
			// if this node matches the function, yield it
			if f(n) {
				if !yield(n) {
					return
				}
			}
		}
	}
}

func AllChildren(n *Node) []*Node {
	return n.Children.Arr()
}

func ObeyExpand(n *Node) []*Node {
	if n.Expand {
		return n.Children.Arr()
	}
	return nil
}

// ChildNode encodes the result of retrieving node at the current path. Node is
// the node found at the path (the furthest ancestor reached when the path does
// not fully resolve, or nil for a nil tree); Rem is the path segments that
// could not be resolved and is empty when the node exists at the full path.
type ChildNode struct {
	Node *Node
	Rem  []string
}

// DFSMulti a DFS implementation that operates across multiple trees simultaneously
func DFSMulti(f func([]string, []ChildNode) error, trees ...*Node) error {
	stack := [][]string{{}}
	var path []string
	nodes := make([]ChildNode, len(trees))
	for {
		if len(stack) == 0 {
			return nil
		}
		path, stack = pop(stack)
		// get children at this path
		for i, tree := range trees {
			n, remainder := GetNodeFromPath(tree, path)
			nodes[i] = ChildNode{Node: n, Rem: remainder}
		}
		// process on nodes
		if err := f(path, nodes); err != nil {
			return err
		}
		// get the paths to children for each tree and add union to stack
		stack = append(stack, getNewPaths(nodes)...)

	}
}

func pop[T any](arr []T) (T, []T) {
	last := len(arr) - 1
	return arr[last], arr[:last]
}

func getNewPaths(nodes []ChildNode) [][]string {
	newPaths := make(map[[32]byte][]string)
	for _, node := range nodes {
		// we only care about nodes that exist at this path already. A nil
		// node (a nil tree, or a path GetNodeFromPath could not resolve)
		// contributes no children.
		if node.Node == nil || len(node.Rem) > 0 {
			continue
		}
		for _, child := range node.Node.Children.Iter() {
			path := GetPathToNode(child)
			newPaths[sha256.Sum256(pathToBytes(path))] = path
		}
	}
	ret := make([][]string, 0, len(newPaths))
	for _, path := range newPaths {
		ret = append(ret, path)
	}
	return ret
}

func pathToBytes(path []string) []byte {
	b := bytes.Buffer{}
	for _, loc := range path {
		b.Write([]byte(loc))
		// separator so ["a","b"] and ["ab"] do not hash to the same key
		b.WriteByte(0)
	}
	return b.Bytes()
}
