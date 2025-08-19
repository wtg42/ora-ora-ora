package config

import (
	"os"
	"path/filepath"
	"testing"
)

// 目標：以 TDD 驅動 Config 載入邏輯
// 約定：
// - 若 path 為空或檔案不存在：回傳預設值且不視為錯誤。
// - 若 YAML 非法：回傳錯誤。
// - 若 YAML 合法：以檔案值覆寫預設值（淺層覆蓋即可通過測試）。
func TestLoad_ConfigDefaultsAndOverlay(t *testing.T) {
	type want struct {
		host     string
		model    string
		notesDir string
		indexDir string
		width    int
	}

	defaultWant := want{
		host:     "http://localhost:11434",
		model:    "llama3",
		notesDir: "data/notes",
		indexDir: "data/index",
		width:    80,
	}

	cases := []struct {
		name     string
		yamlBody string // 空字串表示不寫入檔案
		useFile  bool   // 是否建立檔案
		want     want
		wantErr  bool
	}{
		{
			name:    "empty path returns defaults",
			useFile: false,
			want:    defaultWant,
		},
		{
			name:     "invalid yaml returns error",
			yamlBody: ":: not yaml ::",
			useFile:  true,
			wantErr:  true,
		},
		{
			name: "valid yaml overlays defaults",
			yamlBody: "" +
				"ollamaHost: http://127.0.0.1:11434\n" +
				"model: qwen2.5\n" +
				"data:\n  notesDir: d/notes\n  indexDir: d/index\n" +
				"tui:\n  width: 100\n",
			useFile: true,
			want: want{
				host:     "http://127.0.0.1:11434",
				model:    "qwen2.5",
				notesDir: "d/notes",
				indexDir: "d/index",
				width:    100,
			},
		},
		{
			name:    "non-existent file returns defaults",
			useFile: true, // 會給一個不存在的 path
			// 不寫檔案，直接給不存在路徑
			want: defaultWant,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var path string
			if tc.useFile {
				dir := t.TempDir()
				path = filepath.Join(dir, "config.yaml")
				if tc.yamlBody != "" {
					if err := os.WriteFile(path, []byte(tc.yamlBody), 0o644); err != nil {
						t.Fatalf("write yaml: %v", err)
					}
				} else {
					// 故意不寫檔，模擬不存在檔案
					path = filepath.Join(dir, "missing.yaml")
				}
			}

			cfg, err := Load(path)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if cfg.OllamaHost != tc.want.host {
				t.Fatalf("host: want %q, got %q", tc.want.host, cfg.OllamaHost)
			}
			if cfg.Model != tc.want.model {
				t.Fatalf("model: want %q, got %q", tc.want.model, cfg.Model)
			}
			if cfg.Data.NotesDir != tc.want.notesDir {
				t.Fatalf("notesDir: want %q, got %q", tc.want.notesDir, cfg.Data.NotesDir)
			}
			if cfg.Data.IndexDir != tc.want.indexDir {
				t.Fatalf("indexDir: want %q, got %q", tc.want.indexDir, cfg.Data.IndexDir)
			}
			if cfg.TUI.Width != tc.want.width {
				t.Fatalf("width: want %d, got %d", tc.want.width, cfg.TUI.Width)
			}
		})
	}
}
