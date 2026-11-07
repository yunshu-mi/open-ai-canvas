package app

import "testing"

func TestDecodeLegacyThinkingCapability(t *testing.T) {
	for _, test := range []struct {
		name string
		raw  string
		mode string
	}{
		{"enabled", `true`, ""},
		{"disabled", `false`, ThinkingModeNone},
		{"null", `null`, ""},
		{"new configuration", `{"mode":"none"}`, ThinkingModeNone},
	} {
		t.Run(test.name, func(t *testing.T) {
			config, err := DecodeModelCapabilityConfig(`{"version":1,"text":{"thinking":` + test.raw + `,"streaming":false,"maxOutputTokens":8192,"references":{"promptMaxChars":20000}}}`)
			if err != nil {
				t.Fatal(err)
			}
			normalized, err := NormalizeModelCapabilityConfigForModel("text", "openrouter-chat", "test-model", config)
			if err != nil {
				t.Fatal(err)
			}
			text := normalized.Text
			if test.mode == "" {
				if text.Thinking != nil {
					t.Fatalf("旧开启配置应沿用协议默认值，实际为 %#v", text.Thinking)
				}
			} else if text.Thinking == nil || text.Thinking.Mode != test.mode {
				t.Fatalf("思考模式 = %#v，期望 %s", text.Thinking, test.mode)
			}
			if text.Streaming == nil || *text.Streaming || text.MaxOutputTokens != 8192 || text.References.PromptMaxChars != 20000 {
				t.Fatalf("其他模型能力字段变化：%#v", text)
			}
		})
	}
	if _, err := DecodeModelCapabilityConfig(`{"text":{"thinking":"invalid"}}`); err == nil {
		t.Fatal("非法思考配置必须返回解析错误")
	}
}
