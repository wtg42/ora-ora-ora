package core

import (
    "context"
    "errors"
    "flag"
    "fmt"
    "io"
    "os"
    "strings"
    "time"

    "github.com/wtg42/ora-ora-ora/agent"
    "github.com/wtg42/ora-ora-ora/config"
    "github.com/wtg42/ora-ora-ora/prompt"
    "github.com/wtg42/ora-ora-ora/search"
)

// AskCmd 實作最小 CLI：解析旗標，檢索片段，必要時呼叫 LLM。
// 註：為了易測，採用標準庫 flag 與傳入 argv 形式。
// Option 為 AskCmd 的可選注入設定。
type Option func(*askDeps)

type askDeps struct {
    indexProvider func() (search.Index, error)
    out io.Writer
    err io.Writer
}

// WithIndexProvider 允許呼叫端提供索引建立邏輯（例如先重建再回傳）。
func WithIndexProvider(p func() (search.Index, error)) Option {
    return func(d *askDeps) { d.indexProvider = p }
}

// WithWriters 允許注入輸出 writer，預設為 os.Stdout/os.Stderr。
func WithWriters(out, err io.Writer) Option {
    return func(d *askDeps) {
        d.out = out
        d.err = err
    }
}

func AskCmd(argv []string, cfg *config.Config, opts ...Option) error {
    fs := flag.NewFlagSet("ask", flag.ContinueOnError)

    topK := fs.Int("topk", 20, "number of snippets")
    tagStr := fs.String("tags", "", "comma-separated tags (AND)")
    noLLM := fs.Bool("no-llm", false, "search only, no LLM call")
    templatePath := fs.String("template", "", "path to YAML template")
    model := fs.String("model", cfg.Model, "model name")
    host := fs.String("ollama-host", cfg.OllamaHost, "ollama host")

    // advanced options
    temp := fs.Float64("temp", 0.0, "temperature")
    topP := fs.Float64("top-p", 0.0, "top-p")
    numCtx := fs.Int("num-ctx", 0, "num ctx")
    numPredict := fs.Int("num-predict", 0, "num predict")
    keepAlive := fs.String("keep-alive", "", "keep alive duration (e.g. 30s)")

    if err := fs.Parse(argv); err != nil {
        return err
    }
    args := fs.Args()
    if len(args) == 0 {
        return errors.New("missing question")
    }
    question := strings.TrimSpace(strings.Join(args, " "))

    // env override for host
    if v := os.Getenv("OLLAMA_HOST"); v != "" {
        *host = v
    }

    // 準備索引
    dep := askDeps{out: os.Stdout, err: os.Stderr}
    for _, o := range opts { o(&dep) }
    var (
        idx search.Index
        err error
    )
    if dep.indexProvider != nil {
        idx, err = dep.indexProvider()
    } else {
        // 預設使用 in-memory stub
        idx, err = search.OpenOrCreate("")
    }
    if err != nil {
        return fmt.Errorf("open index: %w", err)
    }
    defer idx.Close()

    // query snippets
    var tags []string
    if strings.TrimSpace(*tagStr) != "" {
        parts := strings.Split(*tagStr, ",")
        for _, p := range parts {
            t := strings.TrimSpace(p)
            if t != "" {
                tags = append(tags, t)
            }
        }
    }
    snippets, err := idx.Query(question, *topK, tags)
    if err != nil {
        return fmt.Errorf("query: %w", err)
    }

    if *noLLM {
        for _, s := range snippets {
            fmt.Fprintln(dep.out, s.NoteID)
        }
        return nil
    }

    if len(snippets) == 0 {
        fmt.Fprintln(dep.out, "找不到相關片段，請調整關鍵詞或移除標籤過濾。")
        return nil
    }

    // build context from snippets (simple join)
    var b strings.Builder
    for i, s := range snippets {
        if i > 0 {
            b.WriteString("\n---\n")
        }
        b.WriteString(s.Excerpt)
    }
    contextText := b.String()

    // load template
    tpl, warn := prompt.LoadAskTemplate(*templatePath)
    if warn != "" {
        fmt.Fprintln(dep.out, warn)
    }
    system := strings.ReplaceAll(tpl.System, "{{question}}", question)
    system = strings.ReplaceAll(system, "{{context}}", contextText)
    user := strings.ReplaceAll(tpl.User, "{{question}}", question)
    user = strings.ReplaceAll(user, "{{context}}", contextText)

    // parse keepAlive
    var ka time.Duration
    if *keepAlive != "" {
        if d, err := time.ParseDuration(*keepAlive); err == nil {
            ka = d
        }
    }

    llm := agent.NewClient(*host, *model)
    out, err := llm.Chat(context.Background(), system, user, agent.Options{
        Temperature: *temp,
        TopP:        *topP,
        NumCtx:      *numCtx,
        NumPredict:  *numPredict,
        KeepAlive:   ka,
    })
    if err != nil {
        return fmt.Errorf("llm: %w", err)
    }
    fmt.Fprintln(dep.out, out)
    return nil
}
