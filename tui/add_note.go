package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// AddNote is a bubbletea model for adding a new note.
type AddNote struct {
	textInput textinput.Model
	quitting  bool
	// Public fields to be accessed after the bubbletea program runs
	Content string
	Tags    []string
}

// NewAddNote initializes a new model for adding a note.
func NewAddNote() AddNote {

	ti := textinput.New()

	ti.Placeholder = "Enter your note, use # for tags..."
	ti.Focus()
	ti.CharLimit = 2048
	ti.Width = 80 // A default width

	return AddNote{
		textInput: ti,
	}
}

// Init command for the model.
func (a AddNote) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles incoming messages.
func (a AddNote) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			// When Enter is pressed, save the content and quit.
			a.Content = a.textInput.Value()
			// Simple tag parsing: words starting with #
			for _, word := range strings.Fields(a.Content) {
				if strings.HasPrefix(word, "#") {
					a.Tags = append(a.Tags, strings.TrimPrefix(word, "#"))
				}
			}
			a.quitting = true
			return a, tea.Quit

		case tea.KeyCtrlC, tea.KeyEsc:
			// Quit without saving.
			a.quitting = true
			return a, tea.Quit
		}
	// Handle window size changes
	case tea.WindowSizeMsg:
		a.textInput.Width = msg.Width - 4 // Adjust for padding
	}

	// Update the text input component
	a.textInput, cmd = a.textInput.Update(msg)
	return a, cmd
}

// View renders the UI.
func (a AddNote) View() string {
	if a.quitting {
		// Don't render anything when quitting. The main program can
		// access the Content and Tags fields.
		return ""
	}

	return fmt.Sprintf(
		"Enter your note below. Use #hashtags to add tags.\n\n%s\n\n(Press Enter to save, Esc to quit)",
		a.textInput.View(),
	)
}
