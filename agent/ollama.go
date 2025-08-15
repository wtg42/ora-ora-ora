package agent

// Ollama HTTP 客戶端，符合 LLM 介面（非串流）。

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
    "time"
)

type Ollama struct {
    Host       string
    Model      string
    HTTPClient *http.Client
}

func NewOllama(host, model string) *Ollama {
    return &Ollama{
        Host:  strings.TrimRight(host, "/"),
        Model: model,
        HTTPClient: &http.Client{
            Timeout: 60 * time.Second,
        },
    }
}

type chatMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type chatRequest struct {
    Model     string                 `json:"model"`
    Messages  []chatMessage          `json:"messages"`
    Stream    bool                   `json:"stream"`
    Options   map[string]any         `json:"options,omitempty"`
    KeepAlive string                 `json:"keep_alive,omitempty"`
}

type chatResponse struct {
    Message struct {
        Role    string `json:"role"`
        Content string `json:"content"`
    } `json:"message"`
    Error string `json:"error,omitempty"`
}

// Chat 呼叫 Ollama /api/chat 取得回覆內容。
func (o *Ollama) Chat(ctx context.Context, system, user string, opts Options) (string, error) {
    if o.Host == "" || o.Model == "" {
        return "", fmt.Errorf("ollama host/model is empty")
    }
    reqBody := chatRequest{
        Model:  o.Model,
        Stream: false,
        Messages: []chatMessage{
            {Role: "system", Content: system},
            {Role: "user", Content: user},
        },
    }
    // map options（僅帶有值的欄位）
    opt := map[string]any{}
    if opts.Temperature > 0 {
        opt["temperature"] = opts.Temperature
    }
    if opts.TopP > 0 {
        opt["top_p"] = opts.TopP
    }
    if opts.NumCtx > 0 {
        opt["num_ctx"] = opts.NumCtx
    }
    if opts.NumPredict > 0 {
        opt["num_predict"] = opts.NumPredict
    }
    if len(opt) > 0 {
        reqBody.Options = opt
    }
    if opts.KeepAlive > 0 {
        reqBody.KeepAlive = fmt.Sprintf("%ds", int(opts.KeepAlive.Seconds()))
    }

    b, _ := json.Marshal(reqBody)
    req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.Host+"/api/chat", bytes.NewReader(b))
    if err != nil {
        return "", err
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := o.HTTPClient.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    if resp.StatusCode >= 400 {
        return "", fmt.Errorf("ollama http %d", resp.StatusCode)
    }
    var cr chatResponse
    if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
        return "", fmt.Errorf("decode response: %w", err)
    }
    if cr.Error != "" {
        return "", fmt.Errorf("ollama error: %s", cr.Error)
    }
    return cr.Message.Content, nil
}

