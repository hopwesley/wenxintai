package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

func (r *SQLRepo) GetInviteForUpdate(ctx context.Context, code string) (*Invite, error) {
	query := `SELECT code, status, reserved_by, reserved_until, redeemed_by, redeemed_at, created_at FROM app.invites WHERE code=$1`
	if _, ok := TxFromContext(ctx); ok {
		query += " FOR UPDATE"
	}
	row := r.getExecer(ctx).QueryRowContext(ctx, query, code)
	var (
		inv           Invite
		reservedBy    sql.NullString
		reservedUntil sql.NullTime
		redeemedBy    sql.NullString
		redeemedAt    sql.NullTime
	)
	if err := row.Scan(&inv.Code, &inv.Status, &reservedBy, &reservedUntil, &redeemedBy, &redeemedAt, &inv.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get invite: %w", err)
	}
	if reservedBy.Valid {
		v := reservedBy.String
		inv.ReservedBy = &v
	}
	if reservedUntil.Valid {
		t := reservedUntil.Time
		inv.ReservedUntil = &t
	}
	if redeemedBy.Valid {
		v := redeemedBy.String
		inv.RedeemedBy = &v
	}
	if redeemedAt.Valid {
		t := redeemedAt.Time
		inv.RedeemedAt = &t
	}
	return &inv, nil
}

func (r *SQLRepo) UpdateInviteReservation(ctx context.Context, code string, reservedBy *string, reservedUntil *time.Time) error {
	execer := r.getExecer(ctx)
	var by any
	if reservedBy != nil {
		by = *reservedBy
	}
	var until any
	if reservedUntil != nil {
		until = *reservedUntil
	}
	_, err := execer.ExecContext(ctx, `UPDATE app.invites SET reserved_by=$2, reserved_until=$3 WHERE code=$1`, code, by, until)
	if err != nil {
		return fmt.Errorf("update invite reservation: %w", err)
	}
	return nil
}

func (r *SQLRepo) RedeemInviteBySession(ctx context.Context, sessionID, redeemedBy string) (bool, error) {
	execer := r.getExecer(ctx)
	res, err := execer.ExecContext(ctx, `UPDATE app.invites SET status='redeemed', redeemed_by=$2, redeemed_at=NOW(), reserved_by=NULL, reserved_until=NULL WHERE reserved_by=$1 AND status='unused' AND (reserved_until IS NULL OR reserved_until>NOW())`, sessionID, redeemedBy)
	if err != nil {
		return false, fmt.Errorf("redeem invite by session: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("redeem invite by session rows: %w", err)
	}
	return affected == 1, nil
}

func (r *SQLRepo) RedeemInviteByCode(ctx context.Context, code, redeemedBy string) (bool, error) {
	execer := r.getExecer(ctx)
	res, err := execer.ExecContext(ctx, `UPDATE app.invites SET status='redeemed', redeemed_by=COALESCE(redeemed_by,$2), redeemed_at=NOW(), reserved_by=NULL, reserved_until=NULL WHERE code=$1 AND status='unused'`, code, redeemedBy)
	if err != nil {
		return false, fmt.Errorf("redeem invite by code: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("redeem invite by code rows: %w", err)
	}
	return affected == 1, nil
}
