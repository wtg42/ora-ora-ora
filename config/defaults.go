package config

// DefaultConfig 回傳專案的預設設定。
// 依 README/AGENTS 的 MVP 合約與現有測試期望：
//   - OllamaHost: http://127.0.0.1:11434
//   - Model:      llama3
//   - Data.NotesDir: data/notes
//   - Data.IndexDir: data/index
//   - TUI.Width:  80
// 注意：回傳的是新值，呼叫端可在其上覆寫使用者設定以避免共享可變狀態。
func DefaultConfig() Config {
    return Config{
        OllamaHost: "http://127.0.0.1:11434",
        Model:      "llama3",
        Data: Data{
            NotesDir: "data/notes",
            IndexDir: "data/index",
        },
        TUI: TUI{
            Width: 80,
        },
    }
}

