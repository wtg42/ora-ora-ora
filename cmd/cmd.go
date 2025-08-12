package cmd

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/wtg42/ora-ora-ora/tui"
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
			m := tui.NewAddNote()

			final, err := tea.NewProgram(m).Run()
			if err != nil {
				fmt.Println("failed to start TUI:", err)
				return
			}

			note, ok := final.(tui.AddNote)
			if !ok {
				fmt.Println("unexpected model result")
				return
			}

			// Output the result so callers can capture it.
			fmt.Println("Note:", note.Content)
			if len(note.Tags) > 0 {
				fmt.Println("Tags:", strings.Join(note.Tags, ","))
			}
		},
	}
}
