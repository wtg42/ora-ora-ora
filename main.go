package main

import (
	"os"

	"github.com/wtg42/ora-ora-ora/cmd"
)

func main() {
	oraCmd := cmd.NewOraCmd()
	oraCmd.RootCmd.AddCommand(oraCmd.StartTui())

	// if no args, default to start-tui
	if len(os.Args) == 1 {
		args := append(os.Args, "start-tui")
		oraCmd.RootCmd.SetArgs(args[1:])
	}

	if _, err := oraCmd.RootCmd.ExecuteC(); err != nil {
		os.Exit(1)
	}
}
