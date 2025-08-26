# 文件關聯說明

在進行任何功能規劃或開發前，請務必先閱讀並參考 **README.md** 的整體說明。

本文件（AGENTS.md）旨在補充 README.md，提供更細緻的開發規範、模組介面契約與測試策略。
若兩份文件內容存在差異或缺漏，應優先以 **README.md** 為準，並提出修訂以確保 AGENTS.md 與 README.md 一致。

# 指派行為：專案進度同步

- 讀取進度：當你說「讀取專案進度」時，我會在 `README.md` 搜尋並讀取一個預先約定的獨立章節，標題固定為「進度同步（AI 專用）」。若該章節存在，僅根據此區塊內容回報專案當前進度；若不存在，將回報尚未建立專用進度區塊。
- 存檔進度：當你說「存檔進度」時，我會在 `README.md` 的同名章節更新內容，只記錄「當前未完成功能/待辦」的精簡摘要，不複製整份狀態。若章節不存在，會在 `README.md` 尾端新增該章節後再寫入。
- 標題與記憶：專用章節標題固定為「進度同步（AI 專用）」，並作為後續讀/寫定位依據（我會記住此標題）。
- 邊界與安全：僅讀寫 `README.md` 中此專用章節；不變更其他內容與任何敏感檔。若處於唯讀或權限不足環境，將回報狀態並提供手動更新指引。

# Ora-Ora-Ora 專案開發指南

這是一本地端運行的個人速記與問答系統。使用者可以透過 CLI 指令快速新增筆記，並以自然語言查詢自己的想法或記錄，由本地 AI 模型回答對應內容。

## 開發模式（TDD + AI 藍圖）

- 原則：採用測試驅動開發（TDD）。AI 負責提供藍圖與測試，開發者專注於實作功能直到測試通過。
- 輸入產物：
  - AI：高階藍圖（Roadmap/介面契約/資料流程）、對應的最小可行測試（table-driven 為主）。
  - 開發者：通過測試的實作、必要的重構與最小設計文件更新（README/AGENTS）。
- 收斂策略：先紅（失敗測試）→ 綠（通過）→ 重構（不改外部行為）。避免一次性大量重構。

### 角色分工

- AI：
  - 你是一位說話非常精簡的工程師，能用簡單易懂的方式解釋原理。
  - 規劃與修訂藍圖、定義介面契約與邊界條件。
  - 提供/更新測試檔與測試資料；維持案例最小但覆蓋關鍵分支。
  - 明確標示風險、替代方案與回滾策略。
- 開發者：
  - 按測試實作功能，必要時回饋介面可行性並與 AI 協調調整。
  - 嚴守小步提交與保守改動，確保每次提交皆可編譯並通過現有測試。
  - 更新文件與 changelog（如有對外 API/CLI 變更）。

### 工作流程（每一迭代）

1) AI 提出小步變更計畫與對應測試檔案清單。
2) 開發者拉取測試 → 驗證紅燈 → 實作最小功能使其轉綠。
3) 重構：在綠燈狀態下進行內部清理，不改對外契約。
4) 品質檢查：`go fmt`, `go vet`, `go test ./... -cover`。
5) 文件同步：更新 README/AGENTS 與任何使用說明。
6) 提交：使用 Conventional Commits；PR 保持單一主題，可回滾。

## 專案結構與模組組織
- **入口檔案**：`main.go` 負責初始化 Cobra 與 CLI 入口。
- **指令模組**：放在 `cmd/`（主指令在 `cmd.go`，進階指令放在 `cmd/core/`）。
- **TUI 元件**：放在 `tui/`（例如 `add_note.go`、`search_result.go`），依 Bubble Tea 架構實作 model/update/view。
- **檢索與 AI 模組**：
  - **全文檢索**：`search/` 資料夾內使用 Bleve，所有索引儲存於專案資料夾內，無外部 DB。
  - **AI 代理**：`agent/` 處理與 Ollama API 的互動（含 prompt 注入與結果解析）。
- **設定檔與環境變數**：
  - 預計建立 `config/` 讀取 `.env` 或 YAML/JSON 設定。
  - 支援 Ollama 模型名稱、索引路徑、TUI 介面選項等設定。

