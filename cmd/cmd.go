package cmd

import (
    "crypto/rand"
    "encoding/hex"
    "errors"
    "fmt"
    "strings"
    "time"

    "github.com/spf13/cobra"

    "github.com/wtg42/ora-ora-ora/config"
    "github.com/wtg42/ora-ora-ora/model"
    "github.com/wtg42/ora-ora-ora/search"
    "github.com/wtg42/ora-ora-ora/storage"
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
            fmt.Println("Starting TUI...")
        },
    }
}

// Add returns the minimal "add" subcommand.
// Usage:
//   ora-ora-ora add "your note content" --tags dev,test --config path/to/config.yaml
func (o *OraCmd) Add() *cobra.Command {
    var (
        cfgPath string
        tagsStr string
    )

    cmd := &cobra.Command{
        Use:   "add [content]",
        Short: "新增一則筆記並更新索引",
        Args:  cobra.MinimumNArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            cfg, err := config.Load(strings.TrimSpace(cfgPath))
            if err != nil {
                return fmt.Errorf("load config: %w", err)
            }

            content := strings.TrimSpace(strings.Join(args, " "))
            if content == "" {
                return errors.New("content is empty")
            }

            note := model.Note{
                ID:        genUUIDv4(),
                Content:   content,
                Tags:      parseTags(tagsStr),
                CreatedAt: time.Now(),
                UpdatedAt: time.Now(),
            }

            // Save to JSONL storage
            st, err := storage.New(cfg.Data.NotesDir)
            if err != nil {
                return fmt.Errorf("init storage: %w", err)
            }
            if err := st.Save(note); err != nil {
                return fmt.Errorf("save note: %w", err)
            }

            // Update index (in-memory/OpenOrCreate semantics per current implementation)
            idx, err := search.OpenOrCreate(cfg.Data.IndexDir)
            if err != nil {
                return fmt.Errorf("open index: %w", err)
            }
            defer func() { _ = idx.Close() }()
            if err := idx.IndexNote(note); err != nil {
                return fmt.Errorf("index note: %w", err)
            }

            fmt.Fprintf(cmd.OutOrStdout(), "added: %s\n", note.ID)
            return nil
        },
    }

    cmd.Flags().StringVar(&cfgPath, "config", "", "YAML config path (optional)")
    cmd.Flags().StringVar(&tagsStr, "tags", "", "comma-separated tags (e.g., dev,test)")
    return cmd
}

// Ask returns the minimal "ask" subcommand which only performs retrieval and prints matched note IDs.
// Usage:
//   ora-ora-ora ask "your question" --topk 10 --tags dev,test --config path/to/config.yaml
func (o *OraCmd) Ask() *cobra.Command {
    var (
        cfgPath string
        tagsStr string
        topK    int
    )

    cmd := &cobra.Command{
        Use:   "ask [query]",
        Short: "檢索相關筆記（不呼叫 LLM）",
        Args:  cobra.MinimumNArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            cfg, err := config.Load(strings.TrimSpace(cfgPath))
            if err != nil {
                return fmt.Errorf("load config: %w", err)
            }

            query := strings.TrimSpace(strings.Join(args, " "))
            tags := parseTags(tagsStr)

            // Build index from storage (small data is fine for now)
            st, err := storage.New(cfg.Data.NotesDir)
            if err != nil {
                return fmt.Errorf("init storage: %w", err)
            }
            notes, err := st.List()
            if err != nil {
                return fmt.Errorf("list notes: %w", err)
            }

            idx, err := search.OpenOrCreate(cfg.Data.IndexDir)
            if err != nil {
                return fmt.Errorf("open index: %w", err)
            }
            defer func() { _ = idx.Close() }()
            for _, n := range notes {
                if err := idx.IndexNote(n); err != nil {
                    return fmt.Errorf("index note %s: %w", n.ID, err)
                }
            }

            // Perform retrieval and print IDs
            if topK <= 0 {
                topK = 20
            }
            res, err := idx.Query(query, topK, tags)
            if err != nil {
                return fmt.Errorf("query: %w", err)
            }
            for _, s := range res {
                fmt.Fprintln(cmd.OutOrStdout(), s.NoteID)
            }
            return nil
        },
    }

    cmd.Flags().StringVar(&cfgPath, "config", "", "YAML config path (optional)")
    cmd.Flags().StringVar(&tagsStr, "tags", "", "comma-separated tags (e.g., dev,test)")
    cmd.Flags().IntVar(&topK, "topk", 20, "max number of results")
    return cmd
}

// genUUIDv4 returns a random UUIDv4 string using crypto/rand (no external deps).
func genUUIDv4() string {
    var b [16]byte
    if _, err := rand.Read(b[:]); err != nil {
        // extremely unlikely; fallback to timestamp-based pseudo-unique
        ts := time.Now().UnixNano()
        return fmt.Sprintf("%x", ts)
    }
    // Set version (4) and variant (10)
    b[6] = (b[6] & 0x0f) | 0x40
    b[8] = (b[8] & 0x3f) | 0x80
    // Format 8-4-4-4-12
    hexs := make([]byte, 32)
    hex.Encode(hexs, b[:])
    return fmt.Sprintf("%s-%s-%s-%s-%s",
        hexs[0:8], hexs[8:12], hexs[12:16], hexs[16:20], hexs[20:32],
    )
}

// parseTags converts a comma-separated string into a cleaned slice of tags.
func parseTags(s string) []string {
    s = strings.TrimSpace(s)
    if s == "" {
        return nil
    }
    parts := strings.Split(s, ",")
    out := make([]string, 0, len(parts))
    for _, p := range parts {
        p = strings.TrimSpace(p)
        if p != "" {
            out = append(out, p)
        }
    }
    if len(out) == 0 {
        return nil
    }
    return out
}
