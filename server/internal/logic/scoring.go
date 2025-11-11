package logic

import (
        "encoding/json"
        "fmt"
)

func ComputeParams(answersS1, answersS2 json.RawMessage) (json.RawMessage, error) {
        type answer struct {
                QuestionID string      `json:"question_id"`
                Value      interface{} `json:"value"`
        }

        var stage1, stage2 []answer
        if len(answersS1) > 0 {
                if err := json.Unmarshal(answersS1, &stage1); err != nil {
                        return nil, fmt.Errorf("parse stage1 answers: %w", err)
                }
        }
        if len(answersS2) > 0 {
                if err := json.Unmarshal(answersS2, &stage2); err != nil {
                        return nil, fmt.Errorf("parse stage2 answers: %w", err)
                }
        }

        summary := map[string]any{
                "stage1_answer_count": len(stage1),
                "stage2_answer_count": len(stage2),
                "total_answer_count": len(stage1) + len(stage2),
        }

        combined := append([]answer{}, stage1...)
        combined = append(combined, stage2...)
        summary["combined_answers"] = combined

        payload, err := json.Marshal(summary)
        if err != nil {
                return nil, fmt.Errorf("marshal params: %w", err)
        }
        return payload, nil
}
