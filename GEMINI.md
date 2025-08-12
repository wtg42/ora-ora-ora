## 📘 ORA-ORA-ORA：CLI AI 筆記系統

這是一本地端運行的個人速記與問答系統。使用者可以透過 CLI 指令快速新增筆記，並以自然語言查詢自己的想法或記錄，由本地 AI 模型回答對應內容。

---

### ⚙️ 技術堆疊

- **Golang**：實作 CLI 工具、處理命令、載入設定、發送 API 請求
- **Bleve**：全文索引工具，儲存與查詢使用者筆記
- **Ollama API Server**：本地 LLM 推論引擎，處理語意與自然語言問答

---

### 🧠 功能流程

1. 使用者下指令新增筆記（`ora add`）
2. 使用者下指令查詢筆記（`ora ask`）
3. Golang 程式：
   - 使用 Bleve 搜尋與提問相關的筆記段落（`context`）
   - 套用 `prompt/ask.zh-tw.yaml` 或 `prompt/ask.en.yaml`
   - 組成符合 Ollama `/api/chat` 規格的 JSON 請求
   - 接收模型回覆後輸出

---

### 📂 設定檔說明

#### `prompt/ask.zh-tw.yaml` 範例

```yaml
system: |
  你是一個筆記助理，會根據使用者的筆記記錄回答他提出的問題。
template: |
  問題：{{question}}
  筆記內容：{{context}}
model: hf.co/bartowski/Llama-3.2-3B-Instruct-GGUF:Q4_K_M
language: zh-tw
```

#### `prompt/weekly.zh-tw.yaml`（週記模式）

```yaml
system: |
  你是一個知識摘要工具，請根據以下筆記整理一週的重點。
template: |
  以下是本週的筆記內容：
  {{context}}
  請整理為條列式摘要，包含主題與要點。
model: hf.co/bartowski/Llama-3.2-3B-Instruct-GGUF:Q4_K_M
```

---

### 📤 送出 Ollama API 請求

```http
POST http://localhost:11434/api/chat
Content-Type: application/json

{
  "model": "llama3",
  "messages": [
    { "role": "system", "content": "你是一個筆記助理..." },
    { "role": "user", "content": "問題：我昨天寫了什麼？
筆記內容：..." }
  ]
}
```

回傳格式為：

```json
{
  "message": {
    "role": "assistant",
    "content": "你昨天寫了關於 Bleve 與 Golang 的索引處理..."
  }
}
```

---

### 🏷 支援功能總結

- ✅ 中文與英文自然語言提問
- ✅ tags / 分群搜尋條件（Bleve 篩選）
- ✅ YAML prompt 模板可調整語氣與行為
- ✅ 可整合 cron 自動執行 `ora weekly`（週記總結）
- ✅ 所有資料本地儲存，無須網路

---

如需為不同角色或應用情境擴充 prompt 模板，只需在 `prompt/*.yaml` 加入對應檔案，並於 Golang 程式中選擇載入即可。

# GEMINI CLI Agent 參考文件（ORA‑ORA‑ORA 專案）

> 此文件提供 **GEMINI CLI Agent** 在本機運作時所需的最小規格（MVP）＋可擴充欄位。Agent 可依本規格完成：檢索 → 組 Prompt → 呼叫本機 LLM（Ollama）→ 回傳答案。所有資料與推論均可在本機完成。

---

## 0. 目的（Purpose）
- 描述 CLI Agent 應如何與本系統互動：指令流程、設定檔格式、API 端點、錯誤處理與擴充點。
- 目標：**先跑得動的 MVP**，並保留未來擴充（多模型路由、Rerank、觀測）的空間。

---

## 1. 系統構成（Stack）
- **Golang CLI**：`ora add` / `ora ask`；載入設定、檢索、組請求、輸出。
- **Bleve**：全文索引與檢索；產生 `context` 片段。
- **Ollama API**（本機）：`/api/chat` 進行回覆生成。

> 預設為本機運行；網路離線亦可使用。

---

## 2. 指令與流程（Flow）
1) `ora add <note>`：寫入筆記並更新索引。
2) `ora ask <question>`：
   - 以 Bleve 查詢相關片段（Top‑k；見 §5）。
   - 套用對應模板（§3）。
   - 呼叫 Ollama `/api/chat`（§4）。
   - 將回答輸出（預設繁體中文，除非模板或旗標覆寫）。

---

## 3. 設定檔格式（YAML）
> 放於 `prompt/*.yaml`；Agent 需讀取並注入變數。可依 `--template` 指定檔名，或依語系自動選擇。

### 3.1 問答模板 `prompt/ask.zh-tw.yaml`
```yaml
system: |
  你是一個筆記助理，只能根據「提供的筆記內容」回答。
  若資訊不足，請誠實說明，勿臆測。
  最終請使用「繁體中文」回答。
template: |
  【問題】\n{{question}}\n\n【相關筆記片段】\n{{context}}\n\n【回答規則】\n- 先給出直接答案，再補充依據（列出來源片段的關鍵句）。\n- 若片段彼此矛盾，請標示「衝突」並各自說明。\n- 若無足夠資訊，請要求我補充關鍵字或筆記連結。
model: hf.co/bartowski/Llama-3.2-3B-Instruct-GGUF:Q4_K_M
language: zh-tw
# 可選：可讓 Agent 覆寫 Ollama options（§4.2）
options:
  temperature: 0.2
  top_p: 0.9
  num_ctx: 8192
  keep_alive: 10m
```

