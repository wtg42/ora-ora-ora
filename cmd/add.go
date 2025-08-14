package cmd

import (
    "fmt"
    "strings"
    "time"

    "github.com/spf13/cobra"
    "github.com/wtg42/ora-ora-ora/config"
    "github.com/wtg42/ora-ora-ora/storage"
)

// AddCmd 建立 `add` 子指令，用於寫入一筆筆記至 JSONL。
func (o *OraCmd) AddCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "add [content]",
        Short: "新增一筆筆記（JSONL）",
        Long:  "新增一筆筆記，支援 #tag 自動解析與 JSONL 儲存格式。",
        Args:  cobra.MinimumNArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            content := strings.TrimSpace(strings.Join(args, " "))
            if content == "" {
                return fmt.Errorf("content is empty")
            }

            id, err := storage.NewID()
            if err != nil {
                return fmt.Errorf("generate id: %w", err)
            }
            tags := storage.ExtractTags(content)

            now := time.Now()
            note := storage.Note{
                ID:        id,
                Content:   content,
                Tags:      tags,
                CreatedAt: now,
                UpdatedAt: now,
            }

            cfg := config.Default()
            fs, err := storage.NewFileStorage(cfg.Data.NotesDir)
            if err != nil {
                return err
            }
            if err := fs.Save(note); err != nil {
                return err
            }
            fmt.Fprintln(cmd.OutOrStdout(), "Saved:")
            fmt.Fprintln(cmd.OutOrStdout(), "  ID:", note.ID)
            if len(note.Tags) > 0 {
                fmt.Fprintln(cmd.OutOrStdout(), "  Tags:", strings.Join(note.Tags, ","))
            }
            return nil
        },
    }
    return cmd
}

