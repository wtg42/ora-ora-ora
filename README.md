# ORA-ORA-ORA

這是一個本機端運行的個人速記與問答系統。使用者可用 CLI/TUI 快速新增筆記，並以自然語言查詢，由本地 AI 模型回覆相關內容。

## 底層技術

- Ollama（本機 LLM 推論 API）
- Bleve（全文檢索；先以 in-memory 介面過渡）

---

## 專案狀態

所有 Golang 原始碼檔案已被刪除。目前的專案結構只剩下文件、設定檔和資料。

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

## 當前開發步驟（Roadmap / Status）

- 里程碑進度：
  - ✅ M1 文件落地：專案指南與介面契約。
  - M2 介面骨架：建立 `storage/`, `search/`, `agent/`, `config/` 最小可用實作（下一步）。
  - M3 指令最小版：`ora add` 寫檔＋索引、`ora ask` 顯示檢索片段（排程中）。
  - M4 LLM 串接：非串流回覆、模板 `prompt/ask.zh-tw.yaml` 參數化（規劃中）。
  - M5 TUI 整合：AddNote 與查詢頁（規劃中）。

- 短期待辦（實作順序建議）：
  - 初始化 Go 模組與資料夾骨架（不動 TUI）。
  - `config.Load`：預設值＋YAML 覆寫，table‑driven 測試。
  - `storage`：JSONL `Save/List`，使用暫存目錄測試 I/O 失敗分支。
  - `search`：定義 `Index` 介面，先提供 in‑memory stub 與測試（之後換 Bleve）。
  - `agent`：`LLM.Chat` 介面與 mock 測試（不依賴網路）。
  - `cmd`：`ora add/ask` 骨架，先接上 storage/search，LLM 之後接。
  - 品質檢查：`go fmt`, `go vet`, `go test ./...`。

- 風險與備註：
  - 目前無 Golang 程式碼：遵循小步提交，自 M2 起逐步恢復功能。
  - 僅使用本機資料與路徑，避免觸碰敏感檔案與外部網路。
  - 模組保持鬆耦合，未來可替換檢索引擎或模型而不影響 CLI/TUI。

---

## 資料與隱私

- 筆記：儲存於 `data/notes/*.jsonl`（一行一筆；`id, content, tags, created_at`）。
- 索引：存於 `data/index/`（導入 Bleve 後）。
- 所有資料與推論皆於本機完成，無雲端上傳。

---

## Troubleshooting

- 無法連到 Ollama：
  - 確認執行 `ollama serve` 並監聽 `11434`；確認模型已 `ollama pull <model>`。
- 權限或路徑問題：
  - 請確認 `data/` 可寫入；在唯讀環境請切換到可寫目錄後重試。
- Windows/WSL 提示：
  - 建議於 WSL2 或 macOS/Linux 直接使用較穩定。
