package tui_test

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/wtg42/ora-ora-ora/tui"
)

// TestAddNote_InitialState 驗證初始狀態是否正確。
func TestAddNote_InitialState(t *testing.T) {
	m := tui.NewAddNoteModel(nil, nil)

	assert.Equal(t, "What's on your mind?", m.ContentInput.Placeholder)
	assert.Equal(t, "dev,go,ai", m.TagsInput.Placeholder)
	assert.True(t, m.ContentInput.Focused())
}

// TestAddNote_Update 驗證 TUI 的狀態轉換邏輯。
func TestAddNote_Update(t *testing.T) {
	// Helper to create a model with pre-filled text
	newModelWithText := func(content, tags string) tui.AddNoteModel {
		m := tui.NewAddNoteModel(nil, nil)
		m.ContentInput.SetValue(content)
		m.TagsInput.SetValue(tags)
		return m
	}

	testCases := []struct {
		name        string
		initial     func() tui.AddNoteModel
		msg         tea.Msg
		assertCmd   func(t *testing.T, cmd tea.Cmd)
		assertModel func(t *testing.T, m tui.AddNoteModel)
	}{
		{
			name:    "Typing in content input",
			initial: func() tui.AddNoteModel { return tui.NewAddNoteModel(nil, nil) },
			msg:     tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("test")},
			assertModel: func(t *testing.T, m tui.AddNoteModel) {
				assert.Equal(t, "test", m.ContentInput.Value())
			},
		},
		{
			name:    "Switch focus with Tab",
			initial: func() tui.AddNoteModel { return tui.NewAddNoteModel(nil, nil) },
			msg:     tea.KeyMsg{Type: tea.KeyTab},
			assertModel: func(t *testing.T, m tui.AddNoteModel) {
				assert.False(t, m.ContentInput.Focused(), "ContentInput should lose focus")
				assert.True(t, m.TagsInput.Focused(), "TagsInput should gain focus")
			},
		},
		{
			name:    "Quit on Esc",
			initial: func() tui.AddNoteModel { return tui.NewAddNoteModel(nil, nil) },
			msg:     tea.KeyMsg{Type: tea.KeyEsc},
			assertCmd: func(t *testing.T, cmd tea.Cmd) {
				assert.NotNil(t, cmd)
			},
		},
		{
			name: "Trigger save on Enter",
			initial: func() tui.AddNoteModel {
				return newModelWithText("some content", "tag1")
			},
			msg: tea.KeyMsg{Type: tea.KeyEnter},
			assertCmd: func(t *testing.T, cmd tea.Cmd) {
				assert.NotNil(t, cmd, "Expected a command to be returned")
			},
			assertModel: func(t *testing.T, m tui.AddNoteModel) {
				assert.Equal(t, "Saving...", m.Status, "Status should show 'Saving...'")
			},
		},
		{
			name:    "Submit with empty content",
			initial: func() tui.AddNoteModel { return tui.NewAddNoteModel(nil, nil) },
			msg:     tea.KeyMsg{Type: tea.KeyEnter},
			assertModel: func(t *testing.T, m tui.AddNoteModel) {
				assert.Equal(t, "Error: Content cannot be empty.", m.Status)
			},
		},
		{
			name:    "Update status on save success",
			initial: func() tui.AddNoteModel { return newModelWithText("content", "tags") },
			msg:     tui.NoteSavedMsg{ID: "note-123"},
			assertModel: func(t *testing.T, m tui.AddNoteModel) {
				assert.Equal(t, "Saved note with ID: note-123", m.Status)
			},
		},
		{
			name:    "Update status on save error",
			initial: func() tui.AddNoteModel { return newModelWithText("content", "tags") },
			msg:     tui.NoteSaveErrorMsg{Err: errors.New("disk full")},
			assertModel: func(t *testing.T, m tui.AddNoteModel) {
				assert.Equal(t, "Error: disk full", m.Status)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			initialModel := tc.initial()
			updatedModel, cmd := initialModel.Update(tc.msg)

			if tc.assertCmd != nil {
				tc.assertCmd(t, cmd)
			}
			if tc.assertModel != nil {
				tc.assertModel(t, updatedModel.(tui.AddNoteModel))
			}
		})
	}
}

// TestAddNote_FinalNoteData 驗證在按下 Enter 後，從 model 中取出的資料是否正確。
func TestAddNote_FinalNoteData(t *testing.T) {
	m := tui.NewAddNoteModel(nil, nil)

	// 模擬輸入
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("some content")})
	m = updatedModel.(tui.AddNoteModel)
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updatedModel.(tui.AddNoteModel)
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("tag1, tag2")})
	m = updatedModel.(tui.AddNoteModel)

	// 取得最終資料
	finalNote := m.FinalNote()

	assert.Equal(t, "some content", finalNote.Content)
	assert.Equal(t, []string{"tag1", "tag2"}, finalNote.Tags)
}
