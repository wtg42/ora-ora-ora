# Repository Guidelines

本文件為 ora-ora-ora 專案的貢獻者指南，旨在打造一個 AI 快速筆記應用，聚焦最小風險與高可維護性。語言以 Go 為主，隨著專案演進請更新本檔。

## Project Structure & Module Organization
- 根目錄：Go 模組與工具設定。
- `cmd/`：可執行程式進入點（每個子資料夾一個 binary）。
- `internal/`：專案內部套件（不可被外部匯入）。
- `pkg/`：穩定可重用的公開套件（如需）。
- `test/` 或各套件內 `*_test.go`：單元/整合測試。
- `assets/`、`configs/`：靜態資源與設定（如需）。

## Build, Test, and Development Commands
```bash
go mod init github.com/user/ora-ora-ora   # 初始化模組（首次）
go mod tidy                               # 安裝/整理依賴
go fmt ./... && go vet ./...              # 格式化與基本靜態檢查
go build ./...                            # 建置所有套件
go test ./... -cover                      # 執行測試與覆蓋率
```

## Coding Style & Naming Conventions
- 使用標準 Go 風格：`go fmt` 強制，`go vet` 輔助。
- 匯入分組：標準庫 / 第三方 / 本地。
- 命名：匯出符號大寫開頭；其餘使用 camelCase；檔名/資料夾使用短小、語義清晰的小寫。
- 錯誤處理：`if err != nil { ... }`，庫碼避免 `panic`，使用 `fmt.Errorf` 包裝情境。

## Testing Guidelines
- 採用 TDD 流程（Red → Green → Refactor）；新增/變更功能先寫失敗測試。
- 使用內建 `testing`；測試檔名 `*_test.go`，函式 `TestXxx`。
- 覆蓋率目標 ≥ 80%；關鍵路徑與錯誤分支需覆蓋。
- 範例：`go test ./path/to/pkg -run '^TestSpecific$' -v`。
- 目前階段以單元測試為主；功能/整合/E2E 測試待 MVP 驗證可行後再補強。

## Commit & Pull Request Guidelines
- Commit：聚焦單一變更；訊息以動詞祈使句開頭（e.g., "Add", "Fix"）。
- PR：提供摘要、動機、影響範圍、測試證據（指令輸出/截圖）、相關 Issue 連結與風險/回滾策略。
- 通過 `go fmt`, `go vet`, `go test` 後再提交。
- MVP 前 PR 至少包含相應單元測試；功能/整合測試可於 MVP 驗證後追加。

## Security & Configuration Tips
- 禁止提交祕密（`.env`、金鑰、憑證）；以範本 `configs/example.yaml` 提供預設。
- 環境差異以環境變數或設定檔處理；避免全域狀態與隱藏副作用。

## Agent-Specific Instructions
- 任務拆分：在實作前，請先將功能拆分為小任務並寫入 `TODO.md`。
- 進度更新：每完成一個任務後，請務必更新 `TODO.md` 中的進度。
- MVP 優先：以最小可行產品（MVP）迭代交付；單次改動聚焦單一能力。
- 避免一次實作過多（尤其使用 LLM 產生大量變更）；切小任務、保持可審查與可回滾（建議精簡 diff，單一責任）。
- 先提案後實作；小步提交；避免破壞性重構。
- 任何依賴升級需附相容性說明與變更摘要。
- **AI 心智註解**：每段程式碼必須添加簡短註解來紀錄 AI 當下心智思維以利後續追蹤。
- **計畫優先**: 任何修改之前先把計畫更新到 `TODO.md` 後，再開始實作。

### TUI Development Guidelines (參考 TUI.md)
- **組件化設計**: 將 TUI 介面拆分為獨立的組件（如：歷史對話區域、輸入區域），每個組件負責特定功能。
- **責任分離**: 每個 TUI 組件應有清晰的職責，例如歷史對話區域負責顯示訊息，輸入區域負責處理用戶輸入。
- **互動模式**: 考慮用戶與 TUI 的互動方式，例如輸入區域的控制鍵（Enter 送出, Ctrl+J 換行）。
- **響應式佈局**: TUI 介面應能適應不同的終端尺寸，例如輸入區域在多行輸入時自動調整高度。
- **提示與說明**: 提供清晰的用戶提示和幫助信息，例如輸入區域的 placeholder 文字和利用 BubbleTea API 產生的 help view。
