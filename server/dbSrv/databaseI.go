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
	QueryUnfinishedTest(ctx context.Context, publicId string) (*TestRecord, error)
	NewTestRecord(ctx context.Context, businessType string, inviteCode *string, weChatId *string) (string, error)
	UpdateBasicInfo(ctx context.Context, publicId string, grade string, mode string, hobby string, status int) (string, error)
	QueryBasicInfo(ctx context.Context, publicId string) (*ai_api.BasicInfo, error)
	FindQASession(ctx context.Context, testType, publicId string) (*QASession, error)
	SaveQuestion(ctx context.Context, testType, publicId string, questionsJSON []byte) error
	SaveAnswer(ctx context.Context, testType, publicId string, answersJSON []byte, status int) error
	FindQASessionsForReport(ctx context.Context, publicId string) ([]*QASession, error)
	SaveTestReportCore(ctx context.Context, publicId, mode string, commonScoreJSON []byte, modeParamJSON []byte) error
	UpdateTestReportAIContent(ctx context.Context, publicId string, aiContentJSON []byte) error
	FindTestReportByPublicId(ctx context.Context, publicId string) (*TestReport, error)
	FindUserProfileByUid(ctx context.Context, uid string) (*UserProfile, error)
	InsertOrUpdateUserProfileBasic(ctx context.Context, id string, name string, url string) error
	UpdateUserProfileExtra(ctx context.Context, uid, mobile, studentId, schoolName, province, city string) error
	FinalizedTest(ctx context.Context, publicID string, businessType string) error
	QueryTestInfos(ctx context.Context, uid string) ([]*TestItem, error)
}
