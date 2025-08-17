# ORA-ORA-ORA：CLI AI 筆記系統

這是一本地端運行的個人速記與問答系統。使用者可以透過 CLI 指令快速新增筆記，並以自然語言查詢自己的想法或記錄，由本地 AI 模型回答對應內容。

---

## 1. 系統構成與專案結構

- **Golang CLI**：使用 `Cobra` 框架實作 `ora add` / `ora ask` 等指令。
- **TUI 元件**：使用 `Bubble Tea` 框架實作互動介面。
- **Bleve**：全文索引與檢索，產生 `context` 片段。
- **Ollama API**（本機）：`/api/chat` 進行回覆生成。

### 專案結構
- `main.go`: 程式入口，初始化 Cobra 與 CLI。
- `cmd/`: 存放 CLI 指令模組。
- `tui/`: 存放 Bubble Tea TUI 元件。
- `storage/`: 負責筆記的檔案儲存（JSONL）。
- `search/`: 負責與 Bleve 互動的全文檢索。
- `agent/`: 處理與 Ollama API 的互動。
- `config/`: 讀取 YAML 設定檔。
- `prompt/`: 存放 LLM 的 Prompt 模板。
- `data/`: 存放使用者資料（筆記、索引）。

---

## 2. 核心流程

1.  **`ora add <note>`**：
    - 透過 `storage` 模組將筆記儲存至 `data/notes/YYYY-MM-DD.jsonl`。
    - 透過 `search` 模組更新 Bleve 索引。
2.  **`ora ask <question>`**：
    - 正規化查詢字串。
    - 使用 `search` 模組以 Bleve 查詢相關片段（Top-K）。
    - 載入 `prompt/*.yaml` 模板。
    - 注入 `{{question}}` 與 `{{context}}` 變數。
    - 透過 `agent` 模組呼叫 Ollama `/api/chat` API。
    - 將模型回覆輸出至終端。

---

## 3. 模組契約與資料模型 (Go 介面)

### 3.1 `storage` (檔案儲存)
- **Note 結構**:
  ```go
  type Note struct {
      ID        string    // UUIDv4
      Content   string
      Tags      []string
      CreatedAt time.Time
      UpdatedAt time.Time
  }
  ```
- **介面**:
  ```go
  type Storage interface {
      Save(note Note) error
      List() ([]Note, error)
  }
  ```

### 3.2 `search` (全文檢索)
- **Snippet 結構**:
  ```go
  type Snippet struct {
      NoteID     string
      Excerpt    string
      Score      float64
      TagMatches []string
  }
  ```
- **介面**:
  ```go
  type Index interface {
      IndexNote(note Note) error
      Query(q string, topK int, tags []string) ([]Snippet, error)
      Close() error
  }
  ```

### 3.3 `agent` (LLM 代理)
- **Options 結構**:
  ```go
  type Options struct {
      Temperature float64
      TopP        float64
      NumCtx      int
      NumPredict  int
      KeepAlive   time.Duration
  }
  ```
- **介面**:
  ```go
  type LLM interface {
      Chat(ctx context.Context, system, user string, opts Options) (string, error)
  }
  ```

### 3.4 `config` (設定檔載入)
- **Config 結構**:
  ```go
  type Config struct {
      OllamaHost string
      Model      string
      Data struct {
          NotesDir string
          IndexDir string
      }
      TUI struct {
          Width int
      }
  }
  ```
- **介面**:
  ```go
  func Load(path string) (Config, error)
  ```

---

## 4. 設定檔格式

### 4.1 Prompt 模板 (`prompt/*.yaml`)
> 用於定義不同任務下 LLM 的行為與輸出格式。

- **`ask.zh-tw.yaml` (問答)**
  ```yaml
  system: |
    你是一個筆記助理，只能根據「提供的筆記內容」回答。
    若資訊不足，請誠實說明，勿臆測。
  template: |
    【問題】
    {{question}}

    【相關筆記片段】
    {{context}}
  model: hf.co/bartowski/Llama-3.2-3B-Instruct-GGUF:Q4_K_M
  language: zh-tw
  options:
    temperature: 0.2
  ```
- **`weekly.zh-tw.yaml` (週記)**
  ```yaml
  system: |
    你是一個知識摘要工具，請將筆記整理成「本週回顧」。
  template: |
    【本週筆記】
    {{context}}

    【輸出格式】
    - 主題（標題句）
    - 關鍵要點（條列）
    - 待辦與下一步（條列）
  model: hf.co/bartowski/Llama-3.2-3B-Instruct-GGUF:Q4_K_M
  ```

### 4.2 系統設定 (`config.yaml`)
> 用於設定 Ollama 位址、預設模型、資料路徑等。由 `config.Load()` 載入。

---

## 5. Ollama API 規格

