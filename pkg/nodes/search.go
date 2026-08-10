package nodes

import "fmt"

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
			n, stack = stack[0], stack[1:]
			if next := conf.NextNodes(n); len(next) > 0 {
				// add to front of stack
				stack = append(next, stack...)
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
