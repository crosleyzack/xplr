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

// GetCommonPath gets the longest common path between two paths
// IE [foo, bar] and [foo, baz] -> [foo]
func GetCommonPath(p1, p2 []string) []string {
	common := make([]string, 0, len(p1))
	size := min(len(p1), len(p2))
	for i := range size {
		if p1[i] == p2[i] {
			common = append(common, p1[i])
			continue
		}
		break
	}
	return common
}

// TrimPath removes path p2 from the end of path p1
func TrimPath(p1, p2 []string) []string {
	for {
		// if either is empty, we cannot trim more
		if len(p2) == 0 || len(p1) == 0 {
			break
		}
		last := len(p2) - 1
		tail := p2[last]
		p2 = p2[:last]
		// if the last element of p1 doesn't match
		// the element of p2, stop iteration
		if p1[len(p1)-1] != tail {
			break
		}
		// remove last item
		p1 = p1[:len(p1)-1]
	}
	return p1
}
