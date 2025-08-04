package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	// rootCmd 代表應用程式的根命令
	var rootCmd = &cobra.Command{
		Use:   "mycli",
		Short: "這是一個簡單的 CLI 示範工具",
		Long:  `這是一個使用 Cobra 建立的簡單 CLI 示範工具。`,
		Run: func(cmd *cobra.Command, args []string) {
			// 在此處撰寫你根命令的邏輯
			fmt.Println("Hello from my CLI!")
		},
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	os.Exit(1)
}

// 先寫一個根命令 然後再添加子命令
