package cmds

import (
	"fmt"
	"io"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/crosleyzack/xplr/pkg/format"
	"github.com/crosleyzack/xplr/pkg/keys"
	"github.com/crosleyzack/xplr/pkg/modules/tree"
	"github.com/crosleyzack/xplr/pkg/nodes"
	"github.com/crosleyzack/xplr/pkg/styles"
	"github.com/crosleyzack/xplr/pkg/tui"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	var layers uint
	var nodeValueRepr string
	var file string
	cmd := &cobra.Command{
		Use:     "xplr [-x <layers>] [-f <file> | data]",
		Version: "0.2.5",
		Short:   "Explore a tree data file with a TUI graphical interface",
		Long:    "Takes in a tree data file (JSON, YAML, TOML) either via flag parameter, first argument, or stdin and produces TUI navigable tree to view and explore the data",
		Example: "xplr -x 2 -f foo.json",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			// get config
			c, err := tui.NewConfig()
			if err != nil {
				return fmt.Errorf("failed to parse config: %w", err)
			}
			// gather every operand from files, arguments and a piped stdin.
			inputs, err := gatherInputs(args, []string{file}, os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to get data: %w", err)
			}
			if len(inputs) != 1 {
				return fmt.Errorf("xplr needs exactly one input, got %d", len(inputs))
			}

			// get data as map[string]any
			m, err := format.Parse(inputs[0])
			if err != nil {
				return fmt.Errorf("failed to parse data: %w", err)
			}
			// parse into node tree
			n := nodes.New(m, layers, nodes.GetRepr(nodeValueRepr))
			// parse configs
			if err = renderTree(c, n); err != nil {
				return fmt.Errorf("failed to render tree: %w", err)
			}
			return nil
		},
	}
	cmd.Flags().UintVarP(&layers, "expand", "x", 0, "number of layers to expand by default")
	cmd.Flags().StringVarP(&file, "file", "f", "", "file to read data from")
	cmd.Flags().StringVar(&nodeValueRepr, "format", nodes.LeafValuesOnlyRepr, "Format to use to represent an expandable node value. Available formats: "+strings.Join(nodes.GetAvailableFormats(), "|"))
	cmd.AddCommand(NewDiffCmd())
	return cmd
}

// gatherInputs gathers operands in a stable order: one entry per file (in the
// order given), then one per positional argument treated as inline data, then
// stdin when it is piped and not empty. stdin is a parameter so callers can test
// it without touching the real os.Stdin.
func gatherInputs(args, files []string, stdin *os.File) ([][]byte, error) {
	var out [][]byte
	for _, f := range files {
		// an unset -f flag arrives as an empty string; skip it so stdin and
		// positional arguments can still supply the data.
		if f == "" {
			continue
		}
		b, err := os.ReadFile(f)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", f, err)
		}
		out = append(out, b)
	}
	for _, a := range args {
		out = append(out, []byte(a))
	}
	if piped(stdin) {
		b, err := io.ReadAll(stdin)
		if err != nil {
			return nil, fmt.Errorf("failed to read from pipe: %w", err)
		}
		if len(b) > 0 {
			out = append(out, b)
		}
	}
	return out, nil
}

// piped reports whether f is a pipe or redirect rather than an interactive
// terminal, meaning it carries data to read.
func piped(f *os.File) bool {
	st, err := f.Stat()
	return err == nil && st.Mode()&os.ModeCharDevice == 0
}

// renderTree takes in a config and a node tree and renders the TUI tree interface
func renderTree(conf *tui.Config, n *nodes.Node) error {
	keyMap := keys.NewKeyMap(&conf.KeyConfig)
	style := styles.NewStyle(&conf.StyleConfig)
	// populate KeyBasedStyles before creating the model so the copy it receives is complete
	if meta := nodes.Child(n, nodes.MetaKey); meta != nil {
		for key, child := range meta.Children.Iter() {
			style.KeyBasedStyles[key] = lipgloss.NewStyle().Background(lipgloss.Color(child.Value))
		}
	}
	format := tree.NewFormat(&conf.TreeConfig)
	model, err := tui.New(format, keyMap, style, n)
	if err != nil {
		return fmt.Errorf("failed to create TUI model: %w", err)
	}
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
