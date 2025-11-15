package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hopwesley/wenxintai/server/comm"
)

type sqlDB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type SQLRepo struct {
	db *sql.DB
}

func NewSQLRepo(db *sql.DB) *SQLRepo {
	return &SQLRepo{db: db}
}

func (r *SQLRepo) WithTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func (r *SQLRepo) CreateAssessment(ctx context.Context, a *Assessment) error {
	execer := r.getExecer(ctx)
	if a.ID == "" {
		a.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	query := `INSERT INTO app.assessments (id, invite_code, wechat_openid, mode, status, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING created_at, updated_at`
	row := execer.QueryRowContext(ctx, query, a.ID, a.InviteCode, a.WechatOpenID, a.Mode, a.Status, now, now)
	if err := row.Scan(&a.CreatedAt, &a.UpdatedAt); err != nil {
		return fmt.Errorf("insert assessment: %w", err)
	}
	return nil
}

func (r *SQLRepo) UpdateAssessmentStatus(ctx context.Context, id string, status int16) error {
	execer := r.getExecer(ctx)
	res, err := execer.ExecContext(ctx, `UPDATE app.assessments SET status=$2, updated_at=NOW() WHERE id=$1`, id, status)
	if err != nil {
		return fmt.Errorf("update assessment status: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if affected == 0 {
		return comm.ErrNotFound
	}
	return nil
}

func (r *SQLRepo) GetAssessmentByID(ctx context.Context, id string) (*Assessment, error) {
	query := `SELECT id, invite_code, wechat_openid, mode, status, created_at, updated_at FROM app.assessments WHERE id=$1`
	if _, ok := TxFromContext(ctx); ok {
		query += " FOR UPDATE"
	}
	row := r.getExecer(ctx).QueryRowContext(ctx, query, id)
	var a Assessment
	if err := row.Scan(&a.ID, &a.InviteCode, &a.WechatOpenID, &a.Mode, &a.Status, &a.CreatedAt, &a.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, comm.ErrNotFound
		}
		return nil, fmt.Errorf("get assessment: %w", err)
	}
	return &a, nil
}

func (r *SQLRepo) CreateQuestionSet(ctx context.Context, qs *QuestionSet) error {
	execer := r.getExecer(ctx)
	if qs.ID == "" {
		qs.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	query := `INSERT INTO app.question_sets (id, assessment_id, stage, questions_json, ai_prompt, ai_raw_response, status, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$8) ON CONFLICT (assessment_id, stage) DO NOTHING RETURNING id, assessment_id, stage, questions_json, ai_prompt, ai_raw_response, status, created_at, updated_at`
	row := execer.QueryRowContext(ctx, query, qs.ID, qs.AssessmentID, qs.Stage, qs.QuestionsJSON, qs.AIPrompt, qs.AIRawResponse, qs.Status, now)
	if err := row.Scan(&qs.ID, &qs.AssessmentID, &qs.Stage, &qs.QuestionsJSON, &qs.AIPrompt, &qs.AIRawResponse, &qs.Status, &qs.CreatedAt, &qs.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			existing, err := r.GetQuestionSetByAssessmentStage(ctx, qs.AssessmentID, qs.Stage)
			if err != nil {
				return err
			}
			*qs = *existing
			return nil
		}
		return fmt.Errorf("insert question set: %w", err)
	}
	return nil
}

func (r *SQLRepo) UpdateQuestionSetStatus(ctx context.Context, id string, status int16) error {
	execer := r.getExecer(ctx)
	res, err := execer.ExecContext(ctx, `UPDATE app.question_sets SET status=$2, updated_at=NOW() WHERE id=$1`, id, status)
	if err != nil {
		return fmt.Errorf("update question set status: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if affected == 0 {
		return comm.ErrNotFound
	}
	return nil
}

func (r *SQLRepo) GetQuestionSetByID(ctx context.Context, id string) (*QuestionSet, error) {
	query := `SELECT id, assessment_id, stage, questions_json, ai_prompt, ai_raw_response, status, created_at, updated_at FROM app.question_sets WHERE id=$1`
	if _, ok := TxFromContext(ctx); ok {
		query += " FOR UPDATE"
	}
	row := r.getExecer(ctx).QueryRowContext(ctx, query, id)
	var qs QuestionSet
	if err := row.Scan(&qs.ID, &qs.AssessmentID, &qs.Stage, &qs.QuestionsJSON, &qs.AIPrompt, &qs.AIRawResponse, &qs.Status, &qs.CreatedAt, &qs.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, comm.ErrNotFound
		}
		return nil, fmt.Errorf("get question set: %w", err)
	}
	return &qs, nil
}

func (r *SQLRepo) GetQuestionSetByAssessmentStage(ctx context.Context, assessmentID string, stage int16) (*QuestionSet, error) {
	query := `SELECT id, assessment_id, stage, questions_json, ai_prompt, ai_raw_response, status, created_at, updated_at FROM app.question_sets WHERE assessment_id=$1 AND stage=$2`
	if _, ok := TxFromContext(ctx); ok {
		query += " FOR UPDATE"
	}
	row := r.getExecer(ctx).QueryRowContext(ctx, query, assessmentID, stage)
	var qs QuestionSet
	if err := row.Scan(&qs.ID, &qs.AssessmentID, &qs.Stage, &qs.QuestionsJSON, &qs.AIPrompt, &qs.AIRawResponse, &qs.Status, &qs.CreatedAt, &qs.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, comm.ErrNotFound
		}
		return nil, fmt.Errorf("get question set by stage: %w", err)
	}
	return &qs, nil
}