## 開發、建置與測試指令
- **本地執行 CLI**：
  ```bash
  go run .               # 啟動 CLI
  go run . start-tui     # 啟動 TUI 介面
  ```
- **編譯執行檔**：
  ```bash
  go build -o ./bin/ora-ora-ora .
  ```
- **程式碼格式化與靜態檢查**：
  ```bash
  go fmt ./...
  go vet ./...
  ```
  若有安裝 `goimports`，可自動整理 import。
- **測試**：
  ```bash
  go test ./... -cover
  ```
  即使暫時沒有測試檔也建議跑一次，確保環境正常。

## 程式碼風格與命名規範
- 使用 `go fmt` 標準格式化；縮排採用 Tab（視覺 2 空格）。
- 套件命名：**短小、全小寫、無底線**（如 `tui`、`core`、`agent`）。
- 檔案命名：小寫，必要時用底線（如 `add_note.go`）。
- 公開識別名稱：UpperCamelCase；私有識別名稱：lowerCamelCase。
- **TUI 元件**：盡量保持狀態不可變，update 函式回傳新 model。
- **Agent 層**：避免將 Bleve 與 Ollama 直接耦合，保持可替換性。

## 測試策略
- 使用標準 `testing` 套件，測試檔與功能檔同資料夾（`xxx_test.go`）。
- **推薦測試範圍**：
  - 關鍵字斷詞與搜尋結果過濾。
  - CLI 參數解析（Cobra）。
  - TUI 狀態轉換與輸入事件。
  - Agent 層與 Ollama API 的互動（可用 mock server 測試）。
- 優先採用 table-driven 測試，方便擴充案例。

## Commit 與 Pull Request 規範
- **Commit 格式**：遵循 Conventional Commits，例如：
  - `feat: 新增 Bleve 搜尋標籤過濾功能`
  - `fix: 修正 TUI 搜尋結果換頁顯示錯誤`
- PR 應保持單一功能或修正，並在描述中：
  - 說明變更目的、作法、影響範圍。
  - 若有 UI/TUI 變更，附上 CLI 輸出或截圖。
- 合併前需確保：
  ```bash
  go fmt ./...
  go vet ./...
  go test ./...
  ```

## 架構與維護建議
- 執行流程：
  ```
  main.go → cmd.NewOraCmd() → (選擇 CLI 功能)
                          → startTui() → TUI event loop
  ```
- **AI 流程建議**：
  - CLI/TUI 接收使用者輸入 → 關鍵詞斷詞 → Bleve 搜尋相關文件 → 注入 Prompt → 呼叫 Ollama API → 顯示回覆。
- **Bleve 儲存**：
  - 單一索引檔可存放於 `data/index/`，可透過 tags 與分群查詢。
  - **設定檔**：
  - 預留 `.env` 或 YAML/JSON 支援不同模型與搜尋策略。
  - 保持 Agent 與搜尋層的模組化，方便未來替換模型或改用其他檢索引擎。

## 模組契約與資料模型（介面先行）

以下定義為 MVP 介面契約，先以文件落地，實作時可逐步完成，並以 table‑driven 測試覆蓋主要分支。

### storage（檔案儲存；建議 JSONL）

Note 結構：

```go
type Note struct {
    ID        string    // UUIDv4
    Content   string
    Tags      []string
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

介面：

```go
type Storage interface {
    Save(note Note) error
    List() ([]Note, error)
}
```

儲存建議：`data/notes/YYYY-MM-DD.jsonl`（一行一筆）。

### search（Bleve 介面層；鬆耦合）

Snippet 結構：

```go
type Snippet struct {
    NoteID     string
    Excerpt    string
    Score      float64
    TagMatches []string
}
```

介面與生命週期：

```go
type Index interface {
    IndexNote(Note) error
    Query(q string, topK int, tags []string) ([]Snippet, error)
    Close() error
}

