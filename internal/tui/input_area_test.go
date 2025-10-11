package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestInputAreaHandlesMultibyteRunes(t *testing.T) {
	// AI 心智註解: 透過逐步輸入與刪除驗證 rune 為單位的編輯流程。
	ia := NewInputArea()

	model, _ := ia.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("漢")})
	ia = model.(InputArea)
	assert.Equal(t, "漢", ia.Text())

	// AI 心智註解: 將游標左移後插入另一個中文字，確保插入位置正確。
	model, _ = ia.Update(tea.KeyMsg{Type: tea.KeyLeft})
	ia = model.(InputArea)
	model, _ = ia.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("字")})
	ia = model.(InputArea)
	assert.Equal(t, "字漢", ia.Text())

	// AI 心智註解: 確認 Backspace 會整個刪除一個 rune，而非部分 byte。
	model, _ = ia.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	ia = model.(InputArea)
	assert.Equal(t, "漢", ia.Text())
}
