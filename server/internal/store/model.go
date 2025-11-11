package store

import (
	"encoding/json"
	"time"
)

type Assessment struct {
	ID           string
	InviteCode   *string
	WechatOpenID *string
	Mode         string
	Status       int16
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type QuestionSet struct {
	ID            string
	AssessmentID  string
	Stage         int16
	QuestionsJSON json.RawMessage
	AIPrompt      *string
	AIRawResponse json.RawMessage
	Status        int16
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Answer struct {
	ID            string
	QuestionSetID string
	AnswerJSON    json.RawMessage
	SubmittedAt   time.Time
}

type ComputedParams struct {
	ID           string
	AssessmentID string
	Stage        int16
	ParamsJSON   json.RawMessage
	CreatedAt    time.Time
}

type Report struct {
	ID           string
	AssessmentID string
	ReportType   int16
	Summary      *string
	FullJSON     json.RawMessage
	CreatedAt    time.Time
}

type Invite struct {
	Code      string
	Status    int16      // 0:unused 1:reserved 2:redeemed 3:disabled
	ExpiresAt *time.Time // 预留过期时间 或 自然过期时间
	UsedBy    *string    // 预留者 session_id 或 核销者 id
	UsedAt    *time.Time // 核销时间
	CreatedAt time.Time
}
