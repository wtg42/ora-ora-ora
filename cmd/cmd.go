package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/wtg42/ora-ora-ora/cmd/core"
	"github.com/wtg42/ora-ora-ora/config"
	"github.com/wtg42/ora-ora-ora/model"
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
				// By default, run start-tui
				startTuiCmd := NewOraCmd().StartTui()
				startTuiCmd.Run(cmd, args)
			},
		},
	}
}

// StartTui creates and returns a Cobra command that starts the TUI interface.
func (o *OraCmd) StartTui() *cobra.Command {
	var cfgPath string
	cmd := &cobra.Command{
		Use:   "start-tui",
		Short: "啟動 TUI 介面",
		Long:  "啟動文字使用者介面 TUI",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(strings.TrimSpace(cfgPath))
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			st, err := storage.New(cfg.Data.NotesDir)
			if err != nil {
				return fmt.Errorf("init storage: %w", err)
			}

			idx, err := search.OpenOrCreate(cfg.Data.IndexDir)
			if err != nil {
				return fmt.Errorf("open index: %w", err)
			}
			defer func() { _ = idx.Close() }()

			addNoteModel := tui.NewAddNoteModel(st, idx)
			p := tea.NewProgram(addNoteModel)

			if _, err := p.Run(); err != nil {
				return fmt.Errorf("TUI error: %w", err)
			}

			// The FinalNote() method in the original test was for extracting data.
			// The updated AddNoteModel in `add_note.go` doesn't have this method anymore.
			// Instead, the saving logic is handled within the Update loop via a command.
			// After the TUI exits, we can check the status.
			

			return nil
		},
	}
	cmd.Flags().StringVar(&cfgPath, "config", "", "YAML config path (optional)")
	return cmd
}

// Add returns the minimal "add" subcommand.
// ... (rest of the file is the same)
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

			// 若提供了 root flag 的 --notes-dir，優先覆蓋 cfg 設定，確保與測試一致
			if v, err := cmd.Root().Flags().GetString("notes-dir"); err == nil && v != "" {
				cfg.Data.NotesDir = v
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

			fmt.Fprintf(cmd.OutOrStdout(), "ID: %s\n", note.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&cfgPath, "config", "", "YAML config path (optional)")
	cmd.Flags().StringVar(&tagsStr, "tags", "", "comma-separated tags (e.g., dev,test)")
	return cmd
}

// Ask returns the minimal "ask" subcommand which only performs retrieval and prints matched note IDs.
// Usage:
//
//	ora-ora-ora ask "your question" --topk 10 --tags dev,test --config path/to/config.yaml
func (o *OraCmd) Ask() *cobra.Command {
	var (
		cfgPath string
		tagsStr string
		topK    int
		model   string
		tplPath string
		noLLM   bool
	)

	cmd := &cobra.Command{
		Use:   "ask [query]",
		Short: "檢索片段，選擇性呼叫 LLM 回答",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(strings.TrimSpace(cfgPath))
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			// 委派給 core.AskCmd，並在此重建索引以符合整合測試期望
            provider := func() (search.Index, error) {
                // 優先使用 root flag 的 --notes-dir 覆蓋 cfg
                if v, err := cmd.Root().Flags().GetString("notes-dir"); err == nil && v != "" {
                    cfg.Data.NotesDir = v
                }
                st, err := storage.New(cfg.Data.NotesDir)
                if err != nil { return nil, err }
                notes, err := st.List()
                if err != nil { return nil, err }
                idx, err := search.OpenOrCreate("")
                if err != nil { return nil, err }
				for _, n := range notes {
					if err := idx.IndexNote(n); err != nil {
						_ = idx.Close()
						return nil, err
					}
				}
				return idx, nil
			}
			argv := []string{}
			if topK > 0 { argv = append(argv, "--topk", fmt.Sprintf("%d", topK)) }
			if tagsStr != "" { argv = append(argv, "--tags", tagsStr) }
			if noLLM { argv = append(argv, "--no-llm") }
			if model != "" { argv = append(argv, "--model", model) }
			if tplPath != "" { argv = append(argv, "--template", tplPath) }
			// 目前 host 覆蓋由環境變數與 core 內部旗標處理，這裡不再重覆
			argv = append(argv, args...)
            return core.AskCmd(argv, cfg,
                core.WithIndexProvider(provider),
                core.WithWriters(cmd.OutOrStdout(), cmd.ErrOrStderr()),
            )
        },
    }

	cmd.Flags().StringVar(&cfgPath, "config", "", "YAML config path (optional)")
	cmd.Flags().StringVar(&tagsStr, "tags", "", "comma-separated tags (e.g., dev,test)")
	cmd.Flags().IntVar(&topK, "topk", 20, "max number of results")
	cmd.Flags().StringVar(&model, "model", "", "LLM model name (override config)")
	cmd.Flags().StringVar(&tplPath, "template", "", "ask prompt template path (YAML)")
	cmd.Flags().BoolVar(&noLLM, "no-llm", false, "print retrieval results only")
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

