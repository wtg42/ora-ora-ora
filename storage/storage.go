package storage

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

