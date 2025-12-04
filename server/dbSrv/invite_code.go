package dbSrv

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Invite struct {
	Code      string
	Status    int16
	PublicID  sql.NullString
	ExpiresAt sql.NullTime
	UsedAt    sql.NullTime
	CreatedAt time.Time
}

const (
	InviteStatusUnused int16 = 0
	InviteStatusUsed   int16 = 1
)

// GetInviteByCode 按 code 查邀请码，不存在时返回 (nil, nil)。
func (pdb *psDatabase) GetInviteByCode(ctx context.Context, code string) (*Invite, error) {
	pdb.log.Debug().Str("code", code).Msg("GetInviteByCode")
	const q = `
		SELECT code, status, expires_at, used_at, created_at, public_id
		FROM app.invites
		WHERE code = $1
	`

	row := pdb.db.QueryRowContext(ctx, q, code)

	var inv Invite
	err := row.Scan(
		&inv.Code,
		&inv.Status,
		&inv.ExpiresAt,
		&inv.UsedAt,
		&inv.CreatedAt,
		&inv.PublicID,
	)

	if errors.Is(err, sql.ErrNoRows) {
		pdb.log.Debug().Err(err).Msg("database query invite code not found")
		return nil, nil
	}
	if err != nil {
		pdb.log.Error().Err(err).Msg("database query error")
		return nil, err
	}
	pdb.log.Debug().Interface("inv", inv).Msg("invite code found")
	return &inv, nil
}

// UpdateInviteStatus 根据邀请码 code 更新其状态字段。
func (pdb *psDatabase) UpdateInviteStatus(ctx context.Context, code string, newStatus int16) error {
	pdb.log.Debug().
		Str("code", code).
		Int16("new_status", newStatus).
		Msg("UpdateInviteStatus: start")

	const q = `
		UPDATE app.invites
		SET status = $2
		WHERE code = $1
	`

	res, err := pdb.db.ExecContext(ctx, q, code, newStatus)
	if err != nil {
		pdb.log.Error().
			Err(err).
			Str("code", code).
			Int16("new_status", newStatus).
			Msg("UpdateInviteStatus: exec error")
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		pdb.log.Error().
			Err(err).
			Str("code", code).
			Msg("UpdateInviteStatus: RowsAffected error")
		return err
	}

	if rows == 0 {
		// 没有这个邀请码，可以按你习惯返回 sql.ErrNoRows 或自定义错误
		pdb.log.Warn().
			Str("code", code).
			Msg("UpdateInviteStatus: no rows affected")
		return sql.ErrNoRows
	}

	pdb.log.Debug().
		Str("code", code).
		Int64("rows", rows).
		Msg("UpdateInviteStatus: success")
	return nil
}

func (pdb *psDatabase) PayByInviteCode(ctx context.Context, publicId string, inviteCode string) error {
	sLog := pdb.log.With().
		Str("public_id", publicId).
		Str("invite_code", inviteCode).
		Logger()
	sLog.Debug().Msg("PayByInviteCode")

	tx, err := pdb.db.BeginTx(ctx, nil)
	if err != nil {
		sLog.Err(err).Msg("begin tx failed")
		return err
	}

	// 1) 更新 tests_record
	const qUpdateTestRecord = `
		UPDATE app.tests_record
		SET 
		    pay_order_id = $1,
		    paid_time    = NOW()
		WHERE public_id = $2
	`
	res1, err := tx.ExecContext(ctx, qUpdateTestRecord, inviteCode, publicId)
	if err != nil {
		sLog.Err(err).Msg("update tests_record failed")
		_ = tx.Rollback()
		return err
	}
	rows1, err := res1.RowsAffected()
	if err != nil {
		sLog.Err(err).Msg("rows affected (tests_record) failed")
		_ = tx.Rollback()
		return err
	}
	if rows1 == 0 {
		sLog.Warn().Msg("no tests_record updated")
		_ = tx.Rollback()
		return sql.ErrNoRows
	}

	// 2) 更新 invites
	const qUpdateInvite = `
		UPDATE app.invites
		SET 
		    status    = $1,
		    used_at   = NOW(),
		    public_id = $2
		WHERE code = $3
	`
	res2, err := tx.ExecContext(ctx, qUpdateInvite, InviteStatusUsed, publicId, inviteCode)
	if err != nil {
		sLog.Err(err).Msg("update invites failed")
		_ = tx.Rollback()
		return err
	}
	rows2, err := res2.RowsAffected()
	if err != nil {
		sLog.Err(err).Msg("rows affected (invites) failed")
		_ = tx.Rollback()
		return err
	}
	if rows2 == 0 {
		sLog.Warn().Msg("no invites updated")
		_ = tx.Rollback()
		return sql.ErrNoRows
	}

	if err = tx.Commit(); err != nil {
		sLog.Err(err).Msg("commit tx failed")
		return err
	}

	sLog.Info().Msg("update record status to paid success")
	return nil
}
