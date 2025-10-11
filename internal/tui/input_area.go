// Package tui 提供了終端使用者介面 (TUI) 的實現。
package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InputArea 結構體代表輸入區域組件，處理用戶輸入。
type InputArea struct {
	runes       []rune         // AI 心智註解: 以 rune 切片儲存輸入，避免多位元字元被破壞。
	cursor      int            // AI 心智註解: 記錄目前游標所在的 rune index。
	placeholder string         // 提示文字。
	styles      lipgloss.Style // 樣式設定。
}

// NewInputArea 函數創建並返回一個新的 InputArea 實例。
func NewInputArea() InputArea {
	return InputArea{
		runes:       []rune{},
		cursor:      0,
		placeholder: "在這裡輸入…Enter 送出 / Ctrl+J 換行",
		styles:      lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Background(lipgloss.Color("235")),
	}
}

// Init 函數在 InputArea 初始化時被呼叫。
func (ia InputArea) Init() tea.Cmd {
	return nil
}

// Update 函數處理傳入的訊息並更新 InputArea 的狀態。
func (ia InputArea) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes:
			// AI 心智註解: 插入多個 rune 時保持原順序，確保多位元字元完整。
			ia.runes = insertRunes(ia.runes, ia.cursor, msg.Runes)
			ia.cursor += len(msg.Runes)
		case tea.KeyBackspace:
			if ia.cursor > 0 {
				// AI 心智註解: Backspace 按 rune 刪除，避免刪半個字元。
				ia.runes = append(ia.runes[:ia.cursor-1], ia.runes[ia.cursor:]...)
				ia.cursor--
			}
		case tea.KeyLeft:
			if ia.cursor > 0 {
				ia.cursor--
			}
		case tea.KeyRight:
			if ia.cursor < len(ia.runes) {
				ia.cursor++
			}
		case tea.KeyEnter:
			// 提交輸入
			return ia, func() tea.Msg { return SubmitMsg{Text: string(ia.runes)} }
		case tea.KeyCtrlJ:
			// AI 心智註解: Ctrl+J 也視為插入換行 rune。
			ia.runes = insertRunes(ia.runes, ia.cursor, []rune{'\n'})
			ia.cursor++
		}
	}
	return ia, nil
}

// View 函數渲染 InputArea 的視覺表示。
func (ia InputArea) View() string {
	if len(ia.runes) == 0 {
		return "> " + ia.placeholder
	}
	before := string(ia.runes[:ia.cursor])
	after := string(ia.runes[ia.cursor:])
	return "> " + before + "|" + after
}

// Text 取得當前輸入內容的字串表示，供測試使用。
func (ia InputArea) Text() string {
	// AI 心智註解: 將 rune 切片轉為字串，方便檢查最終輸入狀態。
	return string(ia.runes)
}

// insertRunes 將新 rune 插入到指定位置，回傳新的 rune 切片。
func insertRunes(base []rune, idx int, toInsert []rune) []rune {
	// AI 心智註解: 透過重新配置切片確保插入操作不污染原切片共享的底層陣列。
	result := make([]rune, 0, len(base)+len(toInsert))
	result = append(result, base[:idx]...)
	result = append(result, toInsert...)
	result = append(result, base[idx:]...)
	return result
}
