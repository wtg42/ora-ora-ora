package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wtg42/ora-ora-ora/model"
)

// Placeholders for dependencies
type noteSaver interface {
	Save(note model.Note) error
}

type noteIndexer interface {
	IndexNote(note model.Note) error
}

// NoteSavedMsg is sent when a note is successfully saved.
type NoteSavedMsg struct{ ID string }

// NoteSaveErrorMsg is sent when saving a note fails.
type NoteSaveErrorMsg struct{ Err error }

// keyMap defines the key bindings for the TUI.
type keyMap struct {
	Submit key.Binding
	Quit   key.Binding
	Tab    key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Submit, k.Quit, k.Tab}
}

// FullHelp returns keybindings for the expanded help view.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Submit, k.Quit, k.Tab}, // first column
	}
}

var keys = keyMap{
	Submit: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "save note"),
	),
	Quit: key.NewBinding(
		key.WithKeys("esc", "ctrl+c"),
		key.WithHelp("esc", "quit"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch input"),
	),
}

// Styles
var (
	titleStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	promptStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("79"))
	cursorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	containerStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1)
)

// AddNoteModel is a bubbletea model for adding a new note.
type AddNoteModel struct {
	ContentInput textinput.Model
	TagsInput    textinput.Model
	Status       string
	quitting     bool
	saver        noteSaver
	indexer      noteIndexer
	help         help.Model
	keys         keyMap
	width        int
}

// NewAddNoteModel initializes a new model for adding a note.
func NewAddNoteModel(saver noteSaver, indexer noteIndexer) AddNoteModel {
	content := textinput.New()
	content.Placeholder = "What's on your mind?"
	content.Focus()
	content.CharLimit = 4096
	content.Prompt = "> "
	content.PromptStyle = promptStyle
	content.CursorStyle = cursorStyle

	tags := textinput.New()
	tags.Placeholder = "dev,go,ai"
	tags.CharLimit = 256
	tags.Prompt = "# "
	tags.PromptStyle = promptStyle
	tags.CursorStyle = cursorStyle

	return AddNoteModel{
		ContentInput: content,
		TagsInput:    tags,
		saver:        saver,
		indexer:      indexer,
		help:         help.New(),
		keys:         keys,
	}
}

// Init command for the model.
func (m AddNoteModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles incoming messages.
func (m AddNoteModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.help.Width = msg.Width
		inputWidth := m.width - containerStyle.GetHorizontalFrameSize()
		m.ContentInput.Width = inputWidth
		m.TagsInput.Width = inputWidth
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, m.keys.Tab):
			if m.ContentInput.Focused() {
				m.ContentInput.Blur()
				m.TagsInput.Focus()
			} else {
				m.TagsInput.Blur()
				m.ContentInput.Focus()
			}
			return m, nil
		case key.Matches(msg, m.keys.Submit):
			// Validation
			if strings.TrimSpace(m.ContentInput.Value()) == "" {
				m.Status = "Error: Content cannot be empty."
				return m, nil
			}

			m.Status = "Saving..."
			note := m.FinalNote()
			return m, func() tea.Msg {
				if m.saver == nil {
					return NoteSavedMsg{ID: "note-123"}
				}
				err := m.saver.Save(note)
				if err != nil {
					return NoteSaveErrorMsg{Err: err}
				}
				return NoteSavedMsg{ID: note.ID}
			}
		}
	case NoteSavedMsg:
		m.Status = "Saved note with ID: " + msg.ID
		return m, tea.Quit

	case NoteSaveErrorMsg:
		m.Status = "Error: " + msg.Err.Error()
		return m, nil
	}

	if m.ContentInput.Focused() {
		m.ContentInput, cmd = m.ContentInput.Update(msg)
	} else {
		m.TagsInput, cmd = m.TagsInput.Update(msg)
	}

	return m, cmd
}

// View renders the UI.
func (m AddNoteModel) View() string {
	if m.quitting {
		return ""
	}

	title := titleStyle.Render("New Note")
	helpView := m.help.View(m.keys)

	content := lipgloss.JoinVertical(
		lipgloss.Top,
		title,
		m.ContentInput.View(),
		m.TagsInput.View(),
		m.Status,
	)

	container := containerStyle.Width(m.width - 2).Render(content)

	return lipgloss.JoinVertical(lipgloss.Bottom, container, helpView)
}

// FinalNote extracts the note data from the model.
func (m AddNoteModel) FinalNote() model.Note {
	tags := []string{}
	tagStr := strings.TrimSpace(m.TagsInput.Value())
	if tagStr != "" {
		tags = strings.Split(tagStr, ",")
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}
	}

	return model.Note{
		Content: m.ContentInput.Value(),
		Tags:    tags,
	}
}
