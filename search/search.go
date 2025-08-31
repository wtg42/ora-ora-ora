package search

// 注意：search 模組對外一律使用 model.Note 作為資料模型，請勿匯入 storage 直接耦合。
// 若內部需要不同儲存結構，應由 storage 轉接層自行處理，避免型別分裂。

import (
	"fmt"
	"github.com/wtg42/ora-ora-ora/model"
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

// OpenOrCreate 依路徑建立或開啟索引。
// 目前為使用者自行實作的骨架；此處僅保留簽名與說明，不變更你的行為。
func OpenOrCreate(path string) (Index, error) {
	if len(path) <= 0 {
		fmt.Printf("%s", path)
		return nil, nil
	}	
	return nil, nil
}
