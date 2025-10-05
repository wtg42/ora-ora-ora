// Package note 提供了筆記相關的資料結構和操作的測試。
package note

import (
	"reflect"
	"testing"
	"time"
)

// TestNewNote 測試 NewNote 函數是否能正確建立 Note 實例。
func TestNewNote(t *testing.T) {
	// 定義測試用的標題、內容和標籤。
	title := "Test Note Title"
	content := "This is the content of the test note."
	tags := []string{"test", "golang"}

	// 呼叫 NewNote 函數建立一個新的筆記。
	n := NewNote(title, content, tags)

	// 檢查筆記的標題是否正確。
	if n.Title != title {
		t.Errorf("預期標題為 %q, 實際得到 %q", title, n.Title)
	}
	// 檢查筆記的內容是否正確。
	if n.Content != content {
		t.Errorf("預期內容為 %q, 實際得到 %q", content, n.Content)
	}
	// 檢查筆記的標籤是否正確。
	if !reflect.DeepEqual(n.Tags, tags) {
		t.Errorf("預期標籤為 %v, 實際得到 %v", tags, n.Tags)
	}

	// 檢查 CreatedAt 是否已設定且時間戳記是最近的。
	if n.CreatedAt.IsZero() {
		t.Error("CreatedAt 不應為零值")
	}
	// 檢查 CreatedAt 是否在合理的時間範圍內（例如，在 1 秒內）。
	if time.Since(n.CreatedAt) > 1*time.Second {
		t.Error("CreatedAt 時間戳記不夠新")
	}
}
