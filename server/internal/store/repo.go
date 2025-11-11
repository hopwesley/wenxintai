package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

type Repo interface {
	WithTx(ctx context.Context, fn func(tx *sql.Tx) error) error

	CreateAssessment(ctx context.Context, a *Assessment) error
	UpdateAssessmentStatus(ctx context.Context, id string, status int16) error
	GetAssessmentByID(ctx context.Context, id string) (*Assessment, error)
	RedeemInviteByCode(ctx context.Context, code, redeemedBy string) (bool, error)
	CreateQuestionSet(ctx context.Context, qs *QuestionSet) error
	UpdateQuestionSetStatus(ctx context.Context, id string, status int16) error
	GetQuestionSetByID(ctx context.Context, id string) (*QuestionSet, error)
	GetQuestionSetByAssessmentStage(ctx context.Context, assessmentID string, stage int16) (*QuestionSet, error)
	UpsertAnswer(ctx context.Context, ans *Answer) (created bool, err error)
	CreateComputedParams(ctx context.Context, cp *ComputedParams) error
	CreateReport(ctx context.Context, r *Report) error
	GetLatestReportByAssessment(ctx context.Context, assessmentID string) (*Report, error)
	GetAnswersByAssessment(ctx context.Context, assessmentID string) (s1, s2 json.RawMessage, err error)
	GetInviteForUpdate(ctx context.Context, code string) (*Invite, error)
	UpdateInviteReservation(ctx context.Context, code string, sessionID string, until time.Time) error
	RedeemInviteBySession(ctx context.Context, sessionID, redeemedBy string) (bool, error)
}

type txKey struct{}

func ContextWithTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func TxFromContext(ctx context.Context) (*sql.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(*sql.Tx)
	return tx, ok
}
