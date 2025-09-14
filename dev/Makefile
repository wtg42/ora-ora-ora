.PHONY: help test test-all test-race tidy vet fmt run-chat ollama-tags ollama-generate ollama-chat ollama-probe demo-beginner demo-beginner-llm detect-ollama-host detect-ollama-model demo-beginner-llm-auto demo-cases-gen ask-run-case ask-run-cases

# 專案根目錄（從 dev/ 相對推算）
ROOT := $(abspath $(CURDIR)/..)
GCACHE := $(ROOT)/.gocache

# 開發用 Makefile：統一本地/CI 測試與工具指令
# 注意：不讀取或改動敏感檔案；測試使用專案內部快取，避免使用者環境限制。

GOCACHE ?= $(PWD)/.gocache
GOFLAGS ?=

help:
	@echo "Available targets:"
	@echo "  make test       - 單元測試（套件級）"
	@echo "  make test-all   - 全域測試（含所有子套件）"
	@echo "  make test-race  - 開啟 race 檢查跑全測"
	@echo "  make vet        - go vet 靜態檢查"
	@echo "  make fmt        - go fmt 格式化（顯示改動）"
	@echo "  make tidy       - go mod tidy（不改 lockfile 排版）"
	@echo "  make run-chat   - 啟動 TUI（chat 頁）"
	@echo "  make ollama-probe    - 測試 Ollama 服務可用性 (/api/version,/api/tags)"
	@echo "  make ollama-tags     - 列出已安裝模型 (/api/tags)"
	@echo "  make ollama-generate - 以 generate 端點測試提示（可覆蓋變數）"
	@echo "  make ollama-chat     - 以 chat 端點測試訊息（可覆蓋變數）"
	@echo "  make demo-beginner   - 一鍵新手流程（新增兩筆→檢索→(可選)LLM）"
	@echo "  make demo-beginner-llm - 一鍵新手（必用 LLM，若無模型則報錯）"
	@echo "  make detect-ollama-host  - 偵測可用 Ollama 服務 Host（stdout），找不到則 exit 1"
	@echo "  make detect-ollama-model - 僅輸出第一個可用模型名（stdout），找不到則 exit 1"
	@echo "  make demo-beginner-llm-auto - 自動偵測模型後執行 demo-beginner-llm"
	@echo "  make demo-cases-gen     - 產生一組可命中的範例筆記"
	@echo "  make ask-run-case       - 執行單一情境（QUESTION=, TAGS=, TOPK=, TEMPLATE= 可覆蓋）"
	@echo "  make ask-run-cases      - 批次執行 dev/ask-cases.txt（需可用模型）"

test:
	GOCACHE=$(GCACHE) go -C $(ROOT) test ./... -count=1 $(GOFLAGS)

test-all: test

test-race:
	GOCACHE=$(GCACHE) go -C $(ROOT) test ./... -race -count=1 $(GOFLAGS)

vet:
	GOCACHE=$(GCACHE) go -C $(ROOT) vet ./...

fmt:
	@echo "Running go fmt (diff only):"
	@go -C $(ROOT) fmt ./...
	@echo "gofmt diff vs workspace (for reference):"
	@gofmt -l -d $(ROOT) | sed 's/^/\t/' || true

tidy:
	GOCACHE=$(GCACHE) go -C $(ROOT) mod tidy

# 快速啟動 TUI（chat 頁）
run-chat:
	GOCACHE=$(GCACHE) go -C $(ROOT) run . start-tui --page chat

# ===== Ollama quick tests =====
# 可用變數：
# - OLLAMA_HOST  預設 http://localhost:11434
# - OLLAMA_MODEL 需與本機模型一致（可用 make ollama-tags 查）
# - PROMPT       generate 用提示字串，預設 "hello"
# - MESSAGE      chat 用訊息字串，預設 "hello"

OLLAMA_HOST ?= http://localhost:11434
OLLAMA_MODEL ?=
PROMPT ?= hello
MESSAGE ?= hello

ollama-probe:
	@echo "[probe] GET $(OLLAMA_HOST)/api/version"
	@curl -fsS $(OLLAMA_HOST)/api/version || true
	@echo "\n[probe] GET $(OLLAMA_HOST)/api/tags"
	@curl -fsS $(OLLAMA_HOST)/api/tags || true