### 3.2 週記模板 `prompt/weekly.zh-tw.yaml`
```yaml
system: |
  你是一個知識摘要工具，請將筆記整理成「本週回顧」。
template: |
  【本週筆記】\n{{context}}\n\n【輸出格式】\n- 主題（標題句）\n- 關鍵要點（條列）\n- 待辦與下一步（條列）
model: hf.co/bartowski/Llama-3.2-3B-Instruct-GGUF:Q4_K_M
language: zh-tw
```

> 變數：`{{question}}`, `{{context}}` 必須可被替換；`language` 建議在 system 內生效（已範例化）。

---

## 4. LLM API 規格（Ollama）
### 4.1 端點
```
POST http://localhost:11434/api/chat
Content-Type: application/json
```

### 4.2 請求（非串流）
```json
{
  "model": "hf.co/bartowski/Llama-3.2-3B-Instruct-GGUF:Q4_K_M",
  "messages": [
    { "role": "system", "content": "你是一個筆記助理...（取自 YAML system）" },
    { "role": "user",   "content": "【問題】...\n【相關筆記片段】...（由 template 渲染）" }
  ],
  "stream": false,
  "options": {
    "temperature": 0.2,
    "top_p": 0.9,
    "num_ctx": 8192,
    "num_predict": 1024,
    "keep_alive": "10m"
  }
}
```

### 4.3 回應（非串流）
```json
{
  "message": {
    "role": "assistant",
    "content": "最終答案..."
  }
}
```

### 4.4 串流（可選）
- 將 `stream` 設為 `true`，回應將分段送出（SSE 風格）。
- Agent 可即時顯示，結尾組合成完整訊息。

---

## 5. 檢索與注入（RAG）
> MVP 可用固定參數，後續可 CLI 旗標覆寫（如 `--top-k=5 --max-tokens=3000`）。

- **Chunk**：建議 600–1,200 字符，overlap 10–20%。
- **Top‑k**：預設 5；依分數排序取前 k 片段。
- **去重**：同檔案連續命中合併；相似度高者去重。
- **截斷**：依 `num_ctx` 估算最大可注入 tokens，超過則尾端截斷並標記「已截斷」。
- **邊界**：以明確分隔符包住每段 context，降低 prompt 注入風險，例如：
  ```
  === BEGIN SNIPPET: <title>#<line-range> ===
  <content>
  === END SNIPPET ===
  ```

---

## 6. 錯誤處理（Agent 指南）
- **Ollama 未啟動**：顯示提示 `請先執行: ollama serve`。可自動重試 1 次（延遲 1s）。
- **模型不存在**：提示 `ollama pull <model>`，並列出目前可用模型清單（若可取得）。
- **逾時**：回報逾時並建議降低 `num_predict`／縮小 context。
- **空檢索**：告知「筆記不足」，建議先 `ora add` 或提供關鍵詞。
- **回應過長**：要求 Agent 轉為串流或降低 `num_predict`。

---

## 7. 安全與治理
- **Prompt 注入防護**：以系統訊息明確規範「只依 context 回答」，外部指令需忽略。
- **來源標註**：回答附來源片段關鍵句（可選）。
- **隱私**：所有資料留在本機，不上傳網路。

---

## 8. CLI Agent 行為建議（即插即用）
- 預設回答語系由模板 system 決定（此處為「繁體中文」）。
- 可支援旗標覆寫：`--model`、`--top-k`、`--stream`、`--num-predict`、`--temperature`。
- 若問題過大或過泛，Agent 應主動要求**關鍵詞**或**縮小範圍**。
- 週記模式：載入 `prompt/weekly.zh-tw.yaml`，context 為近 7 天筆記（或依旗標 `--days`）。

---

## 9. 擴充路線圖（可選）
- **多模型路由**：短問快答→3B；複雜推理→7B/8B。
- **Rerank 模組**：以輕量 cross‑encoder 進行重排序。
- **觀測**：紀錄 latency、token 用量；便於優化 P50/P95。
- **快取**：相同（query, filter）命中直接回傳，降低延遲。

---

> **Agent 實作提示**：直接依此資料結構組裝 JSON 後呼叫 `/api/chat`；若 `stream=true` 則走串流顯示。

---

## 11. 支援功能總結
- ✅ 中英雙語問答（模板驅動）
- ✅ 檢索條件（tags/分群），Top‑k 注入
- ✅ YAML 模板可定義語氣、輸出結構與 LLM 選項
- ✅ 週記模式（cron 可排程）
- ✅ 完全本機化，離線可用

## 12. 開發注意事項
- 修改檔案已一隻檔案為主，若需要修改多個檔案，第二個檔案之後則是先給予建議，不要直接接續修改。
