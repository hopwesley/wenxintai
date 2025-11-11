package ai

import (
        "context"
        "encoding/json"
        "fmt"
        "time"
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
                "prompt":   prompt,
                "questions": prompts,
        }
        rawBytes, err := json.Marshal(raw)
        if err != nil {
                return nil, "", nil, err
        }
        return payload, prompt, rawBytes, nil
}

func InterpretReport(ctx context.Context, params json.RawMessage) (json.RawMessage, *string, error) {
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
        return full, &summary, nil
}
