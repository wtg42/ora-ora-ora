package storage

import (
    "fmt"

    "github.com/wtg42/ora-ora-ora/model"
)
// Storage 定義筆記儲存的介面契約。
// Save：將單筆 Note 以 JSONL 方式寫入；List：讀取所有現有筆記。
// 注意：實作應避免全域狀態，並確保 I/O 錯誤能清楚回報。
type Storage interface {
    Save(note Note) error
    List() ([]Note, error)
}

// NewJSONL 建立以 JSONL 檔案為後端的 Storage 實例。
// baseDir 為資料根目錄（例如 data/notes）。
// TODO: 由實作者完成 JSONL 儲存初始化邏輯（建立必要資料夾、參數檢查等）。
func NewJSONL(baseDir string) Storage { // TODO: return concrete implementation
    return &jsonlStorage{baseDir: baseDir}
}

// ---- 測試相容轉接層 ----
// 某些既有測試使用 model.Note 與 New(dir) 介面。
// 這裡提供轉接，以不影響內部 Storage 介面的前提下滿足測試。

type jsonlAdapter struct{ inner *jsonlStorage }

// New 建立轉接實例，回傳具 Save/List 的物件（以 model.Note 為型別）。
func New(baseDir string) (*jsonlAdapter, error) {
    if baseDir == "" {
        return nil, fmt.Errorf("baseDir is empty")
    }
    return &jsonlAdapter{inner: &jsonlStorage{baseDir: baseDir}}, nil
}

// Save 將 model.Note 轉為 storage.Note 後委派給 jsonlStorage。
func (a *jsonlAdapter) Save(n model.Note) error {
    if a == nil || a.inner == nil {
        return fmt.Errorf("jsonl adapter not initialized")
    }
    sn := Note{
        ID:        n.ID,
        Content:   n.Content,
        Tags:      append([]string(nil), n.Tags...),
        CreatedAt: n.CreatedAt,
        UpdatedAt: n.UpdatedAt,
    }
    return a.inner.Save(sn)
}

// List 讀取 storage.Note 陣列並轉回 model.Note。
func (a *jsonlAdapter) List() ([]model.Note, error) {
    if a == nil || a.inner == nil {
        return nil, fmt.Errorf("jsonl adapter not initialized")
    }
    got, err := a.inner.List()
    if err != nil {
        return nil, err
    }
    out := make([]model.Note, 0, len(got))
    for _, n := range got {
        out = append(out, model.Note{
            ID:        n.ID,
            Content:   n.Content,
            Tags:      append([]string(nil), n.Tags...),
            CreatedAt: n.CreatedAt,
            UpdatedAt: n.UpdatedAt,
        })
    }
    return out, nil
}
