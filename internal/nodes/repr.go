package nodes

import (
	"fmt"
	"strings"
)

const (
	// terminal will never render this many characters
	MaxStringLength = 512
)

// ReprNode is a function that takes a Node and returns its string representation.
type ReprNode func(n *Node) string

// LeafValuesWithBrackets represents a node as the key mapped to sequence of children leaf values with brackets
func LeafValuesWithBrackets(n *Node) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s", n.Key))
	if len(n.Children) > 0 {
		b.WriteString(": {")
		first := true
		for _, child := range n.Children {
			b.WriteString(spacerToken(first))
			b.WriteString(LeafValuesWithBrackets(child))
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

// LeafValuesOnly represents a node as the key mapped to sequence of children leaf values
func LeafValuesOnly(n *Node) string {
	var b strings.Builder
	if len(n.Children) > 0 {
		first := true
		for _, child := range n.Children {
			b.WriteString(spacerToken(first))
			b.WriteString(LeafValuesOnly(child))
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

// spacerToken returns a space if not the first element
func spacerToken(first bool) string {
	if first {
		return ""
	}
	return " "
}
