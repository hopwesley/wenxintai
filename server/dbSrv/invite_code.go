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
	UsedBy    sql.NullString
	UsedAt    sql.NullTime
	CreatedAt time.Time
}

const (
	InviteStatusUnused int16 = 0
	InviteStatusInUse  int16 = 1
	// InviteStatusUsed    int16 = 1
	// InviteStatusBlocked int16 = 3
)

// GetInviteByCode 按 code 查邀请码，不存在时返回 (nil, nil)。
func (pdb *psDatabase) GetInviteByCode(ctx context.Context, code string) (*Invite, error) {
	pdb.log.Debug().Str("code", code).Msg("GetInviteByCode")
	const q = `
		SELECT code, status, expires_at, used_by, used_at, created_at, public_id
		FROM app.invites
		WHERE code = $1
	`

	row := pdb.db.QueryRowContext(ctx, q, code)

	var inv Invite
	err := row.Scan(
		&inv.Code,
		&inv.Status,
		&inv.ExpiresAt,
		&inv.UsedBy,
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
