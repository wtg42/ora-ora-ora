package tui

import (
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AddWizardModel：以聊天模式蒐集 add 所需參數（僅彙整與顯示，不進行儲存）。
type AddWizardModel struct {
	history  viewport.Model
	input    textinput.Model
	keys     chatKeyMap
	width    int
	height   int
	messages []Message
	quitting bool
}

func NewAddWizardModel() AddWizardModel {
	ti := textinput.New()
	ti.Placeholder = "請以自然語言描述你的筆記（可含 #標籤 或 tags: ...）"
	ti.Prompt = "> "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	ti.Focus()
	ti.CharLimit = 4096

	vp := viewport.Model{}
	m := AddWizardModel{
		history:  vp,
		input:    ti,
		keys:     chatKeys,
		messages: nil,
	}
	// 初始系統提問
	m.appendSystem("請描述要新增的筆記內容。\n- 可直接輸入文字；標籤可用 #tag 或 'tags: a,b'。\n- 我會解析內容與標籤並顯示彙整結果供你確認。")
	return m
}

func (m AddWizardModel) Init() tea.Cmd { return textinput.Blink }

func (m AddWizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.input.Width = m.width
		// 以 DockStyle 量測輸入 Dock 實際高度
		dockHeight := lipgloss.Height(DockStyle.Width(m.width).Render(m.input.View()))
		h := m.height - dockHeight
		h = max(h, 3)
		m.history.Width, m.history.Height = m.width, h
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
			// 解析並顯示彙整結果（不儲存）
			sum := summarizeAddInput(content)
			m.appendAssistant(sum)
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
	return m, tea.Batch(cmds...)
}

func (m AddWizardModel) View() string {
	if m.quitting {
		return ""
	}
	top := m.history.View()
	bottom := DockStyle.Width(m.width).Render(m.input.View())
	return lipgloss.JoinVertical(lipgloss.Top, top, bottom)
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

// summarizeAddInput 解析使用者輸入以提取內容與標籤，並以可讀格式回覆。
// 規則（最小可用）：
// - 以 #tag 擷取標籤（英數與 -_）。
// - 以 `tags:` 或 `標籤:` 後接逗號/空白分隔字詞為標籤。
// - 內容 = 原文去除 #tag 與 tags 行後的殘餘文字（trim）。
func summarizeAddInput(s string) string {
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
	content := strings.TrimSpace(raw)
	// 合併標籤
	tags := make([]string, 0, len(tagsSet))
	for k := range tagsSet {
		tags = append(tags, k)
	}
	// 穩定輸出：排序可留待需要再加；目前直接 Join
	tagsStr := ""
	if len(tags) > 0 {
		tagsStr = strings.Join(tags, ", ")
	} else {
		tagsStr = "(無)"
	}

	// 回覆文字（助手樣式顯示）
	var b strings.Builder
	b.WriteString("理解到的新增參數\n")
	b.WriteString("- 內容: ")
	if content == "" {
		b.WriteString("(空)")
	} else {
		b.WriteString(content)
	}
	b.WriteString("\n- 標籤: ")
	b.WriteString(tagsStr)
	b.WriteString("\n\n如需調整，請繼續輸入補充或更正（目前僅顯示彙整，尚未寫入）。")
	return b.String()
}
