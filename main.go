package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wtg42/ora-ora-ora/cmd"
)

func main() {
	var err error
	var c *cobra.Command

	if c, err = cmd.ExecuteRootCommand(); err != nil {
		os.Exit(1)

	}

	fmt.Println(c.Name())
}
