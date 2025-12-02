package dbSrv

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/hopwesley/wenxintai/server/ai_api"
)

type TestRecord struct {
	PublicId     string
	BusinessType string
	InviteCode   sql.NullString
	WeChatID     sql.NullString
	Grade        sql.NullString
	Mode         sql.NullString
	Hobby        sql.NullString
	Status       int16
	CreatedAt    time.Time
	CompletedAt  sql.NullTime
}

func (pdb *psDatabase) QueryTestInProcess(ctx context.Context, uid, businessType string) (*TestRecord, error) {
	log := pdb.log.With().Str("wechat_id", uid).Str("", businessType).Logger()
	log.Debug().Msg("QueryTestInProcess")

	const q = `
        SELECT 
            public_id,
            business_type,
            invite_code,
            wechat_openid,
            grade,
            mode,
            hobby,
            status,
            created_at,
            completed_at
        FROM app.tests_record
        WHERE wechat_openid = $1
      		AND business_type = $2
      		AND completed_at IS NULL
        ORDER BY created_at DESC
        LIMIT 1
    `

	row := pdb.db.QueryRowContext(ctx, q, uid, businessType)

	var rec TestRecord
	err := row.Scan(
		&rec.PublicId,
		&rec.BusinessType,
		&rec.InviteCode,
		&rec.WeChatID,
		&rec.Grade,
		&rec.Mode,
		&rec.Hobby,
		&rec.Status,
		&rec.CreatedAt,
		&rec.CompletedAt,
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
	ctx context.Context, publicId string,
) (*TestRecord, error) {
	log := pdb.log.With().Str("public_id", publicId).Logger()
	log.Debug().Msg("QueryUnfinishedTest")

	const q = `
        SELECT 
            public_id,
            business_type,
            invite_code,
            wechat_openid,
            grade,
            mode,
            hobby,
            status,
            created_at,
            completed_at
        FROM app.tests_record
        WHERE public_id = $1
      		AND completed_at IS NULL
        ORDER BY created_at DESC
        LIMIT 1
    `

	row := pdb.db.QueryRowContext(ctx, q, publicId)

	var rec TestRecord
	err := row.Scan(
		&rec.PublicId,
		&rec.BusinessType,
		&rec.InviteCode,
		&rec.WeChatID,
		&rec.Grade,
		&rec.Mode,
		&rec.Hobby,
		&rec.Status,
		&rec.CreatedAt,
		&rec.CompletedAt,
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
	inviteCode *string,
	weChatId *string,
) (string, error) {
	pdb.log.Debug().
		Str("business_type", businessTyp).
		Msg("NewTestRecord")

	if inviteCode == nil && weChatId == nil {
		return "", errors.New("either inviteCode or weChatId must be non-nil")
	}

	// 通过邀请码创建：需要顺便绑定 invites.public_id
	if inviteCode != nil {
		return pdb.newTestRecordWithInvite(ctx, businessTyp, *inviteCode)
	}

	// 通过 wechat_openid 创建：只写 tests_record
	return pdb.newTestRecordWithWeChat(ctx, businessTyp, *weChatId)
}

func (pdb *psDatabase) newTestRecordWithInvite(
	ctx context.Context,
	businessTyp string,
	inviteCode string,
) (string, error) {
	pdb.log.Debug().
		Str("business_type", businessTyp).
		Str("invite_code", inviteCode).
		Msg("newTestRecordWithInvite")

	const insertSQL = `
		INSERT INTO app.tests_record (business_type, wechat_openid, invite_code)
		VALUES ($1, $2, $3)
		RETURNING public_id
	`

	const updateInviteSQL = `
		UPDATE app.invites
		SET public_id = $2,
		status    = $3,
    		used_at   = now()
		WHERE code = $1
	`

	var publicID string

	err := pdb.WithTx(ctx, func(tx *sql.Tx) error {
		// 1) 插入 tests_record
		if err := tx.QueryRowContext(
			ctx,
			insertSQL,
			businessTyp,
			nil,        // wechat_openid 为空
			inviteCode, // invite_code
		).Scan(&publicID); err != nil {
			pdb.log.Err(err).
				Str("business_type", businessTyp).
				Str("invite_code", inviteCode).
				Msg("newTestRecordWithInvite: insert tests_record failed")
			return err
		}

		// 2) 更新 invites.public_id
		res, err := tx.ExecContext(ctx, updateInviteSQL, inviteCode, publicID, InviteStatusInUse)
		if err != nil {
			pdb.log.Err(err).
				Str("invite_code", inviteCode).
				Str("public_id", publicID).
				Msg("newTestRecordWithInvite: update invites failed")
			return err
		}

		affected, err := res.RowsAffected()
		if err != nil {
			pdb.log.Err(err).
				Str("invite_code", inviteCode).
				Msg("newTestRecordWithInvite: RowsAffected error")
			return err
		}

		if affected == 0 {
			err := fmt.Errorf("invite code not found: %s", inviteCode)
			pdb.log.Err(err).
				Str("invite_code", inviteCode).
				Msg("newTestRecordWithInvite: no invites row updated")
			return err
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	pdb.log.Debug().
		Str("business_type", businessTyp).
		Str("invite_code", inviteCode).
		Str("public_id", publicID).
		Msg("newTestRecordWithInvite created")

	return publicID, nil
}

func (pdb *psDatabase) newTestRecordWithWeChat(
	ctx context.Context,
	businessTyp string,
	weChatId string,
) (string, error) {
	pdb.log.Debug().
		Str("business_type", businessTyp).
		Str("wechat_openid", weChatId).
		Msg("newTestRecordWithWeChat")

	const insertSQL = `
		INSERT INTO app.tests_record (business_type, wechat_openid, invite_code)
		VALUES ($1, $2, $3)
		RETURNING public_id
	`

	var publicID string

	err := pdb.WithTx(ctx, func(tx *sql.Tx) error {
		if err := tx.QueryRowContext(
			ctx,
			insertSQL,
			businessTyp,
			weChatId, // wechat_openid
			nil,      // invite_code 为空
		).Scan(&publicID); err != nil {
			pdb.log.Err(err).
				Str("business_type", businessTyp).
				Str("wechat_openid", weChatId).
				Msg("newTestRecordWithWeChat: insert tests_record failed")
			return err
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	pdb.log.Debug().
		Str("business_type", businessTyp).
		Str("wechat_openid", weChatId).
		Str("public_id", publicID).
		Msg("newTestRecordWithWeChat created")

	return publicID, nil
}

func (pdb *psDatabase) UpdateBasicInfo(
	ctx context.Context,
	publicId string,
	grade string,
	mode string,
	hobby string,
	status int,
) (string, error) {
	const q = `
        UPDATE app.tests_record
        SET grade = $2,
            mode = $3,
            hobby = NULLIF($4, ''),
            status = $5,
            updated_at = now()
        WHERE public_id = $1
        RETURNING business_type
    `

	var businessType string
	err := pdb.db.QueryRowContext(ctx, q,
		publicId, grade, mode, hobby, status,
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
