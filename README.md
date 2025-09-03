# ORA-ORA-ORA

這是一個本機端運行的個人速記與問答系統。使用者可用 CLI/TUI 快速新增筆記，並以自然語言查詢，由本地 AI 模型回覆相關內容。

## 底層技術

- Ollama（本機 LLM 推論 API）
- Bleve（全文檢索；先以 in-memory 介面過渡）

---

## 藍圖（Blueprint）

- 目標：本機端的速記與問答系統，提供 CLI/TUI 新增筆記與以自然語言查詢，回覆由本地 LLM 完成。
- 架構流：`main.go → cobra(cmd) → CLI/TUI`；後端模組包含 `storage(JSONL)`, `search(Bleve)`, `agent(Ollama)`, `config(YAML)`，保持鬆耦合、可替換。
- 主要資料夾（規劃）：
  - `cmd/`：CLI 指令（核心在 `cmd.go`，延伸指令於 `cmd/core/`）。
  - `tui/`：Bubble Tea 模型（如 `add_note.go`, `search_result.go`）。
  - `storage/`：JSONL 檔案儲存（`data/notes/YYYY-MM-DD.jsonl`）。
  - `search/`：Bleve 索引與查詢（索引於 `data/index/`）。
  - `agent/`：Ollama HTTP 客戶端（非串流優先）。
  - `config/`：載入 YAML，含預設值與覆寫。
- 核心流程：
  - 新增（add）：寫入 Note → 更新索引 → 回傳 Note.ID。
  - 詢問（ask）：正規化查詢 → 檢索 Top‑K 片段 → 套模板 → 呼叫 Ollama → 輸出答案。

---

## CLI 使用

- 新增筆記：
  - `go run . add "今天研究 Bleve 查詢語法" --tags dev,search`
  - 會寫入 `data/notes/YYYY-MM-DD.jsonl`，並更新索引；輸出格式：`ID: <uuid>`。
- 查詢片段（不呼叫 LLM）：
  - `go run . ask "golang bleve" --topk 5 --tags dev`
  - 會從儲存讀取所有筆記、重建索引後檢索並逐行列出匹配的 `NoteID`。
  - 範例輸出：
    ```
    a1f3d2c4-...-abcd
    9b2e7a10-...-ef01
    ```
- 設定檔（可選）：
  - `--config path/to/config.yaml`（未提供時使用內建預設）
- 進階旗標（測試/進階用）：
  - `--notes-dir /path/to/dir` 覆寫筆記資料夾（僅 CLI 測試或特殊佈署使用）。

### 啟動 TUI

- `go run . start-tui` 啟動 TUI。預設使用 alternate screen（備用 buffer），退出後終端畫面會自動恢復。
- 若終端不支援或需要除錯，可加入 `--no-alt` 停用 alternate screen：
  - `go run . start-tui --no-alt`
 - 切換頁面：
   - `go run . start-tui --page chat` 啟動對話頁（輸入固定底部、歷史在上方可滾動）。
   - `--page add` 為預設頁（維持現有新增筆記流程）。

### TUI 對話頁布局（規劃）

- 目標：提供沉浸式對話體驗，輸入欄固定於底部，歷史訊息於上方可滾動，並以顏色與邊框樣式加以區分，參考常見的 chat UI。
- 版面配置：
  - 底部輸入：單行/多行自動換行輸入框固定貼齊底部（使用 Bubble Tea layout 與 `lipgloss` 邊框）。
  - 上方歷史：訊息流占據輸入框以上的全部高度，可上下滾動；最新訊息顯示在最下方一列之上。
  - 邊界處理：視窗縮放時即時重算可視高度，保持輸入欄高度與位置穩定。
- 共用策略：`DockStyle` 作為底部輸入區通用樣式，`chat` 與 `add` 皆沿用，維持一致外觀（圓角邊框＋適度內距）。
- 視覺語意：
  - 使用者訊息：以冷色系邊框（例：藍），靠右排列；系統/AI 回覆：以暖色系邊框（例：紫/綠），靠左排列。
  - 內容區塊保持一致的內距與圓角，避免雜訊；時間戳記與標籤採弱化色。
- 互動行為：
  - Enter 送出（Shift+Enter 換行）；Esc 清空或回上層。
  - 送出後將焦點留在輸入框；歷史區自動滾至最新訊息並保留滾動位置記憶。
- 技術要點：
  - 採用 alternate screen（預設開啟，可用 `--no-alt` 關閉），避免污染終端歷史。
  - 以 Bubble Tea 的 model/update/view 管理狀態；訊息流維持不可變 slice，滾動以 viewport 或自製 offset 控制。
  - 色彩與邊框透過 `lipgloss` 統一樣式表，集中定義方便主題化。


### M4：LLM 串接（非串流）

- 功能行為：
  - `ask` 先以 `search.Index` 檢索 Top‑K 片段，合併為 `context`，以模板渲染 `system/user`，最後呼叫 Ollama `/api/chat`（非串流）並輸出回覆。
  - 指定 `--no-llm` 時，僅輸出檢索到的 `NoteID` 清單（維持 M3 行為）。

