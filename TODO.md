# 代辦事項

此檔案用於記錄專案的代辦事項和切分後的任務。
**請注意：每完成一個任務後，請務必在此檔案中更新進度。**

## 待處理任務

### 修正建立視圖初始輸入殘留字元（優先度 P1｜Done 2025-02-14T22:44:00Z）

**背景：** 於列表視圖按下 `n` 切換到建立視圖時，原始鍵盤事件同時被傳入輸入區，導致畫面出現預設字元（例如 `n`）。

**目標：** 切換至建立視圖時輸入框應維持空白，避免誤植。

**方案（最小風險）：**
- 調整 `internal/tui/model.go` 的狀態切換邏輯：在接收到 `'n'` 時立刻重設 `InputArea` 並停止後續輸入事件傳遞。
- 補上一個回歸測試，確認切換建立視圖後輸入框仍為空。

**子任務與進度：**
1. 更新 `model.Update` 切換邏輯，重設輸入區並避免事件重複處理。（Done 2025-02-14T22:43:20Z）
2. 新增回歸測試，驗證建立視圖初始輸入為空字串。（Done 2025-02-14T22:43:45Z）

**驗收準則：**
- 實際操作按 `n` 切換建立視圖時，輸入框不再帶入任何字元。
- 新增/既有測試 `go test ./...` 通過。

**風險與回滾：**
- 變更僅限於 `internal/tui/model.go` 與對應測試檔；若出現非預期行為，可快速回退至先前版本。

### 禁止空內容筆記寫入（優先度 P1｜已完成）

**背景：** 目前 `SaveNote()` 允許空標題與空內容筆記被寫入檔案系統，導致目錄內出現 `YYYYMMDDHHmmss-.md` 等檔名，TUI 顯示為空白列項，造成使用者困惑與雜訊。

**目標：** 當提交的筆記「標題與內容皆為空」或「內容為空」時拒絕寫入，回傳清楚錯誤；TUI 與 CLI 端顯示可理解訊息。

**方案（最小風險）：**
- 在 `internal/storage/notes.go::SaveNote` 增加前置驗證：若 `strings.TrimSpace(n.Content) == ""` 則回傳 `error`（訊息：內容不可為空）。保留允許空標題（不更動既有測試假設）。
- 在 `internal/tui/model.go` 的提交流程接住錯誤並在畫面顯示（不改進入檔系統層邏輯）。
- 不自動清理既有空檔案，僅提供清理指引（避免破壞性行為）。

**子任務與進度：**
1. 在 `SaveNote` 增加內容非空驗證與錯誤訊息（已完成）。
2. 補充單元測試：
    - `TestSaveNote_EmptyContent_ShouldFail`（已完成）。
    - 仍允許空標題但非空內容的成功案例（回歸）（已完成）。
3. TUI 錯誤呈現：在 `SubmitMsg` 流程將儲存失敗訊息顯示於 `errorMessage`（已完成）。
4. 文件更新：在 `README.md`/`GEMINI.md` 補充規則與使用者行為說明（已完成）。
5. 清理指引：在 `README.md` 新增「清理空白筆記指南」（已完成）。

**驗收準則：**
- 嘗試寫入空內容筆記時回傳錯誤，且不產生檔案。
- `go test ./...` 全數通過，含新增測試案例。
- TUI 於提交空內容時顯示易懂錯誤訊息，列表不再新增空白項目。

**風險與回滾：**
- 影響面：僅限 `SaveNote` 行為與 TUI 提交時的錯誤處理，不影響檔案命名或讀取邏輯。
- 回滾策略：若造成非預期影響，可還原 `SaveNote` 驗證段落與對應測試，TUI 僅回退錯誤顯示邏輯。

**備註：** 既有空檔案不自動刪除；文件提供手動清理命令範例以避免誤刪。

### BubbleTea 匯入別名修補與最小互動測試（高優先 P0｜已完成）

**背景：** reviewer 指出 `cmd/ora/main.go` 與 `internal/tui/input_area.go` 參考 `tea`（BubbleTea 別名），但 import 未以 `tea` 別名匯入，導致 `undefined: tea`，目前無法編譯；同時新實作在多位元字元編輯與提交流程上仍有 P1 問題待解。

