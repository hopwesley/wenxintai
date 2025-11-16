package dbSrv

import (
	"context"
	"database/sql"
)

type DbService interface {
	WithTx(ctx context.Context, fn func(tx *sql.Tx) error) error
	Init(cfg any) error
	Shutdown(ctx context.Context) error

	ListHobbies(ctx context.Context) ([]string, error)
	GetInviteByCode(ctx context.Context, code string) (*Invite, error)
	FindRestRecordByUid(ctx context.Context, inviteCode, weChatID string) (*TestRecord, error)
	NewTestRecord(ctx context.Context, testType string, inviteCode *string, weChatId *string) (string, error)
}

type txKey struct{}

func ContextWithTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func TxFromContext(ctx context.Context) (*sql.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(*sql.Tx)
	return tx, ok
}
