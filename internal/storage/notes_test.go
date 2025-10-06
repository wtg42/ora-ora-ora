// Package storage 提供了筆記儲存功能的單元測試。
package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wtg42/ora-ora-ora/internal/note"
)

// TestSaveNote 測試 SaveNote 函數的各種情境，包括成功儲存、標題非法字元和空標題。
func TestSaveNote(t *testing.T) {
	// 設定一個臨時的資料目錄用於測試，確保測試環境的隔離性。
	tempDir := filepath.Join(os.TempDir(), "test-ora-data", fmt.Sprintf("savenote-%d", time.Now().UnixNano()))
	t.Cleanup(func() {
		os.RemoveAll(tempDir) // 清理臨時目錄
	})

	// 定義一系列測試案例。
	testCases := []struct {
		name        string
		note        *note.Note
		expectedErr string
		validate    func(t *testing.T, filePath string)
	}{
		{
			name: "成功儲存且無標籤",
			note: &note.Note{
				Title:     "我的第一篇筆記",
				Content:   "這是我的第一篇筆記的內容。",
				CreatedAt: time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC),
				Tags:      []string{},
			},
			expectedErr: "",
			validate: func(t *testing.T, filePath string) {
				assert.FileExists(t, filePath)
				contentBytes, err := os.ReadFile(filePath)
				assert.NoError(t, err)
				content := string(contentBytes)

				assert.Contains(t, content, "---")
				assert.Contains(t, content, "title: \"我的第一篇筆記\"")
				assert.Contains(t, content, "created_at: \"2023-01-15T10:30:00Z\"")
				assert.NotContains(t, content, "tags:")
				assert.Contains(t, content, "這是我的第一篇筆記的內容。")
			},
		},
		{
			name: "成功儲存且有標籤",
			note: &note.Note{
				Title:     "帶有標籤的筆記",
				Content:   "這篇筆記帶有一些標籤。",
				CreatedAt: time.Date(2023, 2, 20, 14, 0, 0, 0, time.UTC),
				Tags:      []string{"go", "testing", "example"},
			},
			expectedErr: "",
			validate: func(t *testing.T, filePath string) {
				assert.FileExists(t, filePath)
				contentBytes, err := os.ReadFile(filePath)
				assert.NoError(t, err)
				content := string(contentBytes)

				assert.Contains(t, content, "title: \"帶有標籤的筆記\"")
				assert.Contains(t, content, "created_at: \"2023-02-20T14:00:00Z\"")
				assert.Contains(t, content, "tags: [go, testing, example]")
				assert.Contains(t, content, "這篇筆記帶有一些標籤。")
			},
		},
		{
			name: "標題包含非法字元",
			note: &note.Note{
				Title:     "無效/標題",
				Content:   "這篇筆記不應該被儲存。",
				CreatedAt: time.Now(),
			},
			expectedErr: "標題包含非法字元，無法作為檔案名稱: 無效/標題",
			validate: func(t *testing.T, filePath string) {
				assert.NoFileExists(t, filePath) // 檔案不應該被建立
			},
		},
		{
			name: "空標題",
			note: &note.Note{
				Title:     "",
				Content:   "空標題筆記。",
				CreatedAt: time.Now(),
			},
			expectedErr: "", // 允許空標題，檔案名稱將類似 YYYYMMDDHHmmss-.md
			validate: func(t *testing.T, filePath string) {
				assert.FileExists(t, filePath)
				contentBytes, err := os.ReadFile(filePath)
				assert.NoError(t, err)
				content := string(contentBytes)
				assert.Contains(t, content, "title: \"\"")
			},
		},
	}

	// 遍歷所有測試案例並執行。
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 執行 SaveNote 函數。
			err := SaveNote(tc.note)

			// 根據預期的錯誤訊息進行斷言。
			if tc.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr)
			} else {
				assert.NoError(t, err)
				// 獲取資料目錄並構建預期的檔案路徑進行驗證。
				dataDir, err := GetDataDir()
				assert.NoError(t, err)
				filename := fmt.Sprintf("%s-%s.md", tc.note.CreatedAt.Format("20060102150405"), tc.note.Title)
				filePath := filepath.Join(dataDir, filename)
				tc.validate(t, filePath)
			}
		})
	}
}

// TestSaveNote_GetDataDirError 測試當 GetDataDir 返回錯誤時 SaveNote 的行為。
func TestSaveNote_GetDataDirError(t *testing.T) {
	// 添加清理函數以重置 testError。
	t.Cleanup(func() { testError = nil })

	// 在建立筆記實例前，設定 testError 以模擬 GetDataDir 錯誤。
	testError = fmt.Errorf("mocked GetDataDir error")

	// 建立一個筆記實例。
	n := &note.Note{
		Title:     "錯誤測試",
		Content:   "這應該會失敗。",
		CreatedAt: time.Now(),
	}

	// 執行 SaveNote 並斷言它返回錯誤。
	err := SaveNote(n)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "獲取資料目錄失敗: mocked GetDataDir error")
}

// TestSaveNote_WriteFileError 測試當寫入檔案失敗時 SaveNote 的行為。
func TestSaveNote_WriteFileError(t *testing.T) {
	// 設定一個臨時的資料目錄。
	tempDir := filepath.Join(os.TempDir(), "test-ora-data", fmt.Sprintf("writefileerror-%d", time.Now().UnixNano()))
	t.Cleanup(func() { os.RemoveAll(tempDir) })

	// 設定 testDataHome 以使用臨時目錄。
	originalTestDataHome := testDataHome
	testDataHome = tempDir
	defer func() { testDataHome = originalTestDataHome }()

	// 先呼叫 GetDataDir 建立目錄。
	dataDir, err := GetDataDir()
	assert.NoError(t, err)

	// 將資料目錄設定為只讀權限 (0555)，以模擬寫入失敗。
	err = os.Chmod(dataDir, 0555)
	assert.NoError(t, err)

	// 建立一個筆記實例。
	n := &note.Note{
		Title:     "寫入錯誤測試",
		Content:   "這應該會因為寫入權限而失敗。",
		CreatedAt: time.Now(),
	}

	// 執行 SaveNote 並斷言它返回錯誤。
	err = SaveNote(n)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "將筆記寫入檔案")
	// 確切的錯誤訊息可能因作業系統而異，因此只檢查部分訊息。
}
