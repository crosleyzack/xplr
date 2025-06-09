package tree

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/crosleyzack/xplr/pkg/nodes"
)

const (
	spacesAfterKey = 2 // Number of spaces after the key in the tree view
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
		if keyWidth := utf8.RuneCountInString(sibling.Key); keyWidth > maxWidth {
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
		// Generate the correct index for the node
		idx := count
		count++
		if idx < minRow || idx > maxRow {
			return nil // Skip nodes outside the display range
		}
		str := m.getLine(node, layer, idx)
		if len(str) > 0 {
			b.WriteString(str)
		}
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

// getLineShapeStyle returns the shape and style for a node based on its state
func (m *Model) getLineShapeStyle(node *nodes.Node) (string, lipgloss.Style) {
	if len(node.Children) == 0 {
		return m.LeafShape, m.Styles.LeafStyle
	} else if node.Expand {
		return m.ExpandedShape, m.Styles.ExpandedStyle
	} else {
		return m.ExpandableShape, m.Styles.ExpandableStyle
	}
}

// getLine generates a line for the tree corresponding to this node
func (m *Model) getLine(node *nodes.Node, layer int, index int) string {
	var str string
	availableChars := m.Width
	shape, style := m.getLineShapeStyle(node)
	if layer > 0 {
		spaces := (layer - 1) * m.SpacesPerLayer
		str += strings.Repeat(" ", spaces) + style.Render(shape) + " "
		// we need to track runes used to print correct length lines
		availableChars -= spaces + utf8.RuneCountInString(shape) + 1
	}
	// Generate the correct index for the node
	keyStr := replaceAll(node.Key, "\n\r", " ")
	valueStr := replaceAll(node.Value, "\n\r", " ")
	keyWidth := utf8.RuneCountInString(keyStr)
	spacesNeeded := siblingMaxKeyWidth(node, m.Nodes) + spacesAfterKey - keyWidth

	availableChars -= keyWidth + spacesNeeded
	if utf8.RuneCountInString(valueStr) > availableChars {
		// if we have more runes than terminal width, truncate
		valueStr = valueStr[:availableChars-4] + "..."
	}
	// If we are at the cursor, we add the selected style to the string
	styledKey := ""
	styledValue := ""
	if m.cursor == index {
		m.currentNode = node
		styledKey = m.Styles.Selected.Render(keyStr)
		styledValue = m.Styles.Selected.Render(valueStr)
	} else {
		styledKey = m.Styles.Unselected.Render(keyStr)
		styledValue = m.Styles.Unselected.Render(valueStr)
	}
	str += styledKey
	if !node.Expand || !m.hideSummaryWhenExpanded {
		str += strings.Repeat(" ", spacesNeeded) + styledValue
	}
	str += "\n"
	return str
}
