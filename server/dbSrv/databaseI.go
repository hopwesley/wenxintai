package dbSrv

import (
	"context"
	"database/sql"

	"github.com/hopwesley/wenxintai/server/ai_api"
)

type DbService interface {
	WithTx(ctx context.Context, fn func(tx *sql.Tx) error) error
	Init(cfg any) error
	Shutdown(ctx context.Context) error

	ListHobbies(ctx context.Context) ([]string, error)
	GetInviteByCode(ctx context.Context, code string) (*Invite, error)
	FindTestRecordByPublicId(ctx context.Context, publicId string) (*TestRecord, error)
	FindTestRecordByUid(ctx context.Context, inviteCode, weChatID string) (*TestRecord, error)
	NewTestRecord(ctx context.Context, businessType string, inviteCode *string, weChatId *string) (string, error)
	UpdateBasicInfo(ctx context.Context, publicId string, grade string, mode string, hobby string, status int) error
	QueryBasicInfo(ctx context.Context, publicId string) (*ai_api.BasicInfo, error)
	FindQASession(ctx context.Context, businessType, testType, publicId string) (*QASession, error)
	SaveQASession(ctx context.Context, businessType, testType, publicId string, questionsJSON []byte) error
}
