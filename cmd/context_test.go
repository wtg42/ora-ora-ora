package cmd

import (
	"strings"
	"testing"

	"github.com/wtg42/ora-ora-ora/search"
)

func TestBuildAskContext_Basic(t *testing.T) {
	snippets := []search.Snippet{
		{NoteID: "a1", Excerpt: "hello world", TagMatches: []string{"dev"}},
		{NoteID: "a2", Excerpt: "golang test", TagMatches: []string{"test", "dev"}},
	}
	got := buildAskContext(snippets, 80)
	if !strings.Contains(got, "a1") || !strings.Contains(got, "hello world") {
		t.Fatalf("context missing snippet 1: %q", got)
	}
	if !strings.Contains(got, "a2") || !strings.Contains(got, "golang test") {
		t.Fatalf("context missing snippet 2: %q", got)
	}
}

func TestBuildAskContext_Truncate(t *testing.T) {
	snippets := []search.Snippet{{NoteID: "a1", Excerpt: strings.Repeat("x", 200)}}
	got := buildAskContext(snippets, 50)
	if len(got) > 60 { // allow a bit of overhead for ids and labels
		t.Fatalf("expected truncated context, len=%d, got=%q", len(got), got)
	}
}
