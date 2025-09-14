package prompt

import (
	"os"
	"testing"
)

func TestLoadAskTemplate_Default(t *testing.T) {
	tpl, warn := LoadAskTemplate("")
	if warn != "" {
		t.Fatalf("unexpected warn: %s", warn)
	}
	if tpl.System == "" || tpl.User == "" {
		t.Fatalf("expected non-empty default template")
	}
}

func TestLoadAskTemplate_FileMissing(t *testing.T) {
	tpl, warn := LoadAskTemplate("/no/such/file.yaml")
	if warn == "" {
		t.Fatalf("expected warning for missing file")
	}
	if tpl.System == "" || tpl.User == "" {
		t.Fatalf("expected default template when file missing")
	}
}

func TestLoadAskTemplate_InvalidYAML(t *testing.T) {
	tmpFile := t.TempDir() + "/invalid.yaml"
	if err := os.WriteFile(tmpFile, []byte("invalid: yaml: content:"), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	tpl, warn := LoadAskTemplate(tmpFile)
	if warn == "" {
		t.Fatalf("expected warning for invalid YAML")
	}
	if tpl.System == "" || tpl.User == "" {
		t.Fatalf("expected default template when YAML invalid")
	}
}

func TestLoadAskTemplate_MissingKeys(t *testing.T) {
	tmpFile := t.TempDir() + "/missing.yaml"
	if err := os.WriteFile(tmpFile, []byte("other: key"), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	tpl, warn := LoadAskTemplate(tmpFile)
	if warn == "" {
		t.Fatalf("expected warning for missing keys")
	}
	if tpl.System == "" || tpl.User == "" {
		t.Fatalf("expected default template when keys missing")
	}
}
