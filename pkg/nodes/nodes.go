package nodes

import (
	"strconv"

	"github.com/google/uuid"
)

// Node is a node in the tree
type Node struct {
	ID uuid.UUID
	// Key is the value of the node in the tree
	Key string
	// Value is used to store the string representation of the value
	Value string
	// Children is the list of children nodes
	Children []*Node
	// Parent of this node
	Parent *Node
	// Expand indicates if the node is expanded
	Expand bool
}

// Equal returns true if the two nodes are equal
func (n *Node) Equal(other *Node) bool {
	return n.ID == other.ID
}

// New creates a new tree from a JSON object
func New(json map[string]any, displayLayers uint, repr ReprNode) []*Node {
	return makeTree(json, 0, displayLayers, repr)
}

// makeTree creates a tree of nodes from a JSON object
func makeTree(json map[string]any, layer uint, displayLayers uint, repr ReprNode) []*Node {
	nodes := make([]*Node, 0, len(json))
	for k, v := range json {
		node := makeNode(k, v, layer, displayLayers, repr)
		nodes = append(nodes, node)
	}
	return nodes
}

// makeNode creates a new node from a key and value
func makeNode(key string, value any, layer uint, displayLayers uint, repr ReprNode) *Node {
	node := &Node{
		ID:     uuid.New(),
		Key:    key,
		Expand: layer < displayLayers,
	}
	switch v := value.(type) {
	case string:
		node.Value = v
	case int:
		node.Value = strconv.FormatInt(int64(v), 10)
	case float64:
		node.Value = strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		node.Value = strconv.FormatBool(v)
	case []any:
		node.Children = make([]*Node, 0, len(v))
		for i, child := range v {
			childNode := makeNode(strconv.FormatUint(uint64(i), 10), child, layer+1, displayLayers, repr)
			childNode.Parent = node
			node.Children = append(node.Children, childNode)
		}
		node.Value = "[]"
		if len(v) > 0 {
			node.Value = repr(node)
		}
	case map[string]any:
		node.Children = makeTree(v, layer+1, displayLayers, repr)
		for _, n := range node.Children {
			n.Parent = node
		}
		node.Value = "{}"
		if len(v) > 0 {
			node.Value = repr(node)
		}
	}
	return node
}
