package core

import "testing"

func TestSanitizeLLMOutput_BasicPrefixes(t *testing.T) {
    cases := map[string]string{
        "**.assistant\n- 重點1\n- 重點2": "- 重點1\n- 重點2",
        "**assistant 重點" : "重點",
        "assistant: 回覆內容": "回覆內容",
        "Assistant: 回覆內容": "回覆內容",
        "<|assistant|> 回覆": "回覆",
        "assistant 回覆": "回覆",
        "正常內容": "正常內容",
    }
    for in, want := range cases {
        got := sanitizeLLMOutput(in)
        if got != want {
            t.Fatalf("sanitizeLLMOutput(%q)=%q, want %q", in, got, want)
        }
    }
}

