package nodes

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	DisplayedLayers = 2
	MaxStringLength = 150
)

// Node is a node in the tree
type Node struct {
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
	if n.Key != other.Key {
		return false
	}
	if n.Value != other.Value {
		return false
	}
	if n.Expand != other.Expand {
		return false
	}
	if len(n.Children) != len(other.Children) {
		return false
	}
	for i, child := range n.Children {
		if !child.Equal(other.Children[i]) {
			return false
		}
	}
	return true
}

// String returns a string representation of the node
func (n *Node) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s", n.Key))
	if len(n.Children) > 0 {
		b.WriteString(": {")
		first := true
		for _, child := range n.Children {
			b.WriteString(spacerToken(first))
			b.WriteString(child.String())
			first = false
		}
		b.WriteString("}")
	} else if n.Value != "" {
		b.WriteString(fmt.Sprintf(": %s", n.Value))
	}
	s := b.String()
	if len(s) > MaxStringLength {
		return s[:MaxStringLength] + "..."
	}
	return s
}

// ShortString returns a short string representation of the node, only shows leaf nodes
func (n *Node) ShortString() string {
	var b strings.Builder
	if n.Children != nil {
		first := true
		for _, child := range n.Children {
			b.WriteString(spacerToken(first))
			b.WriteString(child.ShortString())
			first = false
		}
	} else {
		b.WriteString(n.Value)
	}
	s := b.String()
	if len(s) > MaxStringLength {
		return s[:MaxStringLength] + "..."
	}
	return s
}

// New creates a new tree from a JSON object
func New(json map[string]any, displayLayers uint) []*Node {
	return makeTree(json, 0, displayLayers)
}

func makeTree(json map[string]any, layer uint, displayLayers uint) []*Node {
	nodes := make([]*Node, 0, len(json))
	for k, v := range json {
		node := makeNode(k, v, layer, displayLayers)
		nodes = append(nodes, node)
	}
	return nodes
}

// makeNode creates a new node from a key and value
func makeNode(key string, value any, layer uint, displayLayers uint) *Node {
	node := &Node{
		Key:    key,
		Expand: layer < displayLayers,
	}
	switch value.(type) {
	case string:
		node.Value = value.(string)
	case int:
		node.Value = strconv.FormatInt(int64(value.(int)), 10)
	case float64:
		node.Value = strconv.FormatFloat(value.(float64), 'f', -1, 64)
	case bool:
		node.Value = strconv.FormatBool(value.(bool))
	case []any:
		node.Children = make([]*Node, 0, len(value.([]any)))
		for i, child := range value.([]any) {
			childNode := makeNode(strconv.FormatUint(uint64(i), 10), child, layer+1, displayLayers)
			childNode.Parent = node
			node.Children = append(node.Children, childNode)
		}
		node.Value = node.ShortString()
	case map[string]any:
		node.Children = makeTree(value.(map[string]any), layer+1, displayLayers)
		for _, n := range node.Children {
			n.Parent = node
		}
		node.Value = node.ShortString()
	}
	return node
}

// spacerToken returns a space if not the first element
func spacerToken(first bool) string {
	if first {
		return ""
	}
	return " "
}
