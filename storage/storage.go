package storage

// Package storage 定義筆記儲存層的資料模型與介面。
// 初版提供 InMemoryStorage 便於開發與測試，未來可替換為檔案或資料庫實作。

import (
    "sort"
    "strings"
    "time"
)

// Note 代表一筆使用者筆記。
// 後續檔案儲存建議為 JSONL，一行一筆。
type Note struct {
    ID        string    // 建議使用 UUIDv4（由上層產生）
    Content   string
    Tags      []string
    CreatedAt time.Time
    UpdatedAt time.Time
}

// Storage 定義儲存層最小契約。
type Storage interface {
    Save(note Note) error
    List() ([]Note, error)
}

// InMemoryStorage 為記憶體版儲存，僅用於開發與測試。
type InMemoryStorage struct {
    items map[string]Note
}

// NewInMemory 建立新的記憶體儲存。
func NewInMemory() *InMemoryStorage {
    return &InMemoryStorage{items: make(map[string]Note)}
}

// Save 覆寫或新增一筆 Note。
func (s *InMemoryStorage) Save(note Note) error {
    s.items[note.ID] = note
    return nil
}

// List 依建立時間排序回傳所有 Note（新到舊）。
func (s *InMemoryStorage) List() ([]Note, error) {
    out := make([]Note, 0, len(s.items))
    for _, n := range s.items {
        out = append(out, n)
    }
    sort.Slice(out, func(i, j int) bool {
        return out[i].CreatedAt.After(out[j].CreatedAt)
    })
    return out, nil
}

// ExtractTags 從內容中以 #tag 解析標籤，供上層或測試使用。
func ExtractTags(content string) []string {
    parts := strings.Fields(content)
    var tags []string
    for _, p := range parts {
        if strings.HasPrefix(p, "#") && len(p) > 1 {
            tags = append(tags, strings.TrimPrefix(p, "#"))
        }
    }
    return tags
}

