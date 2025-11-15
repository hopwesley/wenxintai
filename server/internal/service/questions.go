package service

import (
	"context"

	"github.com/hopwesley/wenxintai/server/comm"
)

// StartQuestionGeneration 暂时只是返回一个占位 assessmentID，后面再接真正的 AI 逻辑。
func (s *Svc) StartQuestionGeneration(_ context.Context, sessionID, mode, grade, hobby string) error {
	if sessionID == "" {
		return &comm.ApiErr{
			Code:    comm.ErrorCodeBadRequest,
			Message: "session_id is required",
		}
	}
	if mode == "" {
		return &comm.ApiErr{
			Code:    comm.ErrorCodeBadRequest,
			Message: "mode is required",
		}
	}

	return nil
}
