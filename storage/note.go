package storage

import "time"

// 注意：此 Note 僅供 storage 模組內部序列化用。
// 專案對外型別請一律使用 model.Note（見 model 套件），避免型別分裂。
// 之所以保留 storage.Note，是為了在 JSONL 或其他後端儲存時使用不同欄位命名與內部最佳化。
// 任何跨模組邊界都應以轉接層或轉型，輸出/輸入 model.Note。
// JSON 欄位命名採用 snake_case 以便與 JSONL 檔案一致。
type Note struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
