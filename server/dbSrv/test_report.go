package dbSrv

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const (
	CurrentEngineVersin = "v1.0.0"
)

type TestReport struct {
	ID            int64           `json:"id"`
	PublicId      string          `json:"public_id"`
	Mode          string          `json:"mode"`
	CommonScore   json.RawMessage `json:"common_score"`
	ModeParam     json.RawMessage `json:"mode_param"`
	AIContent     json.RawMessage `json:"ai_content"`
	EngineVersion string          `json:"engine_version"`
	Status        int16           `json:"status"`
	GeneratedAt   time.Time       `json:"generated_at"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

func (pdb *psDatabase) SaveTestReportCore(
	ctx context.Context,
	publicId string,
	mode string,
	commonScoreJSON []byte,
	modeParamJSON []byte,
) error {
	if publicId == "" {
		return errors.New("publicId must be non-empty")
	}
	if mode == "" {
		return errors.New("mode must be non-empty")
	}
	if len(commonScoreJSON) == 0 {
		return errors.New("commonScoreJSON must be non-empty")
	}
	if len(modeParamJSON) == 0 {
		return errors.New("modeParamJSON must be non-empty")
	}

	log := pdb.log.With().
		Str("public_id", publicId).
		Str("mode", mode).
		Logger()

	log.Debug().Msg("SaveTestReportCore: start")

	const q = `
        INSERT INTO app.test_reports (
            public_id, mode, common_score,  mode_param, engine_version
        )
        VALUES ($1, $2, $3::jsonb, $4::jsonb, $5)
        ON CONFLICT (public_id)
        DO UPDATE SET
            mode        = EXCLUDED.mode,
            common_score = EXCLUDED.common_score,
            mode_param   = EXCLUDED.mode_param,
            engine_version = EXCLUDED.engine_version,
            updated_at   = now()
    `

	_, err := pdb.db.ExecContext(ctx, q,
		publicId,
		mode,
		commonScoreJSON,
		modeParamJSON,
		CurrentEngineVersin,
	)
	if err != nil {
		log.Err(err).Msg("SaveTestReportCore: exec failed")
		return err
	}

	log.Debug().Msg("SaveTestReportCore: done")
	return nil
}

func (pdb *psDatabase) FindTestReportByPublicId(
	ctx context.Context,
	publicId string,
) (*TestReport, error) {
	if publicId == "" {
		return nil, errors.New("publicId must be non-empty")
	}

	log := pdb.log.With().
		Str("public_id", publicId).
		Logger()

	log.Debug().Msg("FindTestReportByPublicId: start")

	const q = `
        SELECT
            id,
            public_id,
            mode,
            common_score,
            mode_param,
            COALESCE(ai_content, 'null'::jsonb) AS ai_content,
            engine_version,
            status,
            generated_at,
            created_at,
            updated_at
        FROM app.test_reports
        WHERE public_id = $1
        ORDER BY generated_at DESC
        LIMIT 1
    `

	row := pdb.db.QueryRowContext(ctx, q, publicId)

	var r TestReport
	if err := row.Scan(
		&r.ID,
		&r.PublicId,
		&r.Mode,
		&r.CommonScore,
		&r.ModeParam,
		&r.AIContent,
		&r.EngineVersion,
		&r.Status,
		&r.GeneratedAt,
		&r.CreatedAt,
		&r.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Warn().Msg("FindTestReportByPublicId: no record")
			return nil, nil
		}
		log.Err(err).Msg("FindTestReportByPublicId: query failed")
		return nil, err
	}

	log.Debug().Msg("FindTestReportByPublicId: done")
	return &r, nil
}

func (pdb *psDatabase) UpdateTestReportAIContent(
	ctx context.Context,
	publicId string,
	aiContentJSON []byte,
) error {
	if publicId == "" {
		return errors.New("publicId must be non-empty")
	}
	if len(aiContentJSON) == 0 {
		return errors.New("aiContentJSON must be non-empty")
	}

	log := pdb.log.With().
		Str("public_id", publicId).
		Logger()

	log.Debug().Msg("UpdateTestReportAIContent: start")

	const q = `
        UPDATE app.test_reports
        SET ai_content   = $2::jsonb,
            generated_at = now(),
            updated_at   = now(),
            status = 1
        WHERE public_id = $1
    `

	res, err := pdb.db.ExecContext(ctx, q, publicId, aiContentJSON)
	if err != nil {
		log.Err(err).Msg("UpdateTestReportAIContent: exec failed")
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		log.Err(err).Msg("UpdateTestReportAIContent: RowsAffected failed")
		return err
	}
	if affected == 0 {
		err := fmt.Errorf("no test_report found for public_id=%s", publicId)
		log.Warn().Err(err).Msg("UpdateTestReportAIContent: not found")
		return err
	}

	log.Debug().Int64("rows", affected).Msg("UpdateTestReportAIContent: done")
	return nil
}

func (pdb *psDatabase) FinalizedTest(ctx context.Context, publicID string, businessType string) error {
	if publicID == "" {
		return errors.New("publicID must be non-empty")
	}
	if businessType == "" {
		return errors.New("businessType must be non-empty")
	}

	log := pdb.log.With().
		Str("public_id", publicID).
		Str("business_type", businessType).
		Logger()

	log.Debug().Msg("FinalizedTest: start")

	const updateRecordSQL = `
		UPDATE app.tests_record
		SET 
			completed_at = now(),
			updated_at   = now()
		WHERE public_id = $1
		  AND business_type = $2
		RETURNING pay_order_id, wechat_openid
	`

	const incReportNoSQL = `
		UPDATE app.user_profile
		SET 
			report_no  = COALESCE(report_no, 0) + 1,
			updated_at = now()
		WHERE uid = $1
	`

	const updateReportSQL = `
		UPDATE app.test_reports
		SET 
			status     = 2,
			updated_at = now()
		WHERE public_id = $1
	`

	err := pdb.WithTx(ctx, func(tx *sql.Tx) error {
		var inviteCode sql.NullString
		var wechatOpenID sql.NullString

		if err := tx.QueryRowContext(
			ctx,
			updateRecordSQL,
			publicID,
			businessType,
		).Scan(&inviteCode, &wechatOpenID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				log.Warn().Msg("FinalizedTest: no tests_record row matched")
			} else {
				log.Err(err).Msg("FinalizedTest: update tests_record failed")
			}
			return err
		}

		// 2) 分支：优先使用 wechat_openid 自增 user_profile.report_no
		if wechatOpenID.Valid && wechatOpenID.String != "" {
			res, err := tx.ExecContext(ctx, incReportNoSQL, wechatOpenID.String)
			if err != nil {
				log.Err(err).
					Str("wechat_openid", wechatOpenID.String).
					Msg("FinalizedTest: update user_profile.report_no failed")
				return err
			}
			if rows, _ := res.RowsAffected(); rows == 0 {
				// 没有 user_profile 的话这里我只是打个日志，不认为是致命错误
				log.Warn().
					Str("wechat_openid", wechatOpenID.String).
					Msg("FinalizedTest: no user_profile row updated")
			}
		}

		// 3) 更新 test_reports.status = 2
		if _, err := tx.ExecContext(ctx, updateReportSQL, publicID); err != nil {
			log.Err(err).Msg("FinalizedTest: update test_reports failed")
			return err
		}

		return nil
	})

	if err != nil {
		log.Err(err).Msg("FinalizedTest: failed")
		return err
	}

	log.Debug().Msg("FinalizedTest: done")
	return nil
}
