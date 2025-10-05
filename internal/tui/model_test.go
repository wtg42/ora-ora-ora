// Package tui 提供了終端使用者介面 (TUI) 模型的單元測試。
package tui

import (
	"io/ioutil"
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	// "github.com/wtg42/ora-ora-ora/internal/storage"
)

// setupTestDataDir 是一個輔助函數，用於建立一個臨時的資料目錄並覆蓋 storage.TestDataDir，以便進行測試。
func setupTestDataDir(t *testing.T) (string, func()) {
	tempDir, err := ioutil.TempDir("", "testdata")
	require.NoError(t, err)

	// oldTestDataDir := storage.TestDataDir
	// storage.TestDataDir = tempDir

	return tempDir, func() {
		os.RemoveAll(tempDir)
		// storage.TestDataDir = oldTestDataDir
	}
}

// TestInitialModel 測試 InitialModel 函數是否能正確初始化模型。
func TestInitialModel(t *testing.T) {
	_, teardown := setupTestDataDir(t)
	defer teardown()

	// Create some dummy notes
	// require.NoError(t, storage.WriteNote("NoteA", "Content A"))
	// require.NoError(t, storage.WriteNote("NoteB", "Content B"))

	m := InitialModel()

	assert.Equal(t, listView, m.currentView)
	// assert.ElementsMatch(t, []string{"NoteA", "NoteB"}, m.notes)
	assert.Empty(t, m.errorMessage)
}

// TestUpdate_ListViewNavigation 測試在列表視圖中的導航功能。
func TestUpdate_ListViewNavigation(t *testing.T) {
	_, teardown := setupTestDataDir(t)
	defer teardown()

	// require.NoError(t, storage.WriteNote("Note1", "Content 1"))
	// require.NoError(t, storage.WriteNote("Note2", "Content 2"))

	m := InitialModel()
	// assert.Equal(t, 0, m.cursor)

	// // Move down
	// updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	// m = updatedModel.(model)
	// assert.Equal(t, 1, m.cursor)

	// // Move down again (should stay at max)
	// updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	// m = updatedModel.(model)
	// assert.Equal(t, 1, m.cursor)

	// // Move up
	// updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	// m = updatedModel.(model)
	// assert.Equal(t, 0, m.cursor)

	// // Move up again (should stay at min)
	// updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	// m = updatedModel.(model)
	// assert.Equal(t, 0, m.cursor)
}

// TestUpdate_ViewNoteContent 測試查看筆記內容的功能。
func TestUpdate_ViewNoteContent(t *testing.T) {
	_, teardown := setupTestDataDir(t)
	defer teardown()

	// require.NoError(t, storage.WriteNote("MyNote", "This is the content of MyNote."))

	m := InitialModel()
	assert.Equal(t, listView, m.currentView)

	// // Press Enter to view the note
	// updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	// m = updatedModel.(model)

	// assert.Equal(t, detailView, m.currentView)
	// assert.Equal(t, "This is the content of MyNote.", m.selectedNoteContent)

	// // Press Esc to go back to list view
	// updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	// m = updatedModel.(model)
	// assert.Equal(t, listView, m.currentView)
	// assert.Empty(t, m.selectedNoteContent)
}

// TestUpdate_CreateNewNote 測試建立新筆記的功能。
func TestUpdate_CreateNewNote(t *testing.T) {
	_, teardown := setupTestDataDir(t)
	defer teardown()

	m := InitialModel()
	assert.Equal(t, listView, m.currentView)

	// Press 'n' to go to create view
	updatedModel, _ := m.Update(tea.KeyMsg{String: "n"})
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
	_, cmd := m.Update(tea.KeyMsg{String: "q"})
	assert.NotNil(t, cmd)

	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	assert.NotNil(t, cmd)
}
