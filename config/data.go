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

type TUI struct {
	Width int
}
