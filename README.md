# ORA-ORA-ORA

這是一個本機端運行的個人速記與問答系統。使用者可用 CLI/TUI 快速新增筆記，並以自然語言查詢，由本地 AI 模型回覆相關內容。

## 底層技術

- Golang（CLI/TUI 與模組化架構）
- Ollama（本機 LLM 推論 API）
- Bleve（全文檢索；先以 in-memory 介面過渡）

---

## Quickstart

- 需求：
  - Go 1.22+（建議最新版）
  - 可選：本機已安裝並啟動 Ollama（預設 http://localhost:11434）。目前僅啟動 TUI 不需要 Ollama。
- 建置與執行：
  - 啟動 TUI 原型：`go run .` 或 `go run . start-tui`
  - 編譯：`go build -o ./bin/ora-ora-ora .`
- 狀態：目前提供「新增筆記」的 TUI 原型；CLI `add/ask` 功能將於後續版本加入（WIP）。

---

## 使用方式

- 啟動 TUI：
  - `go run . start-tui`
  - 在輸入欄輸入內容並按 Enter 結束；輸入中的 `#tag` 會自動解析為標籤。
- 預告 CLI 指令（WIP）：
  - `ora add "今天研究 Bleve #search #golang"`
  - `ora ask "昨天我做了什麼？"`

---

## 藍圖與架構（Blueprint）

- ✅ 入口檔案：`main.go` 初始化 Cobra 與 CLI 入口。
- ✅ 指令模組：`cmd/`（主指令在 `cmd.go`；擴充子指令如 `add/ask`）。
- ✅ TUI 元件：`tui/`（Bubble Tea 架構；目前有 `add_note.go`）。
- 檢索與 AI 模組：
  - ✅ `search/`：目前提供 in-memory Index；後續替換為 Bleve 實作，索引存於 `data/index/`。
  - ✅ `agent/`：目前為 `MockLLM`；後續串接 Ollama `/api/chat`（非串流）。
- ✅ 設定：`config/` 目前回傳預設值；未引入外部 YAML 依賴，保留相容擴充點。

### 資料模型與介面契約（MVP）

```go
// storage
type Note struct {
    ID        string
    Content   string
    Tags      []string
    CreatedAt time.Time
    UpdatedAt time.Time
}

type Storage interface {
    Save(note Note) error
    List() ([]Note, error)
}

// search
type Snippet struct {
    NoteID     string
    Excerpt    string
    Score      float64
    TagMatches []string
}

type Index interface {
    IndexNote(Note) error
    Query(q string, topK int, tags []string) ([]Snippet, error)
    Close() error
}

func OpenOrCreate(path string) (Index, error)

// agent
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

// config
type Config struct {
    OllamaHost string
    Model      string
    Data struct {
        NotesDir string
        IndexDir string
    }
    TUI struct { Width int }
}

func Load(path string) (Config, error)
```

### 流程設計

- 新增（add）：儲存筆記（Storage.Save）→ 更新索引（Index.IndexNote）→ 回傳 Note.ID。
- 查詢（ask）：正規化查詢 → 檢索 Top-K（預設 20，可 `--topk`）→ 渲染模板 → 呼叫 Ollama → 顯示結果。

---

## 開發、建置與測試

```bash
go run .               # 啟動 CLI（預設進入 TUI）
go run . start-tui     # 強制啟動 TUI
go build -o ./bin/ora-ora-ora .
go fmt ./...
go vet ./...
go test ./... -cover
```

程式碼風格：

- `go fmt` 標準格式；縮排 Tab（視覺 2 空格）。
- 套件命名短小、全小寫、無底線（如 `tui`、`search`）。
- 公開識別 UpperCamelCase；私有識別 lowerCamelCase。
- TUI 更新以不可變狀態為主，Update 回傳新 model。
- Agent 與 Search 保持鬆耦合，方便替換。

---

## 設定（預留）

- 預計路徑：`config/config.yaml`（不自動讀 `.env`）
- 預設值：
  - `ollama.host: http://localhost:11434`
  - `model: hf.co/bartowski/Llama-3.2-3B-Instruct-GGUF:Q4_K_M`
  - `data.notes_dir: data/notes/`
  - `data.index_dir: data/index/`
  - `tui.width: 80`

---

## 路線圖（小步、可回滾）

- ✅ M1 文件落地：基本架構與說明文件。
- ✅ M2 介面骨架：`storage/`, `search/`, `agent/`, `config/` 最小 in-memory 實作與測試。
- ✅ M3 指令最小版：
  - ✅ `ora add`：寫檔（JSONL）＋更新 in-memory 索引。
  - ✅ `ora ask`：僅顯示檢索片段（暫不呼叫 LLM）。
- M4 串接 LLM：
  - 加入模板 `prompt/ask.zh-tw.yaml`（無檔則 fallback 內建）。
  - 呼叫 Ollama `/api/chat` 非串流；`--template/--model` 旗標。
- M5 TUI 整合：
  - AddNote 落盤與提示結果；後續再增查詢頁。

---

## M3 完成內容（已完成）

- ✅ `storage`: 檔案型 `FileStorage`（JSONL，`data/notes/YYYY-MM-DD.jsonl`），保留 `InMemoryStorage` 供測試。
- ✅ `search`: 維持 in-memory；啟動查詢前由 `FileStorage.List()` 重建索引。
- ✅ `cmd`：新增 `add` 與 `ask` 子指令（Cobra）。
  - ✅ `add`：接收內容字串、解析 `#tag`、產生 UUID、寫入檔案並即時 Index。
  - ✅ `ask`：接收查詢字串與 `--topk/--tags`，輸出 Snippet 清單（不呼叫 LLM）。
- ✅ `tui`：送出時呼叫 `FileStorage.Save` 與即時 Index，顯示 `ID/Tags`。
- ✅ `config`：維持 Default，延後 YAML 解析至 M4 之後。

驗收條件：

- 可執行 `go run . add "內容 #tag1 #tag2"`，顯示回傳的 `ID`，並在 `data/notes/` 新增記錄。
- 可執行 `go run . ask "關鍵字" --topk 5`，列出 Snippet（NoteID, Excerpt, Score）。
- `go test ./... -cover` 通過現有與新增測試（優先 table-driven）。

風險與回滾：

- 若 `FileStorage` 寫入發生問題，可切回 `InMemoryStorage` 路徑與測試，避免卡住整體開發。
- `search` 仍為 in-memory，不牽涉 Bleve schema，後續替換衝擊小。

---

## 資料與隱私

- 筆記：儲存於 `data/notes/*.jsonl`（一行一筆；`id, content, tags, created_at`）。
- 索引：存於 `data/index/`（導入 Bleve 後）。
- 所有資料與推論皆於本機完成，無雲端上傳。

## Troubleshooting

- 無法連到 Ollama：
  - 確認執行 `ollama serve` 並監聽 `11434`；確認模型已 `ollama pull <model>`。
- 首次建立索引較慢：
  - 正常現象，初次建立完成後為增量更新。
- 權限或路徑問題：
  - 請確認 `data/` 可寫入；在唯讀環境請切換到可寫目錄後重試。
- Windows/WSL 提示：
  - 建議於 WSL2 或 macOS/Linux 直接使用較穩定。
