package main

import (
	"os"

	"github.com/wtg42/ora-ora-ora/cmd"
)

func main() {
	root := cmd.NewOraCmdRoot()
	if _, err := root.ExecuteC(); err != nil {
		os.Exit(1)
	}
}