ollama-tags:
	@echo "[tags] GET $(OLLAMA_HOST)/api/tags"
	@curl -fsS $(OLLAMA_HOST)/api/tags

# 偵測可用的 Ollama Host（優先使用 OLLAMA_HOST，其次嘗試 127.0.0.1 與 localhost）
detect-ollama-host:
	@CANDIDATES="$(OLLAMA_HOST) http://127.0.0.1:11434 http://localhost:11434"; \
	for H in $$CANDIDATES; do \
	  [ -z "$$H" ] && continue; \
	  RESP=$$(curl -fsS "$$H/api/version" 2>/dev/null || true); \
	  if [ -n "$$RESP" ]; then echo "$$H"; exit 0; fi; \
	done; \
	>&2 echo "[detect] 無可用 Ollama 服務，請啟動服務或指定 OLLAMA_HOST"; \
	exit 1

# 僅偵測並回傳第一個模型名稱（stdout），若無則 exit 1
detect-ollama-model:
	@CANDIDATES="$(OLLAMA_HOST) http://127.0.0.1:11434 http://localhost:11434"; \
	for H in $$CANDIDATES; do \
	  [ -z "$$H" ] && continue; \
	  MODEL=$$(curl -fsS "$$H/api/tags" 2>/dev/null | sed 's/\r//g' | grep -o '"name"[[:space:]]*:[[:space:]]*"[^"]*"' | head -n1 | sed 's/.*:\s*"\([^"]*\)"/\1/' || true); \
	  if [ -n "$$MODEL" ]; then echo "$$MODEL"; exit 0; fi; \
	done; \
	>&2 echo "[detect] 無法偵測到模型，請確認 Ollama 是否啟動且已安裝模型"; \
	exit 1

ollama-generate:
	@if [ -z "$(OLLAMA_MODEL)" ]; then \
		echo "[error] OLLAMA_MODEL 未設定。先執行: make -C dev ollama-tags 查詢模型，或以 OLLAMA_MODEL=... 指定"; \
		exit 2; \
	fi
	@echo "[generate] POST $(OLLAMA_HOST)/api/generate model=$(OLLAMA_MODEL) prompt=\"$(PROMPT)\""
	@curl -fsS -H "Content-Type: application/json" \
	  -d '{"model":"$(OLLAMA_MODEL)","prompt":"$(PROMPT)","stream":false}' \
	  $(OLLAMA_HOST)/api/generate | sed 's/\r//g'

ollama-chat:
	@if [ -z "$(OLLAMA_MODEL)" ]; then \
		echo "[error] OLLAMA_MODEL 未設定。先執行: make -C dev ollama-tags 查詢模型，或以 OLLAMA_MODEL=... 指定"; \
		exit 2; \
	fi
	@echo "[chat] POST $(OLLAMA_HOST)/api/chat model=$(OLLAMA_MODEL) message=\"$(MESSAGE)\""
	@curl -fsS -H "Content-Type: application/json" \
	  -d '{"model":"$(OLLAMA_MODEL)","messages":[{"role":"user","content":"$(MESSAGE)"}],"stream":false}' \
	  $(OLLAMA_HOST)/api/chat | sed 's/\r//g'

# ===== Beginner demo =====
# 一鍵流程：
# 1) 在 /tmp/ora-notes 新增兩筆包含/不包含 bleve 的筆記
# 2) 執行僅檢索 ask（--no-llm）比對命中
# 3) 若 OLLAMA_MODEL 有設定，則用低資源參數與嚴格模板呼叫 LLM 回覆（選配）

NOTES_DIR ?= /tmp/ora-notes
BEGINNER_TEMPLATE ?= /tmp/ask.simple.yaml

