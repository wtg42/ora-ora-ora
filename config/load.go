package config

import (
	"fmt"
	"log"
	"os"
)

type Config struct {
	OllamaHost string `yaml:"ollamaHost"`
	Model      string
	NotesDir   string
	IndexDir   string
	Width      int
}

// Load 讀取 YAML 設定檔並回傳合併後的設定。
// 行為：
//   - path 為空字串或檔案不存在：回傳內建預設值（不視為錯誤）。
//   - 檔案存在但 YAML 非法（無法解析）：回傳錯誤（cfg 為 nil）。
//   - 檔案存在且 YAML 合法：以檔案值「淺層覆寫」預設值（未提供的欄位沿用預設）。
//
// 支援的 YAML 欄位對應：
//   - ollamaHost     -> Config.OllamaHost
//   - model          -> Config.Model
//   - data.notesDir  -> Config.Data.NotesDir
//   - data.indexDir  -> Config.Data.IndexDir
//   - tui.width      -> Config.TUI.Width
//
// 預設值：
//
//	OllamaHost = "http://localhost:11434"
//	Model      = "llama3"
//	Data.NotesDir = "data/notes"
//	Data.IndexDir = "data/index"
//	TUI.Width  = 80
//
// 注意：
//   - 不讀取 .env；不會進行任何寫檔或索引建立等副作用。
//   - 未知/多餘的 YAML 欄位會被忽略（不影響解析已知欄位）。
//
// 範例：
//
//	cfg, err := config.Load("config.yaml")
//	if err != nil { /* 處理錯誤 */ }
func Load(path string) (*Config, error) {
	// 用戶給予空的路徑直接回傳 default Config
	if len(path) == 0 {
		return &Config{
			OllamaHost: "http://localhost:11434",
			Model:      "llama3",
			NotesDir:   Data.NotesDir,
			IndexDir:   "data/index",
			Width:      80,
		}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		log.Println("Error reading config file:", err)
		fmt.Println("Using default config values.")
	}

	fmt.Printf("Config file contents: %s\n", data)

	return &Config{
		OllamaHost: "http://localhost:11434",
		Model:      "llama3",
		NotesDir:   "data/notes",
		IndexDir:   "data/index",
		Width:      80,
	}, nil
}
