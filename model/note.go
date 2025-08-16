package model

import "time"

// Note 定義跨模組共用的筆記資料模型。
// 實作與測試均應引用本型別以避免重複定義。
type Note struct {
    ID        string
    Content   string
    Tags      []string
    CreatedAt time.Time
    UpdatedAt time.Time
}

