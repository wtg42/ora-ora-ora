package core

import (
	"github.com/wtg42/ora-ora-ora/search"
	"testing"
)

func TestFallbackFromSnippets(t *testing.T) {
	snips := []search.Snippet{
		{NoteID: "a", Excerpt: "內容A"},
		{NoteID: "b", Excerpt: "內容B"},
	}
	got := fallbackFromSnippets(snips, 2)
	if got == "" || got == "未取得 LLM 回覆。" {
		t.Fatalf("fallback should include snippets, got: %q", got)
	}
}
