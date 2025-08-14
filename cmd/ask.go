package cmd

import (
    "fmt"
    "strings"

    "github.com/spf13/cobra"
    "github.com/wtg42/ora-ora-ora/config"
    "github.com/wtg42/ora-ora-ora/search"
    "github.com/wtg42/ora-ora-ora/storage"
)

// AskCmd 建立 `ask` 子指令，從檔案儲存重建索引並查詢片段。
func (o *OraCmd) AskCmd() *cobra.Command {
    var (
        topK int
        tags []string
    )
    cmd := &cobra.Command{
        Use:   "ask [query]",
        Short: "查詢筆記片段（暫不呼叫 LLM）",
        Long:  "自檔案儲存重建 in-memory 索引後，回傳 Top-K 片段，支援 tag 過濾。",
        Args:  cobra.MinimumNArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            q := strings.TrimSpace(strings.Join(args, " "))
            if q == "" {
                return fmt.Errorf("query is empty")
            }
            cfg := config.Default()
            fs, err := storage.NewFileStorage(cfg.Data.NotesDir)
            if err != nil {
                return err
            }
            notes, err := fs.List()
            if err != nil {
                return err
            }
            idx, err := search.OpenOrCreate(cfg.Data.IndexDir)
            if err != nil {
                return err
            }
            defer idx.Close()
            for _, n := range notes {
                if err := idx.IndexNote(n); err != nil {
                    return err
                }
            }
            snippets, err := idx.Query(q, topK, tags)
            if err != nil {
                return err
            }
            if len(snippets) == 0 {
                fmt.Fprintln(cmd.OutOrStdout(), "No results")
                return nil
            }
            for i, s := range snippets {
                fmt.Fprintf(cmd.OutOrStdout(), "%d. [%s] score=%.2f\n", i+1, s.NoteID, s.Score)
                fmt.Fprintln(cmd.OutOrStdout(), s.Excerpt)
                if len(s.TagMatches) > 0 {
                    fmt.Fprintln(cmd.OutOrStdout(), "tags:", strings.Join(s.TagMatches, ","))
                }
            }
            return nil
        },
    }
    cmd.Flags().IntVar(&topK, "topk", 20, "回傳片段數量")
    cmd.Flags().StringSliceVar(&tags, "tags", nil, "以標籤過濾（可多次指定）")
    return cmd
}

