package dbSrv

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

type QASession struct {
	ID          int64           `json:"id"`
	TestType    string          `json:"test_type"`
	PublicId    string          `json:"public_id"`
	Questions   json.RawMessage `json:"questions"`
	Answers     json.RawMessage `json:"answers,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	CompletedAt *time.Time      `json:"completed_at,omitempty"`
}

func (pdb *psDatabase) FindQASession(
	ctx context.Context,
	testType string,
	publicId string,
) (*QASession, error) {
	if publicId == "" || testType == "" {
		return nil, errors.New("testType and publicId must be non-empty")
	}

	sLog := pdb.log.With().
		Str("testType", testType).
		Str("public_id", publicId).Logger()

	sLog.Debug().Msg("FindQASession: start")

	const q = `
SELECT id, test_type, public_id, questions, COALESCE(answers, 'null'::jsonb) AS answers, created_at, completed_at
FROM app.question_answers
WHERE test_type = $1  AND public_id = $2
`

	var sess QASession

	err := pdb.db.QueryRowContext(ctx, q, testType, publicId).
		Scan(
			&sess.ID,
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
	testType, publicId string,
	questionsJSON []byte,
) error {
	if publicId == "" {
		return errors.New("publicId must be non-empty")
	}
	if testType == "" {
		return errors.New("testType must be non-empty")
	}
	if len(questionsJSON) == 0 {
		return errors.New("questionsJSON must be non-empty")
	}

	sLog := pdb.log.With().Str("public_id", publicId).Str("test_type", testType).Logger()
	sLog.Debug().Msg("SaveQuestion: start")

	const q = `
		INSERT INTO app.question_answers (test_type, public_id, questions)
		VALUES ($1,$2, $3::jsonb)
		ON CONFLICT (test_type, public_id)
		DO UPDATE SET
			questions = EXCLUDED.questions,
			created_at = app.question_answers.created_at
	`

	_, err := pdb.db.ExecContext(ctx, q,
		testType,
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
	testType, publicId string,
	answersJSON []byte, status int,
) error {
	if publicId == "" {
		return errors.New("publicId must be non-empty")
	}
	if testType == "" {
		return errors.New("testType must be non-empty")
	}
	if len(answersJSON) == 0 {
		return errors.New("answersJSON must be non-empty")
	}
	if status <= 0 {
		return errors.New("status must be greater than 0")
	}

	sLog := pdb.log.With().
		Str("public_id", publicId).
		Str("test_type", testType).
		Logger()
	sLog.Debug().Msg("SaveAnswer: start")

	// 1) 更新 app.question_answers.answers / completed_at
	const updateAnswersSQL = `
		UPDATE app.question_answers
		SET
			answers      = $3::jsonb,
			completed_at = now()
		WHERE test_type     = $1
		  AND public_id     = $2
	`

	// 2) 更新 app.tests_record.status
	//    （tests_record 中才有 status 字段）
	const updateTestRecordSQL = `
		UPDATE app.tests_record
		SET 
		    status = GREATEST(status, $2),
		    updated_at =  now()
		WHERE public_id     = $1
	`

	err := pdb.WithTx(ctx, func(tx *sql.Tx) error {
		// --- 更新 question_answers ---
		res1, err := tx.ExecContext(ctx, updateAnswersSQL,
			testType,
			publicId,
			string(answersJSON),
		)
		if err != nil {
			sLog.Err(err).Msg("SaveAnswer: update question_answers failed")
			return err
		}

		rows1, err := res1.RowsAffected()
		if err != nil {
			sLog.Err(err).Msg("SaveAnswer: RowsAffected for question_answers failed")
			return err
		}
		if rows1 == 0 {
			err = errors.New("SaveAnswer: no matching question_answers row")
			sLog.Err(err).Msg("SaveAnswer: no rows updated in question_answers")
			return err
		}

		// --- 更新 tests_record.status ---
		res2, err := tx.ExecContext(ctx, updateTestRecordSQL,
			publicId,
			status,
		)
		if err != nil {
			sLog.Err(err).Msg("SaveAnswer: update tests_record failed")
			return err
		}

		rows2, err := res2.RowsAffected()
		if err != nil {
			sLog.Err(err).Msg("SaveAnswer: RowsAffected for tests_record failed")
			return err
		}
		if rows2 == 0 {
			err = errors.New("SaveAnswer: no matching tests_record row")
			sLog.Err(err).Msg("SaveAnswer: no rows updated in tests_record")
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	sLog.Debug().Msg("SaveAnswer: done")
	return nil
}

// FindQASessionsForReport 按 public_id 查出该用户本次测试下所有阶段的题目与答案
func (pdb *psDatabase) FindQASessionsForReport(
	ctx context.Context,
	publicId string,
) ([]*QASession, error) {
	if publicId == "" {
		return nil, errors.New("businessType and publicId must be non-empty")
	}

	sLog := pdb.log.With().
		Str("public_id", publicId).
		Logger()

	sLog.Debug().Msg("FindQASessionsForReport: start")

	const q = `
SELECT 
    id,
    test_type,
    public_id,
    questions,
    COALESCE(answers, 'null'::jsonb) AS answers,
    created_at,
    completed_at
FROM app.question_answers
WHERE public_id    = $1
ORDER BY test_type, created_at
`

	rows, err := pdb.db.QueryContext(ctx, q, publicId)
	if err != nil {
		sLog.Err(err).Msg("FindQASessionsForReport: query failed")
		return nil, err
	}
	defer rows.Close()

	var result []*QASession

	for rows.Next() {
		var sess QASession
		if err := rows.Scan(
			&sess.ID,
			&sess.TestType,
			&sess.PublicId,
			&sess.Questions,
			&sess.Answers,
			&sess.CreatedAt,
			&sess.CompletedAt,
		); err != nil {
			sLog.Err(err).Msg("FindQASessionsForReport: scan failed")
			return nil, err
		}
		result = append(result, &sess)
	}

	if err := rows.Err(); err != nil {
		sLog.Err(err).Msg("FindQASessionsForReport: rows error")
		return nil, err
	}

	if len(result) == 0 {
		sLog.Info().Msg("FindQASessionsForReport: not found")
		return nil, nil
	}

	sLog.Debug().Int("count", len(result)).Msg("FindQASessionsForReport: done")
	return result, nil
}
