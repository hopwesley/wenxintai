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
	Status       int16
	CreatedAt    time.Time
	PaidTime     sql.NullTime
}

func (pdb *psDatabase) QueryTestInProcess(ctx context.Context, uid, businessType string) (*TestRecord, error) {
	log := pdb.log.With().Str("wechat_id", uid).Str("", businessType).Logger()
	log.Debug().Msg("QueryTestInProcess")

	const q = `
        SELECT 
            public_id,
            business_type,
            pay_order_id,
            wechat_openid,
            grade,
            mode,
            hobby,
            status,
            created_at
        FROM app.tests_record
        WHERE wechat_openid = $1
      		AND business_type = $2
      		AND paid_time IS NULL
        ORDER BY created_at DESC
        LIMIT 1
    `

	row := pdb.db.QueryRowContext(ctx, q, uid, businessType)

	var rec TestRecord
	err := row.Scan(
		&rec.PublicId,
		&rec.BusinessType,
		&rec.PayOrderId,
		&rec.WeChatID,
		&rec.Grade,
		&rec.Mode,
		&rec.Hobby,
		&rec.Status,
		&rec.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		log.Err(err).Msg("no record")
		return nil, nil
	}
	if err != nil {
		log.Err(err).Msg("database query error")
		return nil, err
	}

	log.Debug().Msg("find record")
	return &rec, nil
}

func (pdb *psDatabase) QueryUnfinishedTest(
	ctx context.Context, publicId, uid string,
) (*TestRecord, error) {
	log := pdb.log.With().Str("public_id", publicId).Logger()
	log.Debug().Msg("QueryUnfinishedTest")

	const q = `
        SELECT 
            public_id,
            business_type,
            pay_order_id,
            wechat_openid,
            grade,
            mode,
            hobby,
            status,
            created_at
        FROM app.tests_record
        WHERE public_id = $1
          AND wechat_openid = $2
          AND paid_time IS NULL
        ORDER BY created_at DESC
        LIMIT 1
    `

	row := pdb.db.QueryRowContext(ctx, q, publicId, uid)

	var rec TestRecord
	err := row.Scan(
		&rec.PublicId,
		&rec.BusinessType,
		&rec.PayOrderId,
		&rec.WeChatID,
		&rec.Grade,
		&rec.Mode,
		&rec.Hobby,
		&rec.Status,
		&rec.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		log.Err(err).Msg("no record")
		return nil, nil
	}
	if err != nil {
		log.Err(err).Msg("database query error")
		return nil, err
	}

	log.Debug().Msg("find record")
	return &rec, nil
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

func (pdb *psDatabase) UpdateBasicInfo(
	ctx context.Context,
	publicId string,
	uid string,
	grade string,
	mode string,
	hobby string,
	status int,
) (string, error) {
	const q = `
        UPDATE app.tests_record
	SET grade = $3,
	    mode = $4,
	    hobby = NULLIF($5, ''),
	    status = $6,
	    updated_at = now()
	WHERE public_id = $1
	  AND wechat_openid = $2
	RETURNING business_type
    `

	var businessType string
	err := pdb.db.QueryRowContext(ctx, q,
		publicId, uid, grade, mode, hobby, status,
	).Scan(&businessType)
	if err != nil {
		return "", err
	}
	return businessType, nil
}

func (pdb *psDatabase) QueryBasicInfo(ctx context.Context, publicId string) (*ai_api.BasicInfo, error) {

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

	pdb.log.Debug().
		Str("public_id", publicId).
		Msg("QueryBasicInfo: start")

	err := pdb.db.
		QueryRowContext(ctx, q, publicId).
		Scan(&publicIDDB, &gradeStr, &modeStr, &hobbyStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			pdb.log.Warn().
				Str("public_id", publicId).
				Msg("QueryBasicInfo: no record found")
		} else {
			pdb.log.Err(err).
				Str("public_id", publicId).
				Msg("QueryBasicInfo failed")
		}
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

	pdb.log.Debug().
		Str("public_id", publicId).
		Interface("basic_info", info).
		Msg("QueryBasicInfo: done")

	return info, nil
}

func (pdb *psDatabase) QueryRecordById(ctx context.Context, publicID string) (*TestRecord, error) {
	log := pdb.log.With().
		Str("public_id", publicID).
		Logger()
	log.Debug().Msg("QueryRecordById")

	const q = `
		SELECT 
			public_id,
			business_type,
			pay_order_id,
			wechat_openid,
			grade,
			mode,
			hobby,
			status,
			created_at,
			paid_time
		FROM app.tests_record
		WHERE public_id = $1
		ORDER BY paid_time DESC
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
		&rec.Status,
		&rec.CreatedAt,
		&rec.PaidTime,
	)
	if err != nil {
		// 包括 sql.ErrNoRows 在内，都直接返回给上层
		log.Err(err).Msg("QueryRecordById failed")
		return nil, err
	}

	return &rec, nil
}
