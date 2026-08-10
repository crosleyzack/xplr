package nodes

import (
	"strconv"

	"github.com/crosleyzack/xplr/pkg/omap"
	"github.com/google/uuid"
)

// Node is a node in the tree
type Node struct {
	ID uuid.UUID
	// Key is the value of the node in the tree
	Key string
	// Value is used to store the string representation of the value
	Value string
	// Children is the set of child nodes, keyed by node key and kept in
	// sorted key order.
	Children omap.OMap[string, *Node]
	// Parent of this node
	Parent *Node
	// Expand indicates if the node is expanded
	Expand bool
}

// Equal returns true if the two nodes are equal
func (n *Node) Equal(other *Node) bool {
	return n.ID == other.ID
}

// Child get child node with given key
func Child(n *Node, key string) *Node {
	if child, ok := n.Children.Get(key); ok {
		return child
	}
	return nil
}

// Ancestor for a given node by key identifier
func Ancestor(n *Node, key string) *Node {
	for {
		if n.Key == key {
			return n
		}
		if n.Parent == nil {
			return nil
		}
		n = n.Parent
	}
}

// Siblings get nodes with the same parent as this node
func Siblings(n *Node) []*Node {
	if n == nil || n.Parent == nil {
		return nil
	}

	return n.Parent.Children.Arr()
}

// IsLeaf returns true if the node is a leaf node (has no children)
func IsLeaf(n *Node) bool {
	return n.Children.Len() == 0
}

// IsRoot returns true if the node is the root node
func IsRoot(n *Node) bool {
	return n.Parent == nil
}

// ToMap converts a node back to a map[string]any, which can be used to convert back to JSON
func ToMap(nodes ...*Node) map[string]any {
	m := make(map[string]any)
	for _, n := range nodes {
		if IsLeaf(n) {
			m[n.Key] = n.Value
			continue
		}
		m[n.Key] = ToMap(n.Children.Arr()...)
	}
	return m
}

func addChild(node, child *Node) {
	child.Parent = node
	node.Children.Put(child.Key, child)
}

// New creates a new tree from a JSON object. The returned node is a sentinel
// root: it is not rendered, so the JSON's top-level entries are its children at
// display layer 0.
func New(json map[string]any, displayLayers uint, repr ReprNode) *Node {
	root := &Node{
		ID:       uuid.New(),
		Children: omap.New[string, *Node](),
		Expand:   true,
	}
	for k, v := range json {
		addChild(root, NewNode(k, v, 0, displayLayers, repr))
	}
	return root
}

// makeTree creates a node for a JSON object at the given display layer. Its
// children are one layer deeper.
func makeTree(json map[string]any, layer uint, displayLayers uint, repr ReprNode) *Node {
	tree := &Node{
		ID:       uuid.New(),
		Children: omap.New[string, *Node](),
		Expand:   layer < displayLayers,
	}
	for k, v := range json {
		node := NewNode(k, v, layer+1, displayLayers, repr)
		addChild(tree, node)
	}
	return tree
}

// IsArray checks if a node represents an array (all children have numeric keys)
func IsArray(n *Node) bool {
	if IsLeaf(n) {
		return false
	}
	for key := range n.Children.Keys() {
		if _, err := strconv.Atoi(key); err != nil {
			return false
		}
	}
	return true
}

// IsLeafArray checks if a node represents an array (all children have numeric keys)
// and all children are leaf nodes
func IsLeafArray(n *Node) bool {
	if IsLeaf(n) {
		return false
	}
	for key, child := range n.Children.Iter() {
		if _, err := strconv.Atoi(key); err != nil {
			return false
		}
		if !IsLeaf(child) {
			return false
		}
	}
	return true
}

// NewNode creates a new node from a key and value
func NewNode(key string, value any, layer uint, displayLayers uint, repr ReprNode) *Node {
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
		node.Children = omap.New[string, *Node]()
		for i, child := range v {
			n := NewNode(strconv.FormatUint(uint64(i), 10), child, layer+1, displayLayers, repr)
			addChild(node, n)
		}
		node.Value = "[]"
		if len(v) > 0 {
			node.Value = repr(node)
		}
	case map[string]any:
		// the map node itself sits at `layer`; makeTree places its children one
		// layer deeper.
		node = makeTree(v, layer, displayLayers, repr)
		node.Key = key
	}
	return node
}
