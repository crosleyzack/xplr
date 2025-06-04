package tree

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/crosleyzack/xplr/internal/nodes"
)

// View returns the string representation of the tree
func (m *Model) View() string {
	if m == nil || m.Nodes == nil {
		return "no data"
	}
	treeContent, err := m.renderTree()
	if err != nil {
		return fmt.Sprintf("An error occurred: %v", err)
	}
	return treeContent
}

// renderTree renders the json tree in the component
func (m *Model) renderTree() (string, error) {
	var b strings.Builder
	count := 0
	minRow, maxRow := m.getDisplayRange(m.NumberOfNodes())
	f := func(node *nodes.Node, layer int) error {
		var str string
		availableChars := m.Width
		// If we aren't at the root, we add the arrow shape to the string
		shape := m.LeafShape
		style := m.Styles.LeafShapes
		if len(node.Children) > 0 && !node.Expand {
			shape = m.ExpandShape
			style = m.Styles.ExpandShapes
		}
		if layer > 0 {
			spaces := (layer - 1) * m.SpacesPerLayer
			str += strings.Repeat(" ", spaces) + style.Render(shape) + " "
			// we need to track runes used to print correct length lines
			availableChars -= spaces + utf8.RuneCountInString(shape) + 1
		}
		// Generate the correct index for the node
		idx := count
		count++
		keyStr := replaceAll(node.Key, "\n\r", " ")
		valueStr := replaceAll(node.Value, "\n\r", " ")

		// Apply metadata display logic
		if len(node.Children) > 0 {
			// If node is expanded and has children, don't show the condensed value (only if override is set)
			if m.Styles.MergedObjectExpandOverride != "" && node.Expand {
				valueStr = m.Styles.MergedObjectExpandOverride
			} else if m.Styles.MergedObjectShowMetadata {
				valueStr = m.getMetadataStr(node)
			}
		}

		availableChars -= utf8.RuneCountInString(keyStr) + 8 // +8 from two tabs
		if utf8.RuneCountInString(valueStr) > availableChars {
			// if we have more runes than terminal width, truncate
			valueStr = valueStr[:availableChars-4] + "..."
		}
		// If we are at the cursor, we add the selected style to the string
		if m.cursor == idx {
			m.currentNode = node
			str += fmt.Sprintf("%s\t\t%s\n", m.Styles.Selected.Render(keyStr), m.Styles.Selected.Render(valueStr))
		} else if idx >= minRow && idx <= maxRow {
			str += fmt.Sprintf("%s\t\t%s\n", m.Styles.Unselected.Render(keyStr), m.Styles.Unselected.Render(valueStr))
		} else {
			// If we are not in the display range, we skip this node
			return nil
		}
		b.WriteString(str)
		return nil
	}
	if err := nodes.DFS(m.Nodes, f, nil); err != nil {
		return "", fmt.Errorf("Failed to render tree: %w", err)
	}
	return lipgloss.NewStyle().Height(m.Height).Width(m.Width).Render(b.String()), nil
}

// getDisplayRange returns the range of rows that should be displayed
func (m *Model) getDisplayRange(maxRows int) (int, int) {
	// ensure we show at most maxRows/m.Height rows
	rowsDisplayed := min(m.Height, maxRows)
	// rowsAbove + rowsBelow + 1 should be equal to rowsDisplayed
	rowsAbove := rowsDisplayed / 2
	rowsBelow := rowsDisplayed / 2
	if m.cursor < rowsAbove {
		// If there aren't enough rows above the cursor, we adjust the rows below
		rowsAbove = m.cursor
		rowsBelow = rowsDisplayed - m.cursor - 1
	}
	if m.cursor+rowsBelow > maxRows {
		// If there aren't enough rows below the cursor, we adjust the rows above
		rowsBelow = maxRows - m.cursor
		rowsAbove = rowsDisplayed - rowsBelow
	}
	return m.cursor - rowsAbove, m.cursor + rowsBelow
}

// replaceAll removes all occurrences of the characters in 'old' from the string 's'
func replaceAll(s, old, new string) string {
	if s == "" {
		return s
	}
	for _, char := range old {
		s = strings.ReplaceAll(s, string(char), new)
	}
	return s
}

// getJSONType determines the JSON type of a node's value
func getJSONType(node *nodes.Node) string {
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
func isArray(node *nodes.Node) bool {
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
func getArrayElementTypes(node *nodes.Node) string {
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

// getMetadataStr creates metadata strings for nodes with children
func (m *Model) getMetadataStr(node *nodes.Node) string {
	if len(node.Children) == 0 {
		return ""
	}
	var metadataContent string
	if m.Styles.MergedObjectShowKeyCount && m.Styles.MergedObjectShowKeyNamesWithTypes {
		// Show both count and key names with types
		if isArray(node) {
			arrayType := getArrayElementTypes(node)
			metadataContent = fmt.Sprintf("(%d items: %s)", len(node.Children), arrayType)
		} else {
			keyDetails := make([]string, 0, len(node.Children))
			for _, child := range node.Children {
				childType := getJSONType(child)
				keyDetails = append(keyDetails, fmt.Sprintf("%s:%s", child.Key, childType))
			}
			metadataContent = fmt.Sprintf("(%d keys: %s)", len(node.Children), strings.Join(keyDetails, ", "))
		}
	} else if m.Styles.MergedObjectShowKeyCount {
		// Show only count
		if isArray(node) {
			metadataContent = fmt.Sprintf("(%d items)", len(node.Children))
		} else {
			metadataContent = fmt.Sprintf("(%d keys)", len(node.Children))
		}
	} else if m.Styles.MergedObjectShowKeyNamesWithTypes {
		// Show only key names with types
		if isArray(node) {
			arrayType := getArrayElementTypes(node)
			metadataContent = fmt.Sprintf("(%s)", arrayType)
		} else {
			keyDetails := make([]string, 0, len(node.Children))
			for _, child := range node.Children {
				childType := getJSONType(child)
				keyDetails = append(keyDetails, fmt.Sprintf("%s:%s", child.Key, childType))
			}
			metadataContent = fmt.Sprintf("(%s)", strings.Join(keyDetails, ", "))
		}
	}
	if metadataContent != "" {
		return fmt.Sprintf("%s%s", m.Styles.MergedObjectMetadataPrefix, metadataContent)
	}
	return ""
}
