package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// 目標：以 TDD 驅動 LLM 客戶端與 Chat 邏輯
// 約定：
// - NewClient(host, model) 回傳 LLM 介面實作。
// - Chat(ctx, system, user, opts) 以 Ollama /api/chat 介面送出請求，回傳最終文字。
// - 異常（非 2xx 或 JSON 非法）回傳錯誤。

func TestClient_Chat_SendsExpectedPayloadAndParsesResponse(t *testing.T) {
	// 建立假伺服器驗證請求內容
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if body["model"] != "test-model" {
			t.Fatalf("model mismatch: %v", body["model"])
		}
		msgs, ok := body["messages"].([]any)
		if !ok || len(msgs) != 2 {
			t.Fatalf("messages malformed: %#v", body["messages"])
		}
		// 回傳固定答案
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message": map[string]any{"role": "assistant", "content": "hi"},
		})
	}))
	defer srv.Close()

	cli := NewClient(srv.URL, "test-model")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	out, err := cli.Chat(ctx, "sys", "hello", Options{Temperature: 0})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if out != "hi" {
		t.Fatalf("want 'hi', got %q", out)
	}
}

func TestClient_Chat_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	cli := NewClient(srv.URL, "test-model")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if _, err := cli.Chat(ctx, "sys", "hello", Options{}); err == nil {
		t.Fatalf("expected error, got nil")
	}
}
