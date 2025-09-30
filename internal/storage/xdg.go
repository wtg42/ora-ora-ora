package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

const appName = "ora-ora-ora"

// GetConfigDir 返回 XDG ConfigHome + app 子目錄，用於 TOML config。
func GetConfigDir() (string, error) {
	configHome := xdg.ConfigHome // var (string), no ()
	dir := filepath.Join(configHome, appName)
	if err := ensureDir(dir); err != nil {
		return "", fmt.Errorf("config dir: %w", err)
	}
	return dir, nil
}

// GetDataDir 返回 XDG DataHome + app 子目錄，用於 Markdown 資料。
func GetDataDir() (string, error) {
	dataHome := xdg.DataHome // var (string), no ()
	dir := filepath.Join(dataHome, appName, "notes")
	if err := ensureDir(dir); err != nil {
		return "", fmt.Errorf("data dir: %w", err)
	}
	return dir, nil
}

// ensureDir 確保目錄存在。
func ensureDir(dir string) error {
	return os.MkdirAll(dir, 0700)
}
