// Package tui 提供了終端使用者介面 (TUI) 模型的單元測試。
package tui

import (
	"os"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wtg42/ora-ora-ora/internal/note"
	"github.com/wtg42/ora-ora-ora/internal/storage"
)

// setupTestDataDir 是一個輔助函數，用於建立一個臨時的資料目錄並覆蓋測試資料目錄，以便進行測試。
func setupTestDataDir(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "testdata")
	require.NoError(t, err)

	oldTestDataDir := storage.GetTestDataHome()
	storage.SetTestDataHome(tempDir)

	return tempDir, func() {
		os.RemoveAll(tempDir)
		storage.SetTestDataHome(oldTestDataDir)
	}
}

// writeTestNote 是一個輔助函數，用於在測試中儲存筆記。
func writeTestNote(title, content string) error {
	n := &note.Note{
		Title:     title,
		Content:   content,
		CreatedAt: time.Now(),
	}
	return storage.SaveNote(n)
} // TestInitialModel 測試 InitialModel 函數是否能正確初始化模型。
func TestInitialModel(t *testing.T) {
	_, teardown := setupTestDataDir(t)
	defer teardown()

	// Create some dummy notes
	require.NoError(t, writeTestNote("NoteA", "Content A"))
	require.NoError(t, writeTestNote("NoteB", "Content B"))

	m := InitialModel()

	assert.Equal(t, listView, m.currentView)
	assert.ElementsMatch(t, []string{"NoteA", "NoteB"}, m.notes)
	assert.Empty(t, m.errorMessage)
}

// TestUpdate_ListViewNavigation 測試在列表視圖中的導航功能。
func TestUpdate_ListViewNavigation(t *testing.T) {
	_, teardown := setupTestDataDir(t)
	defer teardown()

	require.NoError(t, writeTestNote("Note1", "Content 1"))
	require.NoError(t, writeTestNote("Note2", "Content 2"))

	m := InitialModel()
	assert.Equal(t, 0, m.cursor)

	// Move down
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updatedModel.(model)
	assert.Equal(t, 1, m.cursor)

	// Move down again (should stay at max)
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updatedModel.(model)
	assert.Equal(t, 1, m.cursor)

	// Move up
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updatedModel.(model)
	assert.Equal(t, 0, m.cursor)

	// Move up again (should stay at min)
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updatedModel.(model)
	assert.Equal(t, 0, m.cursor)
}

// TestUpdate_ViewNoteContent 測試查看筆記內容的功能。
func TestUpdate_ViewNoteContent(t *testing.T) {
	_, teardown := setupTestDataDir(t)
	defer teardown()

	require.NoError(t, writeTestNote("MyNote", "This is the content of MyNote."))

	m := InitialModel()
	assert.Equal(t, listView, m.currentView)

	// Press Enter to view the note
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(model)

	assert.Equal(t, detailView, m.currentView)
	assert.Equal(t, "This is the content of MyNote.", m.selectedNoteContent)

	// Press Esc to go back to list view
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedModel.(model)
	assert.Equal(t, listView, m.currentView)
	assert.Empty(t, m.selectedNoteContent)
}

// TestUpdate_CreateNewNote 測試建立新筆記的功能。
func TestUpdate_CreateNewNote(t *testing.T) {
	_, teardown := setupTestDataDir(t)
	defer teardown()

	m := InitialModel()
	assert.Equal(t, listView, m.currentView)

	// Press 'n' to go to create view
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = updatedModel.(model)
	assert.Equal(t, createView, m.currentView)
	assert.Empty(t, m.newNoteTitle)
	assert.Empty(t, m.newNoteContent)

	// Press Esc to go back to list view
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedModel.(model)
	assert.Equal(t, listView, m.currentView)
}

// TestUpdate_Quit 測試退出應用程式的功能。
func TestUpdate_Quit(t *testing.T) {
	m := InitialModel()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	assert.NotNil(t, cmd)

	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	assert.NotNil(t, cmd)
}

// TestUpdate_BasicKeyInput 測試基本鍵入事件處理，確保 Update 不 panic 並返回有效模型。
func TestUpdate_BasicKeyInput(t *testing.T) {
	m := InitialModel()

	// 測試基本鍵入事件 'n'
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	assert.IsType(t, model{}, updatedModel)
	assert.NotNil(t, updatedModel)
}

func TestSubmitPreservesWhitespace(t *testing.T) {
	// AI 心智註解: 驗證提交時保留使用者輸入的縮排與尾端空白。
	_, teardown := setupTestDataDir(t)
	defer teardown()

	m := InitialModel()
	m.currentView = createView
	m.inputArea = NewInputArea()

	input := "Title\n  leading\n\ntrailing  "
	updatedModel, _ := m.Update(SubmitMsg{Text: input})
	m = updatedModel.(model)

	require.Empty(t, m.errorMessage)

	content, err := storage.ReadNote("Title")
	require.NoError(t, err)
	assert.Equal(t, "  leading\n\ntrailing  ", content)
}
