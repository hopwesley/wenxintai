package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

func (r *SQLRepo) GetInviteForUpdate(ctx context.Context, code string) (*Invite, error) {
	row := r.getExecer(ctx).QueryRowContext(ctx, `
        SELECT code, status, expires_at, used_by, used_at, created_at
        FROM app.invites
        WHERE code = $1
        FOR UPDATE
    `, code)
	var inv Invite
	if err := row.Scan(&inv.Code, &inv.Status, &inv.ExpiresAt, &inv.UsedBy, &inv.UsedAt, &inv.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &inv, nil
}

func (r *SQLRepo) UpdateInviteReservation(ctx context.Context, code, sessionID string, until time.Time) error {
	res, err := r.getExecer(ctx).ExecContext(ctx, `
        UPDATE app.invites
        SET status = 1, used_by = $2, expires_at = $3
        WHERE code = $1
          AND (
                status = 0                              -- 未占用 -> 占用
             OR (status = 1 AND used_by = $2)          -- 自己续期
             OR (status = 1 AND expires_at <= now())   -- 他人占用但过期 -> 抢占
          )
    `, code, sessionID, until)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return ErrConflict
	}
	return nil
}

func (r *SQLRepo) RedeemInviteBySession(ctx context.Context, sessionID, redeemedBy string) (bool, error) {
	res, err := r.getExecer(ctx).ExecContext(ctx, `
        UPDATE app.invites
        SET status = 2,
            used_by = COALESCE(used_by, $2),
            used_at = now(),
            expires_at = NULL
        WHERE status = 1
          AND used_by = $1
          AND (expires_at IS NULL OR expires_at > now())
    `, sessionID, redeemedBy)
	if err != nil {
		return false, fmt.Errorf("redeem invite by session: %w", err)
	}
	n, _ := res.RowsAffected()
	return n == 1, nil
}

func (r *SQLRepo) RedeemInviteByCode(ctx context.Context, code, by string) (bool, error) {
	res, err := r.getExecer(ctx).ExecContext(ctx, `
        UPDATE app.invites
        SET status = 2,
            used_by = COALESCE(used_by, $2),
            used_at = now(),
            expires_at = NULL
        WHERE code = $1
          AND (status = 0 OR status = 1)
    `, code, by)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n == 1, nil
}
