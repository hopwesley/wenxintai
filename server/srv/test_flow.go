package srv

import (
	"encoding/json"
	"net/http"

	"github.com/hopwesley/wenxintai/server/dbSrv"
)

type testFlowRequest struct {
	BusinessType string `json:"business_type"`
}

func (req *testFlowRequest) parseObj(r *http.Request) *ApiErr {
	if r.Method != http.MethodPost {
		return ApiMethodInvalid
	}
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

type testFlowResponse struct {
	TestPublicID string         `json:"public_id"`
	BusinessType string         `json:"business_type"`
	Steps        []TestFlowStep `json:"steps"`         // 全部阶段（key + title）
	CurrentStage string         `json:"current_stage"` // 当前阶段的 stage key，比如 "riasec"
	CurrentIndex int            `json:"current_index"` // 当前阶段在 Steps 里的下标（0-based）
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

	uid, cookieErr := s.currentUserFromCookie(r)
	if cookieErr != nil || len(uid) == 0 {
		sLog.Err(cookieErr).Msgf("fail to get user from cookie")
		writeError(w, ApiInternalErr("请先登录微信", cookieErr))
		return
	}

	record, dbErr := dbSrv.Instance().QueryTestInProcess(ctx, uid, req.BusinessType)
	if dbErr != nil {
		sLog.Err(dbErr).Msg("failed find test record")
		writeError(w, ApiInternalErr("查询文件数据库操作失败", dbErr))
		return
	}

	stageFlow := getTestRoutes(req.BusinessType)
	steps := getTestFlowSteps(req.BusinessType)

	var currentStage = StageBasic
	var currentIndex = 0
	var publicID = ""
	if record == nil {
		pid, dbErr := dbSrv.Instance().NewTestRecord(ctx, req.BusinessType, nil, &uid)
		if dbErr != nil {
			sLog.Err(dbErr).Msg("failed create test record")
			writeError(w, ApiInternalErr("没有问卷相关数据库记录", nil))
			return
		}
		publicID = pid
		currentStage, currentIndex = StageBasic, RecordStatusInit
	} else {
		publicID = record.PublicId
		currentStage, currentIndex = parseStatusToRoute(int(record.Status), stageFlow)
	}

	resp := testFlowResponse{
		TestPublicID: publicID,
		BusinessType: req.BusinessType,
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
	businessTyp, dbErr := dbSrv.Instance().UpdateBasicInfo(
		ctx,
		req.PublicId,
		string(req.Grade),
		string(req.Mode),
		req.Hobby,
		RecordStatusInTest,
	)

	if dbErr != nil {
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
