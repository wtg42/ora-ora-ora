package storage

// Package storage 定義筆記儲存層的資料模型與介面。
// 初版提供 InMemoryStorage 便於開發與測試，未來可替換為檔案或資料庫實作。

import (
    "crypto/rand"
    "fmt"
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

// NewID 產生 UUIDv4 字串，避免額外依賴。
// 注意：此為最小實作，僅供產生唯一識別用。
func NewID() (string, error) {
    var b [16]byte
    if _, err := rand.Read(b[:]); err != nil {
        return "", err
    }
    // 設定 UUIDv4 版本與變體位元
    b[6] = (b[6] & 0x0f) | 0x40
    b[8] = (b[8] & 0x3f) | 0x80
    return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
        uint32(b[0])<<24|uint32(b[1])<<16|uint32(b[2])<<8|uint32(b[3]),
        uint16(b[4])<<8|uint16(b[5]),
        uint16(b[6])<<8|uint16(b[7]),
        uint16(b[8])<<8|uint16(b[9]),
        uint64(b[10])<<40|uint64(b[11])<<32|uint64(b[12])<<24|uint64(b[13])<<16|uint64(b[14])<<8|uint64(b[15]),
    ), nil
}
