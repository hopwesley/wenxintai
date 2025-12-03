package dbSrv

import (
	"context"
	"database/sql"
	"time"

	"github.com/hopwesley/wenxintai/server/ai_api"
)

type DbService interface {
	WithTx(ctx context.Context, fn func(tx *sql.Tx) error) error
	Init(cfg any) error
	Shutdown(ctx context.Context) error

	ListHobbies(ctx context.Context) ([]string, error)
	ListTestPlans(ctx context.Context) ([]TestPlan, error)
	PlanByKey(ctx context.Context, key string) (*TestPlan, error)

	GetInviteByCode(ctx context.Context, code string) (*Invite, error)

	QueryTestInProcess(ctx context.Context, uid, businessType string) (*TestRecord, error)
	QueryUnfinishedTest(ctx context.Context, publicId string) (*TestRecord, error)
	NewTestRecord(ctx context.Context, businessType string, weChatId string) (string, error)
	QueryRecordById(ctx context.Context, publicID string) (*TestRecord, error)

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

	FindWeChatOrderByID(ctx context.Context, id string) (*WeChatOrder, error)
	UpdateWeChatOrderStatus(
		ctx context.Context,
		orderID string,
		tradeState string,
		transactionID *string,
		payerOpenID *string,
		paidAt *time.Time,
		notifyRaw []byte,
	) error
	InsertWeChatOrder(ctx context.Context, d *WeChatOrder) error
	PayByInviteCode(ctx context.Context, publicId string, inviteCode string) error
}
