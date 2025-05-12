package tree

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/crosleyzack/xplr/internal/nodes"
)

const (
	bottomLeft string = " └─"
)

func (m *Model) View() string {
	if m == nil || m.Nodes == nil {
		return "no data"
	}

	availableHeight := m.height
	if availableHeight <= 0 {
		availableHeight = 80 // Default height if not set
	}

	var sections []string

	var help string
	if m.showHelp {
		help = m.helpView()
		availableHeight -= 50
	}

	count := 0 // This is used to keep track of the index of the node we are on
	treeContent := m.renderTree(m.Nodes, 0, &count)
	treeStyle := lipgloss.NewStyle().Height(availableHeight)

	sections = append(sections, treeStyle.Render(treeContent), help)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *Model) helpView() string {
	return m.Styles.Help.Render(m.Help.View(m))
}

// renderTree renders the json tree in the component
func (m *Model) renderTree(remainingNodes []*nodes.Node, indent int, count *int) string {
	var b strings.Builder
	minRow, maxRow := m.getDisplayRange(m.NumberOfNodes())
	for _, node := range remainingNodes {
		var str string
		// If we aren't at the root, we add the arrow shape to the string
		if indent > 0 {
			shape := strings.Repeat(" ", (indent-1)*2) + m.Styles.Shapes.Render(bottomLeft) + " "
			str += shape
		}
		// Generate the correct index for the node
		idx := *count
		*count++
		// Format the string with fixed width for the value and description fields
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
		if node.Children != nil && node.Expand {
			childStr := m.renderTree(node.Children, indent+1, count)
			b.WriteString(childStr)
		}
	}
	return b.String()
}

// getDisplayRange returns the range of rows that should be displayed
func (m *Model) getDisplayRange(maxRows int) (int, int) {
	rowsAbove := m.height / 2
	rowsBelow := m.height / 2
	if m.cursor < rowsAbove {
		rowsAbove = m.cursor
		rowsBelow = m.height - m.cursor
	}
	if m.cursor+rowsBelow > maxRows {
		rowsBelow = maxRows - m.cursor
		rowsAbove = m.height - rowsBelow
	}
	return m.cursor - rowsAbove, m.cursor + rowsBelow
}
