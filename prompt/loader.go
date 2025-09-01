package prompt

import (
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

// AskTemplate 表示問答模板的最小結構。
type AskTemplate struct {
	System string `yaml:"system"`
	User   string `yaml:"user"`
}

// LoadAskTemplate 讀取 YAML 模板；若 path 為空或讀取失敗，回傳內建預設與警示訊息。
func LoadAskTemplate(path string) (AskTemplate, string) {
	if path != "" {
		b, err := os.ReadFile(filepath.Clean(path))
		if err == nil {
			var tpl AskTemplate
			if err := yaml.Unmarshal(b, &tpl); err == nil && (tpl.System != "" || tpl.User != "") {
				return tpl, ""
			}
			return defaultAskTemplate(), "template invalid or missing keys; using default"
		}
		return defaultAskTemplate(), "template not found; using default"
	}
	return defaultAskTemplate(), ""
}

func defaultAskTemplate() AskTemplate {
	return AskTemplate{
		System: "你是助理，請用繁體中文。",
		User:   "根據以下內容回答：\n{{context}}\n問題：{{question}}",
	}
}
