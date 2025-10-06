% ora-ora-ora

這是一個以 Go 語言為主的 AI 快速筆記應用專案骨架。它旨在透過呼叫 shell 方式，利用 stdout 輸出，與 AI CLI agent 互動，以快速查詢和寫入筆記內容。請優先閱讀 `AGENTS.md` 以瞭解貢獻與開發規範。

## 開發方針（重要）
- MVP 優先：以最小可行產品逐步交付，單次改動聚焦單一能力。
- TDD 開發：Red → Green → Refactor。新增/變更功能先寫失敗的單元測試。
- 測試政策（MVP 前）：僅要求單元測試；功能/整合/E2E 測試於 MVP 驗證後逐步補強。
- 變更切小：避免一次性大改（尤其是由 LLM 產生的大量變更）。

## 常用指令
```bash
go mod tidy                   # 安裝/整理依賴
go fmt ./... && go vet ./... # 格式化與靜態檢查
go test ./... -cover         # 單元測試與覆蓋率
```

## 結構概覽
- `cmd/`：可執行程式入口。
- `internal/`：專案內部套件，包含核心邏輯，例如 `internal/storage` 負責檔案儲存。
- `pkg/`：可重用公開套件（如需）。
- `*_test.go`：單元測試檔。

### 筆記與資料儲存
本專案遵循 XDG Base Directory Specification 管理應用程式的配置和資料。筆記內容（Markdown 格式）將儲存在由 `internal/storage/xdg.go` 定義的資料目錄中。
具體來說，筆記會存放在類似 `~/.local/share/ora-ora-ora/notes/`，確保檔案系統的整潔與跨平台相容性。

更多細節請見 `AGENTS.md`（風格、提交流程、安全與回滾策略）。
