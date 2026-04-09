package nodes

func GetPathToNode(n *Node) []string {
	path := []string{}
	for current := n; current != nil; current = current.Parent {
		path = append([]string{current.Key}, path...)
	}
	return path
}

func GetNodeFromPath(root *Node, path []string) *Node {
	current := root
	for _, part := range path {
		found := false
		for _, child := range current.Children {
			if child.Key == part {
				current = child
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}

	return current
}
