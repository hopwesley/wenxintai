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
	Code          string
	Status        string
	ReservedBy    *string
	ReservedUntil *time.Time
	RedeemedBy    *string
	RedeemedAt    *time.Time
	CreatedAt     time.Time
}
