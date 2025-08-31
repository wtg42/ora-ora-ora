# 文件關聯說明

這個專案是本機跑的個人速記 + 問答系統。你可以用 CLI 快速記筆記，再用自然語言問它，由本地 AI 回你重點。

開始前請先看 README.md 的整體說明。這份 AGENTS.md 是補充：寫開發規範、介面契約、測試策略。若和 README 有出入，一律以 README 為準，也請回頭修這份文件讓兩邊一致。

# 指派行為：專案進度同步

- 讀進度：「讀取專案進度」→ 我會在 README.md 找標題「進度同步（AI 專用）」的章節，只看這一段回報現況。沒這段就回報「尚未建立」。
- 存進度：「存檔進度」→ 我會更新同名章節，只寫「目前未完成/待辦」的精簡摘要，不複製整份狀態。沒有這段就加在 README.md 最尾端再寫入。
- 定位標題：固定用「進度同步（AI 專用）」當錨點，我會記住這個標題。
- 邊界：只會讀寫 README.md 的這個章節，不動其他內容與任何敏感檔。若環境唯讀或沒權限，我會回報狀態並附手動操作指引。

## 開發模式（TDD + AI 藍圖）

- 原則：先寫測試再實作（TDD）。AI 負責藍圖與最小可行測試；你專心把紅燈變綠燈。
- 輸入產物：
  - AI：高階藍圖（Roadmap/介面契約/資料流程）＋最小測試（以 table-driven 為主）。
  - 開發者：通過測試的實作、必要的小幅重構、最小文件更新（README/AGENTS）。
- 收斂步驟：紅（測試失敗）→ 綠（通過）→ 重構（不改外部行為）。避免一次性大重構。

### 角色分工
- AI：講重點、白話解釋。規劃/修文件、定義介面與邊界。提供/更新測試與測試資料，案例少但卡關鍵分支。清楚標示風險、替代方案、回滾法。
- 開發者：按測試實作功能，需要時回饋介面可行性再協調調整。小步提交、每次都能編譯並通過現有測試。同步更新文件與 changelog（若有對外變更）。

### 工作流程（每次迭代）
1) AI 提小步變更計畫＋測試檔案清單  
2) 開發者拉測試 → 先紅 → 實作到綠  
3) 綠燈後才重構（不動介面）  
4) 品質檢查：`go fmt`, `go vet`, `go test ./... -cover`  
5) 文件同步：更新 README/AGENTS 與使用說明  
6) 提交：Conventional Commits；PR 單一主題、可回滾

## 專案結構與模組組織
- 入口：`main.go`（初始化 Cobra/CLI）
- 指令：`cmd/`（主指令 `cmd.go`、進階指令放 `cmd/core/`）
- TUI：`tui/`（如 `add_note.go`, `search_result.go`，走 Bubble Tea 的 model/update/view）
- 檢索與 AI：
  - 全文檢索：`search/` 用 Bleve，索引放專案資料夾，無外部 DB
  - AI 代理：`agent/` 負責 Ollama API（含 prompt 注入、結果解析）
- 設定：
  - `config/` 載入 YAML（有預設值，不自動讀 `.env`）
  - 支援選項：模型名稱、索引路徑、TUI 參數…

## 開發、建置與測試
- 本地執行
  - `go run .`            啟動 CLI
  - `go run . start-tui`  啟動 TUI
- 編譯
  - `go build -o ./bin/ora-ora-ora .`
- 格式/檢查
  - `go fmt ./...`
  - `go vet ./...`
  - 有裝 `goimports` 可順便整理 import
- 測試
  - `go test ./... -cover`
  - 就算暫時沒測試檔，也建議先跑一次確認環境 OK

## 程式碼風格與命名
- 格式：`go fmt`；縮排用 Tab（視覺 2 空格）
- 套件命名：短、小寫、無底線（如 `tui`, `core`, `agent`）
- 檔案命名：小寫，必要時可底線（如 `add_note.go`）
- 介面/識別：公開 UpperCamelCase，私有 lowerCamelCase
- TUI：盡量不可變，`update` 回傳新 model
- Agent：別把 Bleve 和 Ollama 綁死，保持可替換

## 測試策略
- 用標準 `testing`；測試檔放同資料夾（`xxx_test.go`）
- 推薦範圍：
  - 斷詞與搜尋結果過濾
  - CLI 參數解析（Cobra）
  - TUI 狀態轉換與輸入事件
  - Agent 與 Ollama API（用 mock server）
- 優先 table-driven，方便擴案例

## Commit 與 PR
- Commit 格式（Conventional Commits）
  - `feat: 新增 Bleve 搜尋標籤過濾功能`
  - `fix: 修正 TUI 搜尋結果換頁顯示錯誤`
- PR：單一主題，說明目的、作法、影響；有 UI/TUI 變更就附 CLI 輸出或截圖
- 合併前確認
  - `go fmt ./...`
  - `go vet ./...`
  - `go test ./...`

## 架構與維護建議
- 執行流程
  - `main.go → cmd.NewOraCmd() → (選 CLI 功能) → startTui() → TUI event loop`
