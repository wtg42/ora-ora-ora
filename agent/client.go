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
type ollamaClient struct {
	host  string
	model string
}

// NewClient creates a new Ollama LLM client
func NewClient(host, model string) LLM {
	return &ollamaClient{
		host:  host,
		model: model,
	}
}

// Chat sends a chat request to Ollama and returns the response
func (c *ollamaClient) Chat(ctx context.Context, system, user string, opts Options) (string, error) {
	payload := map[string]interface{}{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
		"stream": false,
		"options": map[string]interface{}{
			"temperature": opts.Temperature,
			"top_p":       opts.TopP,
			"num_ctx":     opts.NumCtx,
			"num_predict": opts.NumPredict,
		},
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

	client := &http.Client{Timeout: 30 * time.Second}
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
