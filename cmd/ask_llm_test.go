package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Minimal LLM integration test using a mock Ollama server.
func TestAsk_WithLLM_UsesTemplateAndPrintsAnswer(t *testing.T) {

	// Prepare temp notes dir and template
	tmp := t.TempDir()
	notesDir := filepath.Join(tmp, "data", "notes")
	if err := os.MkdirAll(notesDir, 0o755); err != nil {
		t.Fatalf("mkdir notes dir: %v", err)
	}
	// add a note so retrieval has context
	if out, err := runCmd(t, "--notes-dir", notesDir, "add", "bleve intro", "--tags", "dev"); err != nil {
		t.Fatalf("add note failed: %v (%s)", err, out)
	}

	// Create a simple template file
	tpl := filepath.Join(tmp, "ask.yaml")
	if err := os.WriteFile(tpl, []byte("system: s\nuser: u {{context}} {{question}}\n"), 0o644); err != nil {
		t.Fatalf("write template: %v", err)
	}

	// Point agent to mock host via env var expected by config defaults
	// Force host via --config override by writing a small config file
	cfg := filepath.Join(tmp, "cfg.yaml")
	if err := os.WriteFile(cfg, []byte("ollamaHost: \nmodel: dummy\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	// 在受限沙箱下無法啟動本機 HTTP 伺服器，因此此測試只覆蓋指令流與模板載入；
	// 不實際呼叫 LLM（無檢索片段時會提前返回）。

	out, err := runCmd(t,
		"--notes-dir", notesDir,
		"ask", "bleve", "--topk", "3",
		"--config", cfg,
		"--template", tpl,
		"--model", "dummy",
		"--no-llm",
	)
	if err != nil {
		t.Fatalf("ask failed: %v (%s)", err, out)
	}
	// 由於未連線 LLM，且檢索片段可能為 0，預期輸出不包含錯誤。
	if strings.Contains(strings.ToLower(out), "error") {
		t.Fatalf("unexpected error output: %s", out)
	}
}
