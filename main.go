// Command xplr explores a tree data file with a TUI graphical interface.
//
// It takes in a tree data file (JSON, YAML, or TOML) either via the -f flag,
// as the first argument, or piped through stdin, and produces a navigable tree
// to view and explore the data. See [github.com/crosleyzack/xplr/cmds] for the
// command definitions and the main entry point.
package main

import (
	"fmt"
	"os"

	"github.com/crosleyzack/xplr/cmds"
)

func main() {
	if err := cmds.New().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
