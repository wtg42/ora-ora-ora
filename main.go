package main

import (
	"os"

	"github.com/wtg42/ora-ora-ora/cmd"
)

func main() {
    oraCmd := cmd.NewOraCmd()
    oraCmd.RootCmd.AddCommand(oraCmd.StartTui())
    oraCmd.RootCmd.AddCommand(oraCmd.Add())
    oraCmd.RootCmd.AddCommand(oraCmd.Ask())

	if _, err := oraCmd.RootCmd.ExecuteC(); err != nil {
		os.Exit(1)
	}
}
