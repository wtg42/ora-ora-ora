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

// ListNotes 讀取資料目錄中所有 .md 檔案，並從檔案名稱解析筆記標題，返回標題列表。
func ListNotes() ([]string, error) {
	// 獲取資料目錄的路徑。
	dataDir, err := GetDataDir()
	if err != nil {
		return nil, fmt.Errorf("獲取資料目錄失敗: %w", err)
	}

	// 讀取資料目錄中的所有檔案。
	files, err := os.ReadDir(dataDir)
	if err != nil {
		return nil, fmt.Errorf("讀取資料目錄失敗: %w", err)
	}

	var titles []string
	for _, file := range files {
		// 只處理非目錄且以 .md 結尾的檔案。
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
			// 從檔案名稱解析標題：YYYYMMDDHHmmss-Title.md -> Title
			name := strings.TrimSuffix(file.Name(), ".md")
			parts := strings.SplitN(name, "-", 2)
			if len(parts) == 2 {
				title := parts[1]
				titles = append(titles, title)
			}
		}
	}

	return titles, nil
}

// ReadNote 根據給定的筆記標題，讀取對應的 .md 檔案，解析並移除 YAML front matter，返回筆記的純內容。
func ReadNote(title string) (string, error) {
	// 獲取資料目錄的路徑。
	dataDir, err := GetDataDir()
	if err != nil {
		return "", fmt.Errorf("獲取資料目錄失敗: %w", err)
	}

	// 讀取資料目錄中的所有檔案，尋找匹配的檔案。
	files, err := os.ReadDir(dataDir)
	if err != nil {
		return "", fmt.Errorf("讀取資料目錄失敗: %w", err)
	}

	var filePath string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
			name := strings.TrimSuffix(file.Name(), ".md")
			parts := strings.SplitN(name, "-", 2)
			if len(parts) == 2 && parts[1] == title {
				filePath = filepath.Join(dataDir, file.Name())
				break
			}
		}
	}

	if filePath == "" {
		return "", fmt.Errorf("找不到標題為 %s 的筆記", title)
	}

	// 讀取檔案內容。
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("讀取檔案 %s 失敗: %w", filePath, err)
	}
	content := string(contentBytes)

	// 解析並移除 YAML front matter。
	// Front matter 位於第一個 --- 和第二個 --- 之間。
	start := strings.Index(content, "---")
	if start == -1 {
		return "", fmt.Errorf("檔案格式錯誤：缺少 front matter 起始標記")
	}
	end := strings.Index(content[start+3:], "---")
	if end == -1 {
		return "", fmt.Errorf("檔案格式錯誤：缺少 front matter 結束標記")
	}
	end += start + 3 + 3 // adjust for the second ---

	// 移除 front matter 和前後的換行。
	pureContent := content[end:]
	pureContent = strings.TrimLeft(pureContent, "\n")

	return pureContent, nil
}