// buildAskContext 將檢索片段組裝成可讀的上下文字串。
// maxLen 用於限制字串長度，避免 prompt 過長。
func buildAskContext(snippets []search.Snippet, maxLen int) string {
	var b strings.Builder
	for i, s := range snippets {
		if i > 0 {
			b.WriteString("\n\n")
		}
		b.WriteString("- ID: ")
		b.WriteString(s.NoteID)
		if s.Excerpt != "" {
			b.WriteString("\n  Excerpt: ")
			b.WriteString(s.Excerpt)
		}
		if len(s.TagMatches) > 0 {
			b.WriteString("\n  Tags: ")
			b.WriteString(strings.Join(s.TagMatches, ","))
		}
		if maxLen > 0 && b.Len() > maxLen {
			break
		}
	}
	out := b.String()
	if maxLen > 0 && len(out) > maxLen {
		out = out[:maxLen]
	}
	return out
}

// Expose a convenient constructor for tests and main to get a configured root command.
func NewOraCmdRoot() *cobra.Command {
	app := NewOraCmd()
	root := app.RootCmd
	// Global flags: allow overriding notes dir for tests without full config
	var notesDir string
	root.PersistentFlags().String("notes-dir", "", "override notes directory (testing/advanced)")

	// Wire subcommands
	root.AddCommand(app.StartTui())
	root.AddCommand(app.Add())
	root.AddCommand(app.Ask())
	root.AddCommand(app.Diag())

	// Inject notes-dir into config loader via env var adapter (config package already has defaults)
	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// If notes-dir provided, set CONFIG_NOTES_DIR env for config.Load to pick up via defaults override path.
		if v, err := cmd.Flags().GetString("notes-dir"); err == nil {
			notesDir = v
		}
		if notesDir != "" {
			// Best-effort: let config respect this by setting process env the loader reads (if supported).
			// If not supported, storage.New will be called with cfg.Data.NotesDir which comes from config defaults.
			// To ensure override, we set a fallback: if cfg path empty later, commands read notesDir flag directly.
			os.Setenv("ORA_NOTES_DIR", notesDir)
		}
		return nil
	}
	return root
}

// Diag 提供環境診斷：檢查 Ollama 連線、notes 目錄可寫、模板可讀。
func (o *OraCmd) Diag() *cobra.Command {
	var (
		host string
		tpl  string
	)
	cmd := &cobra.Command{
		Use:   "diag",
		Short: "診斷環境（Ollama/路徑/模板）",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			// Ollama host
			if host == "" {
				host = os.Getenv("OLLAMA_HOST")
			}
			if host == "" {
				host = "http://127.0.0.1:11434"
			}
			if err := diagOllama(host); err != nil {
				fmt.Fprintf(out, "Ollama: unreachable (%v)\n", err)
			} else {
				fmt.Fprintln(out, "Ollama: reachable")
			}

			// Notes dir writability (use notes-dir flag if provided)
			notesDir, _ := cmd.Root().Flags().GetString("notes-dir")
			if notesDir == "" {
				// fallback default
				notesDir = "data/notes"
			}
			if err := diagWritable(notesDir); err != nil {
				fmt.Fprintf(out, "NotesDir: not writable (%v)\n", err)
			} else {
				fmt.Fprintln(out, "NotesDir: writable")
			}

			// Template readability
			if tpl != "" {
				if _, err := os.Stat(tpl); err == nil {
					fmt.Fprintln(out, "Template: ok")
				} else {
					fmt.Fprintf(out, "Template: not found (%v)\n", err)
				}
			} else {
				fmt.Fprintln(out, "Template: skipped (no path)")
			}

			return nil
		},
	}
	cmd.Flags().StringVar(&host, "ollama-host", "", "override Ollama host for check")
	cmd.Flags().StringVar(&tpl, "template", "", "template path to verify")
	return cmd
}

func diagOllama(host string) error {
	// simple GET to base URL to verify reachability
	req, err := http.NewRequest("GET", host, nil)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	return nil
}

func diagWritable(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".probe-*")
	if err != nil {
		return err
	}
	name := f.Name()
	_ = f.Close()
	return os.Remove(name)
}