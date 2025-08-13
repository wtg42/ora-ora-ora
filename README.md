# ORA-ORA-ORA

這是一個個人速記筆記程式，用戶可以使用 CLI 介面新增臨時發想到的 idea，之後可以快速查詢自己想要的內容  
用戶只需要依照自己的自然語言提問要搜尋的答案，本機 AI 就會幫忙回答最適當的問題。

## 底層技術

- Golang
- Ollama (API Server)
- Bleve (全文檢索用)
 
## Quickstart

- 需求：
  - Go 1.22+（建議最新版）
  - 可選：本機已安裝並啟動 Ollama（預設 http://localhost:11434）。目前僅啟動 TUI 不需要 Ollama。
- 建置與執行：
  - 啟動 TUI 原型：`go run .` 或 `go run . start-tui`
  - 編譯：`go build -o ./bin/ora-ora-ora .`
- 狀態：目前提供「新增筆記」的 TUI 原型；CLI `add/ask` 功能將於後續版本加入（WIP）。

## 使用方式

- 啟動 TUI：
  - `go run . start-tui`
  - 在輸入欄輸入內容並按 Enter 即可結束；輸入中的 `#tag` 會自動解析為標籤。
- 預告 CLI 指令（WIP，之後版本加入）：
  - `ora add "今天研究 Bleve #search #golang"`
  - `ora ask "昨天我做了什麼？"`

## 設定（預留）

- 預計路徑：`config/config.yaml`（不自動讀取 `.env`，避免敏感資訊外洩）
- 範例鍵值：
  - `ollama.host: "http://localhost:11434"`
  - `model: "hf.co/bartowski/Llama-3.2-3B-Instruct-GGUF:Q4_K_M"`
  - `data.notes_dir: "data/notes/"`
  - `data.index_dir: "data/index/"`
  - `tui.width: 80`

## 資料與隱私

- 筆記：預計儲存於 `data/notes/*.jsonl`（一行一筆；`id, content, tags, created_at`）。
- 索引：預計儲存於 `data/index/`（Bleve 索引；單機檔案、無外部 DB）。
- 所有資料與推論皆可在本機完成，無雲端上傳。

## Troubleshooting

- 無法連到 Ollama：
  - 確認執行 `ollama serve` 並監聽 `11434`；確認模型已 `ollama pull <model>`。
- 首次建立索引較慢：
  - 正常現象，初次建立完成後為增量更新。
- 權限或路徑問題：
  - 請確認 `data/` 可寫入；在唯讀環境請切換到可寫目錄後重試。
- Windows/WSL 提示：
  - 建議於 WSL2 或 macOS/Linux 直接使用較穩定。
