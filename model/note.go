package model

import "time"

// Note 是跨模組的單一真實來源（Single Source of Truth）。
// 專案中的 search、storage、tui、cmd 等所有模組，對外溝通一律使用 model.Note。
// 若個別模組需要不同的內部表示（例如序列化欄位命名），請在模組內部自行轉換，避免洩漏新型別。
type Note struct {
	ID        string
	Content   string
	Tags      []string
	CreatedAt time.Time
	UpdatedAt time.Time
}
