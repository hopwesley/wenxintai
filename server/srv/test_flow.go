package srv

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/hopwesley/wenxintai/server/dbSrv"
)

type TestBasicInfo struct {
	Grade string `json:"grade"`
	Mode  string `json:"mode"`
	Hobby string `json:"hobby,omitempty"`
}

type testFlowRequest struct {
	TestType     string  `json:"test_type"`
	TestPublicID string  `json:"public_id,omitempty"`
	InviteCode   *string `json:"invite_code,omitempty"`
	WechatOpenID *string `json:"wechat_openid,omitempty"`
}

func (req *testFlowRequest) parseObj(r *http.Request) *ApiErr {
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return ApiInvalidReq("invalid request body", err)
	}

	if strings.TrimSpace(req.TestType) == "" {
		return ApiInvalidReq("test_type is required", nil)
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
		writeError(w, err)
		return
	}

	routes := buildTestRoutes(req.TestType)
	if routes == nil {
		writeError(w, NewApiError(http.StatusInternalServerError,
			ErrorCodeInternal, "没有测试类型为："+req.TestType+"的测试卷", nil))
		return
	}
	ctx := r.Context()

	if len(req.TestPublicID) < 4 {
		publicID, dbErr := dbSrv.Instance().NewTestRecord(ctx, req.TestType, req.InviteCode, req.WechatOpenID)
		if dbErr != nil {
			writeError(w, NewApiError(http.StatusInternalServerError,
				ErrorCodeInternal, "创建文件数据库操作失败", dbErr))
			return
		}

		resp := testFlowResponse{
			Routes:       routes,
			TestPublicID: publicID,
		}

		s.log.Debug().Str("public_id", publicID).Msg("no test public id found, init a empty test record")
		writeJSON(w, http.StatusOK, resp)
		return
	}

	var inviteCode, weChatID = req.getUserId()
	record, dbErr := dbSrv.Instance().FindRestRecordByUid(ctx, inviteCode, weChatID)
	if dbErr != nil {
		writeError(w, NewApiError(http.StatusInternalServerError,
			ErrorCodeInternal, "查询文件数据库操作失败", dbErr))
		return
	}

	var nextStep = calculateNextStep(record)
	resp := testFlowResponse{
		Routes:       routes,
		TestPublicID: req.TestPublicID,
		NextRoute:    nextStep,
	}
	s.log.Debug().Str("test_type", req.TestType).Str("invite_code", inviteCode).
		Str("wechat_id", weChatID).Str("next_step", nextStep).Msg("test record found")
	writeJSON(w, http.StatusOK, resp)
}

func calculateNextStep(record *dbSrv.TestRecord) string {
	switch record.Status {
	case RecordStatusInit:
		return StageBasic
	default:
		return StageBasic
	}
}

const (
	RecordStatusInit = 0
)

const (
	StageBasic  = "basic-info"
	StageReport = "report"

	StageRiasec    = "riasec"
	StageRiasecDes = "兴趣测试"

	StageAsc    = "asc"
	StageAscDes = "能力测试"

	StageOcean    = "ocean"
	StageOceanDes = "性格测试"

	StageMotivation    = "motivation"
	StageMotivationDes = "价值观测试"

	TestTypeBasic  = "basic"
	TestTypePro    = "pro"
	TestTypeSchool = "school"
)

func buildTestRoutes(testType string) []testRouteDef {
	basic := testRouteDef{Router: StageBasic, Desc: "基本信息"}
	report := testRouteDef{Router: StageReport, Desc: "测试报告"}

	var middle []testRouteDef

	switch testType {
	case TestTypeBasic:
		middle = []testRouteDef{
			{Router: StageRiasec, Desc: StageRiasecDes},
			{Router: StageAsc, Desc: StageAscDes},
		}
	case TestTypePro:
		middle = []testRouteDef{
			{Router: StageRiasec, Desc: StageRiasecDes},
			{Router: StageAsc, Desc: StageAscDes},
			{Router: StageOcean, Desc: StageOceanDes},
			{Router: StageMotivation, Desc: StageMotivationDes},
		}
	case TestTypeSchool:
		middle = []testRouteDef{
			{Router: StageRiasec, Desc: StageRiasecDes},
			{Router: StageAsc, Desc: StageAscDes},
			{Router: StageOcean, Desc: StageOceanDes},
			{Router: StageMotivation, Desc: StageMotivationDes},
		}

	default:
		return nil
	}

	routes := make([]testRouteDef, 0, len(middle)+2)
	routes = append(routes, basic)
	routes = append(routes, middle...)
	routes = append(routes, report)

	return routes
}
