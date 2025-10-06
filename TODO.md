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

## TUI 測試待修清單 (已完成)

以下問題導致 `internal/tui` 測試編譯失敗，已修復：

- `internal/tui/model.go:78`：未使用變數 `selectedTitle`。實際上變數已被使用，問題已解決。
- `internal/tui/model_test.go:53`：未使用變數 `m`。實際上變數已被使用，問題已解決。
- `internal/tui/model_test.go:110` 與 `:125`：`tea.KeyMsg` 結構初始化錯誤。已修復為使用 `tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}` 等正確格式。
- `internal/tui/model_test.go`：`storage.TestDataDir` 和 `storage.WriteNote` 未定義。已添加 `SetTestDataHome` 和 `GetTestDataHome` 函數，並創建 `writeTestNote` 輔助函數使用 `SaveNote`。

修復後，`internal/tui` 測試通過，覆蓋率 57.1%。

註：上述問題不影響 `internal/note` 與 `internal/storage` 測試結果與覆蓋率。

### 儲存邏輯擴充與測試修復

**目標：** 實作筆記的讀取與列表功能，並修復 `internal/storage` 套件中的測試失敗。

**已完成任務：**

1.  **在 `internal/storage/notes.go` 中實作 `ListNotes` 函數：** (已完成)
    *   讀取資料目錄中所有 `.md` 檔案。
    *   從檔案名稱中解析筆記標題（例如：`YYYYMMDDHHmmss-Title.md` -> `Title`）。
    *   返回所有筆記標題的列表。

2.  **在 `internal/storage/notes.go` 中實作 `ReadNote` 函數：** (已完成)
    *   根據給定的筆記標題，在資料目錄中尋找對應的 `.md` 檔案。
    *   讀取檔案內容。
    *   解析並移除 YAML front matter。
    *   返回筆記的純內容。

3.  **修復 `internal/storage/notes_test.go` 中的測試失敗：** (已完成)
    *   分析 `TestSaveNote_WriteFileError` 失敗的原因（`ensureDir` 重置權限覆蓋測試設定）。
    *   修正測試邏輯，使用臨時目錄並調整 `ensureDir` 以在測試中保留權限。

4.  **重新運行所有測試：** (已完成) 確保 `internal/storage` 測試通過；`internal/tui` 測試因已知編譯錯誤暫未修復。

### 移除 `io/ioutil` 棄用套件

**目標：** 將專案中所有 `io/ioutil` 的使用替換為 `io` 或 `os` 套件中對應的功能，以符合 Go 1.16+ 的最佳實踐。

**待處理任務：**

1.  **識別所有 `io/ioutil` 的使用：** (已完成)
    *   透過程式碼搜尋，找出所有導入和使用 `io/ioutil` 的地方。

2.  **替換為 `io` 或 `os` 的對應功能：** (已完成)
    *   根據具體使用場景，將 `ioutil.ReadFile` 替換為 `os.ReadFile`。
    *   將 `ioutil.WriteFile` 替換為 `os.WriteFile`。
    *   將 `ioutil.TempDir` 替換為 `os.MkdirTemp`。
    *   將 `ioutil.TempFile` 替換為 `os.CreateTemp`。
    *   將 `ioutil.NopCloser` 替換為 `io.NopCloser`。
    *   將 `ioutil.ReadAll` 替換為 `io.ReadAll`。

3.  **更新導入語句：** (已完成)
    *   移除不再需要的 `"io/ioutil"` 導入。

4.  **運行測試並驗證：** (已完成)
    *   執行 `go mod tidy` 清理依賴。
    *   執行 `go fmt ./... && go vet ./...` 格式化與靜態檢查。
    *   執行 `go test ./... -cover` 確保所有測試通過，並且沒有新的編譯錯誤或警告。