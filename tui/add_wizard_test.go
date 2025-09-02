package tui_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wtg42/ora-ora-ora/tui"
)

func driveAddWizard(m tui.AddWizardModel, msgs ...tea.Msg) tui.AddWizardModel {
	var model tea.Model
	model = m
	for _, msg := range msgs {
		model, _ = model.Update(msg)
	}
	return model.(tui.AddWizardModel)
}

func TestAddWizard_InitialPrompt(t *testing.T) {
	m := tui.NewAddWizardModel()
	m = driveAddWizard(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	if !strings.Contains(m.View(), "請描述要新增的筆記內容") {
		t.Fatalf("expected initial system prompt visible")
	}
}

func TestAddWizard_ParseAndSummarize(t *testing.T) {
	m := tui.NewAddWizardModel()
	m = driveAddWizard(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	// type a message and send
	m = driveAddWizard(m,
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("研究 Bleve 搜尋 #dev tags: search,go")},
		tea.KeyMsg{Type: tea.KeyEnter},
	)
	v := m.View()
	if !strings.Contains(v, "理解到的新增參數") {
		t.Fatalf("expected assistant summary message")
	}
	if !strings.Contains(v, "標籤") {
		t.Fatalf("expected tags listed in summary")
	}
}
