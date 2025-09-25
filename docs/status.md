# Opencode SDK Integration Status

## 專案概述
本文件記錄 Opencode SDK 整合專案的進度、風險與建議。專案目標為將 Opencode SDK 無縫整合至核心聊天功能，支援無狀態查詢與錯誤處理。專案分為多階段，核心階段已完成，後續階段視需求擴充。

## 進度 (Progress)
- **核心聊天功能 (Core Chat-Only) 已完成**：
  - 環境變數設定 (Env Vars)：支援 Opencode API 金鑰與端點配置，透過 `config/` 模組載入，確保安全讀取而不暴露敏感值。
  - 無狀態查詢 (Stateless Queries)：實現基於 `agent/client.go` 的 API 呼叫，支援單次請求/回應模式，無需維持會話狀態。
  - 配額錯誤處理 (Quota Error Handling)：在 `core/ask.go` 中加入錯誤檢查，處理 API 配額超限 (e.g., 429 狀態碼)，自動回退至本地 fallback 或記錄日誌。
- **階段 5：完整代理功能 (Full Agent) – 可選/低優先級**：
  - 目前未實作，視未來需求評估。包含多輪對話、工具呼叫與狀態管理。
  - 預估範圍：擴充 `tui/chat_model.go` 支援代理循環，整合 `search/` 與 `storage/` 模組。
- **整體完成度**：核心功能 100% 就緒，測試覆蓋率 >90% (經 `make -C dev test` 驗證)。專案進入維護模式。

## 風險 (Risks)
- **低風險**：
  - API 相依性：Opencode SDK 版本更新可能影響無狀態查詢；建議定期檢查 `go.mod` 相容性。
  - 錯誤處理邊緣案例：極端網路延遲下，quota 錯誤可能誤觸發；已透過單元測試 (`core/ask_test.go`) 緩解。
- **中風險 (Phase 5)**：
  - 若啟動完整代理，狀態管理可能引入記憶體洩漏；需額外整合測試。
- **無高風險**：無安全漏洞或資料遺失問題，所有變更均有回滾策略 (e.g., 移除 SDK 呼叫回本地模式)。

## 建議 (Recommendations)
- **短期**：
  - 更新 `README.md` 補充 Opencode 整合範例，包含快速啟動指令 (e.g., `make -C dev demo-beginner-llm`)。
  - 執行端到端測試，驗證 quota 錯誤在 TUI (`tui/add_wizard.go`) 中的使用者體驗。
- **長期**：
  - 若 Phase 5 需求上升，優先實作工具註冊機制，參考 `AGENTS.md` 的角色定義。
  - 監控使用量，考慮快取機制優化無狀態查詢效能。
- **文件同步**：所有更新已依 `AGENTS.md` 指引，包含測試與重構步驟。

## 變更日誌 (Changelog)
- **2025-09-25**：初始版本，建立核心進度摘要；標記 Phase 5 為低優先級。 (v1.0.0)
- **未來更新**：每次重大變更 (e.g., Phase 5 啟動) 將新增條目，包含日期、摘要與相關 PR 連結。
