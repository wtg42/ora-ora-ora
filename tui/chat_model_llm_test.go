package tui

import (
    "context"
    "testing"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/wtg42/ora-ora-ora/agent"
    "github.com/wtg42/ora-ora-ora/model"
    "github.com/wtg42/ora-ora-ora/search"
)

// fakeLLM implements agent.LLM for testing, returning a fixed answer.
type fakeLLM struct{ answer string }

func (f fakeLLM) Chat(_ context.Context, _ string, _ string, _ agent.Options) (string, error) {
    return f.answer, nil
}

// mockIndex implements search.Index returning one snippet.
type mockIndex struct{}

func (m mockIndex) IndexNote(_ model.Note) error { return nil }
func (m mockIndex) Query(q string, topK int, tags []string) ([]search.Snippet, error) {
    return []search.Snippet{{NoteID: "n1", Excerpt: "e1"}}, nil
}
func (m mockIndex) Close() error { return nil }

// drainCmd executes a tea.Cmd and feeds any returned messages back into Update.
func drainCmd(mod tea.Model, cmd tea.Cmd) tea.Model {
    if cmd == nil { return mod }
    msg := cmd()
    switch v := msg.(type) {
    case nil:
        return mod
    case tea.BatchMsg:
        for _, m := range v {
            mod, _ = mod.Update(m)
        }
        return mod
    default:
        mod, _ = mod.Update(v)
        return mod
    }
}

func TestChat_LLM_On_AppendsAnswer(t *testing.T) {
    m := NewChatModel()
    // enable LLM and inject fakes
    m.llmEnabled = true
    m.llmProvider = func(host, model string) agent.LLM { return fakeLLM{answer: "fake answer"} }
    m.indexProvider = func() (search.Index, error) { return mockIndex{}, nil }

    // 直接執行檢索與 LLM 命令，並把訊息回拋到 Update
    cmd1 := m.queryAndAppend("hello")
    if msg := cmd1(); msg != nil {
        mod, _ := m.Update(msg)
        m = mod.(ChatModel)
    }
    cmd2 := m.maybeLLM("hello")
    if msg := cmd2(); msg != nil {
        switch v := msg.(type) {
        case tea.BatchMsg:
            for _, mmg := range v { mod, _ := m.Update(mmg); m = mod.(ChatModel) }
        default:
            mod, _ := m.Update(v); m = mod.(ChatModel)
        }
    }

    // assert one of assistant messages包含 fake answer
    mm := m
    if len(mm.messages) == 0 {
        t.Fatalf("no messages appended")
    }
    found := false
    for _, msg := range mm.messages {
        if msg.Role == "assistant" && msg.Content == "fake answer" {
            found = true
            break
        }
    }
    if !found {
        t.Fatalf("expected assistant message with fake answer; got %+v", mm.messages)
    }
}