**目標：** 以最小變更修補匯入別名，修正多位元游標處理與提交時的空白裁剪問題，並新增最小測試確保可編譯與基本互動流程。

**子任務與進度：**
1. 在 `cmd/ora/main.go` 將 BubbleTea 匯入改為 `tea "github.com/charmbracelet/bubbletea"`。（已完成）
2. 在 `internal/tui/input_area.go` 將 BubbleTea 匯入改為 `tea "github.com/charmbracelet/bubbletea"`。（已完成）
3. 修正 InputArea 以 rune 為單位處理游標、刪除與插入；新增多位元測試。（已完成）
4. 調整提交流程保留原始前後空白與空行；新增保留空白測試。（已完成）
5. 新增最小互動測試（建構模型與 Program，不跑互動迴圈）。（已完成）
6. 執行 `go test ./... -cover` 於提權模式下驗證通過；後續可在本機以 `go test ./...` 重跑。（已完成）

**驗收準則：**
- 全專案可成功編譯。
- 多位元字元與空白保留測試通過。
- 最小互動測試通過，且不引入額外大型依賴。

**風險與回滾：**
- 變更聚焦在匯入別名、輸入區域內部邏輯與測試，不變更對外 API。
- 若新邏輯導致問題，可回退對 `cmd/ora/main.go`、`internal/tui/input_area.go`、`internal/tui/model.go` 與新增測試的改動。

### TUI 啟動點整合 (優先度第一)

**目標：** 整合 TUI 組件至主程式，新增啟動點以支援互動式筆記管理。

**架構規劃：**

1. **新增 TUI 子命令：** (已完成)
   * 在 `cmd/ora/main.go` 添加 `ora tui` 命令，使用 `tea.NewProgram(tui.InitialModel()).Run()` 啟動 BubbleTea 程式。
   * 確保錯誤處理與退出邏輯。

2. **組件整合至主模型：**
   * 在 `internal/tui/model.go` 嵌入 `InputArea` 結構體，專用於 `createView`。
   * 修改 `Update` 方法委派鍵盤事件至 `inputArea.Update()`，處理文字輸入邏輯。
   * 修改 `View` 方法整合 `inputArea.View()` 輸出至建立視圖。

3. **應用程式流程支援：**
   * 列表視圖：顯示筆記標題，支援導航與選擇。
   * 詳細視圖：顯示選中筆記內容。
   * 建立視圖：使用整合的 `InputArea` 輸入標題與內容，支援多行，提交時儲存筆記並更新列表。

4. **實作重點：**
   * 補全 `internal/tui/input_area.go` 的 `Update` 方法：實作文字編輯（鍵入、刪除、游標移動）、多行支援（Ctrl+J）、提交事件（Enter 返回輸入文字）。
   * 確保模組化，無衝突依賴。
   * 測試擴展至 TUI 互動。

**風險評估：** 低風險，為新增功能，不影響現有 CLI。

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

## TUI 開發任務

1.  **建立組件腳手架:** (已完成)
    *   創建 internal/tui/input_area.go 檔案。
    *   定義 InputArea 結構體，包含當前輸入文字、游標位置等狀態。

2.  **實作基本渲染與顯示:**
    *   實現 InputArea 的基本渲染邏輯，使其能在終端中顯示一個輸入框。
    *   在輸入框開頭顯示 > 提示符號。

3.  **處理文字輸入與編輯:**
    *   處理鍵盤輸入，將用戶鍵入的字元顯示在輸入區域。
    *   實作游標的移動（左右箭頭、Home、End）。
    *   實作文字的插入與刪除（Backspace, Delete）。

4.  **支援多行輸入:**
    *   實作 Ctrl+J 換行功能。
    *   確保輸入區域能正確處理和顯示多行文字。

5.  **實作輸入提交與共享狀態更新:**
    *   當用戶按下 Enter 鍵時，捕獲輸入的全部文字內容。
    *   將捕獲的文字內容更新到 TUI 主模型中預設的搜尋查詢狀態欄位（例如 model.SearchQuery）。

6.  **設定基礎樣式:**
    *   為輸入區域設定基礎的終端樣式（如文字顏色、背景色）。
