package cmds

import (
	"fmt"
	"io"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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
			// get data
			data := []byte{}
			if len(args) > 0 && args[0] != "" {
				data = []byte(args[0])
			} else if len(file) > 0 {
				f, err := os.Open(file)
				if err != nil {
					return fmt.Errorf("failed to open data file: %w", err)
				}
				data, err = io.ReadAll(f)
				if err != nil {
					return fmt.Errorf("failed to read file: %w", err)
				}
			} else {
				data, err = io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("failed to read from pipe: %w", err)
				}
			}
			if len(data) == 0 {
				return fmt.Errorf("no data")
			}
			// get data as map[string]any
			var m map[string]any
			for _, fmt := range []format.Format{format.ParseJson, format.ParseYaml, format.ParseToml} {
				m, err = fmt(data)
				if err == nil {
					break
				}
			}
			if len(m) == 0 {
				return fmt.Errorf("no data")
			}
			// parse into node tree
			n := nodes.New(m, layers, nodes.GetRepr(nodeValueRepr))
			// parse configs
			keyMap := keys.NewKeyMap(&c.KeyConfig)
			style := styles.NewStyle(&c.StyleConfig)
			format := tree.NewFormat(&c.TreeConfig)
			model, err := tui.New(format, keyMap, style, n)
			if err != nil {
				return fmt.Errorf("failed to create TUI model: %w", err)
			}
			p := tea.NewProgram(model)
			if _, err := p.Run(); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.PersistentFlags().UintVarP(&layers, "expand", "x", 0, "number of layers to expand by default")
	cmd.PersistentFlags().StringVarP(&file, "file", "f", "", "file to read data from")
	cmd.PersistentFlags().StringVar(&nodeValueRepr, "format", nodes.LeafValuesOnlyRepr, "Format to use to represent an expandable node value. Available formats: "+strings.Join(nodes.GetAvailableFormats(), "|"))
	return cmd
}
