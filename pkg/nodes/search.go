package nodes

// SearchConfig configuration for performing a search on the tree
type SearchConfig struct {
	SearchAll bool
}

// GetSearchAll retrieve SearchAll accounting for nil config
func (c *SearchConfig) GetSearchAll() bool {
	return c != nil && c.SearchAll
}

// DFS perform depth first search on tree and run f on nodes
func DFS(nodes []*Node, f func(*Node, int) error, conf *SearchConfig) error {
	return dfs(nodes, f, conf, 0)
}

// dfs implementation of dfs
func dfs(nodes []*Node, f func(*Node, int) error, conf *SearchConfig, layer int) error {
	for _, node := range nodes {
		if err := f(node, layer); err != nil {
			return err
		}
		if node.Children != nil && (conf.GetSearchAll() || node.Expand) {
			if err := dfs(node.Children, f, conf, layer+1); err != nil {
				return err
			}
		}
	}
	return nil
}

// DFSIter a DFS implementation as an iterator for efficient searches
func DFSIter(nodes []*Node, f func(*Node) bool) func(func(*Node) bool) {
	stack := nodes
	var n *Node
	return func(yield func(*Node) bool) {
		for len(stack) > 0 {
			n, stack = stack[0], stack[1:]
			if len(n.Children) > 0 {
				// add to front of stack
				stack = append(n.Children, stack...)
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
