package tree

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/crosleyzack/xplr/internal/nodes"
)

// siblingMaxKeyWidth calculates the maximum key width among siblings
func siblingMaxKeyWidth(node *nodes.Node, rootNodes []*nodes.Node) int {
	if node == nil {
		return 0
	}

	var siblings []*nodes.Node
	if node.Parent == nil {
		// For root level nodes, siblings are all root nodes
		siblings = rootNodes
	} else {
		// For other nodes, siblings are parent's children
		siblings = node.Parent.Children
	}

	maxWidth := 0
	for _, sibling := range siblings {
		keyWidth := utf8.RuneCountInString(sibling.Key)
		if keyWidth > maxWidth {
			maxWidth = keyWidth
		}
	}
	return maxWidth
}

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
		if len(node.Children) > 0 {
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

		// If node is expanded and has children, don't show the condensed value (only if override is set)
		if m.Styles.MergedObjectOverride != "" && node.Expand && len(node.Children) > 0 {
			valueStr = m.Styles.MergedObjectOverride
		}

		// Calculate spacing needed to align values within this level
		keyWidth := utf8.RuneCountInString(keyStr)
		// At least two spaces plus difference between key width and maximum.
		spacesNeeded := siblingMaxKeyWidth(node, m.Nodes) + m.SpacesPerLayer - keyWidth

		str += keyStr + strings.Repeat(" ", spacesNeeded)
		availableChars -= keyWidth + spacesNeeded

		if utf8.RuneCountInString(valueStr) > availableChars {
			// if we have more runes than terminal width, truncate
			valueStr = valueStr[:availableChars-4] + "..."
		}
		// If we are at the cursor, we add the selected style to the string
		if m.cursor == idx {
			m.currentNode = node
			str += fmt.Sprintf("%s\n", m.Styles.Selected.Render(valueStr))
		} else if idx >= minRow && idx <= maxRow {
			str += fmt.Sprintf("%s\n", m.Styles.Unselected.Render(valueStr))
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
