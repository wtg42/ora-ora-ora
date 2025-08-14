package tui

import (
	"testing"

	
	tea "github.com/charmbracelet/bubbletea"
)

func TestAddNote_Update(t *testing.T) {
	// 1. Init model
	a := NewAddNote()
	a.textInput.SetValue("hello #world")

	// 2. Send Enter key
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	model, cmd := a.Update(msg)

	// 3. Assert state
	updatedModel, ok := model.(AddNote)
	if !ok {
		t.Fatalf("model is not AddNote")
	}

	if updatedModel.Content != "hello #world" {
		t.Errorf("want content %q, got %q", "hello #world", updatedModel.Content)
	}
	if len(updatedModel.Tags) != 1 || updatedModel.Tags[0] != "world" {
		t.Errorf("want tags %v, got %v", []string{"world"}, updatedModel.Tags)
	}

	// 4. Assert quit command
	if cmd == nil {
		t.Fatalf("want quit command, got nil")
	}
	if cmd() != tea.Quit() {
		t.Errorf("want quit command")
	}
}
