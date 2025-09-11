package search

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/wtg42/ora-ora-ora/model"
)

// 目標：以 TDD 驅動 Index 介面
// 約定：
// - OpenOrCreate(path) 回傳 Index；如不存在則建立，存在則開啟。
// - IndexNote 將 Note 加入索引。
// - Query(q, topK, tags) 回傳 Snippet 陣列，尊重 topK 與 tags 過濾；當 q 為空時允許回傳空或全部，由實作決定，本測試僅覆蓋基本行為。

func TestIndex_IndexAndQuery_WithTagsAndTopK(t *testing.T) {
	tmpDir := t.TempDir()
	indexPath := filepath.Join(tmpDir, "test.bleve")
	idx, err := OpenOrCreate(indexPath)
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
	tmpDir := t.TempDir()
	indexPath := filepath.Join(tmpDir, "test.bleve")
	idx, err := OpenOrCreate(indexPath)
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

func TestBleveIndex_IndexAndQuery_WithTagsAndTopK(t *testing.T) {
	// Create temporary directory for Bleve index
	tmpDir := t.TempDir()
	indexPath := filepath.Join(tmpDir, "test.bleve")

	index, err := OpenOrCreate(indexPath)
	if err != nil {
		t.Fatalf("OpenOrCreate failed: %v", err)
	}
	defer index.Close()

	// Index some notes
	notes := []model.Note{
		{ID: "a1", Content: "golang bleve search", Tags: []string{"dev"}, CreatedAt: time.Now()},
		{ID: "a2", Content: "golang unit test", Tags: []string{"test", "dev"}, CreatedAt: time.Now()},
		{ID: "a3", Content: "python web development", Tags: []string{"web"}, CreatedAt: time.Now()},
	}

	for _, note := range notes {
		if err := index.IndexNote(note); err != nil {
			t.Fatalf("IndexNote failed: %v", err)
		}
	}

	// Test query with content
	results, err := index.Query("golang", 10, nil)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
	expectedIDs := map[string]bool{"a1": true, "a2": true}
	for _, r := range results {
		if !expectedIDs[r.NoteID] {
			t.Errorf("Unexpected NoteID: %s", r.NoteID)
		}
	}

	// Test query with tags
	results, err = index.Query("", 10, []string{"dev"})
	if err != nil {
		t.Fatalf("Query with tags failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results with dev tag, got %d", len(results))
	}

	// Test query with content and tags
	results, err = index.Query("bleve", 10, []string{"dev"})
	if err != nil {
		t.Fatalf("Query with content and tags failed: %v", err)
	}
	if len(results) != 1 || results[0].NoteID != "a1" {
		t.Errorf("Expected 1 result with bleve and dev, got %v", results)
	}

	// Test topK limit
	results, err = index.Query("golang", 1, nil)
	if err != nil {
		t.Fatalf("Query with topK failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result with topK=1, got %d", len(results))
	}
}

func TestBleveIndex_Query_NoMatchReturnsEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	indexPath := filepath.Join(tmpDir, "test.bleve")

	index, err := OpenOrCreate(indexPath)
	if err != nil {
		t.Fatalf("OpenOrCreate failed: %v", err)
	}
	defer index.Close()

	// Index a note
	note := model.Note{ID: "a1", Content: "golang bleve", Tags: []string{"dev"}, CreatedAt: time.Now()}
	if err := index.IndexNote(note); err != nil {
		t.Fatalf("IndexNote failed: %v", err)
	}

	// Query for non-existent content
	results, err := index.Query("nonexistent", 10, nil)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results for nonexistent query, got %d", len(results))
	}

	// Query for non-existent tag
	results, err = index.Query("", 10, []string{"nonexistent"})
	if err != nil {
		t.Fatalf("Query with nonexistent tag failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results for nonexistent tag, got %d", len(results))
	}
}

func TestOpenOrCreate_InMemory(t *testing.T) {
	index, err := OpenOrCreate("")
	if err != nil {
		t.Fatalf("OpenOrCreate with empty path failed: %v", err)
	}
	defer index.Close()

	// Should be inMemoryIndex
	if _, ok := index.(*inMemoryIndex); !ok {
		t.Errorf("Expected inMemoryIndex, got %T", index)
	}
}

func TestOpenOrCreate_Bleve(t *testing.T) {
	tmpDir := t.TempDir()
	indexPath := filepath.Join(tmpDir, "test.bleve")

	index, err := OpenOrCreate(indexPath)
	if err != nil {
		t.Fatalf("OpenOrCreate with path failed: %v", err)
	}
	defer index.Close()

	// Should be bleveIndex
	if _, ok := index.(*bleveIndex); !ok {
		t.Errorf("Expected bleveIndex, got %T", index)
	}

	// Check if index file exists
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Errorf("Bleve index file not created")
	}
}
