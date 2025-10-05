// Package storage 提供了應用程式的資料儲存功能，例如筆記的儲存和讀取。
package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wtg42/ora-ora-ora/internal/note"
)

// SaveNote 將給定的筆記儲存到資料目錄中的 Markdown 檔案。
// 檔案名稱格式為：YYYYMMDDHHmmss-Title.md。
func SaveNote(n *note.Note) error {
	// 獲取資料目錄的路徑。
	dataDir, err := GetDataDir()
	if err != nil {
		return fmt.Errorf("獲取資料目錄失敗: %w", err)
	}

	// 檢查標題中是否存在非法字元，以避免檔案命名問題。
	illegalChars := "/\\:*?\"<>|"
	if strings.ContainsAny(n.Title, illegalChars) {
		return fmt.Errorf("標題包含非法字元，無法作為檔案名稱: %s", n.Title)
	}

	// 根據筆記的建立時間和標題生成檔案名稱。
	filename := fmt.Sprintf("%s-%s.md", n.CreatedAt.Format("20060102150405"), n.Title)
	// 組合資料目錄和檔案名稱，形成完整的檔案路徑。
	filePath := filepath.Join(dataDir, filename)

	// 準備筆記內容，包含 YAML 格式的元資料和筆記本文。
	var contentBuilder strings.Builder
	contentBuilder.WriteString(fmt.Sprintf("---\n"))
	contentBuilder.WriteString(fmt.Sprintf("title: \"%s\"\n", n.Title))
	contentBuilder.WriteString(fmt.Sprintf("created_at: \"%s\"\n", n.CreatedAt.Format("2006-01-02T15:04:05Z07:00")))
	if len(n.Tags) > 0 {
		contentBuilder.WriteString(fmt.Sprintf("tags: [%s]\n", strings.Join(n.Tags, ", ")))
	}
	contentBuilder.WriteString(fmt.Sprintf("---\n\n"))
	contentBuilder.WriteString(n.Content)

	// 將筆記內容寫入檔案。
	err = os.WriteFile(filePath, []byte(contentBuilder.String()), 0644)
	if err != nil {
		return fmt.Errorf("將筆記寫入檔案 %s 失敗: %w", filePath, err)
	}

	return nil
}
