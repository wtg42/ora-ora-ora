package config

import "testing"

func TestLoad_Default(t *testing.T) {
    c, err := Load("")
    if err != nil { t.Fatalf("load: %v", err) }
    if c.OllamaHost == "" || c.Data.NotesDir == "" || c.Data.IndexDir == "" {
        t.Fatalf("default fields should be filled: %+v", c)
    }
}

