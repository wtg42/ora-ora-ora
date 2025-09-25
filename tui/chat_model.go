package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/wtg42/ora-ora-ora/agent"
	"github.com/wtg42/ora-ora-ora/prompt"
	"github.com/wtg42/ora-ora-ora/search"
	"net/http"
)

// Message 表示一則對話訊息。
type ChatMessage struct {
	Role    string // user | assistant | system
	Content string
	TS      time.Time
}

// Message is an alias for ChatMessage to use in tea.Msg
type Message = ChatMessage

// chatKeyMap 定義聊天頁的鍵位。
type chatKeyMap struct {
	Send       key.Binding
	Quit       key.Binding
	ScrollUp   key.Binding
	ScrollDown key.Binding
	Newline    key.Binding // for help display only (Ctrl+j)
	Settings   key.Binding // F2 for settings panel
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
	Newline:    key.NewBinding(key.WithKeys("ctrl+j"), key.WithHelp("ctrl+j", "newline")),
	Settings:   key.NewBinding(key.WithKeys("f2"), key.WithHelp("F2", "settings")),
}

// 樣式：集中管理，之後可主題化。
var (
	chatContainer = lipgloss.NewStyle()
	msgUserStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("69")).Padding(0, 1)
	msgAIStyle    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("135")).Padding(0, 1)
)

// ChatModel 提供「上方歷史、底部輸入」的對話頁。
type ChatModel struct {
	history     viewport.Model
	input       textarea.Model
	keys        chatKeyMap
	width       int
	height      int
	inputHeight int // 動態調整輸入框高度
	messages    []ChatMessage
	quitting    bool
	help        help.Model

	// 檢索依賴
	indexProvider func() (search.Index, error)

	// LLM 設定與依賴
	llmEnabled  bool
	llmProvider func(host, model string) agent.LLM
	ollamaHost  string
	modelName   string
	llmOpts     agent.Options
	topK        int

	// Opencode SDK integration
	opencodeLLM LLM

	// 設定面板狀態
	settingsOpen  bool
	settingsFocus int // 0=model, 1=host
	tmpLLMEnabled bool
	tmpTopK       int
	tmpTemp       float64
	hostInput     textarea.Model
	modelInput    textarea.Model
	settingsErr   string
	connStatus    string
	connCheck     func(host string) string
}

// LLM defines the chat-only interface for opencode SDK.
type LLM interface {
	// NewSession creates a new chat session with optional config (e.g., model name).
	NewSession(ctx context.Context, model string) (Session, error)
	// Generate generates a response for the given prompt in the session.
	Generate(ctx context.Context, session Session, prompt Prompt) (string, error)
}

// Session represents an LLM chat session (mockable state holder).
type Session interface {
	// AppendMessage adds user/assistant messages to session history.
	AppendMessage(role string, content string) error
	// GetHistory returns the current chat history as a slice.
	GetHistory() []LLMMessage
}

// LLMMessage is a simple chat message struct for LLM.
type LLMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Prompt is a templated input for LLM (e.g., system/user prompts).
type Prompt struct {
	System string
	User   string
}

// Common errors.
var (
	ErrInvalidModel = errors.New("invalid LLM model")
	ErrTimeout      = errors.New("LLM request timeout")
	ErrNoResponse   = errors.New("no response from LLM")
)

// NewChatModel 建立新的聊天頁 model。
func NewChatModel() ChatModel {
	ti := textarea.New()
	ti.Placeholder = "Type your message"
	ti.Prompt = "▌ "
	ti.FocusedStyle.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	ti.BlurredStyle.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	ti.ShowLineNumbers = false
	ti.CharLimit = 4096
	ti.SetHeight(1) // 初始高度 1
	ti.Focus()

	vp := viewport.Model{}

	return ChatModel{
		history:     vp,
		input:       ti,
		keys:        chatKeys,
		inputHeight: 1, // 初始高度 1
		messages:    nil,
		help:        help.New(),
		llmEnabled:  false,
		ollamaHost:  "http://127.0.0.1:11434",
		modelName:   "llama3",
		topK:        5,
	}
}

func (m ChatModel) Init() tea.Cmd { return textarea.Blink }

