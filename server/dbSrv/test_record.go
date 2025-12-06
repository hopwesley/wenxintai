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
	bTyp, weChatId string,
	bi *ai_api.BasicInfo,
) (string, error) {
	if len(weChatId) == 0 || len(bTyp) == 0 {
		return "", errors.New("either businessType or weChatId must be non-nil")
	}

	sLog := pdb.log.With().Str("business_type", bTyp).
		Str("wechat_openid", weChatId).Logger()

	sLog.Debug().Msg("NewTestRecord")

	const insertSQL = `
		INSERT INTO app.tests_record (business_type, wechat_openid, grade, mode, hobby, cur_stage)
		VALUES ($1, $2, $3, $4, $5, 1)
		RETURNING public_id
	`

	var publicID string

	err := pdb.WithTx(ctx, func(tx *sql.Tx) error {
		if err := tx.QueryRowContext(
			ctx,
			insertSQL,
			bTyp,
			weChatId,
			bi.Grade,
			bi.Mode,
			bi.Hobby,
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
	publicId, uid string,
	bi *ai_api.BasicInfo,
) (string, error) {
	const q = `
        UPDATE app.tests_record
	SET grade = $2,
	    mode = $3,
	    hobby = NULLIF($4, ''),
	    cur_stage = 1,
	    updated_at = now()
	WHERE public_id = $1
	AND wechat_openid = $5
	RETURNING business_type
    `
	var businessType string
	err := pdb.db.QueryRowContext(ctx, q, publicId, bi.Grade, bi.Mode, bi.Hobby, uid).Scan(&businessType)
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
		Grade: ai_api.Grade(gradeStr),
		Mode:  ai_api.Mode(modeStr),
	}
	if hobbyStr != "" {
		info.Hobby = hobbyStr
	}

	sLog.Debug().Msg("QueryRecordBasicInfo: done")
	return info, nil
}

func (pdb *psDatabase) QueryTestRecord(ctx context.Context, pid, uid string) (*TestRecord, error) {
	sLog := pdb.log.With().
		Str("public_id", pid).
		Str("wechat_id", uid).
		Logger()

	sLog.Debug().Msg("QueryTestRecord")

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
      AND wechat_openid = $2
      ORDER BY created_at DESC
      LIMIT 1
   `

	row := pdb.db.QueryRowContext(ctx, q, pid, uid)

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
		sLog.Err(err).Msg("QueryTestRecord no record")
		return nil, nil
	}

	if err != nil {
		sLog.Err(err).Msg("QueryTestRecord failed")
		return nil, err
	}

	return &rec, nil
}

func (pdb *psDatabase) QueryUnfinishedTestOfUser(ctx context.Context, uid, bType string) (*TestRecord, error) {
	sLog := pdb.log.With().
		Str("wechat_id", uid).
		Str("business_type", bType).
		Logger()

	sLog.Debug().Msg("QueryUnfinishedTestOfUser")

	const q = `
      SELECT 
         t.public_id,
         t.business_type,
         t.pay_order_id,
         t.wechat_openid,
         t.grade,
         t.mode,
         t.hobby,
         t.cur_stage,
         t.created_at,
         t.paid_time
      FROM app.tests_record AS t
      LEFT JOIN app.test_reports AS r
        ON r.public_id = t.public_id
      WHERE t.wechat_openid = $1
        AND t.business_type = $2
        AND (r.id IS NULL OR r.status = 0)
      ORDER BY t.created_at DESC
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
		sLog.Info().Msg("QueryUnfinishedTestOfUser no record")
		return nil, nil
	}

	if err != nil {
		sLog.Err(err).Msg("QueryUnfinishedTestOfUser failed")
		return nil, err
	}

	return &rec, nil
}
