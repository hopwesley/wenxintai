package srv

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/hopwesley/wenxintai/server/ai_api"
)

type CommonRes struct {
	Ok  bool   `json:"ok"`
	Msg string `json:"msg,omitempty"`
}

var publicIDRegex = regexp.MustCompile(`^[a-f0-9]{32}$`)

func IsValidPublicID(publicID string) bool {
	return publicIDRegex.MatchString(publicID)
}

type BasicInfoReq ai_api.BasicInfo

func (bi *BasicInfoReq) parseObj(r *http.Request) *ApiErr {
	if err := json.NewDecoder(r.Body).Decode(bi); err != nil {
		return ApiInvalidReq("无效的请求体", err)
	}
	if !IsValidPublicID(bi.PublicId) {
		return ApiInvalidReq("无效的问卷编号", nil)
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

var testFlowForBasic = []string{StageBasic, StageRiasec, StageAsc, StageReport}
var testFlowForPro = []string{StageBasic, StageRiasec, StageAsc, StageOcean, StageMotivation, StageReport}
var testFlowForSchool = []string{StageBasic, StageRiasec, StageAsc, StageOcean, StageMotivation, StageReport}

func nextRoute(businessTyp, curStage string) (string, error) {
	var flow []string

	switch businessTyp {
	case TestTypeBasic:
		flow = testFlowForBasic
	case TestTypePro:
		flow = testFlowForPro
	case TestTypeSchool:
		flow = testFlowForSchool
	default:
		return "", fmt.Errorf("unknown business type: %s", businessTyp)
	}

	// 在对应流程里查找当前阶段
	idx := -1
	for i, s := range flow {
		if s == curStage {
			idx = i
			break
		}
	}

	if idx == -1 {
		return "", fmt.Errorf("invalid stage %s for business type %s", curStage, businessTyp)
	}

	// 已经是最后一个阶段（通常是 StageReport）——按你的要求返回错误
	if idx == len(flow)-1 {
		return "", fmt.Errorf("stage %s is last stage for business type %s", curStage, businessTyp)
	}

	// 正常返回下一个阶段
	return flow[idx+1], nil
}
