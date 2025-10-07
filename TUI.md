# 對話式筆記 TUI — 組件拆解與功能說明

下面各區塊都是你在最終介面會看到的「組件（component）」或「模組（module）」。每一個組件會講它的責任、互動、邊界條件、以及你可以從 Codex CLI 那類工具借鑑的做法。

---

## 歷史對話區域（Conversation / History Pane）

**責任 / 功能：**

- 顯示對話訊息／筆記條目
- 用戶對話跟系統回應採用不同文字顏色區別

**內容結構：**

每條訊息可包含：

- 內容文字（可能有多行）

---

## 輸入區域（Input / Composer Pane）

**責任 / 功能：**

- 永遠固定在界面底部（無論畫面高度怎樣）  
- 支援多行輸入（textarea）  
- 支援控制鍵：  
  - `Enter` → 送出  
  - `Ctrl+J` → 換行  
- 顯示 placeholder 或提示文字（例如 “在這裡輸入…Enter 送出 / Shift+J 換行”）  
- 當輸入多行時候，此區域會增加高度  
- 利用 BubbleTea API 產生底域 help view 說明
