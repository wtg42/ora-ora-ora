package cmd

import (
    "fmt"
    "os"
    "strings"
    "time"

    "github.com/spf13/cobra"
    "github.com/wtg42/ora-ora-ora/agent"
    "github.com/wtg42/ora-ora-ora/config"
    "github.com/wtg42/ora-ora-ora/search"
    "github.com/wtg42/ora-ora-ora/storage"
)

// AskCmd 建立 `ask` 子指令，從檔案儲存重建索引並查詢片段。
func (o *OraCmd) AskCmd() *cobra.Command {
    var (
        topK int
        tags []string
        useLLM bool
        model string
        host string
        templatePath string
        temperature float64
        topP float64
        numCtx int
        numPredict int
        keepAliveSec int
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

            // 選擇性呼叫 LLM
            if useLLM {
                cfg := config.Default()
                if host == "" { host = cfg.OllamaHost }
                if model == "" { model = cfg.Model }

                tpl := loadAskTemplateOrDefault(templatePath)
                ctxStr := renderContext(snippets)
                sys, usr := renderTemplate(tpl, q, ctxStr)

                llm := agent.NewOllama(host, model)
                ans, err := llm.Chat(cmd.Context(), sys, usr, agent.Options{
                    Temperature: temperature,
                    TopP:        topP,
                    NumCtx:      numCtx,
                    NumPredict:  numPredict,
                    KeepAlive:   time.Duration(keepAliveSec) * time.Second,
                })
                if err != nil {
                    return err
                }
                fmt.Fprintln(cmd.OutOrStdout(), "\n---\nAnswer:\n"+ans)
            }
            return nil
        },
    }
    cmd.Flags().IntVar(&topK, "topk", 20, "回傳片段數量")
    cmd.Flags().StringSliceVar(&tags, "tags", nil, "以標籤過濾（可多次指定）")
    // LLM 相關旗標（預設關閉）
    cmd.Flags().BoolVar(&useLLM, "use-llm", false, "啟用 LLM 產生回答")
    cfg := config.Default()
    cmd.Flags().StringVar(&model, "model", cfg.Model, "Ollama 模型名稱")
    cmd.Flags().StringVar(&host, "ollama-host", cfg.OllamaHost, "Ollama 主機位址")
    cmd.Flags().StringVar(&templatePath, "template", "", "ask 模板路徑（YAML-lite，缺省使用內建）")
    cmd.Flags().Float64Var(&temperature, "temperature", 0, "LLM 溫度 (0~2)")
    cmd.Flags().Float64Var(&topP, "top-p", 0, "LLM Top-P")
    cmd.Flags().IntVar(&numCtx, "num-ctx", 0, "LLM 上下文長度")
    cmd.Flags().IntVar(&numPredict, "num-predict", 0, "LLM 預測 token 數")
    cmd.Flags().IntVar(&keepAliveSec, "keep-alive", 0, "LLM keep-alive 秒數")
    return cmd
}

// Ask 模板（YAML-lite）：若提供路徑，會嘗試解析含 system/user 的內容；否則用預設。
type AskTemplate struct {
    System string
    User   string
}

func defaultAskTemplate() AskTemplate {
    return AskTemplate{
        System: "你是助理，請以繁體中文回答，僅根據提供的相關筆記回答；若無關資訊，請明確說明找不到答案。",
        User:   "問題：{{question}}\n\n相關筆記（若空代表無）：\n{{context}}\n\n請彙整要點、避免臆測。",
    }
}

// 簡易 YAML-lite 載入：支援如下結構（不依賴外部套件）：
// system: |
//   ...多行內容
// user: |
//   ...多行內容
func loadAskTemplateOrDefault(path string) AskTemplate {
    if strings.TrimSpace(path) == "" {
        return defaultAskTemplate()
    }
    b, err := os.ReadFile(path)
    if err != nil {
        return defaultAskTemplate()
    }
    lines := strings.Split(string(b), "\n")
    var sys, usr []string
    var mode string
    for _, ln := range lines {
        trimmed := strings.TrimSpace(ln)
        switch {
        case strings.HasPrefix(trimmed, "system:"):
            mode = "system"
            // 若是以 "|" 結尾，表示多行，忽略本行內容
            continue
        case strings.HasPrefix(trimmed, "user:"):
            mode = "user"
            continue
        default:
            // 收集內容（去掉前導兩空白/縮排）
            content := strings.TrimPrefix(ln, "  ")
            switch mode {
            case "system":
                sys = append(sys, content)
            case "user":
                usr = append(usr, content)
            }
        }
    }
    st := strings.TrimSpace(strings.Join(sys, "\n"))
    ut := strings.TrimSpace(strings.Join(usr, "\n"))
    if st == "" || ut == "" {
        return defaultAskTemplate()
    }
    return AskTemplate{System: st, User: ut}
}

func renderTemplate(tpl AskTemplate, q, ctx string) (string, string) {
    rep := func(s string) string {
        s = strings.ReplaceAll(s, "{{question}}", q)
        s = strings.ReplaceAll(s, "{{context}}", ctx)
        return s
    }
    return rep(tpl.System), rep(tpl.User)
}

func renderContext(snippets []search.Snippet) string {
    var b strings.Builder
    for i, s := range snippets {
        fmt.Fprintf(&b, "[%d] %s (score=%.2f)\n%s\n---\n", i+1, s.NoteID, s.Score, s.Excerpt)
    }
    return b.String()
}
