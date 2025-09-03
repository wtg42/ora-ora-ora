package tui_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wtg42/ora-ora-ora/tui"
)

func pump(m tea.Model, msgs ...tea.Msg) tea.Model {
	for _, msg := range msgs {
		m, _ = m.Update(msg)
	}
	return m
}

func TestChatModel_HelpShownAtBottom(t *testing.T) {
	m := tui.NewChatModel()
	m2 := pump(m, tea.WindowSizeMsg{Width: 80, Height: 20}).(tui.ChatModel)
	v := m2.View()
	// 檢查基本鍵位字樣是否存在
	for _, kw := range []string{"enter", "alt+enter", "esc", "pgup", "pgdn"} {
		if !strings.Contains(v, kw) {
			t.Fatalf("expected help to contain %q in view", kw)
		}
	}
}

func TestAddWizard_HelpShownAtBottom(t *testing.T) {
	m := tui.NewAddWizardModel()
	m2 := pump(m, tea.WindowSizeMsg{Width: 80, Height: 20}).(tui.AddWizardModel)
	v := m2.View()
	for _, kw := range []string{"enter", "alt+enter", "esc", "pgup", "pgdn"} {
		if !strings.Contains(v, kw) {
			t.Fatalf("expected help to contain %q in view", kw)
		}
	}
}
