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

type testRouteDef struct {
	Router string `json:"router"` // 英文路由名，例如 basic-info / riasec / asc / report
	Desc   string `json:"desc"`   // 中文描述，例如 基本信息 / 兴趣测试 / 能力测试 / 测试报告
}

type testFlowResponse struct {
	TestPublicID string         `json:"public_id"`
	Routes       []testRouteDef `json:"routes"`
	NextRoute    string         `json:"nextRoute,omitempty"`
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

	record, dbErr := dbSrv.Instance().FindTestRecordByPublicId(ctx, req.TestPublicID)
	if dbErr != nil {
		sLog.Err(dbErr).Msg("failed find test record")
		writeError(w, ApiInternalErr("查询文件数据库操作失败", dbErr))
		return
	}

	routes := buildTestRoutes(record.BusinessType)
	if routes == nil {
		sLog.Err(err).Msgf("failed build test routes")
		writeError(w, ApiInternalErr("没有测试类型为："+record.BusinessType+"的测试卷", nil))
		return
	}

	var nextStep = calculateNextStep(record, routes)
	resp := testFlowResponse{
		Routes:       routes,
		TestPublicID: req.TestPublicID,
		NextRoute:    nextStep,
	}

	sLog.Debug().Str("next_step", nextStep).Msg("test record found")
	writeJSON(w, http.StatusOK, resp)
}

func calculateNextStep(record *dbSrv.TestRecord, routes []testRouteDef) string {
	if int(record.Status) >= len(routes) {
		return StageReport
	}

	switch record.Status {
	case RecordStatusInit:
		return StageBasic
	case RecordStatusInTest:
		return routes[record.Status].Router

	default:
		return StageBasic
	}
}

func (s *HttpSrv) updateBasicInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, ApiMethodInvalid)
		return
	}

	var req BasicInfoReq
	err := req.parseObj(r)
	if err != nil {
		writeError(w, err)
		return
	}

	ctx := r.Context()
	if err := dbSrv.Instance().UpdateBasicInfo(
		ctx,
		req.PublicId,
		string(req.Grade),
		string(req.Mode),
		req.Hobby,
		RecordStatusInTest,
	); err != nil {
		writeError(w, NewApiError(http.StatusInternalServerError, "db_update_failed", "更新基本信息失败", err))
		return
	}

	writeJSON(w, http.StatusOK, &CommonRes{
		Ok:  true,
		Msg: "更新基本信息成功",
	})
}
