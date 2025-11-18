package srv

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/hopwesley/wenxintai/server/dbSrv"
)

type testFlowRequest struct {
	BusinessType string  `json:"business_type"`
	TestPublicID string  `json:"public_id,omitempty"`
	InviteCode   *string `json:"invite_code,omitempty"`
	WechatOpenID *string `json:"wechat_openid,omitempty"`
}

func (req *testFlowRequest) parseObj(r *http.Request) *ApiErr {
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return ApiInvalidReq("invalid request body", err)
	}

	if strings.TrimSpace(req.BusinessType) == "" {
		return ApiInvalidReq("business_type is required", nil)
	}

	if (req.InviteCode == nil || strings.TrimSpace(*req.InviteCode) == "") &&
		(req.WechatOpenID == nil || strings.TrimSpace(*req.WechatOpenID) == "") {
		return ApiInvalidReq("请先微信登录或者使用邀请码开始测试", nil)
	}

	return nil
}

func (req *testFlowRequest) isInviteCodeUsage() bool {
	return req.InviteCode != nil &&
		len(strings.TrimSpace(*req.InviteCode)) > 2
}

func (req *testFlowRequest) isWeChatUsage() bool {
	return req.WechatOpenID != nil &&
		len(strings.TrimSpace(*req.WechatOpenID)) > 2
}

func (req *testFlowRequest) getUserId() (string, string) {
	var inviteCode, weChatID = "", ""
	if req.isInviteCodeUsage() {
		inviteCode = *req.InviteCode
	}
	if req.isWeChatUsage() {
		weChatID = *req.WechatOpenID
	}

	return inviteCode, weChatID
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
	if r.Method != http.MethodPost {
		writeError(w, ApiMethodInvalid)
		return
	}
	var req testFlowRequest
	err := req.parseObj(r)
	if err != nil {
		s.log.Err(err).Msgf("invalid test flow request")
		writeError(w, err)
		return
	}

	routes := buildTestRoutes(req.BusinessType)
	if routes == nil {
		s.log.Err(err).Msgf("failed build test routes for request:%v", req)
		writeError(w, NewApiError(http.StatusInternalServerError,
			ErrorCodeInternal, "没有测试类型为："+req.BusinessType+"的测试卷", nil))
		return
	}

	ctx := r.Context()

	if len(req.TestPublicID) < 4 {
		publicID, dbErr := dbSrv.Instance().NewTestRecord(ctx, req.BusinessType, req.InviteCode, req.WechatOpenID)
		if dbErr != nil {
			s.log.Err(dbErr).Msgf("create new test record failed, request:%v", req)
			writeError(w, NewApiError(http.StatusInternalServerError,
				ErrorCodeInternal, "创建文件数据库操作失败", dbErr))
			return
		}

		resp := testFlowResponse{
			Routes:       routes,
			TestPublicID: publicID,
		}

		s.log.Info().Str("public_id", publicID).Msg("no test public id found, init a empty test record")
		writeJSON(w, http.StatusOK, resp)
		return
	}

	var inviteCode, weChatID = req.getUserId()
	record, dbErr := dbSrv.Instance().FindTestRecordByUid(ctx, inviteCode, weChatID)
	if dbErr != nil {
		s.log.Err(dbErr).Str("invite_code", inviteCode).Str("wechat_id", weChatID).Msg("failed find test record")
		writeError(w, NewApiError(http.StatusInternalServerError,
			ErrorCodeInternal, "查询文件数据库操作失败", dbErr))
		return
	}

	var nextStep = calculateNextStep(record, routes)
	resp := testFlowResponse{
		Routes:       routes,
		TestPublicID: req.TestPublicID,
		NextRoute:    nextStep,
	}

	s.log.Debug().Str("business_type", req.BusinessType).Str("invite_code", inviteCode).
		Str("wechat_id", weChatID).Str("next_step", nextStep).Msg("test record found")
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
