package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

type OraCmd struct {
	RootCmd *cobra.Command
}

// Init a new command
func NewOraCmd() *OraCmd {
	return &OraCmd{
		RootCmd: &cobra.Command{
			Use:   "ora-ora-ora",
			Short: "個人 AI 快速筆記工具",
			Long:  `將你的靈感、想法、計畫快速紀錄並由 AI 幫你回顧摘要跟重點。`,
			Run: func(cmd *cobra.Command, args []string) {
				// TODO: 啟動 TUI 介面
				fmt.Println("Starting program...")
			},
		},
	}
}

// StartTui creates and returns a Cobra command that starts the TUI interface.
// This allows the caller to add it to the RootCmd using AddCommand().
func (o *OraCmd) StartTui() *cobra.Command {
	return &cobra.Command{
		Use:   "start-tui",
		Short: "啟動 TUI 介面",
		Long:  "啟動文字使用者介面 TUI",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Starting TUI interface...")
		},
	}
}