- AI 流程建議
  - CLI/TUI 收輸入 → 關鍵詞斷詞 → Bleve 找相關文件 → 注入 Prompt → 呼叫 Ollama → 顯示回覆
- Bleve 與設定
  - 索引放 `data/index/`，可用 tags 與分群查
  - 設定用 YAML；保留替換空間（模型/檢索策略）
  - Agent/搜尋層解耦，隨時能換

## 模組契約與資料模型（介面先行）
以下是 MVP 的介面契約。先用文件落地，實作可分段完成，用 table-driven 測試覆蓋主要分支。

### model（跨模組資料模型；唯一來源）
所有模組對外都用 `model.Note`；內部若要不同結構，自行轉換，不外露。

```go
// 套件：github.com/wtg42/ora-ora-ora/model
type Note struct {
    ID        string    // UUIDv4
    Content   string
    Tags      []string
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### storage（檔案儲存；建議 JSONL）
對外契約用 `model.Note`；若內部序列化有差（例如欄位命名），請在內部轉換或做轉接層，不要洩漏到其他模組。

```go
// 對外契約（以 model.Note 為準）
type Storage interface {
    Save(note model.Note) error
    List() ([]model.Note, error)
}
```

儲存建議：`data/notes/YYYY-MM-DD.jsonl`（一行一筆）

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
// 使用 model.Note 作為索引輸入的統一型別
type Index interface {
    IndexNote(note model.Note) error
    Query(q string, topK int, tags []string) ([]Snippet, error)
    Close() error
}

func OpenOrCreate(path string) (Index, error)
```

Mapping 建議：`content`（全文）、`tags`（keyword，不分詞）、`created_at`（sortable）

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

先支援非串流；未來可加 `StreamChat`

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

有預設值、支援 YAML 覆寫；不自動讀 `.env`

## 檢索與推論流程（ask）
- 正規化查詢（trim 空白、大小寫）
- Bleve Top‑K（預設 20，可 `--topk` 覆寫）
- 載入模板 `prompt/ask.zh-tw.yaml`（找不到就用內建預設）
- 渲染 `{{question}}`、`{{context}}`
- 呼叫 Ollama `/api/chat`（非串流），把回答印到終端

## 新增流程（add）
- 儲存筆記（Storage.Save）→ 更新索引（IndexNote）
- 回傳 Note.ID 方便後續引用

## 錯誤處理策略
- Ollama 不可達：提示檢查服務/埠/模型；可用 `--ollama-host` 覆寫
- 索引缺失：首次自動建立，並提示索引會花時間
- 模板缺失：回退內建模板並打印警示
- I/O 失敗：顯示具體路徑與建議（權限/磁碟/唯讀）

## 測試矩陣（table‑driven）
- search：
  - IndexNote/Query 流程、tags 過濾、空結果、分數排序穩定性
  - in‑memory stub 與 Bleve 實作行為一致性
- agent：
  - mock HTTP 驗證 payload、options、錯誤分支與逾時
- config：
  - 預設值合併、YAML 覆寫、非法 YAML 與缺欄位容錯
- tui：
  - Enter/Esc、內容/標籤解析（`#tag` 抽取）、基本狀態轉換
- cli（若開啟）：
  - 參數解析、錯誤訊息、人性化提示

---

## 提交與驗收（Definition of Done）
- 測試：現有測試全綠；新功能要有最小測試，覆蓋關鍵路徑
- 品質：`go fmt`、`go vet` 無錯；不引入不必要相依；避免過早抽象
- 文件：README/AGENTS 同步更新；有對外變更就補使用說明
- 相容：不隨意升級依賴；若必要，附相容性說明與變更摘要
- 風險：提供回滾策略（可單獨 revert）與替代方案

## 安全與邊界（必讀）
- 禁止讀寫或上傳：`.env`, `secrets.*`, `*.key`, `id_*`, `*.pem`, `node_modules/`, `vendor/`, `storage/`, `tmp/`, `.git/`，以及任何 `.gitignore` 忽略的敏感檔
- 只在專案工作目錄內操作；無許可不要連外或下載套件
- 可能破壞性操作（大量重構、刪檔、改 CI）：先提計畫/風險/回滾，再動手
- 「快速執行」情境（依賴版本）：預設採用套件最新穩定版；若要鎖版本，請在文件標明原因與影響
- 預設測試重點：
  - search：索引/查詢、tags 過濾、空結果、分數排序穩定
  - agent：mock HTTP 驗證 payload/options 與錯誤分支
  - config：預設值合併、非法 YAML、缺欄位處理
  - tui：Enter/Esc、Content/Tags 解析（`#tag` 提取）

## 路線圖（小步、可回滾）
- M1 文件落地
- M2 介面骨架：新增 `storage/`, `search/`, `agent/`, `config/` 最小實作（不動 TUI）
- M3 指令最小版：`ora add` 寫檔＋索引、`ora ask` 只顯示檢索片段（暫不呼叫 LLM）
- M4 LLM 串接：非串流回覆＋`--template/--model` 旗標
- M5 TUI 整合：AddNote 寫檔與提示結果；之後再加查詢頁
