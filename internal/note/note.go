// Package note 提供了筆記相關的資料結構和操作。
package note

import (
	"time"
)

// Note 結構體代表應用程式中的單一筆記條目。
type Note struct {
	Title     string    `json:"title"`       // 筆記的標題。
	Content   string    `json:"content"`     // 筆記的內容。
	Tags      []string  `json:"tags,omitempty"` // 筆記的標籤，可選。
	CreatedAt time.Time `json:"created_at"`  // 筆記的建立時間。
}

// NewNote 函數建立一個新的 Note 實例。
// 它接受標題、內容和標籤，並將建立時間設定為當前時間。
func NewNote(title, content string, tags []string) *Note {
	return &Note{
		Title:     title,
		Content:   content,
		Tags:      tags,
		CreatedAt: time.Now(),
	}
}
