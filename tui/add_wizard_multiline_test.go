package tui_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wtg42/ora-ora-ora/tui"
)

// drive helper is reused from other tests when available; redefine minimal here.
func driveModel(m tea.Model, msgs ...tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	for _, msg := range msgs {
		m, cmd = m.Update(msg)
	}
	return m, cmd
}

func TestAddWizard_Multiline_AltEnterThenSend(t *testing.T) {
	m := tui.NewAddWizardModel()
	mm, _ := driveModel(m, tea.WindowSizeMsg{Width: 80, Height: 20})
	m = mm.(tui.AddWizardModel)

	// type 'hello', Ctrl+j to insert newline, then 'world', then Enter to summarize
	mm, _ = driveModel(m,
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("hello")},
		tea.KeyMsg{Type: tea.KeyCtrlJ}, // newline
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("world")},
		tea.KeyMsg{Type: tea.KeyEnter}, // send
	)
	m = mm.(tui.AddWizardModel)

	v := m.View()
	if !strings.Contains(v, "理解到的新增參數") {
		t.Fatalf("expected summary")
	}
	// content should include both lines; UI 裝飾可能插入邊框與空白，放寬判定
	if !(strings.Contains(v, "內容: hello") && strings.Contains(v, "world")) {
		t.Fatalf("expected multiline content parts in view, got: %q", v)
	}
}

func TestAddWizard_Tags_NormalizeAndDedupAndSort(t *testing.T) {
	m := tui.NewAddWizardModel()
	mm, _ := driveModel(m, tea.WindowSizeMsg{Width: 80, Height: 20})
	m = mm.(tui.AddWizardModel)

	// Include duplicates and different cases
	input := "demo #Dev #dev tags: DEV, go, dev"
	mm, _ = driveModel(m,
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(input)},
		tea.KeyMsg{Type: tea.KeyEnter},
	)
	m = mm.(tui.AddWizardModel)
	v := m.View()
	// After normalization and sorting, expect 'dev, go'
	if !strings.Contains(v, "標籤: dev, go") {
		t.Fatalf("expected normalized and sorted tags, got: %q", v)
	}
}

func TestAddWizard_Tags_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "chinese tags",
			input:    "學習 tags: chinese,learn,program",
			expected: "chinese, learn, program",
		},
		{
			name:     "special chars in tags",
			input:    "#test_123 #test-456 tags: test_789,test-000",
			expected: "test-000, test-456, test_123, test_789",
		},
		{
			name:     "empty tags",
			input:    "content only # tags: , ,",
			expected: "(無)",
		},
		{
			name:     "multiple tags sections",
			input:    "#first tags: second,third 標籤: fourth",
			expected: "first, fourth, second, third",
		},
		{
			name:     "mixed hash and tags",
			input:    "#a tags: b,c #d",
			expected: "a, b, c, d",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tui.NewAddWizardModel()
			mm, _ := driveModel(m, tea.WindowSizeMsg{Width: 80, Height: 20})
			m = mm.(tui.AddWizardModel)

			mm, _ = driveModel(m,
				tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.input)},
				tea.KeyMsg{Type: tea.KeyEnter},
			)
			m = mm.(tui.AddWizardModel)
			v := m.View()
			if !strings.Contains(v, "標籤: "+tt.expected) {
				t.Fatalf("expected tags '%s', got view: %q", tt.expected, v)
			}
		})
	}
}
