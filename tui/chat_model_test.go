package tui_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wtg42/ora-ora-ora/tui"
)

// helper: drive a model with a message and return updated concrete model
func updateChat(m tui.ChatModel, msg tea.Msg) tui.ChatModel {
	updated, _ := m.Update(msg)
	return updated.(tui.ChatModel)
}

func TestChatModel_EnterSendsMessageAndKeepsPrompt(t *testing.T) {
	m := tui.NewChatModel()
	// set window size to allocate viewport and input
	m = updateChat(m, tea.WindowSizeMsg{Width: 80, Height: 24})

	// type content
	m = updateChat(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("hello world")})
	// press enter to send
	m = updateChat(m, tea.KeyMsg{Type: tea.KeyEnter})

	// verify rendered history contains the sent text
	if !strings.Contains(m.View(), "hello world") {
		t.Fatalf("expected history to contain sent message")
	}

	// input should be cleared; prompt should still be visible at the end
	v := m.View()
	if !strings.Contains(v, "▌ ") {
		t.Fatalf("expected prompt to be visible in view")
	}
}

func TestChatModel_LayoutWithMargins(t *testing.T) {
	m := tui.NewChatModel()
	// 設置窗口大小，應減去容器邊距
	m = updateChat(m, tea.WindowSizeMsg{Width: 80, Height: 24})

	// 檢查 View 是否應用了容器樣式
	v := m.View()
	lines := strings.Split(v, "\n")
	if len(lines) == 0 {
		t.Fatal("expected view to have content")
	}
	// 第一行應有左邊距（空格），Margin(1,2) 在左側添加2個空格
	if !strings.HasPrefix(lines[0], "  ") {
		t.Fatalf("expected view to start with margin spaces, got: %q", lines[0])
	}
}
