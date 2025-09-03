package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Message 表示一則對話訊息。
type Message struct {
	Role    string // user | assistant | system
	Content string
	TS      time.Time
}

// chatKeyMap 定義聊天頁的鍵位。
type chatKeyMap struct {
    Send       key.Binding
    Quit       key.Binding
    ScrollUp   key.Binding
    ScrollDown key.Binding
    Newline    key.Binding // for help display only (Alt+Enter)
}

// ShortHelp implements help.KeyMap for compact help view.
func (k chatKeyMap) ShortHelp() []key.Binding {
    return []key.Binding{k.Send, k.Newline, k.Quit, k.ScrollUp, k.ScrollDown}
}

// FullHelp implements help.KeyMap for expanded help view.
func (k chatKeyMap) FullHelp() [][]key.Binding {
    return [][]key.Binding{{k.Send, k.Newline, k.Quit, k.ScrollUp, k.ScrollDown}}
}

var chatKeys = chatKeyMap{
    Send:       key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "send")),
    Quit:       key.NewBinding(key.WithKeys("esc", "ctrl+c"), key.WithHelp("esc", "quit")),
    ScrollUp:   key.NewBinding(key.WithKeys("pgup", "up"), key.WithHelp("pgup", "scroll up")),
    ScrollDown: key.NewBinding(key.WithKeys("pgdn", "down"), key.WithHelp("pgdn", "scroll down")),
    Newline:    key.NewBinding(key.WithKeys("alt+enter"), key.WithHelp("alt+enter", "newline")),
}

// 樣式：集中管理，之後可主題化。
var (
	chatContainer = lipgloss.NewStyle()
	msgUserStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("69")).Padding(0, 1)
	msgAIStyle    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("135")).Padding(0, 1)
)

// ChatModel 提供「上方歷史、底部輸入」的對話頁。
type ChatModel struct {
	history viewport.Model
	input   textinput.Model
	keys    chatKeyMap
	width   int
	height  int
	// 固定輸入欄高度（單行輸入；之後可改為 textarea 多行）。
	inputHeight int
	messages    []Message
	quitting    bool
	help        help.Model
}

// NewChatModel 建立新的聊天頁 model。
func NewChatModel() ChatModel {
	ti := textinput.New()
	ti.Placeholder = "Type your message"
	ti.Prompt = "> "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	ti.Focus()

	// 單行輸入：由我們攔截 Enter 送出，避免插入換行。
	// （如需 Shift+Enter 換行，未來可改用 textarea.Model）
	ti.CharLimit = 4096

	vp := viewport.Model{}

	return ChatModel{
		history:     vp,
		input:       ti,
		keys:        chatKeys,
		inputHeight: 1, // 單行
		messages:    nil,
		help:        help.New(),
	}
}

func (m ChatModel) Init() tea.Cmd { return textinput.Blink }

// Update 處理鍵盤與視窗事件。
func (m ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.input.Width = m.width
		m.help.Width = msg.Width
		// 量測 Dock 後的實際底部高度，避免與歷史區重疊
		dockHeight := lipgloss.Height(DockStyle.Width(m.width).Render(m.input.View()))
		h := m.height - dockHeight
		h = max(h, 3)
		m.history.Width = m.width
		m.history.Height = h
		m.refreshHistory()
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, m.keys.Send):
			content := strings.TrimSpace(m.input.Value())
			if content == "" {
				return m, nil
			}
			m.appendUser(content)
			m.input.Reset()
			m.scrollToBottom()
			return m, nil
		case key.Matches(msg, m.keys.ScrollUp):
			m.history.LineUp(1)
			return m, nil
		case key.Matches(msg, m.keys.ScrollDown):
			m.history.LineDown(1)
			return m, nil
		}
	}

	// 同時讓 viewport 與 input 處理事件（支援 PgUp/PgDn/滑鼠滾動）
	var cmds []tea.Cmd
	var cmd tea.Cmd
	m.history, cmd = m.history.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	m.input, cmd = m.input.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

// View 組裝畫面：上方歷史，下方輸入欄。
func (m ChatModel) View() string {
	if m.quitting {
		return ""
	}
	top := m.history.View()
	// 以 DockStyle 統一底部輸入區的外觀（與 add 頁共用）
	bottom := DockStyle.Width(m.width).Render(m.input.View())
    helpView := m.help.View(m.keys)
    return lipgloss.JoinVertical(lipgloss.Top, chatContainer.Render(top), bottom, helpView)
}

// appendUser 追加使用者訊息並刷新歷史內容（AI 回覆整合留待後續）。
func (m *ChatModel) appendUser(content string) {
	m.messages = append(m.messages, Message{Role: "user", Content: content, TS: time.Now()})
	m.refreshHistory()
}

// refreshHistory 依據 messages 重新渲染歷史區內容。
func (m *ChatModel) refreshHistory() {
	var b strings.Builder
	for i, msg := range m.messages {
		if i > 0 {
			b.WriteString("\n\n")
		}
		if msg.Role == "user" {
			b.WriteString(msgUserStyle.Render(msg.Content))
		} else {
			b.WriteString(msgAIStyle.Render(msg.Content))
		}
	}
	m.history.SetContent(b.String())
}

// scrollToBottom 將 viewport 捲到最底。
func (m *ChatModel) scrollToBottom() {
	m.history.GotoBottom()
}
