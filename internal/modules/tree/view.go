package tree

import (
	"fmt"
	"strings"

	"github.com/crosleyzack/xplr/internal/nodes"
)

const (
	bottomLeft string = " └─"
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
		// If we aren't at the root, we add the arrow shape to the string
		if layer > 0 {
			shape := strings.Repeat(" ", (layer-1)*2) + m.Styles.Shapes.Render(bottomLeft) + " "
			str += shape
		}
		// Generate the correct index for the node
		idx := count
		count++
		keyWidth := 10
		valueWidth := 20
		keyStr := strings.ReplaceAll(fmt.Sprintf("%-*s", keyWidth, node.Key), "\n", " ")
		valueStr := strings.ReplaceAll(fmt.Sprintf("%-*s", valueWidth, node.Value), "\n", " ")
		// If we are at the cursor, we add the selected style to the string
		if m.cursor == idx {
			m.currentNode = node
			str += fmt.Sprintf("%s\t\t%s\n", m.Styles.Selected.Render(keyStr), m.Styles.Selected.Render(valueStr))
		} else if idx >= minRow && idx <= maxRow {
			str += fmt.Sprintf("%s\t\t%s\n", m.Styles.Unselected.Render(keyStr), m.Styles.Unselected.Render(valueStr))
		} else {
			// nothing to do here
		}
		b.WriteString(str)
		return nil
	}
	if err := nodes.DFS(m.Nodes, f, nil); err != nil {
		return "", fmt.Errorf("Failed to render tree: %w", err)
	}
	return b.String(), nil
}

// getDisplayRange returns the range of rows that should be displayed
func (m *Model) getDisplayRange(maxRows int) (int, int) {
	rowsAbove := m.Height / 2
	rowsBelow := m.Height / 2
	if m.cursor < rowsAbove {
		rowsAbove = m.cursor
		rowsBelow = m.Height - m.cursor
	}
	if m.cursor+rowsBelow > maxRows {
		rowsBelow = maxRows - m.cursor
		rowsAbove = m.Height - rowsBelow
	}
	return m.cursor - rowsAbove, m.cursor + rowsBelow
}
