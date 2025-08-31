package storage

import (
	"testing"
	"time"

	"github.com/wtg42/ora-ora-ora/model"
)

// 目標：以 TDD 驅動 Storage（JSONL）最小行為
// 約定：
// - New(dir) 建立檔案儲存實例；Save 追加到 `YYYY-MM-DD.jsonl`。
// - List 回傳所有 Note（不強制排序，本測試只檢查數量與內容存在）。

func TestStorage_SaveAndList(t *testing.T) {
	dir := t.TempDir()
	s, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	now := time.Now().UTC()
	notes := []model.Note{
		{ID: "n1", Content: "hello world", Tags: []string{"work"}, CreatedAt: now, UpdatedAt: now},
		{ID: "n2", Content: "second line", Tags: []string{"life", "work"}, CreatedAt: now.Add(time.Second), UpdatedAt: now.Add(time.Second)},
	}

	for _, n := range notes {
		if err := s.Save(n); err != nil {
			t.Fatalf("Save(%s): %v", n.ID, err)
		}
	}

	got, err := s.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != len(notes) {
		t.Fatalf("List length: want %d, got %d", len(notes), len(got))
	}

	// 驗證內容存在（不檢查順序）
	found := map[string]bool{}
	for _, n := range got {
		found[n.ID] = true
	}
	for _, n := range notes {
		if !found[n.ID] {
			t.Fatalf("missing note %s", n.ID)
		}
	}
}
