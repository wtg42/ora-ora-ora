package tui

import "github.com/charmbracelet/lipgloss"

// DockStyle 定義底部輸入區（對話框/輸入欄）的通用外觀樣式。
// 目的：
// - 在不同頁面（chat/add）維持一致的「底部固定輸入區」視覺語言。
// - 邊框與內距集中管理，方便未來主題化與調整。
var DockStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("63")).
	Padding(0, 1)
