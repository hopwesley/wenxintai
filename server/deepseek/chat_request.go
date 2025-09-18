package deepseek

type ChatRequest struct {
	Model            string    `json:"model"` // 使用 "deepseek-chat"
	Messages         []Message `json:"messages"`
	Temperature      float64   `json:"temperature"`       // 建议: 0.7-0.8 (创造性适中)
	MaxTokens        int       `json:"max_tokens"`        // 建议: 1024-2048
	TopP             float64   `json:"top_p"`             // 建议: 0.9-0.95
	FrequencyPenalty float64   `json:"frequency_penalty"` // 建议: 0.1-0.5 (减少重复)
	PresencePenalty  float64   `json:"presence_penalty"`  // 建议: 0.1-0.5 (鼓励新话题)
	Stop             []string  `json:"stop,omitempty"`    // 可设置停止词
}