- 使用範例：
  - `go run . ask "如何使用 bleve？" --topk 5 --model llama3 --template prompt/ask.zh-tw.yaml`
  - 未指定 `--template` 時使用內建預設；讀檔失敗或格式不符時回退預設並打印警示。

- 旗標一覽（ask）：
  - `--topk int`：檢索片段數（預設 20）。
  - `--tags string`：以逗號分隔之標籤過濾（AND 行為）。
  - `--no-llm`：僅檢索不呼叫 LLM。
  - `--template path`：指定 YAML 模板檔（例：`prompt/ask.zh-tw.yaml`）。
  - `--model string`：覆寫模型名稱（預設取自設定檔 `model`）。
  - `--ollama-host string`：覆寫 Ollama 位址（預設取自設定檔 `ollamaHost`）。
  - 進階 LLM 參數（對應 agent.Options）：
    - `--temp float`（temperature）
    - `--top-p float`
    - `--num-ctx int`
    - `--num-predict int`
    - `--keep-alive duration`（例如 `30s`、`5m`）

- 環境變數與優先序：
  - `OLLAMA_HOST` 可覆寫連線位址（例：`http://127.0.0.1:11434`）。
  - 優先序：CLI 旗標 > 環境變數 > 設定檔（`config.yaml`）> 內建預設。
  - 需先啟動 `ollama serve` 並確保模型可用（例：`ollama pull llama3`）。

- 模板格式（YAML）：
  - 鍵值：`system`, `user`（至少其一非空）。
  - 變數：`{{question}}`, `{{context}}`。
  - 範例（`prompt/ask.zh-tw.yaml`）：
    ```yaml
    system: |
      你是嚴謹的技術助理，僅根據下方「相關筆記」作答。
      - 請以繁體中文輸出，條列重點。
      - 若找不到相關內容，請直接說明無法回答。
    user: |
      問題：{{question}}

      相關筆記：
      {{context}}

      要求：
      - 摘要相關內容並回覆
      - 勿臆測、勿捏造
    ```
  - Fallback 規則：
    - 檔案不存在：使用預設模板並打印 `template not found; using default`。
    - YAML 非法或缺鍵：使用預設模板並打印 `template invalid or missing keys; using default`。

- 輸出與錯誤處理：
  - 成功：輸出 LLM 回覆文字（純文字，不含多餘標註）。
  - Ollama 不可達/逾時：提示檢查服務、埠與模型（可用 `--ollama-host` 覆寫）；可加上 `--no-llm` 退回檢索模式。
  - 檢索為空：明確提示找不到相關片段，建議放寬關鍵詞或移除標籤過濾。

- 設定檔對應（config.yaml）：
  - `ollamaHost`: `http://127.0.0.1:11434`
  - `model`: `llama3`
  - `data.notesDir`: `data/notes`、`data.indexDir`: `data/index`
  - `tui.width`: `80`
  - 未提供時採用內建預設；可參考 `config.example.yaml`。

- 測試（TDD 重點）：
  - agent：mock HTTP 驗證 `/api/chat` payload（system/user/messages、model、options）、正/誤回應解析與逾時處理。
  - prompt：模板載入、非法模板 fallback 與警示訊息。
  - config：YAML 覆寫預設、缺欄位容錯。
  - search：Top‑K 與 tags 過濾、空結果行為；排名穩定性以最小保證撰寫。

---

## 技術架構

流程總覽（概念示意）：

```
[User Query]
    │
    ├─ 正規化（trim/case）
    │
    ├─ 檢索（Index）
    │      └─ 目前：in-memory stub（M2/M3）
    │      └─ 未來：Bleve 持久化索引（M4/M5）
    │
    ├─ 片段整合（Top‑K → context）
    │
    ├─ 模板渲染（prompt/ask.zh-tw.yaml）
    │
    └─ LLM 推論（Ollama /api/chat） → 回覆
```

Bleve 索引結構（依本專案 Note 資料模型調整）：

```
[ Index ]   ←——（最小儲存與搜尋單位）
    │
    ├── Document #a1 (NoteID: "a1")
    │       ├── content : "golang bleve search"
    │       └── tags    : ["dev"]
    │
    ├── Document #a2 (NoteID: "a2")
    │       ├── content : "golang unit test"
    │       └── tags    : ["test", "dev"]
    │
    └── Inverted Index（倒排索引）
            ├── "golang" → [a1.content, a2.content]
            ├── "bleve"  → [a1.content]
            ├── "test"   → [a2.content, a2.tags]
            └── "dev"    → [a1.tags, a2.tags]
```

欄位對應（Mapping 建議）：
- content: 全文（full-text，參與斷詞與排名）
- tags: 關鍵字（keyword，不分詞，支援過濾）
- created_at: 可排序欄位（sortable），便於時間序查詢

說明：目前測試階段以 in‑memory `Index` 介面支撐功能與 TDD，待介面穩定後替換為 Bleve 實作；對外介面保持不變。

---

## Bleve 導入計畫（從 in-memory 過渡）

