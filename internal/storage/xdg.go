// Package storage 提供了應用程式的資料儲存功能，例如筆記的儲存和讀取。
package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

// appName 定義了應用程式的名稱，用於構建 XDG 相容的路徑。
const appName = "ora-ora-ora"

// testConfigHome 和 testDataHome 用於測試，以覆蓋預設的 XDG 路徑。
var testConfigHome string
var testDataHome string

// testError 用於測試中模擬 GetDataDir 的錯誤。
var testError error

// GetConfigDir 返回 ~/.config + app 子目錄，用於 TOML 配置檔案。
// 如果目錄不存在，它會嘗試建立該目錄。
func GetConfigDir() (string, error) {
	var baseConfigHome string
	// 如果設定了 testConfigHome，則使用它，否則使用預設的 ~/.config。
	if testConfigHome != "" {
		baseConfigHome = testConfigHome
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("get user home dir: %w", err)
		}
		baseConfigHome = filepath.Join(home, ".config")
	}
	// 組合基礎配置目錄和應用程式名稱，形成完整的配置目錄路徑。
	dir := filepath.Join(baseConfigHome, appName)
	// 確保目錄存在，如果不存在則建立它。
	if err := ensureDir(dir); err != nil {
		return "", fmt.Errorf("config dir: %w", err)
	}
	return dir, nil
}

// GetDataDir 返回 ~/.local/share + app 子目錄 + "notes" 子目錄，用於 Markdown 資料。
// 如果目錄不存在，它會嘗試建立該目錄。
func GetDataDir() (string, error) {
	// 如果設定了 testError，則返回錯誤（用於測試）。
	if testError != nil {
		return "", testError
	}
	var baseDataHome string
	// 如果設定了 testDataHome，則使用它，否則使用預設的 ~/.local/share。
	if testDataHome != "" {
		baseDataHome = testDataHome
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("get user home dir: %w", err)
		}
		baseDataHome = filepath.Join(home, ".local", "share")
	}
	// 組合基礎資料目錄、應用程式名稱和 "notes" 子目錄，形成完整的資料目錄路徑。
	dir := filepath.Join(baseDataHome, appName, "notes")
	// 確保目錄存在，如果不存在則建立它。
	if err := ensureDir(dir); err != nil {
		return "", fmt.Errorf("data dir: %w", err)
	}
	return dir, nil
}

// ensureDir 是一個輔助函數，用於確保給定的目錄存在。
// 如果目錄不存在，它會以 0700 的權限建立它。
func ensureDir(dir string) error {
	return os.MkdirAll(dir, 0700)
}
