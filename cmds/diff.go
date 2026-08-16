package cmds

import (
	"fmt"
	"os"

	"github.com/crosleyzack/wndr/pkg/diff"
	"github.com/crosleyzack/wndr/pkg/format"
	"github.com/crosleyzack/wndr/pkg/nodes"
	"github.com/crosleyzack/wndr/pkg/tui"
	"github.com/spf13/cobra"
)

// NewDiffCmd builds the `diff` command. It compares N trees drawn from any mix
// of files (-f), positional arguments, and a piped stdin, walking them together
// with DFSMulti so nodes present in only some trees still appear in the diff.
func NewDiffCmd() *cobra.Command {
	var files, keys []string
	var output string
	var nilValue string
	cmd := &cobra.Command{
		Use:     "diff [-f <file>]... [data]...",
		Aliases: []string{"d"},
		Version: version,
		Short:   "Diff two or more tree data files with a TUI graphical interface",
		Long:    "Takes in two or more tree data sources (JSON, YAML, TOML) via file flags, positional arguments, or a piped stdin and compares them.",
		Example: "wndr diff -f foo.json -f bar.json",
		Args:    cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			// get config
			c, err := tui.NewConfig()
			if err != nil {
				return fmt.Errorf("failed to parse config: %w", err)
			}
			// gather every operand from files, arguments and a piped stdin.
			inputs, err := gatherInputs(args, files, os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to get data: %w", err)
			}
			if len(inputs) < 2 {
				return fmt.Errorf("diff needs at least two inputs, got %d", len(inputs))
			}

			// keys default to _f1.._fN; when supplied, one is required per input.
			if len(keys) == 0 {
				keys = defaultKeys(len(inputs))
			}
			if len(keys) != len(inputs) {
				return fmt.Errorf("need one key per input: got %d keys for %d inputs", len(keys), len(inputs))
			}

			// parse each input into its own tree.
			trees := make([]*nodes.Node, len(inputs))
			for i, in := range inputs {
				m, err := format.Parse(in)
				if err != nil {
					return fmt.Errorf("failed to parse input %d: %w", i+1, err)
				}
				trees[i] = nodes.New(m, 0, nodes.EmptyRepr)
			}

			diffTree, err := diff.Diff(
				trees,
				diff.WithKeys(keys...),
				diff.WithNilValue(nilValue),
			)
			if err != nil {
				return fmt.Errorf("failed to create diff tree: %w", err)
			}

			// output the diff tree in the requested format, or render the TUI.
			switch output {
			case "json":
				err := printOutput(diffTree, format.AsJson)
				if err != nil {
					return fmt.Errorf("failed to print output: %w", err)
				}
			case "yaml":
				err := printOutput(diffTree, format.AsYaml)
				if err != nil {
					return fmt.Errorf("failed to print output: %w", err)
				}
			case "toml":
				err := printOutput(diffTree, format.AsToml)
				if err != nil {
					return fmt.Errorf("failed to print output: %w", err)
				}
			default:
				if err := renderTree(c, diffTree); err != nil {
					return fmt.Errorf("failed to render tree: %w", err)
				}
			}
			return nil
		},
	}
	cmd.Flags().StringSliceVarP(&files, "file", "f", nil, "files to read data from")
	cmd.Flags().StringVarP(&output, "out", "o", "", "what to output the diff to (defaults to tree display)")

	cmd.Flags().StringSliceVarP(&keys, "key", "k", nil, "key to label each input in the diff (one per input, defaults to _f1.._fN)")
	cmd.Flags().StringVar(&nilValue, "nilValue", "nil", "what to use as value for missing nodes in one tree")
	return cmd
}

// defaultKeys returns the default label for each of n inputs: _f1, _f2, ... _fn.
func defaultKeys(n int) []string {
	keys := make([]string, n)
	for i := range keys {
		keys[i] = fmt.Sprintf("_f%d", i+1)
	}
	return keys
}

func printOutput(diffTree *nodes.Node, formatter func(map[string]any) ([]byte, error)) error {
	// diffTree is a sentinel root with an empty key; passing it to ToMap would
	// nest the whole output under a "" key. Map its children directly instead so
	// the top-level entries sit at the document root.
	b, err := formatter(nodes.ToMap(diffTree.Children.Arr()...))
	if err != nil {
		return fmt.Errorf("failed to convert diff tree to json: %w", err)
	}
	fmt.Println(string(b))
	return nil
}
