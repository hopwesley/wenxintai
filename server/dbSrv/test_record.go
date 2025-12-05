package dbSrv

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/hopwesley/wenxintai/server/ai_api"
)

type TestRecord struct {
	PublicId     string
	BusinessType string
	PayOrderId   sql.NullString
	WeChatID     sql.NullString
	Grade        sql.NullString
	Mode         sql.NullString
	Hobby        sql.NullString
	CurStage     int16
	CreatedAt    time.Time
	PaidTime     sql.NullTime
}

func (pdb *psDatabase) NewTestRecord(
	ctx context.Context,
	businessTyp string,
	weChatId string,
) (string, error) {
	if len(weChatId) == 0 || len(businessTyp) == 0 {
		return "", errors.New("either businessType or weChatId must be non-nil")
	}

	sLog := pdb.log.With().Str("business_type", businessTyp).
		Str("wechat_openid", weChatId).Logger()

	sLog.Debug().Msg("NewTestRecord")

	const insertSQL = `
		INSERT INTO app.tests_record (business_type, wechat_openid)
		VALUES ($1, $2)
		RETURNING public_id
	`

	var publicID string

	err := pdb.WithTx(ctx, func(tx *sql.Tx) error {
		if err := tx.QueryRowContext(
			ctx,
			insertSQL,
			businessTyp,
			weChatId,
		).Scan(&publicID); err != nil {
			sLog.Err(err).Msg("newTestRecordWithWeChat: insert tests_record failed")
			return err
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	sLog.Debug().Str("public_id", publicID).Msg("newTestRecordWithWeChat created")
	return publicID, nil
}

func (pdb *psDatabase) UpdateRecordBasicInfo(
	ctx context.Context,
	uid string,
	bi *ai_api.BasicInfo,
	status int16,
) (string, error) {
	const q = `
        UPDATE app.tests_record
	SET grade = $2,
	    mode = $3,
	    hobby = NULLIF($4, ''),
	    cur_stage = $5,
	    updated_at = now()
	WHERE public_id = $1
	AND wechat_openid = $6
	RETURNING business_type
    `

	var businessType string
	err := pdb.db.QueryRowContext(ctx, q,
		bi.PublicId, bi.Grade, bi.Mode, bi.Hobby, status, uid,
	).Scan(&businessType)
	if err != nil {
		return "", err
	}
	return businessType, nil
}

func (pdb *psDatabase) QueryRecordBasicInfo(ctx context.Context, publicId string) (*ai_api.BasicInfo, error) {

	const q = `
        SELECT public_id, grade, mode, COALESCE(hobby, '')
        FROM app.tests_record
        WHERE public_id = $1
    `

	var (
		publicIDDB string
		gradeStr   string
		modeStr    string
		hobbyStr   string
	)
	sLog := pdb.log.With().Str("public_id", publicId).Logger()

	sLog.Debug().Msg("QueryRecordBasicInfo: start")

	err := pdb.db.
		QueryRowContext(ctx, q, publicId).
		Scan(&publicIDDB, &gradeStr, &modeStr, &hobbyStr)
	if err != nil {
		sLog.Err(err).Msg("QueryRecordBasicInfo failed")
		return nil, err
	}

	info := &ai_api.BasicInfo{
		PublicId: publicIDDB,
		Grade:    ai_api.Grade(gradeStr),
		Mode:     ai_api.Mode(modeStr),
	}
	if hobbyStr != "" {
		info.Hobby = hobbyStr
	}

	sLog.Debug().Msg("QueryRecordBasicInfo: done")
	return info, nil
}

func (pdb *psDatabase) QueryRecordByPid(ctx context.Context, publicID string) (*TestRecord, error) {
	sLog := pdb.log.With().
		Str("public_id", publicID).
		Logger()

	sLog.Debug().Msg("QueryRecordByPid")

	const q = `
      SELECT 
         public_id,
         business_type,
         pay_order_id,
         wechat_openid,
         grade,
         mode,
         hobby,
         cur_stage,
         created_at,
         paid_time
      FROM app.tests_record
      WHERE public_id = $1
      ORDER BY created_at DESC
      LIMIT 1
   `

	row := pdb.db.QueryRowContext(ctx, q, publicID)

	var rec TestRecord
	err := row.Scan(
		&rec.PublicId,
		&rec.BusinessType,
		&rec.PayOrderId,
		&rec.WeChatID,
		&rec.Grade,
		&rec.Mode,
		&rec.Hobby,
		&rec.CurStage,
		&rec.CreatedAt,
		&rec.PaidTime,
	)

	if errors.Is(err, sql.ErrNoRows) {
		sLog.Err(err).Msg("QueryRecordByPid no record")
		return nil, nil
	}

	if err != nil {
		sLog.Err(err).Msg("QueryRecordByPid failed")
		return nil, err
	}

	return &rec, nil
}

func (pdb *psDatabase) QueryRecordOfUser(ctx context.Context, uid, bType string) (*TestRecord, error) {
	sLog := pdb.log.With().
		Str("wechat_id", uid).
		Str("business_type", bType).
		Logger()

	sLog.Debug().Msg("QueryRecordOfUser")

	const q = `
      SELECT 
         public_id,
         business_type,
         pay_order_id,
         wechat_openid,
         grade,
         mode,
         hobby,
         cur_stage,
         created_at,
         paid_time
      FROM app.tests_record
      WHERE wechat_openid = $1
      AND business_type = $2
      ORDER BY created_at DESC
      LIMIT 1
   `

	row := pdb.db.QueryRowContext(ctx, q, uid, bType)

	var rec TestRecord
	err := row.Scan(
		&rec.PublicId,
		&rec.BusinessType,
		&rec.PayOrderId,
		&rec.WeChatID,
		&rec.Grade,
		&rec.Mode,
		&rec.Hobby,
		&rec.CurStage,
		&rec.CreatedAt,
		&rec.PaidTime,
	)

	if errors.Is(err, sql.ErrNoRows) {
		sLog.Info().Msg("QueryRecordOfUser no record")
		return nil, nil
	}

	if err != nil {
		sLog.Err(err).Msg("QueryRecordOfUser failed")
		return nil, err
	}

	return &rec, nil
}
