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

// LLMProvider interface for opencode SDK (mocked)
type LLMProvider interface {
	Prompt(ctx context.Context, prompt string, options ...LLMOption) (*LLMResponse, error)
	Close() error
}

// LLMOption for LLM
type LLMOption func(*llmOptions)

type llmOptions struct {
	Model string
}

func WithLLMModel(model string) LLMOption {
	return func(o *llmOptions) {
		o.Model = model
	}
}

// LLMResponse for LLM
type LLMResponse struct {
	Content string
}

// LLMSession for core ask with opencode integration (mocked for stateless chat-only)
type LLMSession struct {
	Provider        LLMProvider
	Prompt          string
	Mode            string // e.g., "chat-only"
	FallbackEnabled bool
}

// ChatOnlyMode constant
const LLMChatOnlyMode = "chat-only"
const LLMDefaultMode = "default"

// AskLLM performs chat-only query with opencode SDK or fallback
func (s *LLMSession) AskLLM(ctx context.Context) (*LLMResponse, error) {
	if s.Mode != LLMChatOnlyMode {
		return nil, errors.New("mode not supported")
	}
	if s.Provider == nil {
		return nil, errors.New("no LLM provider")
	}
	resp, err := s.Provider.Prompt(ctx, s.Prompt, WithLLMModel("default"))
	if err != nil && s.FallbackEnabled {
		// Fallback to local echo or Ollama
		return &LLMResponse{Content: "Fallback response: Unable to reach LLM"}, nil
	}
	return resp, err
}

// MockLLMProvider for testing
type MockLLMProvider struct {
	Responses []string
	Errors    []error
	idx       int
}

func (m *MockLLMProvider) Prompt(ctx context.Context, prompt string, options ...LLMOption) (*LLMResponse, error) {
	// Ignore options for mock
	if m.idx < len(m.Errors) {
		err := m.Errors[m.idx]
		m.idx++
		return nil, err
	}
	if m.idx < len(m.Responses) {
		resp := &LLMResponse{Content: m.Responses[m.idx]}
		m.idx++
		return resp, nil
	}
	return &LLMResponse{Content: "Default mock"}, nil
}

func (m *MockLLMProvider) Close() error {
	return nil
}

// AskCmd 實作最小 CLI：解析旗標，檢索片段，必要時呼叫 LLM。
// 註：為了易測，採用標準庫 flag 與傳入 argv 形式。
// Option 為 AskCmd 的可選注入設定。
type Option func(*askDeps)

type askDeps struct {
	indexProvider func() (search.Index, error)
	llmProvider   func(host, model string) agent.LLM
	out           io.Writer
	err           io.Writer
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

// WithLLMProvider 允許注入 LLM 客戶端建立邏輯。
func WithLLMProvider(p func(host, model string) agent.LLM) Option {
	return func(d *askDeps) { d.llmProvider = p }
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
	for _, o := range opts {
		o(&dep)
	}
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
		// 警示訊息走 stderr，以符合文件預期與 CLI 慣例
		fmt.Fprintln(dep.err, warn)
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

	var llm agent.LLM
	if dep.llmProvider != nil {
		llm = dep.llmProvider(*host, *model)
	} else {
		llm = agent.NewClient(*host, *model)
	}
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
	cleaned := sanitizeLLMOutput(out)
	if strings.TrimSpace(cleaned) == "" {
		// LLM 回覆為空或僅殘留角色標記時，回退輸出最相關片段，避免使用者以為無結果。
		fmt.Fprintln(dep.out, fallbackFromSnippets(snippets, 3))
		return nil
	}
	fmt.Fprintln(dep.out, cleaned)
	return nil
}

// sanitizeLLMOutput 移除常見的聊天模板殘留前綴（例如 assistant 標籤），避免污染 CLI 輸出。
// 僅針對開頭少量已知片段做保守處理，避免過度刪除有效內容。
func sanitizeLLMOutput(s string) string {
	t := strings.TrimLeft(s, "\n\r \t")
	// 常見殘留："**.assistant", "**assistant", "assistant:", "Assistant:", "<|assistant|>", "assistant"
	prefixes := []string{"**.assistant", "**assistant", "assistant:", "Assistant:", "<|assistant|>", "assistant"}
	for _, p := range prefixes {
		if strings.HasPrefix(t, p) {
			t = strings.TrimLeft(t[len(p):], "\n\r \t:>")
			break
		}
	}
	return t
}

// fallbackFromSnippets 在 LLM 無有效回覆時，輸出前 n 個檢索片段的簡短清單。
func fallbackFromSnippets(snips []search.Snippet, n int) string {
	if len(snips) == 0 {
		return "未取得 LLM 回覆。"
	}
	if n <= 0 {
		n = 3
	}
	if n > len(snips) {
		n = len(snips)
	}
	var b strings.Builder
	b.WriteString("未取得 LLM 回覆；以下為相關片段：\n")
	for i := 0; i < n; i++ {
		txt := strings.TrimSpace(snips[i].Excerpt)
		if txt == "" {
			txt = snips[i].NoteID
		}
		b.WriteString("- ")
		b.WriteString(txt)
		if i < n-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}
