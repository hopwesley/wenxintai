package dbSrv

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/hopwesley/wenxintai/server/ai_api"
)

type TestRecord struct {
	PublicId    string
	TestType    string
	InviteCode  sql.NullString
	WeChatID    sql.NullString
	Status      int16
	CreatedAt   time.Time
	CompletedAt sql.NullTime
}

func (pdb *psDatabase) FindRestRecordByUid(
	ctx context.Context,
	inviteCode, weChatID string,
) (*TestRecord, error) {
	pdb.log.Debug().
		Str("invite code", inviteCode).
		Str("wechat_openid", weChatID).
		Msg("FindRestRecordByUid")

	if inviteCode == "" && weChatID == "" {
		return nil, errors.New("either inviteCode or weChatID must be non-empty")
	}

	var (
		q   string
		arg string
	)

	if inviteCode != "" {
		q = `
            SELECT public_id, test_type, invite_code, status, created_at, completed_at
            FROM app.tests_record
            WHERE invite_code = $1
            ORDER BY created_at DESC
            LIMIT 1
        `
		arg = inviteCode
	} else {
		q = `
            SELECT public_id, test_type, invite_code, status, created_at, completed_at
            FROM app.tests_record
            WHERE wechat_openid = $1
            ORDER BY created_at DESC
            LIMIT 1
        `
		arg = weChatID
	}

	row := pdb.db.QueryRowContext(ctx, q, arg)

	var rec TestRecord
	err := row.Scan(
		&rec.PublicId,
		&rec.TestType,
		&rec.InviteCode,
		&rec.Status,
		&rec.CreatedAt,
		&rec.CompletedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		pdb.log.Err(err).
			Str("invite code", inviteCode).
			Str("wechat_openid", weChatID).
			Msg("no record")
		return nil, nil
	}
	if err != nil {
		pdb.log.Err(err).
			Str("invite code", inviteCode).
			Str("wechat_openid", weChatID).
			Msg("database query error")
		return nil, err
	}

	pdb.log.Debug().
		Str("invite code", inviteCode).
		Str("wechat_openid", weChatID).
		Str("public_id", rec.PublicId).
		Msg("find record")

	return &rec, nil
}

func (pdb *psDatabase) NewTestRecord(ctx context.Context, testType string, inviteCode *string, weChatId *string) (string, error) {
	pdb.log.Debug().Str("test_type", testType).Msg("NewTestRecord")

	if inviteCode == nil && weChatId == nil {
		return "", errors.New("either inviteCode or weChatId must be non-nil")
	}

	var inviteVal interface{}
	var wechatVal interface{}

	if inviteCode != nil {
		inviteVal = *inviteCode
	} else {
		inviteVal = nil
	}

	if weChatId != nil {
		wechatVal = *weChatId
	} else {
		wechatVal = nil
	}

	// language=SQL
	const q = `
		INSERT INTO app.tests_record (test_type, wechat_openid, invite_code)
		VALUES ($1, $2, $3)
		RETURNING public_id
	`

	var publicID string
	err := pdb.db.QueryRowContext(ctx, q,
		testType,
		wechatVal,
		inviteVal,
	).Scan(&publicID)
	if err != nil {
		pdb.log.Err(err).
			Str("test_type", testType).
			Msg("NewTestRecord insert failed")
		return "", err
	}

	pdb.log.Debug().
		Str("test_type", testType).
		Str("public_id", publicID).
		Msg("NewTestRecord created")

	return publicID, nil
}

func (pdb *psDatabase) UpdateBasicInfo(ctx context.Context, publicId string, grade string, mode string, hobby string) error {
	const q = `
        UPDATE app.tests_record
        SET grade = $2,
            "mode" = $3,
            hobby = NULLIF($4, ''),
            updated_at = now()
        WHERE public_id = $1
    `
	_, err := pdb.db.ExecContext(ctx, q, publicId, grade, mode, hobby)
	return err
}

func (pdb *psDatabase) QueryBasicInfo(ctx context.Context, publicId string) (*ai_api.BasicInfo, error) {

	const q = `
        SELECT public_id, grade, "mode", COALESCE(hobby, '')
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
