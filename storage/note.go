package storage

import "time"

// Note 表示一則筆記的資料模型。
// 依文件契約：
// - ID: UUIDv4
// - Content: 內文
// - Tags: 標籤列表
// - CreatedAt/UpdatedAt: 時間戳
// JSON 欄位命名採用 snake_case 以便與 JSONL 檔案一致。
type Note struct {
    ID        string    `json:"id"`
    Content   string    `json:"content"`
    Tags      []string  `json:"tags"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

