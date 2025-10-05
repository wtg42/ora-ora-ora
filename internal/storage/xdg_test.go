// Package storage 提供了 XDG 目錄相關功能的單元測試。
package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetConfigDir 測試 GetConfigDir 函數是否能正確返回配置目錄。
func TestGetConfigDir(t *testing.T) {
	// 呼叫 GetConfigDir 函數。
	dir, err := GetConfigDir()
	// 斷言沒有錯誤發生。
	assert.NoError(t, err)
	// 斷言返回的目錄路徑不為空。
	assert.NotEmpty(t, dir)

	// 檢查目錄是否存在。
	_, err = os.Stat(dir)
	assert.NoError(t, err)

	// 檢查目錄是否為絕對路徑。
	assert.True(t, filepath.IsAbs(dir))
}

// TestGetDataDir 測試 GetDataDir 函數是否能正確返回資料目錄。
func TestGetDataDir(t *testing.T) {
	// 呼叫 GetDataDir 函數。
	dir, err := GetDataDir()
	// 斷言沒有錯誤發生。
	assert.NoError(t, err)
	// 斷言返回的目錄路徑不為空。
	assert.NotEmpty(t, dir)

	// 檢查目錄是否存在。
	_, err = os.Stat(dir)
	assert.NoError(t, err)

	// 檢查目錄是否為絕對路徑。
	assert.True(t, filepath.IsAbs(dir))
}

// TestEnsureDir 測試 ensureDir 函數是否能正確建立目錄。
func TestEnsureDir(t *testing.T) {
	// 建立臨時目錄進行測試，並在測試結束後清理。
	tempDir := filepath.Join(os.TempDir(), "test-ora-dir")
	defer os.RemoveAll(tempDir)

	// 呼叫 ensureDir 函數建立目錄。
	err := ensureDir(tempDir)
	// 斷言沒有錯誤發生。
	assert.NoError(t, err)

	// 檢查目錄是否存在。
	_, err = os.Stat(tempDir)
	assert.NoError(t, err)
}
