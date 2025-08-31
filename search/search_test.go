package search

import (
	"testing"

	"github.com/wtg42/ora-ora-ora/model"
)

// 目標：以 TDD 驅動 Index 介面
// 約定：
// - OpenOrCreate(path) 回傳 Index；如不存在則建立，存在則開啟。
// - IndexNote 將 Note 加入索引。
// - Query(q, topK, tags) 回傳 Snippet 陣列，尊重 topK 與 tags 過濾；當 q 為空時允許回傳空或全部，由實作決定，本測試僅覆蓋基本行為。

func TestIndex_IndexAndQuery_WithTagsAndTopK(t *testing.T) {
	dir := t.TempDir()
	idx, err := OpenOrCreate(dir)
	if err != nil {
		t.Fatalf("OpenOrCreate: %v", err)
	}
	defer func() {
		_ = idx.Close()
	}()

	notes := []model.Note{
		{ID: "a1", Content: "golang bleve search", Tags: []string{"dev"}},
		{ID: "a2", Content: "golang unit test", Tags: []string{"test", "dev"}},
		{ID: "a3", Content: "cooking recipe", Tags: []string{"life"}},
	}
	for _, n := range notes {
		if err := idx.IndexNote(n); err != nil {
			t.Fatalf("IndexNote(%s): %v", n.ID, err)
		}
	}

	// 僅取 dev tag，topK=1
	got, err := idx.Query("golang", 1, []string{"dev"})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("want topK=1, got %d", len(got))
	}
	if got[0].NoteID != "a1" && got[0].NoteID != "a2" { // 任一 dev 且含 golang 的筆記皆可
		t.Fatalf("unexpected note id: %s", got[0].NoteID)
	}
}

func TestIndex_Query_NoMatchReturnsEmpty(t *testing.T) {
	dir := t.TempDir()
	idx, err := OpenOrCreate(dir)
	if err != nil {
		t.Fatalf("OpenOrCreate: %v", err)
	}
	defer func() { _ = idx.Close() }()

	// 未索引任何資料，或查詢無匹配應回空陣列
	got, err := idx.Query("nope", 5, nil)
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("want 0, got %d", len(got))
	}
}
