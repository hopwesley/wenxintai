package service

import (
	"context"
)

// StartQuestionGeneration 暂时只是返回一个占位 assessmentID，后面再接真正的 AI 逻辑。
func (s *Svc) StartQuestionGeneration(_ context.Context, sessionID, mode, grade, hobby string) error {
	if sessionID == "" {
		return &Error{
			Code:    ErrorCodeBadRequest,
			Message: "session_id is required",
		}
	}
	if mode == "" {
		return &Error{
			Code:    ErrorCodeBadRequest,
			Message: "mode is required",
		}
	}

	return nil
}
