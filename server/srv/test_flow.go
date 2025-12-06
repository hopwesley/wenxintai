package srv

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/hopwesley/wenxintai/server/ai_api"
	"github.com/hopwesley/wenxintai/server/dbSrv"
)

type testFlowRequest struct {
	BusinessType string `json:"business_type"`
	PublicId     string `json:"public_id,omitempty"`
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

type TestRecordDTO struct {
	PublicId     string     `json:"public_id"`
	BusinessType string     `json:"business_type"`
	PayOrderId   string     `json:"pay_order_id,omitempty"`
	WeChatID     string     `json:"wechat_id,omitempty"`
	Grade        string     `json:"grade,omitempty"`
	Mode         string     `json:"mode,omitempty"`
	Hobby        string     `json:"hobby,omitempty"`
	CurStage     int16      `json:"cur_stage"`
	CreatedAt    time.Time  `json:"created_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}

type testFlowResponse struct {
	Record       TestRecordDTO  `json:"record"`
	Steps        []TestFlowStep `json:"steps"`         // 全部阶段（key + title）
	CurrentStage string         `json:"current_stage"` // 当前阶段的 stage key，比如 "riasec"
	CurrentIndex int16          `json:"current_index"` // 当前阶段在 Steps 里的下标（0-based）
}

func toRecordDTO(rec dbSrv.TestRecord) TestRecordDTO {
	var completed *time.Time
	return TestRecordDTO{
		PublicId:     rec.PublicId,
		BusinessType: rec.BusinessType,
		PayOrderId:   nullToString(rec.PayOrderId),
		WeChatID:     nullToString(rec.WeChatID),
		Grade:        nullToString(rec.Grade),
		Mode:         nullToString(rec.Mode),
		Hobby:        nullToString(rec.Hobby),
		CurStage:     rec.CurStage,
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

	sLog := s.log.With().Str("business_type", req.BusinessType).Str("public_id", req.PublicId).Logger()
	sLog.Info().Msg("start test flow")

	uid := userIDFromContext(ctx)

	var (
		dto    TestRecordDTO
		record *dbSrv.TestRecord
		dbErr  error = nil
	)

	if len(req.PublicId) > 0 {
		if !IsValidPublicID(req.PublicId) {
			sLog.Error().Msg("invalid public id")
			writeError(w, ApiInvalidReq("问卷变化错误", nil))
			return
		}
		record, dbErr = dbSrv.Instance().QueryTestRecord(ctx, req.PublicId, uid)
		if record == nil || dbErr != nil {
			sLog.Err(dbErr).Msg("query data by public id failed")
			writeError(w, ApiInvalidNoTestRecord(dbErr))
			return
		}

	} else {
		record, dbErr = dbSrv.Instance().QueryUnfinishedTestOfUser(ctx, uid, req.BusinessType)
		if dbErr != nil {
			sLog.Err(dbErr).Msg("failed find test record by user and type")
			writeError(w, ApiInternalErr("访问数据库失败", dbErr))
			return
		}
	}

	steps := getTestFlowSteps(req.BusinessType)

	var currentStage = StageBasic
	var currentIndex int16 = 0
	if record == nil {
		dto.BusinessType = req.BusinessType
		currentStage, currentIndex = StageBasic, RecordStatusInit
		sLog.Info().Msg("need new test record")
	} else {
		dto = toRecordDTO(*record)
		currentIndex = dto.CurStage
		currentStage = getStageIndex(steps, currentIndex)
		sLog.Info().Msg("find test record in database")
	}

	resp := testFlowResponse{
		Record:       dto,
		Steps:        steps,
		CurrentStage: string(currentStage),
		CurrentIndex: currentIndex,
	}

	sLog.Debug().
		Str("current_stage", string(currentStage)).
		Int16("current_index", currentIndex).
		Msg("test record proceed success in test flow")

	writeJSON(w, http.StatusOK, resp)
}

func (s *HttpSrv) updateBasicInfo(w http.ResponseWriter, r *http.Request) {

	var req BasicInfoReq
	err := req.parseObj(r)
	if err != nil {
		writeError(w, err)
		return
	}
	ctx := r.Context()
	uid := userIDFromContext(ctx)

	slog := s.log.With().Str("public_id", req.PublicId).Str("business_type", req.BusinessType).Str("uid", uid).Logger()
	slog.Info().Msg("prepare to update basic info")
	aiBasic := &ai_api.BasicInfo{
		Grade: req.Grade,
		Mode:  req.Mode,
		Hobby: req.Hobby,
	}

	var newPublicId = ""
	var dbErr error = nil
	if len(req.PublicId) > 0 {
		_, dbErr = dbSrv.Instance().UpdateRecordBasicInfo(ctx, req.PublicId, uid, aiBasic)
	} else {
		newPublicId, dbErr = dbSrv.Instance().NewTestRecord(ctx, req.BusinessType, uid, aiBasic)
	}

	if dbErr != nil {
		slog.Err(dbErr).Msg("更新基本信息失败")
		writeError(w, NewApiError(http.StatusInternalServerError, "db_update_failed", "更新基本信息失败", err))
		return
	}

	nextStage, nextStageIdx, rErr := nextRoute(req.BusinessType, StageBasic)
	if rErr != nil {
		slog.Err(rErr).Msg("获取下一级路由失败")
		writeError(w, ApiInvalidTestSequence(err))
		return
	}

	writeJSON(w, http.StatusOK, &CommonRes{
		Ok:          true,
		NewPublicID: newPublicId,
		Msg:         "更新基本信息成功",
		NextRoute:   string(nextStage),
		NextRid:     nextStageIdx,
	})

	slog.Info().Str("next-route", string(nextStage)).Int("next-route-index", nextStageIdx).Msg("update basic info success")
}
