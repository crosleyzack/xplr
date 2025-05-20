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

	treeContent, err := m.renderTree()
	if err != nil {
		return fmt.Sprintf("An error occurred: %v", err)
	}
	treeStyle := lipgloss.NewStyle().Height(availableHeight)

	sections = append(sections, treeStyle.Render(treeContent), help)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *Model) helpView() string {
	return m.Styles.Help.Render(m.Help.View(m))
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
	if err := nodes.DFS(m.Nodes, f, 0); err != nil {
		return "", fmt.Errorf("Failed to render tree: %w", err)
	}
	return b.String(), nil
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
