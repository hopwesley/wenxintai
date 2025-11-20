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

type testFlowResponse struct {
	TestPublicID string   `json:"public_id"`
	Routes       []string `json:"routes"`
	NextRoute    string   `json:"next_route"`
	NextRid      int16    `json:"next_route_id"`
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

	record, dbErr := dbSrv.Instance().FindTestRecordByPublicId(ctx, req.TestPublicID)
	if dbErr != nil {
		sLog.Err(dbErr).Msg("failed find test record")
		writeError(w, ApiInternalErr("查询文件数据库操作失败", dbErr))
		return
	}

	routes := getTestRoutesDes(record.BusinessType)
	if routes == nil {
		sLog.Err(err).Msgf("failed build test routes")
		writeError(w, ApiInternalErr("没有测试类型为："+record.BusinessType+"的测试卷", nil))
		return
	}
	resp := testFlowResponse{
		Routes:       routes,
		TestPublicID: req.TestPublicID,
		NextRoute:    calculateNextStep(record, getTestRoutes(record.BusinessType)),
		NextRid:      record.Status,
	}

	sLog.Debug().Strs("routes", routes).Msg("test record found")
	writeJSON(w, http.StatusOK, resp)
}

func calculateNextStep(record *dbSrv.TestRecord, routes []string) string {
	status := int(record.Status)
	if status >= len(routes) {
		return StageReport
	}
	switch {
	case record.Status == RecordStatusInit:
		return StageBasic

	case record.Status >= RecordStatusInTest && record.Status < RecordStatusInReport:
		return routes[status]

	default:
		return StageBasic
	}
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
