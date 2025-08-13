package search

// Package search 定義全文檢索介面。此檔提供最小的記憶體實作，
// 以便在未引入 Bleve 前即可進行 CLI/TUI 端對接與測試。

import (
    "strings"

    "github.com/wtg42/ora-ora-ora/storage"
)

// Snippet 代表查詢結果的一小段摘要，供組 Prompt 當作 context。
type Snippet struct {
    NoteID     string
    Excerpt    string
    Score      float64
    TagMatches []string
}

// Index 定義檢索索引介面。
type Index interface {
    IndexNote(storage.Note) error
    Query(q string, topK int, tags []string) ([]Snippet, error)
    Close() error
}

// OpenOrCreate 回傳記憶體索引實作。之後可替換為 Bleve 實作。
func OpenOrCreate(_ string) (Index, error) { // path 先忽略
    return &memoryIndex{notes: make(map[string]storage.Note)}, nil
}

// memoryIndex 是最小可用的 in-memory 搜尋實作。
type memoryIndex struct {
    notes map[string]storage.Note
}

func (m *memoryIndex) IndexNote(n storage.Note) error {
    m.notes[n.ID] = n
    return nil
}

func (m *memoryIndex) Query(q string, topK int, tags []string) ([]Snippet, error) {
    q = strings.TrimSpace(strings.ToLower(q))
    // 簡單條件：內容包含 q；若指定 tags 則需包含任一指定標籤。
    var out []Snippet
    for _, n := range m.notes {
        if q != "" && !strings.Contains(strings.ToLower(n.Content), q) {
            continue
        }
        if len(tags) > 0 {
            if !anyTagMatch(n.Tags, tags) {
                continue
            }
        }
        out = append(out, Snippet{
            NoteID:     n.ID,
            Excerpt:    excerpt(n.Content, q, 80),
            Score:      1.0, // in-memory 先固定分數
            TagMatches: intersect(n.Tags, tags),
        })
        if topK > 0 && len(out) >= topK {
            break
        }
    }
    return out, nil
}

func (m *memoryIndex) Close() error { return nil }

func anyTagMatch(a, b []string) bool {
    if len(a) == 0 || len(b) == 0 {
        return false
    }
    set := make(map[string]struct{}, len(a))
    for _, t := range a {
        set[strings.ToLower(t)] = struct{}{}
    }
    for _, t := range b {
        if _, ok := set[strings.ToLower(t)]; ok {
            return true
        }
    }
    return false
}

func intersect(a, b []string) []string {
    if len(a) == 0 || len(b) == 0 {
        return nil
    }
    set := make(map[string]struct{}, len(a))
    for _, t := range a {
        set[strings.ToLower(t)] = struct{}{}
    }
    var out []string
    for _, t := range b {
        if _, ok := set[strings.ToLower(t)]; ok {
            out = append(out, t)
        }
    }
    return out
}

func excerpt(content, q string, width int) string {
    if q == "" || width <= 0 {
        if len(content) <= width || width <= 0 {
            return content
        }
        if width > 0 && len(content) > width {
            return content[:width]
        }
    }
    lower := strings.ToLower(content)
    ql := strings.ToLower(q)
    idx := strings.Index(lower, ql)
    if idx < 0 {
        if len(content) > width && width > 0 {
            return content[:width]
        }
        return content
    }
    start := idx - width/2
    if start < 0 {
        start = 0
    }
    end := start + width
    if end > len(content) {
        end = len(content)
    }
    return content[start:end]
}

