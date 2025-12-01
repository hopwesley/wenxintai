package srv

import (
	"encoding/json"
	"net/http"

	"github.com/hopwesley/wenxintai/server/dbSrv"
)

type testFlowRequest struct {
	TestPublicID string `json:"public_id"`
}

func (req *testFlowRequest) parseObj(r *http.Request) *ApiErr {
	if r.Method != http.MethodPost {
		return ApiMethodInvalid
	}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return ApiInvalidReq("invalid request body", err)
	}
	if !IsValidPublicID(req.TestPublicID) {
		return ApiInvalidReq("无效的问卷编号", nil)
	}
	return nil
}

type TestFlowStep struct {
	Stage string `json:"stage"` // 路由用的 key，例如 "basic-info" / "riasec" / ...
	Title string `json:"title"` // 展示给用户的标题，例如 "基础信息" / "兴趣测试"
}

type testFlowResponse struct {
	TestPublicID string `json:"public_id"`
	BusinessType string `json:"business_type"`

	Steps        []TestFlowStep `json:"steps"`         // 全部阶段（key + title）
	CurrentStage string         `json:"current_stage"` // 当前阶段的 stage key，比如 "riasec"
	CurrentIndex int            `json:"current_index"` // 当前阶段在 Steps 里的下标（0-based）

	Routes    []string `json:"routes"`
	NextRoute string   `json:"next_route"`
	NextRid   int16    `json:"next_route_id"`
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

	sLog := s.log.With().Str("public_id", req.TestPublicID).Logger()
	s.log.Info().Msg("query test flow")

	record, dbErr := dbSrv.Instance().QueryUnfinishedTest(ctx, req.TestPublicID)
	if dbErr != nil {
		sLog.Err(dbErr).Msg("failed find test record")
		writeError(w, ApiInternalErr("查询文件数据库操作失败", dbErr))
		return
	}

	if record == nil {
		sLog.Err(dbErr).Msg("no such test record")
		writeError(w, ApiInternalErr("没有问卷相关数据库记录", nil))
		return
	}

	// 1. 取出完整流程的 stage 列表 & 描述列表
	stageFlow := getTestRoutes(record.BusinessType)    // e.g. ["basic-info","riasec",...]
	titleFlow := getTestRoutesDes(record.BusinessType) // e.g. ["基础信息","兴趣测试",...]

	if stageFlow == nil || titleFlow == nil {
		sLog.Error().Msgf("failed build test routes for business_type=%s", record.BusinessType)
		writeError(w, ApiInternalErr("没有测试类型为："+record.BusinessType+"的测试卷", nil))
		return
	}

	// 2. 当前应该在哪个阶段（基于 status 计算）
	currentStage := parseStatusToRoute(record, stageFlow)

	// 3. 在流程中找到当前阶段的 index（0-based）
	currentIndex := 0
	for i, sName := range stageFlow {
		if sName == currentStage {
			currentIndex = i
			break
		}
	}

	// 4. 组装 steps（stage + title）
	steps := getTestFlowSteps(record.BusinessType)

	resp := testFlowResponse{
		TestPublicID: req.TestPublicID,
		BusinessType: record.BusinessType,

		Steps:        steps,
		CurrentStage: currentStage,
		CurrentIndex: currentIndex,

		// 兼容旧字段
		Routes:    titleFlow,
		NextRoute: currentStage,
		NextRid:   int16(currentIndex),
	}

	sLog.Debug().
		Str("business_type", record.BusinessType).
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
