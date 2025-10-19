package nodes

func GetPathToNode(n *Node) string {
	b := ""
	for current := n; current != nil; current = current.Parent {
		// if we already have something in the path, we need a dot separator
		if len(b) > 0 {
			b = "." + b
		}
		b = current.Key + b
	}
	return b
}
