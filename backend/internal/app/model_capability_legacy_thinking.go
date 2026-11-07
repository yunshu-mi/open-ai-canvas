package app

import (
	"encoding/json"
	"strings"
)

// UnmarshalJSON 将本仓库旧版的推理能力开关转成上游思考等级结构。
// true 沿用上游未配置时的协议默认行为，false 保留管理员禁止思考的选择。
func (config *TextCapabilityConfig) UnmarshalJSON(data []byte) error {
	type textCapabilityConfig TextCapabilityConfig
	value := struct {
		*textCapabilityConfig
		Thinking json.RawMessage `json:"thinking"`
	}{textCapabilityConfig: (*textCapabilityConfig)(config)}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	switch strings.TrimSpace(string(value.Thinking)) {
	case "", "null", "true":
		config.Thinking = nil
	case "false":
		config.Thinking = &TextThinkingConfig{Mode: ThinkingModeNone}
	default:
		return json.Unmarshal(value.Thinking, &config.Thinking)
	}
	return nil
}
