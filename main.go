package main

import (
	"fmt"
	"os"

	"github.com/wtg42/ora-ora-ora/cmd"
)

func main() {
	oraCmd := cmd.NewOraCmd()
	oraCmd.RootCmd.AddCommand(oraCmd.StartTui())

	if c, err := oraCmd.RootCmd.ExecuteC(); err != nil {
		os.Exit(1)
	} else {
		fmt.Println(c.Name())
	}
}
