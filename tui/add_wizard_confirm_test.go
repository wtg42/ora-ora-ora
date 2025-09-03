package tui_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/wtg42/ora-ora-ora/model"
	"github.com/wtg42/ora-ora-ora/tui"
)

// fakeSaver implements the AddWizard noteSaver dependency using model.Note.
type fakeSaver struct {
	last model.Note
	err  error
}

func (f *fakeSaver) Save(n model.Note) error {
	f.last = n
	return f.err
}

// fakeIndexer implements the AddWizard noteIndexer dependency using model.Note.
type fakeIndexer struct{ err error }

func (f *fakeIndexer) IndexNote(n model.Note) error { return f.err }

// drive applies a sequence of tea.Msg to the model and returns the updated model and last cmd.
func drive(m tea.Model, msgs ...tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	for _, msg := range msgs {
		m, cmd = m.Update(msg)
	}
	return m, cmd
}

func TestAddWizard_Confirm_OK_SavesAndQuits(t *testing.T) {
	fs := &fakeSaver{}
	fi := &fakeIndexer{}
	m := tui.NewAddWizardModelWithDeps(fs, fi)

	// allocate sizes
	mm, _ := drive(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m = mm.(tui.AddWizardModel)

	// user enters content with tags, then Enter to summarize
	mm, _ = drive(m,
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("研究 Bleve 搜尋 #Dev tags: search,go")},
		tea.KeyMsg{Type: tea.KeyEnter},
	)
	m = mm.(tui.AddWizardModel)
	v := m.View()
	if !strings.Contains(v, "理解到的新增參數") {
		t.Fatalf("expected summary after first enter")
	}

	// confirm OK to save, run returned command to simulate async save
	mm, cmd := drive(m,
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ok")},
		tea.KeyMsg{Type: tea.KeyEnter},
	)
	m = mm.(tui.AddWizardModel)
	if cmd == nil {
		t.Fatalf("expected a save command on OK confirm")
	}
	// Run command to get resulting message and feed back into Update
	msg := cmd()
	mm, _ = m.Update(msg)
	m = mm.(tui.AddWizardModel)

	// The view should show a saved message (and model will quit on next frame)
	out := m.View()
	// 視圖可能在 Quit 前為空（本實作在收到 Saved 後立即 Quit），允許任一情況
	if !(out == "" || strings.Contains(out, "Saved note with ID:")) {
		t.Fatalf("expected saved message or quit view, got: %q", out)
	}

	// Saver should have received a reasonable note
	assert.NotEmpty(t, fs.last.ID)
	assert.Equal(t, "研究 Bleve 搜尋", fs.last.Content)
	// tags are lower-cased in parser; order not guaranteed
	got := map[string]bool{}
	for _, tg := range fs.last.Tags {
		got[tg] = true
	}
	assert.True(t, got["dev"])    // from #Dev
	assert.True(t, got["search"]) // from tags: search,go
	assert.True(t, got["go"])     // from tags: search,go
	assert.WithinDuration(t, time.Now(), fs.last.CreatedAt, time.Second)
}

func TestAddWizard_Confirm_Cancel_Quits(t *testing.T) {
	m := tui.NewAddWizardModelWithDeps(&fakeSaver{}, &fakeIndexer{})
	// summarize first
	mm, _ := drive(m,
		tea.WindowSizeMsg{Width: 80, Height: 24},
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("demo #x")},
		tea.KeyMsg{Type: tea.KeyEnter},
	)
	m = mm.(tui.AddWizardModel)

	// cancel
	mm, _ = drive(m,
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("cancel")},
		tea.KeyMsg{Type: tea.KeyEnter},
	)
	m = mm.(tui.AddWizardModel)
	// model quits → View becomes empty string
	assert.Equal(t, "", m.View())
}

func TestAddWizard_Save_Error_ShowsMessage_Stay(t *testing.T) {
	fs := &fakeSaver{err: errors.New("disk full")}
	fi := &fakeIndexer{}
	m := tui.NewAddWizardModelWithDeps(fs, fi)

	// prepare summary stage
	mm, _ := drive(m,
		tea.WindowSizeMsg{Width: 80, Height: 24},
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("abc #t")},
		tea.KeyMsg{Type: tea.KeyEnter},
	)
	m = mm.(tui.AddWizardModel)

	// confirm OK, get command
	mm, cmd := drive(m,
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ok")},
		tea.KeyMsg{Type: tea.KeyEnter},
	)
	m = mm.(tui.AddWizardModel)
	if cmd == nil {
		t.Fatalf("expected save command on OK confirm")
	}
	// run command and feed error back
	msg := cmd()
	mm, _ = m.Update(msg)
	m = mm.(tui.AddWizardModel)

	v := m.View()
	assert.Contains(t, v, "Error: disk full")
	// should not quit (view not empty and prompt visible)
	assert.NotEqual(t, "", v)
	assert.Contains(t, v, "> ")
}
