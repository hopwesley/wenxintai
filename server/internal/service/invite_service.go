package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/hopwesley/wenxintai/server/internal/store"
)

type InviteService struct {
	repo store.Repo
	ttl  time.Duration
	now  func() time.Time
}

type VerifyInviteResult struct {
	SessionID     string
	Status        string
	ReservedUntil time.Time
}

func NewInviteService(repo store.Repo) *InviteService {
	return &InviteService{
		repo: repo,
		ttl:  15 * time.Minute,
		now:  time.Now,
	}
}

const (
	inviteUnused   int16 = 0
	inviteReserved int16 = 1
	inviteRedeemed int16 = 2
	inviteDisabled int16 = 3
)

func (s *InviteService) Verify(ctx context.Context, code string, sessionID *string) (*VerifyInviteResult, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, newError(ErrorCodeBadRequest, "code is required", nil)
	}

	var result *VerifyInviteResult
	err := s.repo.WithTx(ctx, func(tx *sql.Tx) error {
		txCtx := store.ContextWithTx(ctx, tx)
		inv, err := s.repo.GetInviteForUpdate(txCtx, code)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				return newError(ErrorCodeNotFound, "invite not found", err)
			}
			return err
		}

		now := s.now()

		// 自然过期（如果你希望 unused 也能自然过期，这里可判断 inv.ExpiresAt != nil && inv.ExpiresAt.Before(now)）
		if inv.Status == inviteDisabled {
			return newError(ErrorCodeInviteDisabled, "invite disabled", nil)
		}
		if inv.Status == inviteRedeemed {
			return newError(ErrorCodeInviteRedeemed, "invite already redeemed", nil)
		}

		// 选择/生成 sessionID
		var sess string
		if sessionID != nil && strings.TrimSpace(*sessionID) != "" {
			sess = strings.TrimSpace(*sessionID)
		} else {
			sess = uuid.NewString()
		}
		until := now.Add(s.ttl)

		switch inv.Status {
		case inviteUnused:
			// 直接占用
			if err := s.repo.UpdateInviteReservation(txCtx, inv.Code, sess, until); err != nil {
				if err == store.ErrConflict {
					return newError(ErrorCodeInviteReserved, "invite is reserved", nil)
				}
				return err
			}

		case inviteReserved:
			// 仍被占用但是否可续期/抢占？
			// 1) 自己占用且未过期 -> 续期
			if inv.UsedBy != nil && *inv.UsedBy == sess && inv.ExpiresAt != nil && inv.ExpiresAt.After(now) {
				if err := s.repo.UpdateInviteReservation(txCtx, inv.Code, sess, until); err != nil {
					if err == store.ErrConflict {
						return newError(ErrorCodeInviteReserved, "invite is reserved", nil)
					}
					return err
				}
			} else {
				// 2) 他人占用已过期 -> 抢占
				if inv.ExpiresAt != nil && inv.ExpiresAt.Before(now) {
					if err := s.repo.UpdateInviteReservation(txCtx, inv.Code, sess, until); err != nil {
						if err == store.ErrConflict {
							return newError(ErrorCodeInviteReserved, "invite is reserved", nil)
						}
						return err
					}
				} else {
					// 3) 他人占用未过期 -> 返回占用中
					return newError(ErrorCodeInviteReserved, "invite is reserved", nil)
				}
			}

		default:
			// 其他非法状态
			return newError(ErrorCodeBadRequest, "invalid invite status", nil)
		}

		result = &VerifyInviteResult{
			SessionID:     sess,
			Status:        "reserved",
			ReservedUntil: until,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

type RedeemInviteResult struct {
	Status string
}

func (s *InviteService) Redeem(ctx context.Context, sessionID string) (*RedeemInviteResult, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, newError(ErrorCodeBadRequest, "session_id is required", nil)
	}

	var ok bool
	err := s.repo.WithTx(ctx, func(tx *sql.Tx) error {
		txCtx := store.ContextWithTx(ctx, tx)
		var err error
		ok, err = s.repo.RedeemInviteBySession(txCtx, sessionID, sessionID)
		if err != nil {
			return err
		}
		if !ok {
			return newError(ErrorCodeInviteReserved, "invite not reserved by session", nil)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, newError(ErrorCodeInviteReserved, "invite not reserved by session", nil)
	}
	return &RedeemInviteResult{Status: "redeemed"}, nil
}
