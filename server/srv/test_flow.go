package srv

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/hopwesley/wenxintai/server/dbSrv"
)

type testFlowRequest struct {
	BusinessType string `json:"business_type"`
}

func (req *testFlowRequest) parseObj(r *http.Request) *ApiErr {
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return ApiInvalidReq("invalid request body", err)
	}
	if !isValidBusinessType(req.BusinessType) {
		return ApiInvalidReq("无效的测试类型", nil)
	}
	return nil
}

type TestFlowStep struct {
	Stage string `json:"stage"` // 路由用的 key，例如 "basic-info" / "riasec" / ...
	Title string `json:"title"` // 展示给用户的标题，例如 "基础信息" / "兴趣测试"
}

type TestRecordDTO struct {
	PublicId     string     `json:"public_id"`
	BusinessType string     `json:"business_type"`
	PayOrderId   string     `json:"pay_order_id,omitempty"`
	WeChatID     string     `json:"wechat_id,omitempty"`
	Grade        string     `json:"grade,omitempty"`
	Mode         string     `json:"mode,omitempty"`
	Hobby        string     `json:"hobby,omitempty"`
	Status       int16      `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}

type testFlowResponse struct {
	Record       TestRecordDTO  `json:"record"`
	Steps        []TestFlowStep `json:"steps"`         // 全部阶段（key + title）
	CurrentStage string         `json:"current_stage"` // 当前阶段的 stage key，比如 "riasec"
	CurrentIndex int            `json:"current_index"` // 当前阶段在 Steps 里的下标（0-based）
}

func toRecordDTO(rec dbSrv.TestRecord) TestRecordDTO {
	var completed *time.Time
	if rec.CompletedAt.Valid {
		completed = &rec.CompletedAt.Time
	}
	return TestRecordDTO{
		PublicId:     rec.PublicId,
		BusinessType: rec.BusinessType,
		PayOrderId:   nullToString(rec.PayOrderId),
		WeChatID:     nullToString(rec.WeChatID),
		Grade:        nullToString(rec.Grade),
		Mode:         nullToString(rec.Mode),
		Hobby:        nullToString(rec.Hobby),
		Status:       rec.Status,
		CreatedAt:    rec.CreatedAt,
		CompletedAt:  completed,
	}
}

func (s *HttpSrv) handleTestFlow(w http.ResponseWriter, r *http.Request) {

	var req testFlowRequest
	err := req.parseObj(r)
	if err != nil {
		s.log.Err(err).Msgf("invalid test flow request")
		writeError(w, err)
		return
	}
	ctx := r.Context()

	sLog := s.log.With().Str("business_type", req.BusinessType).Logger()
	sLog.Info().Msg("start test flow")

	uid := userIDFromContext(ctx)

	record, dbErr := dbSrv.Instance().QueryTestInProcess(ctx, uid, req.BusinessType)
	if dbErr != nil {
		sLog.Err(dbErr).Msg("failed find test record")
		writeError(w, ApiInvalidNoTestRecord(dbErr))
		return
	}

	stageFlow := getTestRoutes(req.BusinessType)
	steps := getTestFlowSteps(req.BusinessType)

	var currentStage = StageBasic
	var currentIndex = 0
	var dto TestRecordDTO
	if record == nil {
		pid, dbErr := dbSrv.Instance().NewTestRecord(ctx, req.BusinessType, uid)
		if dbErr != nil {
			sLog.Err(dbErr).Msg("failed create test record")
			writeError(w, ApiInternalErr("没有问卷相关数据库记录", nil))
			return
		}
		dto.PublicId = pid
		dto.BusinessType = req.BusinessType
		currentStage, currentIndex = StageBasic, RecordStatusInit
	} else {
		currentStage, currentIndex = parseStatusToRoute(int(record.Status), stageFlow)
		dto = toRecordDTO(*record)
	}

	resp := testFlowResponse{
		Record:       dto,
		Steps:        steps,
		CurrentStage: currentStage,
		CurrentIndex: currentIndex,
	}

	sLog.Debug().
		Str("current_stage", currentStage).
		Int("current_index", currentIndex).
		Msg("test record found")

	writeJSON(w, http.StatusOK, resp)
}

func (s *HttpSrv) updateBasicInfo(w http.ResponseWriter, r *http.Request) {

	var req BasicInfoReq
	err := req.parseObj(r)
	if err != nil {
		writeError(w, err)
		return
	}

	slog := s.log.With().Str("public_id", req.PublicId).Logger()
	slog.Info().Msg("prepare to update basic info")

	ctx := r.Context()
	uid := userIDFromContext(ctx)
	businessTyp, dbErr := dbSrv.Instance().UpdateBasicInfo(
		ctx,
		req.PublicId,
		uid,
		string(req.Grade),
		string(req.Mode),
		req.Hobby,
		RecordStatusInTest,
	)

	if dbErr != nil || len(businessTyp) == 0 {
		slog.Err(dbErr).Msg("更新基本信息失败")
		writeError(w, NewApiError(http.StatusInternalServerError, "db_update_failed", "更新基本信息失败", err))
		return
	}

	nri, nextR, rErr := nextRoute(businessTyp, StageBasic)
	if rErr != nil {
		slog.Err(rErr).Msg("获取下一级路由失败")
		writeError(w, NewApiError(http.StatusInternalServerError, "db_update_failed", "未找到一下", err))
		return
	}

	writeJSON(w, http.StatusOK, &CommonRes{
		Ok:        true,
		Msg:       "更新基本信息成功",
		NextRoute: nextR,
		NextRid:   nri,
	})

	slog.Info().Str("next-route", nextR).Int("next-route-index", nri).Msg("update basic info success")
}
