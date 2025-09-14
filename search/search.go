package search

// 注意：search 模組對外一律使用 model.Note 作為資料模型，請勿匯入 storage 直接耦合。
// 若內部需要不同儲存結構，應由 storage 轉接層自行處理，避免型別分裂。

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/blevesearch/bleve/v2/search/query"
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
			NoteID:  id,
			Excerpt: makeExcerpt(note.Content, 160),
			Score:   1.0, // dummy score
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

// bleveIndex implements Index using Bleve
type bleveIndex struct {
	index bleve.Index
}

// newBleveIndexMapping creates the mapping for Bleve index
func newBleveIndexMapping() *mapping.IndexMappingImpl {
	indexMapping := bleve.NewIndexMapping()

	// Content field: full-text, participates in ranking
	contentMapping := bleve.NewTextFieldMapping()
	contentMapping.Store = true
	contentMapping.Index = true
	contentMapping.IncludeInAll = true

	// Tags field: keyword, no tokenization, for filtering
	tagsMapping := bleve.NewTextFieldMapping()
	tagsMapping.Store = true
	tagsMapping.Index = true
	tagsMapping.IncludeInAll = false
	tagsMapping.Analyzer = "keyword"

	// CreatedAt field: sortable
	createdAtMapping := bleve.NewDateTimeFieldMapping()
	createdAtMapping.Store = true
	createdAtMapping.Index = true
	createdAtMapping.IncludeInAll = false

	docMapping := bleve.NewDocumentMapping()
	docMapping.AddFieldMappingsAt("content", contentMapping)
	docMapping.AddFieldMappingsAt("tags", tagsMapping)
	docMapping.AddFieldMappingsAt("created_at", createdAtMapping)

	indexMapping.AddDocumentMapping("_default", docMapping)
	indexMapping.TypeField = "type"
	indexMapping.DefaultType = "_default"

	return indexMapping
}

// IndexNote indexes a note using Bleve
func (b *bleveIndex) IndexNote(note model.Note) error {
	doc := map[string]interface{}{
		"content":    note.Content,
		"tags":       note.Tags, // Store as array for proper indexing
		"created_at": note.CreatedAt,
	}
	return b.index.Index(note.ID, doc)
}

// Query searches using Bleve
func (b *bleveIndex) Query(q string, topK int, tags []string) ([]Snippet, error) {
	var queries []query.Query

	// Content query
	if qTrim := strings.TrimSpace(q); qTrim != "" {
		contentQuery := bleve.NewMatchQuery(qTrim)
		contentQuery.SetField("content")
		queries = append(queries, contentQuery)
	}

	// Tags query: AND all specified tags
	if len(tags) > 0 {
		for _, tag := range tags {
			tagQuery := bleve.NewTermQuery(tag)
			tagQuery.SetField("tags")
			queries = append(queries, tagQuery)
		}
	}

	var searchQuery query.Query
	if len(queries) == 0 {
		searchQuery = bleve.NewMatchAllQuery()
	} else if len(queries) == 1 {
		searchQuery = queries[0]
	} else {
		searchQuery = bleve.NewConjunctionQuery(queries...)
	}

	searchRequest := bleve.NewSearchRequest(searchQuery)
	searchRequest.Size = topK
	searchRequest.SortBy([]string{"-_score", "created_at"}) // Default relevance, then by date
    // 取回 content 以產生 excerpt
    searchRequest.Fields = []string{"content"}

	searchResult, err := b.index.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	var results []Snippet
	for _, hit := range searchResult.Hits {
        var content string
        if f, ok := hit.Fields["content"]; ok {
            if s, ok2 := f.(string); ok2 {
                content = s
            }
        }
        results = append(results, Snippet{
            NoteID:  hit.ID,
            Excerpt: makeExcerpt(content, 160),
            Score:   hit.Score,
        })
    }

	return results, nil
}

// Close closes the Bleve index
func (b *bleveIndex) Close() error {
	return b.index.Close()
}

// OpenOrCreate creates index based on path: if path is empty, use in-memory; otherwise use Bleve
func OpenOrCreate(path string) (Index, error) {
	if path == "" {
		return &inMemoryIndex{}, nil
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create index dir %s: %w", filepath.Dir(path), err)
	}

	var index bleve.Index
	var err error
	if _, err = os.Stat(path); os.IsNotExist(err) {
		// Create new index
		mapping := newBleveIndexMapping()
		index, err = bleve.New(path, mapping)
		if err != nil {
			return nil, fmt.Errorf("create bleve index at %s: %w", path, err)
		}
	} else {
		// Open existing index
		index, err = bleve.Open(path)
		if err != nil {
			return nil, fmt.Errorf("open bleve index at %s: %w", path, err)
		}
	}

	return &bleveIndex{index: index}, nil
}

// makeExcerpt 以 rune 安全方式從內容擷取前 n 個字元（中文安全），避免空白 context。
func makeExcerpt(s string, n int) string {
    s = strings.TrimSpace(s)
    if s == "" || n <= 0 {
        return ""
    }
    rs := []rune(s)
    if len(rs) <= n {
        return string(rs)
    }
    return string(rs[:n])
}
