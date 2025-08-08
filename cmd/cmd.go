package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func ExecuteRootCommand() (*cobra.Command, error) {
	// rootCmd 代表應用程式的根命令
	var rootCmd = &cobra.Command{
		Use:   "ora-ora-ora",
		Short: "個人 AI 快速筆記工具",
		Long:  `將你的靈感、想法、計畫快速紀錄並由 AI 幫你回顧摘要跟重點。`,
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: 啟動 TUI 介面
			fmt.Println("Starting program...")

			// fmt.Printf("%v", args)
		},
	}

	return rootCmd.ExecuteC()
}
