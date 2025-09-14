package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Options defines parameters for LLM chat requests
type Options struct {
	Temperature float64
	TopP        float64
	NumCtx      int
	NumPredict  int
	KeepAlive   time.Duration
}

// LLM defines the interface for language model interactions
type LLM interface {
	Chat(ctx context.Context, system, user string, opts Options) (string, error)
}

// ollamaClient implements LLM for Ollama API
// doer 是 http.Client 的最小依賴，用於測試時注入假實作，避免綁定本機埠。
type doer interface {
    Do(req *http.Request) (*http.Response, error)
}

type ollamaClient struct {
    host  string
    model string
    httpc doer
}

// NewClient creates a new Ollama LLM client
func NewClient(host, model string) LLM {
    return &ollamaClient{host: host, model: model, httpc: &http.Client{Timeout: 30 * time.Second}}
}

// NewClientWithHTTP 提供測試用建構子，可注入自訂 http 客戶端（如 fake RoundTripper）。
func NewClientWithHTTP(host, model string, httpc doer) LLM {
    if httpc == nil {
        httpc = &http.Client{Timeout: 30 * time.Second}
    }
    return &ollamaClient{host: host, model: model, httpc: httpc}
}

// Chat sends a chat request to Ollama and returns the response
func (c *ollamaClient) Chat(ctx context.Context, system, user string, opts Options) (string, error) {
    // 構建 options：避免 num_predict=0（代表不生成）的情況寫入，導致空回覆。
    opt := map[string]interface{}{}
    // 溫度與 top_p 若為 0 代表明確設定，仍允許傳入。
    if opts.Temperature != 0 {
        opt["temperature"] = opts.Temperature
    } else {
        // 若為 0，讓伺服端採用預設即可（不放入鍵）。
    }
    if opts.TopP != 0 {
        opt["top_p"] = opts.TopP
    }
    if opts.NumCtx > 0 {
        opt["num_ctx"] = opts.NumCtx
    }
    if opts.NumPredict > 0 {
        opt["num_predict"] = opts.NumPredict
    }

    payload := map[string]interface{}{
        "model": c.model,
        "messages": []map[string]string{
            {"role": "system", "content": system},
            {"role": "user", "content": user},
        },
        "stream": false,
    }
    if len(opt) > 0 {
        payload["options"] = opt
    }

	if opts.KeepAlive > 0 {
		payload["keep_alive"] = opts.KeepAlive.String()
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.host+"/api/chat", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

    client := c.httpc
    if client == nil {
        client = &http.Client{Timeout: 30 * time.Second}
    }
    resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("server error %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	return response.Message.Content, nil
}
