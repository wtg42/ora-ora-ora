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

// ContainerStyle 定義全局容器樣式，提供邊距避免 UI 貼近邊緣。
// 目的：
// - 適應 tmux 等多工器佔用空間。
// - 統一應用到所有頁面根容器。
var ContainerStyle = lipgloss.NewStyle().
	Margin(1, 2)