demo-beginner:
	@echo "[demo] notes dir: $(NOTES_DIR)"
	@echo "[demo] 建立簡潔模板: $(BEGINNER_TEMPLATE)"
	@printf '%s\n' \
	  'system: |' \
	  '  你是嚴謹的技術助理。請直接用繁體中文回答，條列 1–3 點重點。不要輸出任何角色名稱或多餘標記。' \
	  'user: |' \
	  '  問題：{{question}}' \
	  '' \
	  '  相關筆記：' \
	  '  {{context}}' \
	  '' \
	  '  要求：' \
	  '  - 僅根據相關筆記回答' \
	  '  - 請避免無關內容，勿臆測' \
	  '  - 若無法明確回答，請逐條複述以上相關筆記中最關鍵的 1–3 句原文。' > $(BEGINNER_TEMPLATE)
	@echo "[demo] 新增兩筆示範筆記"
	@GOCACHE=$(GCACHE) go -C $(ROOT) run . --notes-dir $(NOTES_DIR) add "今天研究 Bleve 查詢語法與使用情境" --tags dev,search
	@GOCACHE=$(GCACHE) go -C $(ROOT) run . --notes-dir $(NOTES_DIR) add "Golang unit test 筆記，未涉及 bleve" --tags dev,test
	@echo "[demo] 僅檢索（不呼叫 LLM）：TopK=3 關鍵字 bleve"
	@GOCACHE=$(GCACHE) go -C $(ROOT) run . --notes-dir $(NOTES_DIR) ask "bleve" --topk 3 --no-llm || true
	@echo "[demo] 可選 LLM：若已設定 OLLAMA_MODEL 則執行低資源摘要（temp=0, top-p=0.1, ctx=1024, predict=128）"
	@MODEL="$(OLLAMA_MODEL)"; \
	if [ -z "$$MODEL" ]; then \
	  echo "[demo] 嘗試自動偵測 OLLAMA 模型 (GET $(OLLAMA_HOST)/api/tags)"; \
	  JSON=$$(curl -fsS $(OLLAMA_HOST)/api/tags 2>/dev/null || true); \
	  PY3=$$(command -v python3 2>/dev/null || true); \
	  [ -z "$$PY3" ] && [ -x /opt/homebrew/bin/python3 ] && PY3=/opt/homebrew/bin/python3; \
	  [ -z "$$PY3" ] && [ -x /usr/local/bin/python3 ] && PY3=/usr/local/bin/python3; \
	  [ -z "$$PY3" ] && [ -x /usr/bin/python3 ] && PY3=/usr/bin/python3; \
	  if [ -n "$$PY3" ]; then \
	    AUTO_MODEL=$$(printf '%s' "$$JSON" | "$$PY3" -c 'import sys,json; s=sys.stdin.read();\ntry:\n    d=json.loads(s)\n    m=d.get("models") or []\n    print(m[0].get("name","")) if m else print("")\nexcept Exception:\n    print("")' 2>/dev/null); \
	  elif command -v jq >/dev/null 2>&1; then \
	    AUTO_MODEL=$$(printf '%s' "$$JSON" | jq -r '.models[0].name // empty'); \
	  else \
	    AUTO_MODEL=$$(printf '%s' "$$JSON" | tr -d '\n' | sed -n 's/.*"models"[[:space:]]*:[[:space:]]*\[[[:space:]]*{[^}]*"name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p'); \
	  fi; \
	  if [ -n "$$AUTO_MODEL" ]; then \
	    echo "[demo] 偵測到模型: $$AUTO_MODEL"; \
	    MODEL="$$AUTO_MODEL"; \
	  fi; \
	fi; \
	if [ -n "$$MODEL" ]; then \
	  echo "[demo] 使用模型: $$MODEL  host: $(OLLAMA_HOST)"; \
	  GOCACHE=$(GCACHE) go -C $(ROOT) run . --notes-dir $(NOTES_DIR) ask "bleve 查詢語法" --topk 3 --model "$$MODEL" --template $(BEGINNER_TEMPLATE) || true; \
	else \
	  echo "[demo] 跳過 LLM 步驟：未設定 OLLAMA_MODEL 且未偵測到模型（可用 make -C dev ollama-tags 查模型）"; \
	fi

