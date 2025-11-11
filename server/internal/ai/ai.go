package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/hopwesley/wenxintai/server/assessment"
)

type question struct {
	ID      string `json:"id"`
	Prompt  string `json:"prompt"`
	Stage   int16  `json:"stage"`
	Created string `json:"created_at"`
}

func GenerateQuestions(ctx context.Context, stage int16, mode string, userCtx map[string]string) (json.RawMessage, string, json.RawMessage, error) {
	prompts := []question{
		{
			ID:      fmt.Sprintf("%s-%d-1", mode, stage),
			Prompt:  fmt.Sprintf("请在阶段 %d (%s) 中分享你最近一次的学习经历。", stage, mode),
			Stage:   stage,
			Created: time.Now().UTC().Format(time.RFC3339),
		},
		{
			ID:      fmt.Sprintf("%s-%d-2", mode, stage),
			Prompt:  fmt.Sprintf("阶段 %d: 描述一个让你印象深刻的问题，并说明原因。", stage),
			Stage:   stage,
			Created: time.Now().UTC().Format(time.RFC3339),
		},
	}

	payload, err := json.Marshal(prompts)
	if err != nil {
		return nil, "", nil, err
	}
	prompt := fmt.Sprintf("mode=%s stage=%d user=%v", mode, stage, userCtx)
	raw := map[string]any{
		"prompt":    prompt,
		"questions": prompts,
	}
	rawBytes, err := json.Marshal(raw)
	if err != nil {
		return nil, "", nil, err
	}
	return payload, prompt, rawBytes, nil
}

type StreamCallback func(string) error

func InterpretReport(ctx context.Context, params json.RawMessage, onToken StreamCallback) (json.RawMessage, *string, error) {
	if onToken == nil {
		onToken = func(string) error { return nil }
	}

	apiKey := strings.TrimSpace(os.Getenv("DS_API_KEY"))
	apiBase := strings.TrimSpace(os.Getenv("DS_API_BASE"))
	model := strings.TrimSpace(os.Getenv("DS_MODEL"))
	if apiBase == "" {
		apiBase = "https://api.deepseek.com"
	}
	if model == "" {
		model = "deepseek-chat"
	}

	if apiKey == "" {
		return fallbackReport(params, onToken)
	}

	systemPrompt := "你是一名教育咨询顾问，请根据提供的 JSON 数据生成结构化的选科战略报告，输出必须是合法 JSON。"
	userPrompt := fmt.Sprintf("以下是学生的原始数据：\n%s\n请输出包含 common_section、mode_section 与 final_report 的 JSON 对象。", string(params))

	reqBody := map[string]any{
		"model":           model,
		"temperature":     0.4,
		"stream":          true,
		"response_format": map[string]string{"type": "json_object"},
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
	}

	content, err := assessment.StreamChatCompletion(ctx, apiBase, apiKey, reqBody, func(delta string) error {
		return onToken(delta)
	})
	if err != nil {
		return nil, nil, err
	}

	raw := strings.TrimSpace(content)
	if raw == "" {
		return nil, nil, fmt.Errorf("AI 返回空内容")
	}
	if err := validateJSON(raw); err != nil {
		return nil, nil, fmt.Errorf("AI 返回非合法 JSON: %w", err)
	}

	summary := extractSummary(raw)
	return json.RawMessage(raw), &summary, nil
}

func fallbackReport(params json.RawMessage, onToken StreamCallback) (json.RawMessage, *string, error) {
	summary := "系统根据你的回答生成了初步解读，请结合实际情况理解结果。"
	var paramsMap map[string]any
	if err := json.Unmarshal(params, &paramsMap); err != nil {
		paramsMap = map[string]any{"raw_params": json.RawMessage(params)}
	}
	report := map[string]any{
		"summary": summary,
		"params":  paramsMap,
	}
	full, err := json.Marshal(report)
	if err != nil {
		return nil, nil, err
	}
	_ = onToken(summary)
	return full, &summary, nil
}

func extractSummary(raw string) string {
	var payload map[string]any
	if err := json.Unmarshal([]byte(raw), &payload); err == nil {
		if finalSection, ok := payload["final_report"].(map[string]any); ok {
			if conclusion, ok := finalSection["strategic_conclusion"].(string); ok && strings.TrimSpace(conclusion) != "" {
				return conclusion
			}
		}
	}
	trimmed := strings.ReplaceAll(raw, "\n", " ")
	trimmed = strings.ReplaceAll(trimmed, "\t", " ")
	trimmed = strings.Join(strings.Fields(trimmed), " ")
	if utf8.RuneCountInString(trimmed) > 160 {
		r := []rune(trimmed)
		trimmed = string(r[:160]) + "…"
	}
	if trimmed == "" {
		trimmed = "报告已生成，请查看完整内容。"
	}
	return trimmed
}

func validateJSON(raw string) error {
	var payload map[string]any
	return json.Unmarshal([]byte(raw), &payload)
}
