## 📘 ORA-ORA-ORA：CLI AI 筆記系統

這是一本地端運行的個人速記與問答系統。使用者可以透過 CLI 指令快速新增筆記，並以自然語言查詢自己的想法或記錄，由本地 AI 模型回答對應內容。

---

### ⚙️ 技術堆疊

- **Golang**：實作 CLI 工具、處理命令、載入設定、發送 API 請求
- **Bleve**：全文索引工具，儲存與查詢使用者筆記
- **Ollama API Server**：本地 LLM 推論引擎，處理語意與自然語言問答

---

### 🧠 功能流程

1. 使用者下指令新增筆記（`ora add`）
2. 使用者下指令查詢筆記（`ora ask`）
3. Golang 程式：
   - 使用 Bleve 搜尋與提問相關的筆記段落（`context`）
   - 套用 `prompt/ask.zh-tw.yaml` 或 `prompt/ask.en.yaml`
   - 組成符合 Ollama `/api/chat` 規格的 JSON 請求
   - 接收模型回覆後輸出

---

### 📂 設定檔說明

#### `prompt/ask.zh-tw.yaml` 範例

```yaml
system: |
  你是一個筆記助理，會根據使用者的筆記記錄回答他提出的問題。
template: |
  問題：{{question}}
  筆記內容：{{context}}
model: hf.co/bartowski/Llama-3.2-3B-Instruct-GGUF:Q4_K_M
language: zh-tw
```

#### `prompt/weekly.zh-tw.yaml`（週記模式）

```yaml
system: |
  你是一個知識摘要工具，請根據以下筆記整理一週的重點。
template: |
  以下是本週的筆記內容：
  {{context}}
  請整理為條列式摘要，包含主題與要點。
model: hf.co/bartowski/Llama-3.2-3B-Instruct-GGUF:Q4_K_M
```

---

### 🔧 Golang 封裝邏輯（簡化）

```go
type PromptTemplate struct {
	System   string `yaml:"system"`
	Template string `yaml:"template"`
	Model    string `yaml:"model"`
}

func BuildOllamaPayload(tmpl PromptTemplate, question string, context string) map[string]interface{} {
	prompt := strings.ReplaceAll(tmpl.Template, "{{question}}", question)
	prompt = strings.ReplaceAll(prompt, "{{context}}", context)

	return map[string]interface{}{
		"model": tmpl.Model,
		"messages": []map[string]string{
			{"role": "system", "content": tmpl.System},
			{"role": "user", "content": prompt},
		},
	}
}
```

---

### 📤 送出 Ollama API 請求

```http
POST http://localhost:11434/api/chat
Content-Type: application/json

{
  "model": "llama3",
  "messages": [
    { "role": "system", "content": "你是一個筆記助理..." },
    { "role": "user", "content": "問題：我昨天寫了什麼？
筆記內容：..." }
  ]
}
```

回傳格式為：

```json
{
  "message": {
    "role": "assistant",
    "content": "你昨天寫了關於 Bleve 與 Golang 的索引處理..."
  }
}
```

---

### 🏷 支援功能總結

- ✅ 中文與英文自然語言提問
- ✅ tags / 分群搜尋條件（Bleve 篩選）
- ✅ YAML prompt 模板可調整語氣與行為
- ✅ 可整合 cron 自動執行 `ora weekly`（週記總結）
- ✅ 所有資料本地儲存，無須網路

---

如需為不同角色或應用情境擴充 prompt 模板，只需在 `prompt/*.yaml` 加入對應檔案，並於 Golang 程式中選擇載入即可。
