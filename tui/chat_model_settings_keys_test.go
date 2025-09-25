package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wtg42/ora-ora-ora/search"
)

// searchIndex 是為了型別簡化而取自 ChatModel.indexProvider 的回傳型別。
// reuse mockIndex from llm test file; implements search.Index

func TestSettings_Apply_TopK_IncrementsOnBracket(t *testing.T) {
	m := NewChatModel()
	// Open settings via F2
	mod, _ := m.Update(tea.KeyMsg{Type: tea.KeyF2})
	m = mod.(ChatModel)
	if !m.settingsOpen {
		t.Fatalf("settings should be open")
	}
	// Press ']' to increment tmpTopK, then Enter to apply
	mod, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{']'}})
	m = mod.(ChatModel)
	mod, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mod.(ChatModel)
	// After apply, settings closed and topK increased (default 5 -> 6)
	if m.settingsOpen {
		t.Fatalf("settings should be closed after apply")
	}
	if m.topK != 6 {
		t.Fatalf("expected topK=6, got %d", m.topK)
	}
	// Set a mock index so query returns a snippet
	m.indexProvider = func() (search.Index, error) { return mockIndex{}, nil }
	// Run a query to see header reflects new TopK
	cmd := m.queryAndAppend("q")
	if msg := cmd(); msg != nil {
		mod, _ = m.Update(msg)
		m = mod.(ChatModel)
	}
	if len(m.messages) == 0 {
		t.Fatalf("no messages")
	}
	got := m.messages[len(m.messages)-1].Content
	if !strings.Contains(got, "Top 6") {
		t.Fatalf("expected header to contain 'Top 6', got %q", got)
	}
}

func TestSettings_ToggleLLM_WithApply(t *testing.T) {
	m := NewChatModel()
	if m.llmEnabled {
		t.Fatalf("llm should be off by default")
	}
	// Open settings and toggle 'l', then apply
	mod, _ := m.Update(tea.KeyMsg{Type: tea.KeyF2})
	m = mod.(ChatModel)
	mod, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	m = mod.(ChatModel)
	mod, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mod.(ChatModel)
	if !m.llmEnabled {
		t.Fatalf("expected llmEnabled=true after apply")
	}
}

func TestSettings_TabFocus_And_EmptyHostError(t *testing.T) {
	m := NewChatModel()
	// 開啟設定面板
	mod, _ := m.Update(tea.KeyMsg{Type: tea.KeyF2})
	m = mod.(ChatModel)
	// 預設焦點在 Model 欄，輸入 'X' -> 應進入 modelInput
	mod, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'X'}})
	m = mod.(ChatModel)
	if got := m.modelInput.Value(); len(got) == 0 || got[len(got)-1] != 'X' {
		t.Fatalf("expected modelInput ends with X, got %q", m.modelInput.Value())
	}
	// Tab 切到 Host 欄，輸入 'Y'
	mod, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = mod.(ChatModel)
	mod, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'Y'}})
	m = mod.(ChatModel)
	if got := m.hostInput.Value(); len(got) == 0 || got[len(got)-1] != 'Y' {
		t.Fatalf("expected hostInput ends with Y, got %q", m.hostInput.Value())
	}
	// 清空 Host 後按 Enter -> 應顯示錯誤並保持面板開啟
	m.hostInput.SetValue("")
	mod, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mod.(ChatModel)
	if !m.settingsOpen {
		t.Fatalf("expected settings remain open on error")
	}
	// 面板渲染應包含錯誤文字
	view := m.renderSettings()
	if view == "" || (view != "" && !contains(view, "Host 必填")) {
		t.Fatalf("expected error hint in settings view, got: %q", view)
	}
}

func TestSettings_ConnTest_ShowsResult(t *testing.T) {
	m := NewChatModel()
	// override connection check to deterministic message
	m.connCheck = func(host string) string { return "OK: " + host }
	// open
	mod, _ := m.Update(tea.KeyMsg{Type: tea.KeyF2})
	m = mod.(ChatModel)
	// press 't' to test
	mod, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	m = mod.(ChatModel)
	view := m.renderSettings()
	if !contains(view, "OK:") {
		t.Fatalf("expected connection OK in view, got: %q", view)
	}
	// 僅驗證顯示結果，錯誤分支另有主機必填測試覆蓋
}

// contains: minimal helper to avoid importing strings again
func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || (len(sub) > 0 && (indexOf(s, sub) >= 0)))
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
