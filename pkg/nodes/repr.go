package nodes

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	// terminal will never render this many characters
	MaxStringLength = 512

	// Representation type constants with examples

	// LeafValuesWithBracketsRepr: "{childKey1: childValue1, childKey2: {grandchildKey1: grandchildValue1}}"
	LeafValuesWithBracketsRepr = "full"
	// LeafValuesOnlyRepr: "childValue1 childValue2 grandchildValue1 grandchildValue2"
	LeafValuesOnlyRepr = "values"
	// DirectChildrenKeysRepr: "childKey1 childKey2 childKey3"
	DirectChildrenKeysRepr = "keys"
	// KeyCountOnlyRepr: "(3 keys)" or "(3 items)"
	KeyCountOnlyRepr = "key-count"
	// KeyNamesWithTypesRepr: "(array of objects)" or (childKey1:object, childKey2:integer, childKey3:string)"
	KeyNamesWithTypesRepr = "key-names"
	// KeyCountAndTypesRepr: "(3 items: array of objects)" or "(2 keys: childKey1:object, childKey2:integer)"
	KeyCountAndTypesRepr = "key-names-with-count"
	// Default metadata prefix for representations that do not show values
	MetadataPrefix = "ⓘ "
)

// ReprNode is a function that takes a Node and returns its string representation.
type ReprNode func(n *Node) string

// GetRepr returns a ReprNode based on the provided string representation type.
func GetRepr(repr string) ReprNode {
	switch repr {
	case LeafValuesWithBracketsRepr:
		return LeafValuesWithBrackets
	case LeafValuesOnlyRepr:
		return LeafValuesOnly
	case DirectChildrenKeysRepr:
		return DirectChildrenKeys
	case KeyCountOnlyRepr:
		return KeyCountOnly
	case KeyNamesWithTypesRepr:
		return KeyNamesWithTypes
	case KeyCountAndTypesRepr:
		return KeyCountAndTypes
	default:
		return LeafValuesOnly // default representation
	}
}

// GetAvailableFormats returns a list of available formats
func GetAvailableFormats() []string {
	return []string{
		LeafValuesWithBracketsRepr,
		LeafValuesOnlyRepr,
		DirectChildrenKeysRepr,
		KeyCountOnlyRepr,
		KeyNamesWithTypesRepr,
		KeyCountAndTypesRepr,
	}
}

// LeafValuesWithBrackets represents a node as the key mapped to sequence of children leaf values with brackets
// Example: "parent: {child1: value1 child2: value2}"
func LeafValuesWithBrackets(n *Node) string {
	return leafValuesWithBracketsHelper(n, false)
}

func leafValuesWithBracketsHelper(n *Node, includeKey bool) string {
	var b strings.Builder
	if includeKey {
		b.WriteString(n.Key)
	}
	if len(n.Children) > 0 {
		if includeKey {
			b.WriteString(": ")
		}
		b.WriteString("{")
		first := true
		for _, child := range n.Children {
			b.WriteString(spacerToken(first))
			b.WriteString(leafValuesWithBracketsHelper(child, true))
			first = false
		}
		b.WriteString("}")
	} else if n.Value != "" {
		if includeKey {
			b.WriteString(fmt.Sprintf(": %s", n.Value))
		} else {
			b.WriteString(n.Value)
		}
	}
	return truncateIfNeeded(b.String())
}

// LeafValuesOnly represents a node as the sequence of children leaf values only
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
	return truncateIfNeeded(b.String())
}

// DirectChildrenKeys represents a node as the keys of its direct children
func DirectChildrenKeys(n *Node) string {
	var b strings.Builder
	if len(n.Children) > 0 {
		first := true
		for _, child := range n.Children {
			b.WriteString(spacerToken(first))
			b.WriteString(child.Key)
			first = false
		}
	} else {
		b.WriteString("{}")
	}
	return truncateIfNeeded(b.String())
}

// KeyCountOnly represents a node showing only the count of keys/items
func KeyCountOnly(n *Node) string {
	var b strings.Builder

	if len(n.Children) > 0 {
		var metadata string
		if isArray(n) {
			metadata = fmt.Sprintf("(%d items)", len(n.Children))
		} else {
			metadata = fmt.Sprintf("(%d keys)", len(n.Children))
		}

		b.WriteString(MetadataPrefix)
		b.WriteString(metadata)
	}

	return truncateIfNeeded(b.String())
}

// KeyNamesWithTypes represents a node showing key names with their types
func KeyNamesWithTypes(n *Node) string {
	var b strings.Builder
	if len(n.Children) > 0 {
		var metadata string
		if isArray(n) {
			arrayType := getArrayElementTypes(n)
			metadata = fmt.Sprintf("(%s)", arrayType)
		} else {
			keyDetails := make([]string, 0, len(n.Children))
			for _, child := range n.Children {
				childType := getJSONType(child)
				keyDetails = append(keyDetails, fmt.Sprintf("%s:%s", child.Key, childType))
			}
			metadata = fmt.Sprintf("(%s)", strings.Join(keyDetails, ", "))
		}

		b.WriteString(MetadataPrefix)
		b.WriteString(metadata)
	}

	return truncateIfNeeded(b.String())
}

// KeyCountAndTypes represents a node showing both count and key names with types
func KeyCountAndTypes(n *Node) string {
	var b strings.Builder
	if len(n.Children) > 0 {
		var metadata string
		if isArray(n) {
			arrayType := getArrayElementTypes(n)
			metadata = fmt.Sprintf("(%d items: %s)", len(n.Children), arrayType)
		} else {
			keyDetails := make([]string, 0, len(n.Children))
			for _, child := range n.Children {
				childType := getJSONType(child)
				keyDetails = append(keyDetails, fmt.Sprintf("%s:%s", child.Key, childType))
			}
			metadata = fmt.Sprintf("(%d keys: %s)", len(n.Children), strings.Join(keyDetails, ", "))
		}

		b.WriteString(MetadataPrefix)
		b.WriteString(metadata)
	}

	return truncateIfNeeded(b.String())
}

// getJSONType determines the JSON type of a node's value
func getJSONType(node *Node) string {
	if len(node.Children) > 0 {
		if isArray(node) {
			return "array"
		}
		return "object"
	}

	value := node.Value
	if value == "" {
		return "null"
	}
	if value == "true" || value == "false" {
		return "boolean"
	}
	if _, err := strconv.ParseInt(value, 10, 64); err == nil {
		return "integer"
	}
	if _, err := strconv.ParseFloat(value, 64); err == nil {
		return "number"
	}
	return "string"
}

// isArray checks if a node represents an array (all children have numeric keys)
func isArray(node *Node) bool {
	if len(node.Children) == 0 {
		return false
	}
	for _, child := range node.Children {
		if _, err := strconv.Atoi(child.Key); err != nil {
			return false
		}
	}
	return true
}

// getArrayElementTypes analyzes array elements and returns a descriptive type string
func getArrayElementTypes(node *Node) string {
	if !isArray(node) || len(node.Children) == 0 {
		return "array"
	}

	// Count types of all elements
	typeCounts := make(map[string]int)
	for _, child := range node.Children {
		childType := getJSONType(child)
		typeCounts[childType]++
	}

	// If all elements are the same type
	if len(typeCounts) == 1 {
		for elemType := range typeCounts {
			if elemType == "integer" || elemType == "number" {
				return "array of numbers"
			}
			return fmt.Sprintf("array of %ss", elemType)
		}
	}

	// If mixed types, just return "array"
	return "array"
}

// truncateIfNeeded truncates string if it exceeds MaxStringLength
func truncateIfNeeded(s string) string {
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