# 嚴格版：一定要使用 LLM（若偵測不到模型則直接報錯）
demo-beginner-llm:
	@echo "[demo-llm] notes dir: $(NOTES_DIR)"
	@echo "[demo-llm] 建立簡潔模板: $(BEGINNER_TEMPLATE)"
	@printf '%s\n' \
	  'system: |' \
	  '  你是嚴謹的技術助理。請直接用繁體中文回答，條列 1–3 點重點。不要輸出任何角色名稱或多餘標記。' \
	  'user: |' \
	  '  問題：{{question}}' \
	  '' \
	  '  相關筆記：' \
	  '  {{context}}' \
	  '' \
	  '  要求：' \
	  '  - 僅根據相關筆記回答' \
	  '  - 請避免無關內容，勿臆測' > $(BEGINNER_TEMPLATE)
	@echo "[demo-llm] 新增兩筆示範筆記"
	@GOCACHE=$(GCACHE) go -C $(ROOT) run . --notes-dir $(NOTES_DIR) add "今天研究 Bleve 查詢語法與使用情境" --tags dev,search
	@GOCACHE=$(GCACHE) go -C $(ROOT) run . --notes-dir $(NOTES_DIR) add "Golang unit test 筆記，未涉及 bleve" --tags dev,test
	@echo "[demo-llm] 解析與確認可用模型..."
	@MODEL="$(OLLAMA_MODEL)"; \
	if [ -z "$$MODEL" ]; then \
	  echo "[demo-llm] 嘗試自動偵測 OLLAMA 模型 (GET $(OLLAMA_HOST)/api/tags)"; \
	  JSON=$$(curl -fsS $(OLLAMA_HOST)/api/tags 2>/dev/null || true); \
	  PY3=$$(command -v python3 2>/dev/null || true); \
	  [ -z "$$PY3" ] && [ -x /opt/homebrew/bin/python3 ] && PY3=/opt/homebrew/bin/python3; \
	  [ -z "$$PY3" ] && [ -x /usr/local/bin/python3 ] && PY3=/usr/local/bin/python3; \
	  [ -z "$$PY3" ] && [ -x /usr/bin/python3 ] && PY3=/usr/bin/python3; \
	  if [ -n "$$PY3" ]; then \
	    AUTO_MODEL=$$(printf '%s' "$$JSON" | "$$PY3" -c 'import sys,json; s=sys.stdin.read();\ntry:\n    d=json.loads(s)\n    m=d.get("models") or []\n    print(m[0].get("name","")) if m else print("")\nexcept Exception:\n    print("")' 2>/dev/null); \
	  elif command -v jq >/dev/null 2>&1; then \
	    AUTO_MODEL=$$(printf '%s' "$$JSON" | jq -r '.models[0].name // empty'); \
	  else \
	    AUTO_MODEL=$$(printf '%s' "$$JSON" | tr -d '\n' | sed -n 's/.*"models"[[:space:]]*:[[:space:]]*\[[[:space:]]*{[^}]*"name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p'); \
	  fi; \
	  if [ -n "$$AUTO_MODEL" ]; then \
	    echo "[demo-llm] 偵測到模型: $$AUTO_MODEL"; \
	    MODEL="$$AUTO_MODEL"; \
	  fi; \
	fi; \
	if [ -z "$$MODEL" ]; then \
	  echo "[error] 無法取得可用模型。請先：1) 啟動 Ollama 服務；2) 安裝/拉取至少一個模型；或手動以 OLLAMA_MODEL=<model> 指定。可用 make -C dev ollama-tags 檢視。"; \
	  exit 2; \
	fi; \
	echo "[demo-llm] 使用模型: $$MODEL  host: $(OLLAMA_HOST)"; \
	GOCACHE=$(GCACHE) go -C $(ROOT) run . --notes-dir $(NOTES_DIR) ask "bleve 查詢語法" --topk 3 --model "$$MODEL" --template $(BEGINNER_TEMPLATE)

# 自動偵測模型後執行 demo-beginner-llm（若偵測失敗則報錯）
demo-beginner-llm-auto:
	@HOST=$$($(MAKE) -s detect-ollama-host) || exit $$?; \
	MODEL=$$($(MAKE) -s OLLAMA_HOST="$$HOST" detect-ollama-model) || exit $$?; \
	echo "[demo-llm-auto] 使用 host: $$HOST  model: $$MODEL"; \
	OLLAMA_HOST="$$HOST" OLLAMA_MODEL="$$MODEL" $(MAKE) demo-beginner-llm