func (r *SQLRepo) UpsertAnswer(ctx context.Context, ans *Answer) (bool, error) {
	execer := r.getExecer(ctx)
	if ans.ID == "" {
		ans.ID = uuid.NewString()
	}
	if ans.SubmittedAt.IsZero() {
		ans.SubmittedAt = time.Now().UTC()
	}
	query := `INSERT INTO app.answers (id, question_set_id, answer_json, submitted_at)
VALUES ($1,$2,$3,$4) ON CONFLICT (question_set_id) DO NOTHING RETURNING id, submitted_at`
	row := execer.QueryRowContext(ctx, query, ans.ID, ans.QuestionSetID, ans.AnswerJSON, ans.SubmittedAt)
	var insertedID string
	if err := row.Scan(&insertedID, &ans.SubmittedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			existing, err := r.getExistingAnswer(ctx, ans.QuestionSetID)
			if err != nil {
				return false, err
			}
			ans.ID = existing.ID
			ans.SubmittedAt = existing.SubmittedAt
			ans.AnswerJSON = existing.AnswerJSON
			return false, nil
		}
		return false, fmt.Errorf("insert answer: %w", err)
	}
	ans.ID = insertedID
	return true, nil
}

func (r *SQLRepo) getExistingAnswer(ctx context.Context, questionSetID string) (*Answer, error) {
	row := r.getExecer(ctx).QueryRowContext(ctx, `SELECT id, question_set_id, answer_json, submitted_at FROM app.answers WHERE question_set_id=$1`, questionSetID)
	var ans Answer
	if err := row.Scan(&ans.ID, &ans.QuestionSetID, &ans.AnswerJSON, &ans.SubmittedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, comm.ErrNotFound
		}
		return nil, err
	}
	return &ans, nil
}

func (r *SQLRepo) CreateComputedParams(ctx context.Context, cp *ComputedParams) error {
	execer := r.getExecer(ctx)
	if cp.ID == "" {
		cp.ID = uuid.NewString()
	}
	if cp.CreatedAt.IsZero() {
		cp.CreatedAt = time.Now().UTC()
	}
	_, err := execer.ExecContext(ctx, `INSERT INTO app.computed_params (id, assessment_id, stage, params_json, created_at)
VALUES ($1,$2,$3,$4,$5)`, cp.ID, cp.AssessmentID, cp.Stage, cp.ParamsJSON, cp.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert computed params: %w", err)
	}
	return nil
}

func (r *SQLRepo) CreateReport(ctx context.Context, rp *Report) error {
	execer := r.getExecer(ctx)
	if rp.ID == "" {
		rp.ID = uuid.NewString()
	}
	if rp.CreatedAt.IsZero() {
		rp.CreatedAt = time.Now().UTC()
	}
	row := execer.QueryRowContext(ctx, `INSERT INTO app.reports (id, assessment_id, report_type, summary, full_json, created_at)
VALUES ($1,$2,$3,$4,$5,$6) RETURNING created_at`, rp.ID, rp.AssessmentID, rp.ReportType, rp.Summary, rp.FullJSON, rp.CreatedAt)
	if err := row.Scan(&rp.CreatedAt); err != nil {
		return fmt.Errorf("insert report: %w", err)
	}
	return nil
}

func (r *SQLRepo) GetLatestReportByAssessment(ctx context.Context, assessmentID string) (*Report, error) {
	row := r.getExecer(ctx).QueryRowContext(ctx, `SELECT id, assessment_id, report_type, summary, full_json, created_at FROM app.reports WHERE assessment_id=$1 ORDER BY created_at DESC LIMIT 1`, assessmentID)
	var rp Report
	if err := row.Scan(&rp.ID, &rp.AssessmentID, &rp.ReportType, &rp.Summary, &rp.FullJSON, &rp.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, comm.ErrNotFound
		}
		return nil, fmt.Errorf("get latest report: %w", err)
	}
	return &rp, nil
}

func (r *SQLRepo) GetAnswersByAssessment(ctx context.Context, assessmentID string) (json.RawMessage, json.RawMessage, error) {
	rows, err := r.getExecer(ctx).QueryContext(ctx, `SELECT qs.stage, ans.answer_json FROM app.question_sets qs
LEFT JOIN app.answers ans ON ans.question_set_id = qs.id
WHERE qs.assessment_id=$1 AND qs.stage IN (1,2)`, assessmentID)
	if err != nil {
		return nil, nil, fmt.Errorf("get answers: %w", err)
	}
	defer rows.Close()
	var s1, s2 json.RawMessage
	for rows.Next() {
		var stage int16
		var ans json.RawMessage
		if err := rows.Scan(&stage, &ans); err != nil {
			return nil, nil, fmt.Errorf("scan answers: %w", err)
		}
		if stage == 1 {
			s1 = ans
		} else if stage == 2 {
			s2 = ans
		}
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("iterate answers: %w", err)
	}
	return s1, s2, nil
}

func (r *SQLRepo) getExecer(ctx context.Context) sqlDB {
	if tx, ok := TxFromContext(ctx); ok && tx != nil {
		return tx
	}
	return r.db
}
