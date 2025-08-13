package storage

import (
    "testing"
    "time"
)

func TestInMemoryStorage_SaveAndList(t *testing.T) {
    s := NewInMemory()
    n1 := Note{ID: "1", Content: "hello #world", Tags: []string{"world"}, CreatedAt: time.Now()}
    n2 := Note{ID: "2", Content: "golang #notes", Tags: []string{"notes"}, CreatedAt: time.Now().Add(1 * time.Minute)}
    if err := s.Save(n1); err != nil { t.Fatalf("save1: %v", err) }
    if err := s.Save(n2); err != nil { t.Fatalf("save2: %v", err) }

    got, err := s.List()
    if err != nil { t.Fatalf("list: %v", err) }
    if len(got) != 2 { t.Fatalf("want 2 notes, got %d", len(got)) }
    if got[0].ID != "2" { t.Fatalf("want newest first (2), got %s", got[0].ID) }
}

func TestExtractTags(t *testing.T) {
    tags := ExtractTags("a #b c #d #中文")
    if len(tags) != 3 { t.Fatalf("want 3 tags, got %d", len(tags)) }
}

