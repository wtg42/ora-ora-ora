// Package tui 提供了終端使用者介面 (TUI) 的實現。
package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wtg42/ora-ora-ora/internal/note"
	"github.com/wtg42/ora-ora-ora/internal/storage"
)

// viewState 是一個整數類型，用於表示 TUI 的當前視圖狀態。
type viewState int

const (
	listView   viewState = iota // 列表視圖，顯示所有筆記的標題。
	detailView                  // 詳細視圖，顯示單個筆記的內容。
	createView                  // 建立視圖，用於建立新筆記。
)

// SubmitMsg 訊息表示用戶提交了輸入。
type SubmitMsg struct {
	Text string
}

// model 結構體包含了 TUI 應用程式的所有狀態。
type model struct {
	notes               []string  // 筆記標題列表。
	cursor              int       // 當前選中的筆記索引。
	currentView         viewState // 當前的視圖狀態。
	selectedNoteContent string    // 當前查看的筆記內容。
	newNoteTitle        string    // 新筆記的標題。
	newNoteContent      string    // 新筆記的內容。
	errorMessage        string    // 錯誤訊息，用於顯示給使用者。
	inputArea           InputArea // 輸入區域組件。
}

// InitialModel 函數返回一個初始化的 model 實例。
// 它是 TUI 應用程式的起始狀態。
func InitialModel() model {
	notes, err := storage.ListNotes()
	if err != nil {
		return model{
			currentView:  listView,
			errorMessage: fmt.Sprintf("Failed to load notes: %v", err),
			inputArea:    NewInputArea(),
		}
	}
	return model{
		notes:       notes,
		currentView: listView,
		inputArea:   NewInputArea(),
	}
}

// Init 函數在 TUI 應用程式啟動時被呼叫。
// 它返回一個 tea.Cmd，用於執行初始操作，例如載入資料。
func (m model) Init() tea.Cmd {
	return nil
}

// Update 函數處理傳入的訊息並更新 model 的狀態。
// 它是 TUI 應用程式的核心邏輯。
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.currentView == listView {
				if m.cursor > 0 {
					m.cursor--
				}
			}

		case "down", "j":
			if m.currentView == listView {
				if m.cursor < len(m.notes)-1 {
					m.cursor++
				}
			}

		case "enter":
			if m.currentView == listView && len(m.notes) > 0 {
				selectedTitle := m.notes[m.cursor]
				content, err := storage.ReadNote(selectedTitle)
				if err != nil {
					m.errorMessage = fmt.Sprintf("Failed to read note: %v", err)
				} else {
					m.selectedNoteContent = content
					m.currentView = detailView
				}
			}
		case "esc":
			if m.currentView == detailView {
				m.currentView = listView
				m.selectedNoteContent = ""
			} else if m.currentView == createView {
				m.currentView = listView
			}
		case "n": // New note
			if m.currentView == listView {
				m.currentView = createView
				m.newNoteTitle = ""
				m.newNoteContent = ""
			}
		}

		if m.currentView == createView {
			newIA, cmd := m.inputArea.Update(msg)
			m.inputArea = newIA.(InputArea)
			return m, cmd
		}

	case tea.WindowSizeMsg:
		// 處理視窗大小調整事件。

	case SubmitMsg:
		lines := strings.Split(msg.Text, "\n")
		if len(lines) == 0 {
			lines = []string{""}
		}
		title := strings.TrimSpace(lines[0])
		// AI 心智註解: 保留使用者原始內容，不裁剪前後空白，只做換行拆分後重組。
		content := strings.Join(lines[1:], "\n")
		if title == "" {
			m.errorMessage = "筆記標題不能為空"
			return m, nil
		}
		n := note.NewNote(title, content, nil)
		err := storage.SaveNote(n)
		if err != nil {
			m.errorMessage = fmt.Sprintf("儲存筆記失敗: %v", err)
		} else {
			m.notes, err = storage.ListNotes()
			if err != nil {
				m.errorMessage = fmt.Sprintf("重新載入筆記失敗: %v", err)
			} else {
				m.currentView = listView
				m.inputArea = NewInputArea()
			}
		}
	}

	return m, nil
}

// View 函數根據 model 的當前狀態渲染 TUI 介面。
// 它返回一個字串，代表要顯示在終端上的內容。
func (m model) View() string {
	// 如果有錯誤訊息，則顯示錯誤訊息並提示使用者退出。
	if m.errorMessage != "" {
		return fmt.Sprintf("錯誤: %s\n按下 q 鍵退出。", m.errorMessage)
	}

	// 根據當前視圖狀態渲染不同的介面。
	switch m.currentView {
	case listView:
		s := "您的筆記:\n\n"

		// 如果沒有筆記，則提示使用者建立新筆記。
		if len(m.notes) == 0 {
			s += "沒有找到筆記。按下 'n' 鍵建立新筆記。\n"
		} else {
			// 遍歷筆記列表，顯示每個筆記的標題，並標記當前選中的筆記。
			for i, note := range m.notes {
				cursor := " "
				if m.cursor == i {
					cursor = ">"
				}
				s += fmt.Sprintf("%s %s\n", cursor, note)
			}
		}
		s += "\n按下 'n' 鍵建立新筆記，'enter' 鍵查看，'q' 鍵退出。\n"
		return s

	case detailView:
		// 顯示選中筆記的內容。
		s := fmt.Sprintf("筆記內容:\n\n%s\n\n按下 'esc' 鍵返回，'q' 鍵退出。\n", m.selectedNoteContent)
		return s

	case createView:
		// 顯示建立新筆記的介面。
		s := "建立新筆記:\n\n" + m.inputArea.View() + "\n\n按下 'esc' 鍵取消，'q' 鍵退出。\n"
		return s
	}
	return ""
}