- **端點**: `POST http://localhost:11434/api/chat`
- **請求 (非串流)**:
  ```json
  {
    "model": "hf.co/bartowski/Llama-3.2-3B-Instruct-GGUF:Q4_K_M",
    "messages": [
      { "role": "system", "content": "..." },
      { "role": "user",   "content": "..." }
    ],
    "stream": false,
    "options": {
      "temperature": 0.2,
      "top_p": 0.9,
      "num_ctx": 8192
    }
  }
  ```
- **回應**:
  ```json
  {
    "message": {
      "role": "assistant",
      "content": "最終答案..."
    }
  }
  ```

---

## 6. 檢索與注入 (RAG)

- **Top‑K**: 預設 20；可由 `--topk` 旗標覆寫。
- **截斷**: 依模型 `num_ctx` 上下文視窗大小估算最大可注入 tokens，超過則尾端截斷。
- **邊界**: 建議以明確分隔符（如 `=== BEGIN SNIPPET ===`）包住每段 context，降低注入風險。

---

## 7. 錯誤處理指南

- **Ollama 未啟動/不可達**: 提示 `請先執行: ollama serve`，或檢查 `--ollama-host` 旗標。
- **模型不存在**: 提示 `ollama pull <model>`。
- **索引缺失**: 初次使用時應自動建立索引。
- **模板缺失**: 回退至內建預設模板並顯示警告。
- **空檢索**: 告知「筆記不足」，建議先 `ora add` 或更換關鍵詞。
- **I/O 失敗**: 顯示具體檔案路徑與權限/磁碟空間問題建議。

---

## 8. Agent 行為建議

- **旗標覆寫**: 應支援 `--model`、`--top-k`、`--template` 等核心參數覆寫。
- **互動**: 若問題模糊，應主動要求使用者提供**關鍵詞**或**縮小範圍**。
- **週記模式**: 執行 `ora weekly` 時，載入 `prompt/weekly.zh-tw.yaml`，`context` 為近 7 天筆記（可由 `--days` 調整）。

---

## 9. 支援功能總結

- ✅ **完全本機化**: 所有資料與模型推論皆在本地完成，支援離線使用。
- ✅ **自然語言問答**: 透過 RAG 技術，結合筆記內容回答問題。
- ✅ **標籤過濾**: 可在查詢時使用 `tags:tagA,tagB` 語法篩選特定筆記。
- ✅ **可自訂 Prompt**: 透過 `prompt/*.yaml` 檔案調整模型語氣、行為與輸出格式。
- ✅ **多樣化任務**: 支援一般問答、週記總結等不同應用情境。

---

## 10. 路線圖 (Roadmap)

- **M1**: 文件落地，定義專案架構、介面與流程。
- **M2**: 建立 `storage/`, `search/`, `agent/`, `config/` 模組的介面骨架與最小化實作。
- **M3**: 完成 `ora add` 與 `ora ask` 指令的 CLI 最小功能（`ask` 僅顯示檢索結果）。
- **M4**: 完整串接 Ollama API，實現基於 RAG 的問答功能。
- **M5**: 整合 TUI 介面，提供更豐富的互動體驗。

---

## 11. 安全與邊界 (必讀)

- **檔案系統**: 禁止讀寫或上傳 `.env`, `secrets.*`, `*.key`, `id_*`, `*.pem`, `node_modules/`, `vendor/`, `storage/`, `tmp/`, `.git/`，以及任何被 `.gitignore` 忽略的敏感檔案。
- **工作目錄**: 所有操作應限制在專案工作目錄內。
- **網路**: 未經明確許可，不得任意連接外部網路或下載套件。
- **破壞性操作**: 對於大量重構、刪除檔案、修改 CI/CD 等操作，需先提出變更計畫、風險與回滾策略。

---

## 12. 開發注意事項

- **小步修改**: 優先專注於修改單一檔案。若需修改多個檔案，建議分次進行或先提出計畫，避免一次性引入大量變更。

---

## 13. 建置與測試

- **程式碼格式化與檢查**:
  ```bash
  go fmt ./...
  go vet ./...
  ```
- **執行測試**:
  ```bash
  go test ./... -cover
  ```
- **本地執行 CLI**:
  ```bash
  go run . --help
  ```
- **編譯執行檔**:
  ```bash
  go build -o ./bin/ora-ora-ora .
  ```

---

## 14. 測試指導方針

- **`search` 模組**: 應測試索引建立、關鍵字查詢、`tags` 過濾、空結果處理、分數排序等。
- **`agent` 模組**: 使用 mock server 驗證 API payload、options、錯誤處理與逾時機制。
- **`config` 模組**: 應測試預設值載入、YAML 檔案覆寫、以及對非法格式的容錯能力。
- **`tui` 模組**: 應測試狀態轉換、使用者輸入事件（如 Enter/Esc）、以及內容與標籤的解析。