package config

// Package config 提供最小設定結構與載入函式。
// 初版不引入外部 YAML 套件，Load 會先回傳預設值；若未來需要可再加入 YAML 解析。

import (
    "os"
)

// Config 代表系統設定。YAML/JSON 鍵建議與欄位對應。
type Config struct {
    OllamaHost string
    Model      string
    Data       struct {
        NotesDir string
        IndexDir string
    }
    TUI struct {
        Width int
    }
}

// Default 產生預設設定。
func Default() Config {
    var c Config
    c.OllamaHost = "http://localhost:11434"
    c.Model = "hf.co/bartowski/Llama-3.2-3B-Instruct-GGUF:Q4_K_M"
    c.Data.NotesDir = "data/notes/"
    c.Data.IndexDir = "data/index/"
    c.TUI.Width = 80
    return c
}

// Load 嘗試載入設定檔。為了避免引入依賴，暫時忽略檔案內容並回傳 Default。
// 未來可改為 YAML/JSON 解析（保留相容性）。
func Load(path string) (Config, error) {
    c := Default()
    if path == "" {
        return c, nil
    }
    // 若檔案存在，目前僅檢查存在性，不讀取內容。
    if _, err := os.Stat(path); err == nil {
        // TODO: 後續加入 YAML/JSON 解析並合併到 c。
        return c, nil
    }
    return c, nil
}

