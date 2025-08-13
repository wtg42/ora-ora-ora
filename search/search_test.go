package search

import (
    "testing"
    "time"

    "github.com/wtg42/ora-ora-ora/storage"
)

func TestMemoryIndex_QueryByTextAndTag(t *testing.T) {
    idx, err := OpenOrCreate("")
    if err != nil { t.Fatalf("open: %v", err) }
    defer idx.Close()

    notes := []storage.Note{
        {ID: "1", Content: "learn bleve and golang", Tags: []string{"golang"}, CreatedAt: time.Now()},
        {ID: "2", Content: "write tests for search", Tags: []string{"test"}, CreatedAt: time.Now()},
    }
    for _, n := range notes {
        if err := idx.IndexNote(n); err != nil { t.Fatalf("index: %v", err) }
    }

    // 關鍵字
    rs, err := idx.Query("golang", 10, nil)
    if err != nil { t.Fatalf("query: %v", err) }
    if len(rs) != 1 || rs[0].NoteID != "1" { t.Fatalf("want 1 result for golang") }

    // 標籤
    rs, err = idx.Query("", 10, []string{"test"})
    if err != nil { t.Fatalf("query tag: %v", err) }
    if len(rs) != 1 || rs[0].NoteID != "2" { t.Fatalf("want 1 result tag=test") }
}

