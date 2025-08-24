package storage

import (
    "context"
)

// jsonlStorage 為 JSONL 檔案後端的最小骨架實作。
// 實作細節（路徑、檔名、錯誤處理）請見下方 TODO。
type jsonlStorage struct {
    baseDir string
}

// Save 將單筆 Note 以 JSONL 方式寫入。
// 建議行為：
// - 目的路徑：<baseDir>/YYYY-MM-DD.jsonl（依 note.CreatedAt）
// - 旗標：os.O_CREATE|os.O_APPEND；自動建立目錄
// - 錯誤處理：將 I/O/序列化錯誤包裝回傳，供上層提示
// TODO: 實作寫入邏輯（建立資料夾、編碼 JSON、逐行寫入）。
func (s *jsonlStorage) Save(note Note) error { // TODO: implement
    return nil
}

// List 讀取所有 JSONL 檔案並反序列化為 Note 列表。
// 建議行為：
// - 掃描 baseDir 下所有 *.jsonl，依時間或寫入順序確保穩定性
// - 單行壞資料：建議直接回傳錯誤（或以選項決定是否跳過）
// - 回傳完整 Note 列表
// TODO: 實作讀取與反序列化邏輯（遍歷檔案、逐行 decode）。
func (s *jsonlStorage) List() ([]Note, error) { // TODO: implement
    return nil, nil
}

// 可選：保留 context 擴充點，以便未來支援取消/逾時或背景作業。
// 目前未使用，僅示意未來擴充介面位置。
var _ = context.TODO

