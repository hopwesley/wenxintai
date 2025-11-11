package service

import (
	"context"
	"database/sql"
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

func (s *InviteService) Verify(ctx context.Context, code string, sessionID *string) (*VerifyInviteResult, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, newError(ErrorCodeBadRequest, "code is required", nil)
	}

	var result *VerifyInviteResult
	err := s.repo.WithTx(ctx, func(tx *sql.Tx) error {
		txCtx := store.ContextWithTx(ctx, tx)
		invite, err := s.repo.GetInviteForUpdate(txCtx, code)
		if err != nil {
			if err == store.ErrNotFound {
				return newError(ErrorCodeNotFound, "invite not found", err)
			}
			return err
		}

		switch invite.Status {
		case "unused":
			// continue
		case "disabled":
			return newError(ErrorCodeInviteDisabled, "invite disabled", nil)
		default:
			return newError(ErrorCodeInviteRedeemed, "invite already redeemed", nil)
		}

		now := s.now()
		var targetSession string
		var expiresAt time.Time

		if invite.ReservedBy != nil && invite.ReservedUntil != nil && invite.ReservedUntil.After(now) {
			if sessionID != nil && *sessionID == *invite.ReservedBy {
				targetSession = *sessionID
				expiresAt = now.Add(s.ttl)
				if err := s.repo.UpdateInviteReservation(txCtx, invite.Code, &targetSession, &expiresAt); err != nil {
					return err
				}
			} else {
				return newError(ErrorCodeInviteReserved, "invite is reserved", nil)
			}
		} else {
			if sessionID != nil && strings.TrimSpace(*sessionID) != "" {
				targetSession = strings.TrimSpace(*sessionID)
			} else {
				targetSession = uuid.NewString()
			}
			expiresAt = now.Add(s.ttl)
			if err := s.repo.UpdateInviteReservation(txCtx, invite.Code, &targetSession, &expiresAt); err != nil {
				return err
			}
		}

		result = &VerifyInviteResult{
			SessionID:     targetSession,
			Status:        "reserved",
			ReservedUntil: expiresAt,
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
