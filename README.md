# ORA-ORA-ORA

這是一個本機端運行的個人速記與問答系統。使用者可用 CLI/TUI 快速新增筆記，並以自然語言查詢，由本地 AI 模型回覆相關內容。

## 底層技術

- Ollama（本機 LLM 推論 API）
- Bleve（全文檢索；先以 in-memory 介面過渡）

---

## 專案狀態

所有 Golang 原始碼檔案已被刪除。目前的專案結構只剩下文件、設定檔和資料。

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
