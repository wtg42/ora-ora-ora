package tui

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wtg42/ora-ora-ora/model"
)

// AddWizardModel：以聊天模式蒐集 add 所需參數（僅彙整與顯示，不進行儲存）。
type AddWizardModel struct {
	history  viewport.Model
	input    textarea.Model
	keys     chatKeyMap
	width    int
	height   int
	messages []Message
	quitting bool
	// dependencies for saving/indexing
	saver   noteSaver
	indexer noteIndexer
	// parsed state for confirmation flow
	parsedContent  string
	parsedTags     []string
	confirmPending bool
	help           help.Model
}

// NewAddWizardModel 建立不注入儲存依賴的精簡精靈（僅彙整顯示）。
func NewAddWizardModel() AddWizardModel {
	ti := textarea.New()
	ti.Placeholder = "請以自然語言描述你的筆記（可含 #標籤 或 tags: ...）"
	ti.Prompt = "> "
	ti.FocusedStyle.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	ti.BlurredStyle.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	ti.ShowLineNumbers = false
	ti.CharLimit = 4096
	ti.SetHeight(3) // 預設 3 行，Alt+Enter 可換行
	ti.Focus()

	vp := viewport.Model{}
	m := AddWizardModel{
		history:  vp,
		input:    ti,
		keys:     chatKeys,
		messages: nil,
		help:     help.New(),
	}
	// 初始系統提問
	m.appendSystem("請描述要新增的筆記內容。\n- 可直接輸入文字；標籤可用 #tag 或 'tags: a,b'。\n- 我會解析內容與標籤並顯示彙整結果供你確認。")
	return m
}

// NewAddWizardModelWithDeps 建立可儲存的精靈，注入儲存與索引依賴。
func NewAddWizardModelWithDeps(saver noteSaver, indexer noteIndexer) AddWizardModel {
	m := NewAddWizardModel()
	m.saver = saver
	m.indexer = indexer
	return m
}

func (m AddWizardModel) Init() tea.Cmd { return textarea.Blink }

func (m AddWizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.input.SetWidth(m.width)
		m.help.Width = msg.Width
		// 以 DockStyle 量測輸入 Dock 實際高度
		dockHeight := lipgloss.Height(DockStyle.Width(m.width).Render(m.input.View()))
		h := m.height - dockHeight
		h = max(h, 3)
		m.history.Width, m.history.Height = m.width, h
		m.refreshHistory()
		return m, nil

	case tea.KeyMsg:
		// Alt+Enter: 插入換行並繼續編輯（不觸發送出）
		if msg.Type == tea.KeyEnter && msg.Alt {
			m.input.SetValue(m.input.Value() + "\n")
			return m, nil
		}
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, m.keys.Send):
			content := strings.TrimSpace(m.input.Value())
			// Alt+Enter：改為插入換行（交由 textarea 處理）。
			if msg.Alt {
				// 明確插入換行於末端（簡化處理，不考慮游標位置）。
				m.input.SetValue(m.input.Value() + "\n")
				return m, nil
			}
			if content == "" {
				return m, nil
			}
			// 若正在確認階段，解讀指令
			if m.confirmPending {
				low := strings.ToLower(content)
				switch low {
				case "ok", "yes", "y":
					m.input.Reset()
					// 需要依賴才能儲存
					if m.saver == nil || m.indexer == nil {
						m.appendAssistant("Error: storage/index not configured")
						m.confirmPending = false
						return m, nil
					}
					// 建立 note 並回傳保存命令
					note := model.Note{
						ID:        genID(),
						Content:   m.parsedContent,
						Tags:      append([]string(nil), m.parsedTags...),
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}
					saver := m.saver
					indexer := m.indexer
					return m, func() tea.Msg {
						if err := saver.Save(note); err != nil {
							return NoteSaveErrorMsg{Err: err}
						}
						if err := indexer.IndexNote(note); err != nil {
							return NoteSaveErrorMsg{Err: err}
						}
						return NoteSavedMsg{ID: note.ID}
					}
				case "cancel", "no", "n":
					m.input.Reset()
					m.appendAssistant("Cancelled.")
					m.quitting = true
					return m, tea.Quit
				default:
					// 視為新內容，重新進入彙整
					m.appendUser(content)
					m.input.Reset()
					c, tags := parseAddInput(content)
					m.parsedContent, m.parsedTags = c, tags
					sum := summarizeParsed(c, tags)
					m.appendAssistant(sum)
					m.confirmPending = true
					m.scrollToBottom()
					return m, nil
				}
			}
			// 一般輸入：彙整並進入確認階段
			m.appendUser(content)
			m.input.Reset()
			c, tags := parseAddInput(content)
			m.parsedContent, m.parsedTags = c, tags
			sum := summarizeParsed(c, tags)
			m.appendAssistant(sum)
			m.confirmPending = true
			m.scrollToBottom()
			return m, nil
		}
	}
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
	// handle async save results
	switch msg := msg.(type) {
	case NoteSavedMsg:
		m.appendAssistant("Saved note with ID: " + msg.ID)
		m.quitting = true
		return m, tea.Quit
	case NoteSaveErrorMsg:
		m.appendAssistant("Error: " + msg.Err.Error())
		// 停留在確認階段，可再試一次
		m.confirmPending = true
		m.scrollToBottom()
		return m, nil
	}
	return m, tea.Batch(cmds...)
}

