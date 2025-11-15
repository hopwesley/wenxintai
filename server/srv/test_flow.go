package srv

import (
	"encoding/json"
	"net/http"
	"strings"
)

// 前端发来的请求体
type testFlowRequest struct {
	TestType     string  `json:"test_type"`
	TestRecordID string  `json:"record_id,omitempty"`
	InviteCode   *string `json:"invite_code,omitempty"`
	WechatOpenID *string `json:"wechat_openid,omitempty"`
}

// 单个测试步骤
type testRouteDef struct {
	Router string `json:"router"` // 英文路由名，例如 basic-info / riasec / asc / report
	Desc   string `json:"desc"`   // 中文描述，例如 基本信息 / 兴趣测试 / 能力测试 / 测试报告
}

// 下一步题目从哪里来
type questionSource string

// 下一步路由信息
type nextRouteInfo struct {
	Router string         `json:"router"` // 英文路由名
	Source questionSource `json:"source"` // 'db' or 'ai'
}

// /api/test_flow 响应体
type testFlowResponse struct {
	TestType  string         `json:"test_type"`
	Routes    []testRouteDef `json:"routes"`
	NextRoute *nextRouteInfo `json:"nextRoute,omitempty"`
}

// /api/test_flow: 根据 test_type + 身份信息，返回整套测试流程 + 下一步要进入的路由
func (s *HttpSrv) handleTestFlow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, ApiMethodInvalid)
		return
	}

	var req testFlowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, ApiInvalidReq("invalid request body", err))
		return
	}

	// test_type 必填
	if strings.TrimSpace(req.TestType) == "" {
		writeError(w, ApiInvalidReq("test_type is required", nil))
		return
	}

	// 身份：邀请码或微信至少有一个（后面我们用它去查 tests_record）
	if (req.InviteCode == nil || strings.TrimSpace(*req.InviteCode) == "") &&
		(req.WechatOpenID == nil || strings.TrimSpace(*req.WechatOpenID) == "") {
		writeError(w, ApiInvalidReq("invite_code or wechat_openid is required", nil))
		return
	}

	// 1) 根据 test_type 构建完整的测试流程 routes
	routes := buildTestRoutes(req.TestType)

	// 2) 计算 nextRoute：当前先不查数据库，统一返回 nil
	//    等 tests_record + 各量表运行记录表建好后，在这里接 service 计算实际进度。
	var next *nextRouteInfo

	// 示例：将来可以大致这样写（这里只是结构示意，暂不启用）：
	// next, err := h.svc.ComputeNextRoute(r.Context(), req.TestType, req.InviteCode, req.WechatOpenID)
	// if err != nil { ... }

	resp := testFlowResponse{
		TestType:  req.TestType,
		Routes:    routes,
		NextRoute: next,
	}
	writeJSON(w, http.StatusOK, resp)
}

// 根据测试类型构建完整的测试流程。
// 规则：basic-info 一定在最前，report 一定在最后，中间根据 test_type 不同插入不同的量表。
func buildTestRoutes(testType string) []testRouteDef {
	// 基本信息 & 报告，这两个是固定的
	basic := testRouteDef{Router: "basic-info", Desc: "基本信息"}
	report := testRouteDef{Router: "report", Desc: "测试报告"}

	var middle []testRouteDef

	switch testType {
	case "basic":
		// 基础版：RIASEC + ASC
		middle = []testRouteDef{
			{Router: "riasec", Desc: "RIASEC"},
			{Router: "asc", Desc: "ASC"},
		}
	case "pro":
		// 举例：专业版增加更多量表，后面可以按实际再调整
		middle = []testRouteDef{
			{Router: "riasec", Desc: "RIASEC"},
			{Router: "asc", Desc: "ASC"},
			{Router: "big5", Desc: "OCEAN"},
			{Router: "motivation", Desc: "MOTIVATION"},
		}
	default:
		// 默认：至少给出基础流程，避免前端拿不到任何路由
		middle = []testRouteDef{
			{Router: "riasec", Desc: "兴趣测试"},
			{Router: "asc", Desc: "能力测试"},
		}
	}

	routes := make([]testRouteDef, 0, len(middle)+2)
	routes = append(routes, basic)
	routes = append(routes, middle...)
	routes = append(routes, report)
	return routes
}