func OpenOrCreate(path string) (Index, error)
```

Mapping 建議：`content`（全文）、`tags`（keyword，不分詞）、`created_at`（sortable）。

### agent（Ollama HTTP 客戶端）

Options 與介面：

```go
type Options struct {
    Temperature float64
    TopP        float64
    NumCtx      int
    NumPredict  int
    KeepAlive   time.Duration
}

type LLM interface {
    Chat(ctx context.Context, system, user string, opts Options) (string, error)
}
```

先支援非串流；未來可加上 `StreamChat`。

### config（YAML 載入）

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

func Load(path string) (Config, error)
```

預設值＋YAML 覆寫；不自動讀 `.env`。

## 檢索與推論流程（ask）

- 正規化查詢（Trim 空白、大小寫正常化）。
- Bleve Top‑K（預設 20；可用 `--topk` 覆寫）。
- 載入模板 `prompt/ask.zh-tw.yaml`（不存在則使用內建預設）。
- 渲染變數 `{{question}}`、`{{context}}`。
- 呼叫 Ollama `/api/chat`（非串流），將回答輸出為終端文字。

## 新增流程（add）

- 儲存筆記（Storage.Save）→ 更新索引（IndexNote）。
- 回傳 Note.ID 以便後續引用。

## 錯誤處理策略

- Ollama 不可達：提示檢查服務、埠、模型；提供 `--ollama-host` 覆寫。
- 索引缺失：初次自動建立並提示索引時間成本。
- 模板缺失：回退到內建模板並打印警示訊息。
- I/O 失敗：顯示具體路徑與建議（權限/磁碟/唯讀環境）。

## 測試矩陣（table‑driven）

- search：
  - IndexNote/Query 流程、tags 過濾、空結果、分數排序穩定性。
  - in‑memory stub 與未來 Bleve 實作行為一致性。
- agent：
  - mock HTTP 驗證 payload、options、錯誤分支與逾時處理。
- config：
  - 預設值合併、YAML 覆寫、非法 YAML 與缺欄位容錯。
- tui：
  - Enter/Esc 行為、內容/標籤解析（`#tag` 抽取）、基本狀態轉換。
- cli（若已啟用）：
  - 參數解析、錯誤訊息、人性化提示。

---

## 提交與驗收規範（Definition of Done）

- 測試：所有現有測試綠燈，新增功能須附最小測試；覆蓋關鍵路徑。
- 品質：`go fmt`、`go vet` 無錯；不引入不必要相依；避免提前抽象。
- 文件：README/AGENTS 同步更新；若有用戶面向變更，補充使用說明。
- 相容：不隨意升級依賴；若必要，附相容性說明與變更摘要。
- 風險：提供回滾策略（如單獨 revert 的提交）與替代方案。

## 安全與邊界（必讀）

- 禁止讀寫或上傳：`.env`, `secrets.*`, `*.key`, `id_*`, `*.pem`, `node_modules/`, `vendor/`, `storage/`, `tmp/`, `.git/`，及任何被 `.gitignore` 忽略的敏感檔案。
- 僅在專案工作目錄內作業；無特別許可不得連外或下載套件。
- 對可能破壞性操作（大量重構、刪檔、改 CI）：先提出變更計畫、風險與回滾，再行動。
- 「快速執行」情境**依賴版本**：為確保專案能受益於最新的功能與安全修正，開發時應優先選用函式庫的最新穩定版本。若有版本鎖定之需求，需於文件中明確記錄其原因與影響
- search：索引/查詢、tags 過濾、空結果、分數排序穩定性。
- agent：mock HTTP 驗證 payload、options 與錯誤分支。
- config：預設值合併、非法 YAML、缺欄位處理。
- tui：Enter/Esc 行為、Content/Tags 解析（`#tag` 提取）。

## 路線圖（小步、可回滾）

- M1 文件落地。
- M2 介面骨架：新增 `storage/`, `search/`, `agent/`, `config/` 套件與最小實作（不動 TUI）。
- M3 指令最小版：`ora add` 寫檔＋索引、`ora ask` 僅顯示檢索片段（暫不呼叫 LLM）。
- M4 LLM 串接：非串流回覆＋`--template/--model` 旗標。
- M5 TUI 整合：AddNote 寫檔與提示結果；後續再增查詢頁。
