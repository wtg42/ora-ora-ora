package prompt

import "testing"

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