// Update 處理鍵盤與視窗事件。
func (m ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// 考慮容器邊距的可用空間
		effectiveWidth := msg.Width - ContainerStyle.GetHorizontalFrameSize()
		effectiveHeight := msg.Height - ContainerStyle.GetVerticalFrameSize()
		m.width, m.height = effectiveWidth, effectiveHeight
		m.input.SetWidth(m.width)
		m.help.Width = msg.Width // help 使用總寬度來決定顯示模式
		// 量測輸入區和 help 的高度
		inputHeight := lipgloss.Height(m.input.View())
		helpView := m.help.View(m.keys)
		helpHeight := lipgloss.Height(helpView)
		h := m.height - inputHeight - helpHeight
		h = max(h, 1) // 歷史區最小 1 行
		m.history.Width = m.width
		m.history.Height = h
		m.refreshHistory()
		// 調整歷史區高度以緊跟內容
		contentHeight := lipgloss.Height(m.history.View())
		if contentHeight < h {
			m.history.Height = contentHeight
		}
		return m, nil

	case tea.KeyMsg:
		// 若設定面板開啟，攔截面板鍵位
		if m.settingsOpen {
			km := msg
			// Enter: 套用；Esc: 關閉；l: 切換 LLM；[/]: TopK 調整；</>: 溫度調整
			switch km.Type {
			case tea.KeyEnter:
				// 套用暫存值
				m.settingsErr = ""
				host := strings.TrimSpace(m.hostInput.Value())
				if host == "" {
					m.settingsErr = "Host 必填"
					return m, nil
				}
				m.llmEnabled = m.tmpLLMEnabled
				if m.tmpTopK > 0 {
					m.topK = m.tmpTopK
				}
				if m.tmpTemp < 0 {
					m.tmpTemp = 0
				}
				if m.tmpTemp > 1 {
					m.tmpTemp = 1
				}
				m.llmOpts.Temperature = m.tmpTemp
				m.ollamaHost = host
				m.modelName = m.modelInput.Value()
				m.settingsOpen = false
				return m, nil
			case tea.KeyEsc:
				// 取消並關閉
				m.settingsOpen = false
				return m, nil
			case tea.KeyRunes:
				if len(km.Runes) == 1 {
					r := km.Runes[0]
					switch r {
					case 'l', 'L':
						m.tmpLLMEnabled = !m.tmpLLMEnabled
						return m, nil
					case 't', 'T':
						host := strings.TrimSpace(m.hostInput.Value())
						if m.connCheck != nil {
							m.connStatus = m.connCheck(host)
						} else {
							m.connStatus = connProbe(host)
						}
						return m, nil
					case '[':
						if m.tmpTopK <= 0 {
							m.tmpTopK = m.topK
						}
						if m.tmpTopK > 1 {
							m.tmpTopK--
						}
						return m, nil
					case ']':
						if m.tmpTopK <= 0 {
							m.tmpTopK = m.topK
						}
						m.tmpTopK++
						return m, nil
					case '<':
						m.tmpTemp -= 0.05
						if m.tmpTemp < 0 {
							m.tmpTemp = 0
						}
						return m, nil
					case '>':
						m.tmpTemp += 0.05
						if m.tmpTemp > 1 {
							m.tmpTemp = 1
						}
						return m, nil
					}
				}
			}
			// 未處理的鍵：編輯當前焦點輸入框
			if km.Type == tea.KeyTab {
				m.settingsFocus = (m.settingsFocus + 1) % 2
				if m.settingsFocus == 0 {
					m.modelInput.Focus()
					m.hostInput.Blur()
				} else {
					m.modelInput.Blur()
					m.hostInput.Focus()
				}
				return m, nil
			}
			// Enter 套用：放在 runes 分支之外
			if km.Type == tea.KeyEnter {
				// 預設保持面板開啟，僅在成功套用後關閉
				m.settingsOpen = true
				m.settingsErr = ""
				host := strings.TrimSpace(m.hostInput.Value())
				model := strings.TrimSpace(m.modelInput.Value())
				if model == "" {
					m.settingsErr = "Model 必填"
					m.settingsOpen = true
					return m, nil
				}
				if host == "" {
					m.settingsErr = "Host 必填"
					m.settingsOpen = true
					return m, nil
				}
				if !(strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://")) {
					m.settingsErr = "Host 需以 http:// 或 https:// 開頭"
					m.settingsOpen = true
					return m, nil
				}
				m.llmEnabled = m.tmpLLMEnabled
				if m.tmpTopK > 0 {
					m.topK = m.tmpTopK
				}
				if m.tmpTemp < 0 {
					m.tmpTemp = 0
				}
				if m.tmpTemp > 1 {
					m.tmpTemp = 1
				}
				m.llmOpts.Temperature = m.tmpTemp
				m.ollamaHost = host
				m.modelName = model
				m.settingsOpen = false
				return m, nil
			}
			if km.Type == tea.KeyShiftTab {
				m.settingsFocus = (m.settingsFocus + 1) % 2 // 兩欄位循環
				if m.settingsFocus == 0 {
					m.modelInput.Focus()
					m.hostInput.Blur()
				} else {
					m.modelInput.Blur()
					m.hostInput.Focus()
				}
				return m, nil
			}
			if m.settingsFocus == 0 {
				m.modelInput, _ = m.modelInput.Update(msg)
			} else {
				m.hostInput, _ = m.hostInput.Update(msg)
			}
			return m, nil
		}
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, m.keys.Settings):
			if !m.settingsOpen {
				m.openSettings()
			} else {
				m.settingsOpen = false
			}
			return m, nil
		case key.Matches(msg, m.keys.Newline):
			// Ctrl+j: 插入換行並繼續編輯，增加高度
			m.input.SetValue(m.input.Value() + "\n")
			m.inputHeight++
			m.input.SetHeight(m.inputHeight)
			return m, nil
		case key.Matches(msg, m.keys.Send):
			content := strings.TrimSpace(m.input.Value())
			if content == "" {
				return m, nil
			}
			m.appendUser(content)
			m.input.SetValue("") // 清除內容
			m.inputHeight = 1    // 重置高度
			m.input.SetHeight(1)
			m.scrollToBottom()
			// 檢索並顯示結果；若開啟 LLM，追加 LLM 回覆
			return m, tea.Batch(m.queryAndAppend(content), m.maybeLLM(content))
		case key.Matches(msg, m.keys.ScrollUp):
			m.history.LineUp(1)
			return m, nil
		case key.Matches(msg, m.keys.ScrollDown):
			m.history.LineDown(1)
			return m, nil
		}
	case Message:
		m.messages = append(m.messages, msg)
		m.refreshHistory()
		m.scrollToBottom()
		return m, nil
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
	helpView := m.help.View(m.keys) + " | F2: Settings"
	body := lipgloss.JoinVertical(lipgloss.Top, top, m.input.View(), helpView)
	if m.settingsOpen {
		panel := m.renderSettings()
		body = lipgloss.JoinVertical(lipgloss.Top, body, panel)
	}
	// 應用容器邊距，寬度為有效寬度 + 邊距
	totalWidth := m.width + ContainerStyle.GetHorizontalFrameSize()
	return ContainerStyle.Width(totalWidth).Render(body)
}

