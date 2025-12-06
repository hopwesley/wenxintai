package dbSrv

import (
	"context"
	"database/sql"
	"strings"
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

	NewTestRecord(ctx context.Context, bType, weChatId string, bi *ai_api.BasicInfo) (string, error)
	QueryTestRecord(ctx context.Context, pid, uid string) (*TestRecord, error)
	QueryUnfinishedTestOfUser(ctx context.Context, uid, bType string) (*TestRecord, error)
	UpdateRecordBasicInfo(ctx context.Context, publicID, uid string, bi *ai_api.BasicInfo) (string, error)
	QueryRecordBasicInfo(ctx context.Context, publicId string) (*ai_api.BasicInfo, error)

	FindQASession(ctx context.Context, testType, publicId string) (*QASession, error)
	SaveQuestion(ctx context.Context, testType, publicId string, questionsJSON []byte) error
	SaveAnswer(ctx context.Context, testType, publicId, uid string, answersJSON []byte, status int) error
	FindQASessionsForReport(ctx context.Context, publicId string) ([]*QASession, error)

	SaveReportCore(ctx context.Context, publicId, mode string, commonScoreJSON []byte, modeParamJSON []byte) error
	UpdateReportAIContent(ctx context.Context, publicId string, aiContentJSON []byte) error
	QueryReportByPublicId(ctx context.Context, publicId string) (*TestReport, error)

	QueryUserProfileUid(ctx context.Context, uid string) (*UserProfile, error)
	InsertOrUpdateWeChatInfo(ctx context.Context, id string, name string, url string) error
	UpdateUserProfileExtra(ctx context.Context, uid string, extra UsrProfileExtra) error
	QueryTestInfos(ctx context.Context, uid string) ([]*TestItem, error)

	QueryWeChatOrderByOrderID(ctx context.Context, oid string) (*WeChatOrder, error)
	QueryUnfinishedOrder(ctx context.Context, pid string, timeout time.Time) (*WeChatOrder, error)
	UpdateWeChatOrderStatus(
		ctx context.Context,
		orderID string,
		tradeState int16,
		transactionID string,
		payerOpenID string,
		paidAt time.Time,
		notifyRaw []byte,
	) error
	InsertWeChatOrder(ctx context.Context, d *WeChatOrder) error
	PayByInviteCode(ctx context.Context, publicId string, inviteCode string) error
}

func maskMobile(m string) string {
	m = strings.TrimSpace(m)
	if len(m) < 7 {
		// 太短的就不处理，直接返回
		return m
	}
	// 默认认为是类似 11 位手机号：保留前三位 + **** + 剩余
	start := 3
	end := 7
	if len(m) < end {
		end = len(m)
	}
	return m[:start] + "****" + m[end:]
}
