package config

import (
	"fmt"
	"log"
	"os"

	"github.com/goccy/go-yaml"
)

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
// 預設值（與 DefaultConfig 一致）：
//
//	OllamaHost   = "http://127.0.0.1:11434"
//	Model        = "llama3"
//	Data.NotesDir = "data/notes"
//	Data.IndexDir = "data/index"
//	TUI.Width    = 80
//
// 注意：
//   - 不讀取 .env；不會進行任何寫檔或索引建立等副作用。
//   - 未知/多餘的 YAML 欄位目前會被忽略（不影響解析已知欄位）。
//
// 範例：
//
//	cfg, err := config.Load("config.yaml")
//	if err != nil { /* 處理錯誤 */ }
func Load(path string) (*Config, error) {
	if len(path) == 0 {
		conf := DefaultConfig()
		return &conf, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Error reading config file: %s", path)
		conf := DefaultConfig()
		return &conf, nil
	}

	// Unmarshal yml file here.
	conf := Config{}
	if err = yaml.Unmarshal(data, &conf); err != nil {
		fmt.Printf("%v", err)
		return nil, err
	}

	fmt.Printf("Config file contents: %v\n", conf)
	// Manual Merge config
	merged_conf := yamlConfigOverride(conf)

	return &merged_conf, nil
}

// Manually merge default YAML config and user's config
func yamlConfigOverride(src Config) Config {
	default_c := DefaultConfig()

	if default_c.OllamaHost != src.OllamaHost {
		default_c.OllamaHost = src.OllamaHost
	}

	if default_c.Data != src.Data {
		default_c.Data = src.Data
	}

	if default_c.Data.NotesDir != src.Data.NotesDir {
		default_c.Data.NotesDir = src.Data.NotesDir
	}

	if default_c.Data.IndexDir != src.Data.IndexDir {
		default_c.Data.IndexDir = src.Data.IndexDir
	}

	if default_c.Model != src.Model {
		default_c.Model = src.Model
	}

	if default_c.TUI.Width != src.TUI.Width {
		default_c.TUI.Width = src.TUI.Width
	}

	return default_c
}