// appendUser 追加使用者訊息並刷新歷史內容（AI 回覆整合留待後續）。
func (m *ChatModel) appendUser(content string) {
	m.messages = append(m.messages, ChatMessage{Role: "user", Content: content, TS: time.Now()})
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

// queryAndAppend 執行檢索並將結果以 assistant 訊息追加到歷史。
func (m *ChatModel) queryAndAppend(q string) tea.Cmd {
	return func() tea.Msg {
		// 注意：目前不做 tags 萃取，僅以空白分詞關鍵字查詢。
		var (
			idx search.Index
			err error
		)
		if m.indexProvider != nil {
			idx, err = m.indexProvider()
		} else {
			idx, err = search.OpenOrCreate("")
		}
		if err != nil {
			return Message{Role: "assistant", Content: "檢索初始化失敗: " + err.Error(), TS: time.Now()}
		}
		defer idx.Close()
		// 簡單查詢，不過濾 tags
		tk := m.topK
		if tk <= 0 {
			tk = 5
		}
		snippets, err := idx.Query(q, tk, nil)
		if err != nil {
			return Message{Role: "assistant", Content: "檢索錯誤: " + err.Error(), TS: time.Now()}
		}
		if len(snippets) == 0 {
			return Message{Role: "assistant", Content: "找不到相關片段。", TS: time.Now()}
		}
		var b strings.Builder
		b.WriteString(fmt.Sprintf("相關片段（Top %d）：\n", tk))
		for _, s := range snippets {
			b.WriteString("- ")
			b.WriteString(s.NoteID)
			b.WriteString("\n")
		}
		return Message{Role: "assistant", Content: b.String(), TS: time.Now()}
	}
}

// maybeLLM 在 LLM 開啟時呼叫 Ollama 產生回覆。
func (m *ChatModel) maybeLLM(question string) tea.Cmd {
	if !m.llmEnabled {
		return func() tea.Msg { return nil }
	}
	host := m.ollamaHost
	model := m.modelName
	opts := m.llmOpts
	topK := m.topK
	if topK <= 0 {
		topK = 5
	}
	return func() tea.Msg {
		idx, err := func() (search.Index, error) {
			if m.indexProvider != nil {
				return m.indexProvider()
			}
			return search.OpenOrCreate("")
		}()
		if err != nil {
			return Message{Role: "assistant", Content: "LLM 前置檢索錯誤: " + err.Error(), TS: time.Now()}
		}
		defer idx.Close()
		snippets, err := idx.Query(question, topK, nil)
		if err != nil {
			return Message{Role: "assistant", Content: "LLM 前置檢索錯誤: " + err.Error(), TS: time.Now()}
		}
		var ctxb strings.Builder
		for i, s := range snippets {
			if i > 0 {
				ctxb.WriteString("\n---\n")
			}
			if s.Excerpt != "" {
				ctxb.WriteString(s.Excerpt)
			} else {
				ctxb.WriteString(s.NoteID)
			}
		}
		tpl, _ := prompt.LoadAskTemplate("")
		system := strings.ReplaceAll(strings.ReplaceAll(tpl.System, "{{question}}", question), "{{context}}", ctxb.String())
		user := strings.ReplaceAll(strings.ReplaceAll(tpl.User, "{{question}}", question), "{{context}}", ctxb.String())
		var llm agent.LLM
		if m.llmProvider != nil {
			llm = m.llmProvider(host, model)
		} else {
			llm = agent.NewClient(host, model)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		answer, err := llm.Chat(ctx, system, user, opts)
		if err != nil {
			return Message{Role: "assistant", Content: "LLM 錯誤: " + err.Error() + "（可關閉 LLM 僅用檢索）", TS: time.Now()}
		}
		return Message{Role: "assistant", Content: answer, TS: time.Now()}
	}
}

func (m *ChatModel) openSettings() {
	m.tmpLLMEnabled = m.llmEnabled
	if m.topK <= 0 {
		m.topK = 5
	}
	m.tmpTopK = m.topK
	m.tmpTemp = m.llmOpts.Temperature
	hi := textarea.New()
	hi.SetWidth(80)
	hi.SetHeight(1)
	hi.SetValue(m.ollamaHost)
	mi := textarea.New()
	mi.SetWidth(40)
	mi.SetHeight(1)
	mi.SetValue(m.modelName)
	mi.Focus()
	hi.Blur()
	m.hostInput = hi
	m.modelInput = mi
	m.settingsOpen = true
	m.settingsFocus = 0
	m.settingsErr = ""
}

func (m *ChatModel) renderSettings() string {
	lines := []string{
		"[設定面板] Enter: 套用 | Esc: 關閉 | l: 切換 LLM | [: TopK- | ]: TopK+ | <: Temp- | >: Temp+ | t: 測試連線",
		fmt.Sprintf("LLM: %v    TopK: %d    Temp: %.2f", m.tmpLLMEnabled, m.tmpTopK, m.tmpTemp),
		"Model:" + focusMark(m.settingsFocus == 0), m.modelInput.View(),
		"Ollama Host:" + focusMark(m.settingsFocus == 1), m.hostInput.View(),
	}
	if m.settingsErr != "" {
		lines = append(lines, "Error: "+m.settingsErr)
	}
	if m.connStatus != "" {
		lines = append(lines, "Conn: "+m.connStatus)
	}
	panel := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).Width(m.width).Render(strings.Join(lines, "\n"))
	return panel
}

func focusMark(active bool) string {
	if active {
		return " [*]"
	}
	return ""
}

// connProbe 進行最小連線測試：嘗試 GET 到 host 根路徑。
func connProbe(host string) string {
	if strings.TrimSpace(host) == "" {
		return "host empty"
	}
	client := &http.Client{Timeout: 2 * time.Second}
	req, err := http.NewRequest("GET", host, nil)
	if err != nil {
		return "invalid host"
	}
	resp, err := client.Do(req)
	if err != nil {
		return "unreachable"
	}
	_ = resp.Body.Close()
	return "reachable"
}
