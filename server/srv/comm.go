package srv

import (
	"encoding/json"
	"net/http"

	"github.com/hopwesley/wenxintai/server/ai_api"
)

type CommonRes struct {
	Ok  bool   `json:"ok"`
	Msg string `json:"msg,omitempty"`
}

type BasicInfoReq ai_api.BasicInfo

func (bi *BasicInfoReq) parseObj(r *http.Request) *ApiErr {
	if err := json.NewDecoder(r.Body).Decode(bi); err != nil {
		return ApiInvalidReq("无效的请求体", err)
	}

	if len(bi.PublicId) <= 4 {
		return ApiInvalidReq("缺少有效的测试 ID", nil)
	}
	if !bi.Grade.IsValid() {
		return ApiInvalidReq("年级不合法，只能是：初二、初三、高一", nil)
	}
	if !bi.Mode.IsValid() {
		return ApiInvalidReq("模式不合法，只能是：Mode33 或 Mode312", nil)
	}

	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}

func writeError(w http.ResponseWriter, err *ApiErr) {
	writeJSON(w, err.status, err)
}

const (
	RecordStatusInit   = 0
	RecordStatusInTest = 1
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
