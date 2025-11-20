package srv

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/hopwesley/wenxintai/server/ai_api"
)

type CommonRes struct {
	Ok        bool   `json:"ok"`
	Msg       string `json:"msg,omitempty"`
	NextRoute string `json:"next_route,omitempty"`
	NextRid   int    `json:"next_route_index,omitempty"`
}

var publicIDRegex = regexp.MustCompile(`^[a-f0-9]{32}$`)

func IsValidPublicID(publicID string) bool {
	return publicIDRegex.MatchString(publicID)
}

type BasicInfoReq ai_api.BasicInfo

func (bi *BasicInfoReq) parseObj(r *http.Request) *ApiErr {

	if r.Method != http.MethodPost {
		return ApiMethodInvalid
	}
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

	RecordStatusInReport = 100
)

const (
	StageBasic    = "basic-info"
	StageBasicDes = "基础信息"

	StageReport    = "report"
	StageReportDes = "测评报告"

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

var testFlowForBasic = []string{StageBasic, StageRiasec, StageAsc, StageReport}
var testFlowForPro = []string{StageBasic, StageRiasec, StageAsc, StageOcean, StageMotivation, StageReport}
var testFlowForSchool = []string{StageBasic, StageRiasec, StageAsc, StageOcean, StageMotivation, StageReport}

func nextRoute(businessTyp, curStage string) (int, string, error) {
	var flow []string

	switch businessTyp {
	case TestTypeBasic:
		flow = testFlowForBasic
	case TestTypePro:
		flow = testFlowForPro
	case TestTypeSchool:
		flow = testFlowForSchool
	default:
		return -1, "", fmt.Errorf("unknown business type: %s", businessTyp)
	}

	if len(curStage) == 0 {
		return 0, flow[0], nil
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
		return -1, "", fmt.Errorf("invalid stage %s for business type %s", curStage, businessTyp)
	}

	// 已经是最后一个阶段（通常是 StageReport）——按你的要求返回错误
	if idx == len(flow)-1 {
		return -1, "", fmt.Errorf("stage %s is last stage for business type %s", curStage, businessTyp)
	}

	return idx + 1, flow[idx+1], nil
}

func getTestRoutes(testType string) []string {
	switch testType {
	case TestTypeBasic:
		return testFlowForBasic
	case TestTypePro:
		return testFlowForPro
	case TestTypeSchool:
		return testFlowForSchool
	default:
		return nil
	}
}

var testFlowDescForBasic = []string{
	StageBasicDes,
	StageRiasecDes,
	StageAscDes,
	StageReportDes,
}

var testFlowDescForPro = []string{
	StageBasicDes,
	StageRiasecDes,
	StageAscDes,
	StageOceanDes,
	StageMotivationDes,
	StageReportDes,
}

var testFlowDescForSchool = []string{
	StageBasicDes,
	StageRiasecDes,
	StageAscDes,
	StageOceanDes,
	StageMotivationDes,
	StageReportDes,
}

func getTestRoutesDes(testType string) []string {
	switch testType {
	case TestTypeBasic:
		return testFlowDescForBasic
	case TestTypePro:
		return testFlowDescForPro
	case TestTypeSchool:
		return testFlowDescForSchool
	default:
		return nil
	}
}
