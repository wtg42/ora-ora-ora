package tui

import (
    "testing"
    "time"
)

// Test that when llmEnabled=false (default), sending a message appends a retrieval reply.
func TestChat_Settings_LLMOff_RetrievalOnly(t *testing.T) {
    m := NewChatModel()
    // 直接呼叫 queryAndAppend 檢查訊息型態即可（indexProvider 為 nil 時會建立 in-memory index，回空結果訊息）
    // Directly call queryAndAppend to avoid complex key events in tests
    cmd := m.queryAndAppend("hello")
    msg := cmd()
    // The message should be assistant type Message
    mm, ok := msg.(Message)
    if !ok {
        t.Fatalf("expected Message, got %T", msg)
    }
    if mm.Role != "assistant" {
        t.Fatalf("expected assistant role, got %s", mm.Role)
    }
    _ = time.Now() // silence unused imports on some toolchains
}

// Note: A full LLM-on flow test will be added with actual llmProvider hook once implemented in ChatModel.
