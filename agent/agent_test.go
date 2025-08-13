package agent

import (
    "context"
    "strings"
    "testing"
)

func TestMockLLM_Chat(t *testing.T) {
    m := MockLLM{}
    out, err := m.Chat(context.Background(), "sys", "hello", Options{})
    if err != nil { t.Fatalf("chat: %v", err) }
    if !strings.Contains(out, "[mock-answer]") { t.Fatalf("unexpected output: %s", out) }
}

