package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

// 目標：以 TDD 驅動 LLM 客戶端與 Chat 邏輯
// 約定：
// - NewClient(host, model) 回傳 LLM 介面實作。
// - Chat(ctx, system, user, opts) 以 Ollama /api/chat 介面送出請求，回傳最終文字。
// - 異常（非 2xx 或 JSON 非法）回傳錯誤。

func TestClient_Chat_SendsExpectedPayloadAndParsesResponse(t *testing.T) {
	// 使用 fake RoundTripper 攔截請求，避免啟動本機埠
	rt := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/api/chat" {
			t.Fatalf("unexpected path: %s", req.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
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
		respBody, _ := json.Marshal(map[string]any{
			"message": map[string]any{"role": "assistant", "content": "hi"},
		})
		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewReader(respBody)),
			Header:     make(http.Header),
		}, nil
	})
	httpc := &http.Client{Transport: rt}
	cli := NewClientWithHTTP("http://example", "test-model", httpc)
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
	rt := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 500, Body: ioutil.NopCloser(bytes.NewBufferString("boom"))}, nil
	})
	httpc := &http.Client{Transport: rt}
	cli := NewClientWithHTTP("http://example", "test-model", httpc)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if _, err := cli.Chat(ctx, "sys", "hello", Options{}); err == nil {
		t.Fatalf("expected error, got nil")
	}
}

// roundTripFunc 讓函式實作 http.RoundTripper
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestClient_Chat_OmitsZeroNumPredict(t *testing.T) {
	rt := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if opt, ok := body["options"].(map[string]any); ok {
			if _, exists := opt["num_predict"]; exists {
				t.Fatalf("expected no num_predict in options, got: %#v", opt)
			}
		}
		respBody, _ := json.Marshal(map[string]any{
			"message": map[string]any{"role": "assistant", "content": "ok"},
		})
		return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader(respBody)), Header: make(http.Header)}, nil
	})
	httpc := &http.Client{Transport: rt}
	cli := NewClientWithHTTP("http://example", "test-model", httpc)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err := cli.Chat(ctx, "sys", "u", Options{NumPredict: 0}); err != nil {
		t.Fatalf("Chat error: %v", err)
	}
}

func TestClient_Chat_IncludesNumPredictWhenPositive(t *testing.T) {
	rt := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		opt, ok := body["options"].(map[string]any)
		if !ok {
			t.Fatalf("options missing: %#v", body)
		}
		v, ok := opt["num_predict"]
		if !ok {
			t.Fatalf("num_predict missing in options")
		}
		// JSON 數值解碼為 float64
		if vv, ok := v.(float64); !ok || vv != 128 {
			t.Fatalf("num_predict want 128, got %#v", v)
		}
		respBody, _ := json.Marshal(map[string]any{
			"message": map[string]any{"role": "assistant", "content": "ok"},
		})
		return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader(respBody)), Header: make(http.Header)}, nil
	})
	httpc := &http.Client{Transport: rt}
	cli := NewClientWithHTTP("http://example", "test-model", httpc)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err := cli.Chat(ctx, "sys", "u", Options{NumPredict: 128}); err != nil {
		t.Fatalf("Chat error: %v", err)
	}
}
