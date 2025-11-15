package dbSrv

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type TestRecord struct {
	ID          int64
	TestType    string
	InviteCode  sql.NullString
	Status      int16
	CreatedAt   time.Time
	CompletedAt sql.NullTime
}

// FindLatestTestRecordByInvite 根据 invite_code 查最近的一条测试记录。
// 没查到时返回 (nil, nil)，出错返回 (nil, err)。
func (pdb *psDatabase) FindLatestTestRecordByInvite(ctx context.Context, inviteCode string) (*TestRecord, error) {
	pdb.log.Debug().Str("invite code", inviteCode).Msg("FindLatestTestRecordByInvite")
	const q = `
		SELECT id, test_type, invite_code, status, created_at, completed_at
		FROM app.tests_record
		WHERE invite_code = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	row := pdb.db.QueryRowContext(ctx, q, inviteCode)

	var rec TestRecord
	err := row.Scan(
		&rec.ID,
		&rec.TestType,
		&rec.InviteCode,
		&rec.Status,
		&rec.CreatedAt,
		&rec.CompletedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		pdb.log.Err(err).Str("invite code", inviteCode).Msg("no record")
		return nil, nil
	}
	if err != nil {
		pdb.log.Err(err).Str("invite code", inviteCode).Msg("database query error")
		return nil, err
	}
	pdb.log.Debug().Str("invite code", inviteCode).Msg("find record")
	return &rec, nil
}
