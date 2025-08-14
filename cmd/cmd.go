package cmd

import (
    "fmt"
    "strings"
    "time"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/spf13/cobra"
    "github.com/wtg42/ora-ora-ora/config"
    "github.com/wtg42/ora-ora-ora/search"
    "github.com/wtg42/ora-ora-ora/storage"
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

            // 儲存筆記到檔案並即時索引（in-memory）。
            id, err := storage.NewID()
            if err != nil {
                fmt.Println("failed to generate id:", err)
                return
            }
            now := time.Now()
            saved := storage.Note{
                ID:        id,
                Content:   note.Content,
                Tags:      note.Tags,
                CreatedAt: now,
                UpdatedAt: now,
            }

            cfg := config.Default()
            fs, err := storage.NewFileStorage(cfg.Data.NotesDir)
            if err != nil {
                fmt.Println("failed to open storage:", err)
                return
            }
            if err := fs.Save(saved); err != nil {
                fmt.Println("failed to save note:", err)
                return
            }

            // 即時索引（目前為 in-memory，主要用於後續流程一致）
            idx, err := search.OpenOrCreate(cfg.Data.IndexDir)
            if err == nil {
                _ = idx.IndexNote(saved)
                _ = idx.Close()
            }

            fmt.Println("Saved Note ID:", saved.ID)
            if len(saved.Tags) > 0 {
                fmt.Println("Tags:", strings.Join(saved.Tags, ","))
            }
        },
    }
}
