package search

// 注意：search 模組對外一律使用 model.Note 作為資料模型，請勿匯入 storage 直接耦合。
// 若內部需要不同儲存結構，應由 storage 轉接層自行處理，避免型別分裂。

import (
	"github.com/wtg42/ora-ora-ora/model"
	"strings"
)

// Snippet 為查詢回傳的最小片段資訊。
// 目前測試只檢查 NoteID；其他欄位未使用但保留以利未來擴充。
type Snippet struct {
	NoteID     string
	Excerpt    string
	Score      float64
	TagMatches []string
}

// Index 定義搜尋索引的對外介面契約。
// 一律使用 model.Note 作為索引輸入的統一型別，避免跨模組型別分裂。
type Index interface {
	IndexNote(model.Note) error
	Query(q string, topK int, tags []string) ([]Snippet, error)
	Close() error
}

// inMemoryIndex implements Index for testing purposes
type inMemoryIndex struct {
	notes map[string]model.Note
}

// IndexNote adds a note to the in-memory index
func (i *inMemoryIndex) IndexNote(note model.Note) error {
	if i.notes == nil {
		i.notes = make(map[string]model.Note)
	}
	i.notes[note.ID] = note
	return nil
}

// Query searches notes by content and tags, limited by topK
func (i *inMemoryIndex) Query(q string, topK int, tags []string) ([]Snippet, error) {
	var results []Snippet
	for id, note := range i.notes {
		// Check tags: must have all specified tags
		if len(tags) > 0 {
			hasAllTags := true
			for _, tag := range tags {
				found := false
				for _, ntag := range note.Tags {
					if ntag == tag {
						found = true
						break
					}
				}
				if !found {
					hasAllTags = false
					break
				}
			}
			if !hasAllTags {
				continue
			}
		}
		// Check content: split q by spaces and require all tokens to be present (case-insensitive AND)
		if qTrim := strings.TrimSpace(q); qTrim != "" {
			contentLower := strings.ToLower(note.Content)
			tokens := strings.Fields(strings.ToLower(qTrim))
			ok := true
			for _, tok := range tokens {
				if !strings.Contains(contentLower, tok) {
					ok = false
					break
				}
			}
			if !ok {
				continue
			}
		}
		results = append(results, Snippet{
			NoteID: id,
			Score:  1.0, // dummy score
		})
	}
	// Limit to topK
	if len(results) > topK {
		results = results[:topK]
	}
	return results, nil
}

// Close does nothing for in-memory index
func (i *inMemoryIndex) Close() error {
	return nil
}

// OpenOrCreate creates an in-memory index (ignores path for now)
func OpenOrCreate(path string) (Index, error) {
	return &inMemoryIndex{}, nil
}
