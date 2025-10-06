# 代辦事項

此檔案用於記錄專案的代辦事項和切分後的任務。
**請注意：每完成一個任務後，請務必在此檔案中更新進度。**

## 待處理任務

### 筆記寫入功能

1.  **定義筆記結構**: (已完成)
    *   決定筆記的格式（例如：帶有 metadata 的 Markdown）。
    *   考慮包含標題、內容、標籤、建立時間戳記等欄位。

2.  **CLI 筆記寫入指令**: (已完成)
    *   建立一個新的 CLI 指令（例如：`ora note new` 或 `ora note add`）。
    *   允許使用者輸入筆記內容（例如：透過 stdin 或開啟編輯器）。

3.  **儲存邏輯整合**: (已完成)
    *   利用 `internal/storage/xdg.go` 取得資料目錄路徑。
    *   實作將筆記內容儲存到資料目錄中檔案的邏輯。
    *   考慮筆記檔案的命名慣例（例如：`timestamp-title.md`）。

4.  **基本錯誤處理**: (已完成)
    *   處理資料目錄無法建立或存取的情況。
    *   處理檔案寫入錯誤。

5.  **測試**: (已完成)
    *   已為筆記寫入與儲存邏輯編寫單元測試（`internal/note` 覆蓋率 100.0%，`internal/storage` 覆蓋率 86.7%）。
    *   待辦：`internal/tui` 測試編譯失敗需修復（不影響筆記相關模組覆蓋率）。

### XDG 路徑重構 (已完成)

**目標：** 確保應用程式在所有作業系統（包括 macOS）上，配置檔案統一存放於 `~/.config/ora-ora-ora`，資料檔案統一存放於 `~/.local/share/ora-ora-ora/notes`。此舉旨在提供更一致的跨平台使用者體驗，並遵循 Linux 環境下更普遍的 XDG 慣例。

**原因：** 儘管 `adrg/xdg` 函式庫在 macOS 上會預設將路徑導向 `~/Library/Application Support`，但使用者明確要求所有平台統一使用 `~/.config` 和 `~/.local/share` 作為基礎目錄。

**實作完成：**

1.  **修改 `internal/storage/xdg.go`：**
    *   移除對 `github.com/adrg/xdg` 函式庫的依賴。
    *   在 `GetConfigDir()` 函式中，使用 `os.UserHomeDir()` 和 `filepath.Join` 建構 `~/.config` 作為基礎配置目錄。
    *   在 `GetDataDir()` 函式中，使用 `os.UserHomeDir()` 和 `filepath.Join` 建構 `~/.local/share` 作為基礎資料目錄。
    *   保持應用程式名稱 (`ora-ora-ora`) 和 `notes` 子目錄的附加邏輯不變。

2.  **驗證與測試：**
    *   執行所有單元測試 (`go test ./... -cover`)，確保路徑建構邏輯正確。
    *   手動驗證在 macOS 上，`GetConfigDir()` 返回 `~/.config/ora-ora-ora`，`GetDataDir()` 返回 `~/.local/share/ora-ora-ora/notes`。

## TUI 測試待修清單

以下問題導致 `internal/tui` 測試編譯失敗，需後續修復：

- `internal/tui/model.go:78`：未使用變數 `selectedTitle`。
- `internal/tui/model_test.go:53`：未使用變數 `m`。
- `internal/tui/model_test.go:110` 與 `:125`：`tea.KeyMsg` 結構初始化使用不存在欄位 `String`，需依目前 bubbletea 版本改用正確欄位（如 `Type`、`Rune` 等）或以事件幫手函式建立鍵盤訊息。

註：上述問題不影響 `internal/note` 與 `internal/storage` 測試結果與覆蓋率。
