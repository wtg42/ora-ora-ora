package cmd

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
)

func TestDiag_Basic(t *testing.T) {
    tmp := t.TempDir()
    notesDir := filepath.Join(tmp, "notes")
	if err := os.MkdirAll(notesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	tpl := filepath.Join(tmp, "ask.yaml")
	if err := os.WriteFile(tpl, []byte("system: s\nuser: u"), 0o644); err != nil {
		t.Fatal(err)
	}

    out, err := runCmd(t, "--notes-dir", notesDir, "diag", "--template", tpl)
	if err != nil {
		t.Fatalf("diag returned error: %v (%s)", err, out)
	}
    // 在受限環境下，Ollama 預設不可連線，接受 unreachable 訊息
	if !strings.Contains(out, "NotesDir: writable") {
		t.Fatalf("expected notes writable, got: %s", out)
	}
	if !strings.Contains(out, "Template: ok") {
		t.Fatalf("expected template ok, got: %s", out)
	}
}