func (m AddWizardModel) View() string {
	if m.quitting {
		return ""
	}
	top := m.history.View()
	bottom := DockStyle.Width(m.width).Render(m.input.View())
	helpView := m.help.View(m.keys)
	return lipgloss.JoinVertical(lipgloss.Top, top, bottom, helpView)
}

// --- internal helpers ---

func (m *AddWizardModel) appendUser(content string) {
	m.messages = append(m.messages, Message{Role: "user", Content: content, TS: time.Now()})
	m.refreshHistory()
}
func (m *AddWizardModel) appendAssistant(content string) {
	m.messages = append(m.messages, Message{Role: "assistant", Content: content, TS: time.Now()})
	m.refreshHistory()
}
func (m *AddWizardModel) appendSystem(content string) {
	m.messages = append(m.messages, Message{Role: "system", Content: content, TS: time.Now()})
	m.refreshHistory()
}
func (m *AddWizardModel) refreshHistory() {
	var b strings.Builder
	for i, msg := range m.messages {
		if i > 0 {
			b.WriteString("\n\n")
		}
		switch msg.Role {
		case "user":
			b.WriteString(msgUserStyle.Render(msg.Content))
		case "assistant":
			b.WriteString(msgAIStyle.Render(msg.Content))
		default:
			// system：採用較中性的框線
			b.WriteString(lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("244")).Padding(0, 1).Render(msg.Content))
		}
	}
	m.history.SetContent(b.String())
}
func (m *AddWizardModel) scrollToBottom() { m.history.GotoBottom() }

// parseAddInput 將原始輸入解析為內容與標籤。
// 規則（最小可用）：
// - 以 #tag 擷取標籤（英數與 -_，轉為小寫）。
// - 以 `tags:` 或 `標籤:` 後接逗號/空白分隔字詞為標籤。
// - 內容 = 原文去除 #tag 與 tags 行後的殘餘文字（trim）。
func parseAddInput(s string) (content string, tags []string) {
	raw := s
	tagsSet := map[string]struct{}{}

	// 1) #hashtags
	reHash := regexp.MustCompile(`#([A-Za-z0-9_-]+)`) // 簡化規則
	for _, m := range reHash.FindAllStringSubmatch(raw, -1) {
		if len(m) > 1 {
			tagsSet[strings.ToLower(m[1])] = struct{}{}
		}
	}
	raw = reHash.ReplaceAllString(raw, "")

	// 2) tags: a,b / 標籤: a b
	reTags := regexp.MustCompile(`(?i)(tags|標籤)\s*[:：]\s*([#A-Za-z0-9_\-\s,]+)`) // 寬鬆擷取
	if m := reTags.FindStringSubmatch(s); len(m) == 3 {
		list := strings.FieldsFunc(m[2], func(r rune) bool { return r == ',' || r == ' ' || r == '\t' })
		for _, t := range list {
			t = strings.TrimSpace(strings.TrimPrefix(t, "#"))
			if t != "" {
				tagsSet[strings.ToLower(t)] = struct{}{}
			}
		}
		raw = strings.ReplaceAll(raw, m[0], "")
	}
	// 內容清理
	content = strings.TrimSpace(raw)
	// 合併標籤
	tags = make([]string, 0, len(tagsSet))
	for k := range tagsSet {
		tags = append(tags, k)
	}
	sort.Strings(tags)
	return content, tags
}

// summarizeParsed 基於解析結果輸出可讀摘要，提示確認。
func summarizeParsed(content string, tags []string) string {
	tagsStr := "(無)"
	if len(tags) > 0 {
		tagsStr = strings.Join(tags, ", ")
	}
	var b strings.Builder
	b.WriteString("理解到的新增參數\n")
	b.WriteString("- 內容: ")
	if strings.TrimSpace(content) == "" {
		b.WriteString("(空)")
	} else {
		b.WriteString(content)
	}
	b.WriteString("\n- 標籤: ")
	b.WriteString(tagsStr)
	b.WriteString("\n\n輸入 'ok' 以儲存，或 'cancel' 取消。")
	return b.String()
}

// genID 產生簡易唯一 ID（避免跨套件依賴）。
func genID() string {
	// 時間戳 + 隨機尾碼（弱唯一即可，無跨程序需求）
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
