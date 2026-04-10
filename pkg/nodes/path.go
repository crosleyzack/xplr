package nodes

func GetPathToNode(n *Node) []string {
	path := []string{}
	for current := n; current != nil; current = current.Parent {
		path = append([]string{current.Key}, path...)
	}
	return path
}

// GetNodeFromPath walks down the tree from root following the path and returns
// the node at the end of the path and any remaining path if it doesn't exist
func GetNodeFromPath(root *Node, path []string) (*Node, []string) {
	if root == nil {
		return nil, path
	}
	if len(path) == 0 {
		return root, nil
	}
	current := root
	remaining := path
	for _, part := range path {
		found := false
		for _, child := range current.Children {
			if child.Key == part {
				current = child
				remaining = remaining[1:]
				found = true
				break
			}
		}
		if !found {
			return current, remaining
		}
	}

	return current, nil
}

// GetNodeFromTree walks down the tree from root following the path and returns
// the node at the end of the path and any remaining path if it doesn't exist
func GetNodeFromTree(root []*Node, path []string) (*Node, []string) {
	if len(path) == 0 {
		return nil, nil
	}
	if len(root) == 0 {
		return nil, path
	}
	for _, node := range root {
		if node.Key == path[0] {
			return GetNodeFromPath(node, path[1:])
		}
	}
	return nil, path
}
