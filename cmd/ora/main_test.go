// Package main 包含 Ora 應用程式的進入點及其相關測試。
package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wtg42/ora-ora-ora/internal/tui"
)

// TestRunApp 測試 runApp 函數是否能正確返回配置和資料目錄。
func TestRunApp(t *testing.T) {
	// 呼叫 runApp 函數以獲取配置和資料目錄。
	configDir, dataDir, err := runApp()
	// 檢查是否有錯誤返回。
	if err != nil {
		t.Fatalf("runApp() 返回錯誤: %v", err)
	}

	// 檢查配置目錄是否為空。
	if configDir == "" {
		t.Error("runApp() 返回空的配置目錄")
	}

	// 檢查資料目錄是否為空。
	if dataDir == "" {
		t.Error("runApp() 返回空的資料目錄")
	}
}

// TestTUIBuild 測試 TUI 組件是否能正確建置而不報錯。
func TestTUIBuild(t *testing.T) {
	// 建構初始模型，確保不 panic。
	model := tui.InitialModel()

	// 建立程式實例，確保建置成功。
	p := tea.NewProgram(model)
	if p == nil {
		t.Error("NewProgram() 返回 nil")
	}
}
