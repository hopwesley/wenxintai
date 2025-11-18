package dbSrv

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

type QuestionRecord interface {
	GetQuestions() json.RawMessage
	GetAnswers() json.RawMessage
}

type RiasecSession struct {
	ID           int64           `json:"id"`
	BusinessType string          `json:"business_type"`
	PublicId     string          `json:"public_id"`
	Questions    json.RawMessage `json:"questions"`
	Answers      json.RawMessage `json:"answers,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	CompletedAt  *time.Time      `json:"completed_at,omitempty"`
}

func (rs *RiasecSession) GetQuestions() json.RawMessage {
	return rs.Questions
}

func (rs *RiasecSession) GetAnswers() json.RawMessage {
	return rs.Answers
}

func (pdb *psDatabase) FindRiasecSession(
	ctx context.Context,
	businessType string,
	publicId string,
) (*RiasecSession, error) {
	if businessType == "" || publicId == "" {
		return nil, errors.New("businessType and publicId must be non-empty")
	}

	pdb.log.Debug().
		Str("business_type", businessType).
		Str("public_id", publicId).
		Msg("FindRiasecSession: start")

	const q = `
SELECT id, business_type, public_id, questions, COALESCE(answers, 'null'::jsonb) AS answers, created_at, completed_at
FROM app.riasec_sessions
WHERE business_type = $1 AND public_id = $2
`

	var sess RiasecSession

	err := pdb.db.QueryRowContext(ctx, q, businessType, publicId).
		Scan(
			&sess.ID,
			&sess.BusinessType,
			&sess.PublicId,
			&sess.Questions,
			&sess.Answers,
			&sess.CreatedAt,
			&sess.CompletedAt,
		)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			pdb.log.Info().
				Str("business_type", businessType).
				Str("public_id", publicId).
				Msg("FindRiasecSession: not found")
			return nil, nil
		}

		pdb.log.Err(err).
			Str("business_type", businessType).
			Str("public_id", publicId).
			Msg("FindRiasecSession failed")
		return nil, err
	}

	pdb.log.Debug().
		Str("business_type", businessType).
		Str("public_id", publicId).
		Msg("FindRiasecSession: found")

	return &sess, nil
}

func (pdb *psDatabase) SaveRiasecSession(
	ctx context.Context,
	publicId, businessType string,
	questionsJSON []byte,
) error {
	if publicId == "" {
		return errors.New("publicId must be non-empty")
	}
	if businessType == "" {
		return errors.New("businessType must be non-empty")
	}
	if len(questionsJSON) == 0 {
		return errors.New("questionsJSON must be non-empty")
	}

	pdb.log.Debug().
		Str("public_id", publicId).
		Str("business_type", businessType).
		Msg("SaveRiasecSession: start")

	const q = `
		INSERT INTO app.riasec_sessions (business_type, public_id, questions)
		VALUES ($1, $2, $3::jsonb)
		ON CONFLICT (business_type, public_id)
		DO UPDATE SET
			questions = EXCLUDED.questions,
			-- 保留最初创建时间，只更新题目内容
			created_at = app.riasec_sessions.created_at
	`

	_, err := pdb.db.ExecContext(ctx, q,
		businessType,
		publicId,
		string(questionsJSON),
	)
	if err != nil {
		pdb.log.Err(err).
			Str("public_id", publicId).
			Msg("SaveRiasecSession failed")
		return err
	}

	pdb.log.Debug().
		Str("public_id", publicId).
		Msg("SaveRiasecSession: done")

	return nil
}

// UpdateRiasecAnswers 更新指定 publicId 对应的 RIASEC 测试答案 JSON。
func (pdb *psDatabase) UpdateRiasecAnswers(
	ctx context.Context,
	publicId string,
	answersJSON []byte,
) error {
	pdb.log.Debug().
		Str("public_id", publicId).
		Msg("UpdateRiasecAnswers")

	if publicId == "" {
		return errors.New("publicId must be non-empty")
	}
	if len(answersJSON) == 0 {
		return errors.New("answersJSON must be non-empty")
	}

	const q = `
		UPDATE app.riasec_sessions
		SET
			answers      = $2::jsonb,
			completed_at = now()
		WHERE public_id = $1
	`

	_, err := pdb.db.ExecContext(ctx, q,
		publicId,
		string(answersJSON),
	)

	if err != nil {
		pdb.log.Err(err).
			Str("public_id", publicId).
			Msg("UpdateRiasecAnswers failed")
		return err
	}

	return nil
}
