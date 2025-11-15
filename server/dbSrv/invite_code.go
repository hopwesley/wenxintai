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
	ExpiresAt sql.NullTime
	UsedBy    sql.NullString
	UsedAt    sql.NullTime
	CreatedAt time.Time
}

const (
	InviteStatusUnused int16 = 0
	// 其它状态你可以按实际再扩展：
	// InviteStatusUsed    int16 = 1
	// InviteStatusExpired int16 = 2
	// InviteStatusBlocked int16 = 3
)

// GetInviteByCode 按 code 查邀请码，不存在时返回 (nil, nil)。
func (pdb *psDatabase) GetInviteByCode(ctx context.Context, code string) (*Invite, error) {
	pdb.log.Debug().Str("code", code).Msg("GetInviteByCode")
	const q = `
		SELECT code, status, expires_at, used_by, used_at, created_at
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
