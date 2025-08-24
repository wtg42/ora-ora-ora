package config

// Data 定義資料目錄設定，用於指定筆記與索引儲存位置。
// 對應 YAML:
//
//	data:
//	  notesDir: ...
//	  indexDir: ...
type Data struct {
	NotesDir string `yaml:"notesDir"`
	IndexDir string `yaml:"indexDir"`
}

// TUI 設定，目前僅包含寬度。
// 對應 YAML:
//
//	tui:
//	  width: 80
type TUI struct {
	Width int `yaml:"width"`
}

// Config 為整體設定結構，符合 README/AGENTS 合約。
// 注意：為避免額外相依，本專案不引入 YAML 套件，
// 由 Load 以極簡解析器處理必要欄位。
type Config struct {
	OllamaHost string `yaml:"ollamaHost"`
	Model      string `yaml:"model"`
	Data       Data   `yaml:"data"`
	TUI        TUI    `yaml:"tui"`
}
