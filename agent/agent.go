package agent

// Package agent 定義與 LLM（如 Ollama）互動的最小介面與選項。
// 目前提供 MockLLM 以便在未串接 HTTP 前即可驗證整體流程。

import (
    "context"
    "fmt"
    "time"
)

// Options 映射到常見的 Ollama 參數，僅保留必要欄位。
type Options struct {
    Temperature float64
    TopP        float64
    NumCtx      int
    NumPredict  int
    KeepAlive   time.Duration
}

// LLM 定義最小互動介面（非串流）。
type LLM interface {
    Chat(ctx context.Context, system, user string, opts Options) (string, error)
}

// MockLLM 為測試與開發時的假回應。
type MockLLM struct{}

func (m MockLLM) Chat(_ context.Context, system, user string, _ Options) (string, error) {
    // 回傳簡單的可預期字串，方便測試斷言。
    return fmt.Sprintf("[mock-answer]\nsystem: %s\nuser: %s", system, user), nil
}

