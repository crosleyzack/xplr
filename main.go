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