目標：在不改動對外 `search.Index` 介面的前提下，將目前的 in‑memory 索引替換為 Bleve，並保留快速回滾能力。

- 推進步驟（小步、可回滾）：
  - 保持介面：維持 `type Index interface { IndexNote(model.Note) error; Query(...); Close() }` 不變。
  - 新增實作：在 `search/` 內新增 `bleveIndex`（後台採用 Bleve），`OpenOrCreate(path)` 依 `path` 決定回傳何種實作：
    - `path == ""` 或不可寫 → 回退為 in‑memory。
    - `path != ""` → 嘗試開啟/建立 Bleve 索引。
  - 一次性重建：首次導入時，由呼叫端（CLI/服務啟動流程）以 `storage.List()` 讀筆記並 `IndexNote(...)` 進行重建（小量資料足夠）。
  - 設定整合：透過 `config.Config.Data.IndexDir` 指定索引路徑；留空即採 in‑memory。

- Bleve 實作要點：
  - Mapping：`content`（全文、參與排名）、`tags`（keyword 不分詞）、`created_at`（sortable）。
  - OpenOrCreate：不存在→建立索引並套用 mapping；存在→開啟。
  - IndexNote：以 `note.ID` 作為 DocID，寫入 `content/tags/created_at` 欄位。
  - Query：
    - `q` 非空→全文查詢（可用 `MatchQuery` 或 `QueryStringQuery`）。
    - `tags` 非空→以 AND 串接多個 `TermQuery` 過濾。
    - `topK`→限制回傳數量；排序採預設相關度，必要時加次序欄位確保穩定性。
  - Close：委派至 Bleve 的 `index.Close()`。

- 風險與緩解：
  - 排序/斷詞差異：不同分析器可能改變召回與排名 → 測試僅檢查最小行為（長度、NoteID 存在），降低耦合。
  - I/O/權限：索引路徑不可寫時應自動回退 in‑memory 並打印警示。
  - 中文分析：若需更佳中文斷詞，再以 analyzer 做為可選優化，不作為導入前置條件。

- 回滾策略：
  - 組態回滾：`IndexDir` 置空立即回復 in‑memory。
  - 程式回滾：`bleveIndex` 與 in‑memory 實作分檔獨立，移除/停用 Bleve 檔案即可。
  - 資料安全：Bleve 索引可重建，無需遷移；原始資料以 JSONL 為準。

說明：依專案「避免過早抽象」原則，先以 in‑memory 驗證介面與測試，再按上述步驟換成 Bleve。全程不更動對外介面，可隨時回退。


## 第三方相依（建議）

- 必備/計畫引入：
  - YAML 解析：`github.com/goccy/go-yaml`（選定，用於 `config.Load`；取代已封存的 `gopkg.in/yaml.v3`）。
  - 全文檢索：`blevesearch/bleve`（M4/M5 導入，用於 `search` 實作）。
  - TUI 框架：`github.com/charmbracelet/bubbletea`、`github.com/charmbracelet/bubbles`、`github.com/charmbracelet/lipgloss`（M5 導入）。
- 建議（可選）：
  - CLI：`github.com/spf13/cobra`（M3 視情況加入，作為薄層 shim）。
  - UUID v4：`github.com/google/uuid`（或以 `crypto/rand` 自行實作 v4）。
  - 檔案監控：`github.com/fsnotify/fsnotify`（如需設定/資料熱更新）。
  - 中文斷詞/分析器：`gojieba` 或 Bleve 專用 analyzer（需提升中文召回/精度時）。
  - 終端樣式（非 TUI 場景）：`github.com/fatih/color`。
- 依賴策略：
  - 初期盡量少依賴：`search` 以 in‑memory stub 起步，待介面穩定再導入 Bleve。
  - 封裝點：在 `config` 內集中 `unmarshalYAML`，未來若需替換 YAML 套件只動一處。

---

## 資料與隱私

- 筆記：儲存於 `data/notes/*.jsonl`（一行一筆；`id, content, tags, created_at`）。
- 索引：存於 `data/index/`（導入 Bleve 後）。

---

## Troubleshooting

- 無法連到 Ollama：
  - 確認執行 `ollama serve` 並監聽 `11434`；確認模型已 `ollama pull <model>`。
- 權限或路徑問題：
  - 請確認 `data/` 可寫入；在唯讀環境請切換到可寫目錄後重試。
- Windows/WSL 提示：
  - 建議於 WSL2 或 macOS/Linux 直接使用較穩定。

---

## 進度同步（AI 專用）

- 下一步（待辦，精簡）：
  - Chat：整合 ask/LLM 回覆（保留 `--no-llm`），加 mock 測試。
  - 文件：更新 README 的 TUI 操作與鍵位（Enter/Alt+Enter/Esc/PgUp/PgDn）與底部 help 列，必要時附示意圖。
  - AddWizard：擴充標籤解析案例與邊界測試，文件化正規化/排序規則。
  - 效能：評估儲存/索引延遲初始化以降低啟動成本（僅在確認儲存時初始化）。
  - TUI UX：儲存成功後是否暫留提示再退出、錯誤後提示可重試的文案。
