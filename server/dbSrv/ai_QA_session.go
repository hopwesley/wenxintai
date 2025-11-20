package dbSrv

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

type QASession struct {
	ID           int64           `json:"id"`
	BusinessType string          `json:"business_type"`
	TestType     string          `json:"test_type"`
	PublicId     string          `json:"public_id"`
	Questions    json.RawMessage `json:"questions"`
	Answers      json.RawMessage `json:"answers,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	CompletedAt  *time.Time      `json:"completed_at,omitempty"`
}

func (pdb *psDatabase) FindQASession(
	ctx context.Context,
	businessType string,
	testType string,
	publicId string,
) (*QASession, error) {
	if businessType == "" || publicId == "" || testType == "" {
		return nil, errors.New("businessType and publicId must be non-empty")
	}

	sLog := pdb.log.With().Str("business_type", businessType).
		Str("testType", testType).
		Str("public_id", publicId).Logger()

	sLog.Debug().Msg("FindQASession: start")

	const q = `
SELECT id, business_type,test_type, public_id, questions, COALESCE(answers, 'null'::jsonb) AS answers, created_at, completed_at
FROM app.question_answers
WHERE business_type = $1 AND test_type = $2  AND public_id = $3
`

	var sess QASession

	err := pdb.db.QueryRowContext(ctx, q, businessType, testType, publicId).
		Scan(
			&sess.ID,
			&sess.BusinessType,
			&sess.TestType,
			&sess.PublicId,
			&sess.Questions,
			&sess.Answers,
			&sess.CreatedAt,
			&sess.CompletedAt,
		)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sLog.Info().Msg("FindQASession: not found")
			return nil, nil
		}

		sLog.Err(err).Msg("FindQASession failed")
		return nil, err
	}

	sLog.Debug().Msg("FindQASession: found")

	return &sess, nil
}

func (pdb *psDatabase) SaveQuestion(
	ctx context.Context,
	businessType, testType, publicId string,
	questionsJSON []byte,
) error {
	if publicId == "" {
		return errors.New("publicId must be non-empty")
	}
	if businessType == "" {
		return errors.New("businessType must be non-empty")
	}
	if testType == "" {
		return errors.New("testType must be non-empty")
	}
	if len(questionsJSON) == 0 {
		return errors.New("questionsJSON must be non-empty")
	}

	sLog := pdb.log.With().Str("public_id", publicId).Str("test_type", testType).
		Str("business_type", businessType).Logger()
	sLog.Debug().Msg("SaveQuestion: start")

	const q = `
		INSERT INTO app.question_answers (business_type,test_type, public_id, questions)
		VALUES ($1, $2,$3, $4::jsonb)
		ON CONFLICT (business_type, test_type, public_id)
		DO UPDATE SET
			questions = EXCLUDED.questions,
			-- 保留最初创建时间，只更新题目内容
			created_at = app.question_answers.created_at
	`

	_, err := pdb.db.ExecContext(ctx, q,
		businessType, testType,
		publicId,
		string(questionsJSON),
	)
	if err != nil {
		sLog.Err(err).Msg("SaveQuestion failed")
		return err
	}

	sLog.Debug().Msg("SaveQuestion: done")

	return nil
}

func (pdb *psDatabase) SaveAnswer(
	ctx context.Context,
	businessType, testType, publicId string,
	answersJSON []byte,
) error {
	if publicId == "" {
		return errors.New("publicId must be non-empty")
	}
	if businessType == "" {
		return errors.New("businessType must be non-empty")
	}
	if testType == "" {
		return errors.New("testType must be non-empty")
	}
	if len(answersJSON) == 0 {
		return errors.New("answersJSON must be non-empty")
	}

	sLog := pdb.log.With().
		Str("public_id", publicId).
		Str("test_type", testType).
		Str("business_type", businessType).
		Logger()
	sLog.Debug().Msg("SaveAnswer: start")

	// 这里直接 UPDATE，假设前面 SaveQuestion 一定已经插入过对应记录
	const q = `
		UPDATE app.question_answers
		SET
			answers      = $4::jsonb,
			completed_at = now()
		WHERE business_type = $1
		  AND test_type     = $2
		  AND public_id     = $3
	`

	res, err := pdb.db.ExecContext(ctx, q,
		businessType,
		testType,
		publicId,
		string(answersJSON),
	)
	if err != nil {
		sLog.Err(err).Msg("SaveAnswer failed")
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		sLog.Err(err).Msg("SaveAnswer: RowsAffected failed")
		return err
	}
	if rows == 0 {
		// 说明没有找到对应的问题记录（可能 SaveQuestion 没成功）
		err = errors.New("SaveAnswer: no matching question_answers row")
		sLog.Err(err).Msg("SaveAnswer: no rows updated")
		return err
	}

	sLog.Debug().Msg("SaveAnswer: done")
	return nil
}
