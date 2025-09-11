package storage

import (
	"fmt"

	"github.com/wtg42/ora-ora-ora/model"
)

// Storage 定義筆記儲存的介面契約（內部版本）。
// 說明：
// - 專案對外統一型別為 model.Note（見下方轉接層）。
// - 內部 storage 實作可使用 storage.Note 作為序列化結構，避免洩漏到其他模組。
// - 請避免全域狀態，並確保 I/O 錯誤能清楚回報。
type Storage interface {
	Save(note Note) error
	List() ([]Note, error)
}

// NewJSONL 建立以 JSONL 檔案為後端的 Storage 實例。
// baseDir 為資料根目錄（例如 data/notes）。
func NewJSONL(baseDir string) Storage {
	return &jsonlStorage{baseDir: baseDir}
}

// ---- 測試相容轉接層與對外契約 ----
// 某些測試與上層模組（例如 CLI/TUI）僅認 model.Note。
// 因此提供轉接層，讓外部以 model.Note 互動，內部維持 storage.Note 以符合序列化需求。

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