# ===== Demo cases (guaranteed hit) =====
# 產生一組涵蓋不同主題的示範筆記，確保檢索能命中。
demo-cases-gen:
	@echo "[cases] notes dir: $(NOTES_DIR)"
	@GOCACHE=$(GCACHE) go -C $(ROOT) run . --notes-dir $(NOTES_DIR) add "今天研究 Bleve 查詢語法與使用情境：MatchQuery、TermQuery 與標籤 AND 檢索。" --tags dev,search
	@GOCACHE=$(GCACHE) go -C $(ROOT) run . --notes-dir $(NOTES_DIR) add "Golang unit test 筆記：table-driven 測試、fake http client 注入。" --tags dev,test
	@GOCACHE=$(GCACHE) go -C $(ROOT) run . --notes-dir $(NOTES_DIR) add "TUI 設定面板（F2）：切換模型、調整溫度與 top-p、保存設定。" --tags tui,doc
	@GOCACHE=$(GCACHE) go -C $(ROOT) run . --notes-dir $(NOTES_DIR) add "索引與檢索流程：先重建索引，再以 Bleve 查詢並產出 excerpt。" --tags search,arch
	@echo "[cases] done."

# 單題執行：需 QUESTION，其他參數可選（TAGS, TOPK, TEMPLATE）
ask-run-case:
	@HOST=$$($(MAKE) -s detect-ollama-host) || exit $$?; \
	MODEL=$$($(MAKE) -s OLLAMA_HOST="$$HOST" detect-ollama-model) || exit $$?; \
	Q="$(QUESTION)"; [ -n "$$Q" ] || { echo "[error] QUESTION 未設定"; exit 2; }; \
	TAGS_OPT=""; [ -z "$(TAGS)" ] || TAGS_OPT="--tags $(TAGS)"; \
	TOPK_OPT=""; [ -z "$(TOPK)" ] || TOPK_OPT="--topk $(TOPK)"; \
	TPL_OPT="--template $(BEGINNER_TEMPLATE)"; [ -z "$(TEMPLATE)" ] || TPL_OPT="--template $(TEMPLATE)"; \
	GOCACHE=$(GCACHE) OLLAMA_HOST="$$HOST" OLLAMA_MODEL="$$MODEL" \
	  go -C $(ROOT) run . --notes-dir $(NOTES_DIR) ask "$$Q" $$TAGS_OPT $$TOPK_OPT $$TPL_OPT --model "$$MODEL" || true

# 批次執行預設情境列表（會自動偵測 host+model；若偵測不到則報錯）
ask-run-cases:
	@HOST=$$($(MAKE) -s detect-ollama-host) || exit $$?; \
	MODEL=$$($(MAKE) -s OLLAMA_HOST="$$HOST" detect-ollama-model) || exit $$?; \
	OUT=out/ask-$$(date +%Y%m%d-%H%M%S).log; mkdir -p out; \
	echo "[cases] 使用 host: $$HOST  model: $$MODEL  log: $$OUT"; \
	while IFS='|' read -r QUESTION TAGS TOPK; do \
	  [ -z "$$QUESTION" ] && continue; \
	  echo "\n[cases] Q: $$QUESTION | tags=$$TAGS | topk=$$TOPK" | tee -a "$$OUT"; \
	  TAGS_OPT=""; [ -z "$$TAGS" ] || TAGS_OPT="--tags $$TAGS"; \
	  TOPK_OPT=""; [ -z "$$TOPK" ] || TOPK_OPT="--topk $$TOPK"; \
	  GOCACHE=$(GCACHE) OLLAMA_HOST="$$HOST" OLLAMA_MODEL="$$MODEL" \
	    go -C $(ROOT) run . --notes-dir $(NOTES_DIR) ask "$$QUESTION" $$TAGS_OPT $$TOPK_OPT --template $(BEGINNER_TEMPLATE) --model "$$MODEL" 2>&1 | tee -a "$$OUT"; \
	done < $(CURDIR)/ask-cases.txt; \
	echo "[cases] 完成，輸出於 $$OUT"
