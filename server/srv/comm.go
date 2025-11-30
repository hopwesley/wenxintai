package srv

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/hopwesley/wenxintai/server/ai_api"
	"github.com/hopwesley/wenxintai/server/dbSrv"
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
	TestTypeAdv    = "adv"
	TestTypeSchool = "school"
)

var testFlowForBasic = []string{StageBasic, StageRiasec, StageAsc, StageReport}
var testFlowForPro = []string{StageBasic, StageRiasec, StageAsc, StageReport}
var testFlowForAdv = []string{StageBasic, StageRiasec, StageAsc, StageOcean, StageMotivation, StageReport}
var testFlowForSchool = []string{StageBasic, StageRiasec, StageAsc, StageOcean, StageMotivation, StageReport}

func nextRoute(businessTyp, curStage string) (int, string, error) {
	var flow []string

	switch businessTyp {
	case TestTypeBasic:
		flow = testFlowForBasic
	case TestTypePro:
		flow = testFlowForPro
	case TestTypeAdv:
		flow = testFlowForAdv
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
	case TestTypeAdv:
		return testFlowForAdv
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
	StageReportDes,
}

var testFlowDescForAdv = []string{
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
	case TestTypeAdv:
		return testFlowDescForAdv
	case TestTypeSchool:
		return testFlowDescForSchool
	default:
		return nil
	}
}

func parseStatusToRoute(record *dbSrv.TestRecord, routes []string) string {
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

func (s *HttpSrv) checkTestSequence(ctx context.Context, publicID, testType string) (*dbSrv.TestRecord, error) {
	record, dbErr := dbSrv.Instance().FindTestRecordByPublicId(ctx, publicID)
	if dbErr != nil {
		return nil, dbErr
	}

	flow := getTestRoutes(record.BusinessType)
	if len(flow) == 0 {
		return nil, fmt.Errorf("no test flow configured for business type %s", record.BusinessType)
	}

	// 1. 根据当前记录状态解析出“当前阶段”
	currentStage := parseStatusToRoute(record, flow)

	// 2. 计算当前阶段在流程中的下标
	currentIdx := -1
	for i, stage := range flow {
		if stage == currentStage {
			currentIdx = i
			break
		}
	}
	if currentIdx == -1 {
		return nil, fmt.Errorf("invalid current stage %s for business type %s", currentStage, record.BusinessType)
	}

	// 3. 计算请求阶段在流程中的下标
	reqIdx := -1
	for i, stage := range flow {
		if stage == testType {
			reqIdx = i
			break
		}
	}
	if reqIdx == -1 {
		return nil, fmt.Errorf("invalid requested stage %s for business type %s", testType, record.BusinessType)
	}

	// 4. 只允许访问当前阶段及之前的阶段，禁止越级访问未来阶段
	if reqIdx > currentIdx {
		return nil, fmt.Errorf(
			"mismatched route, need stage index <= %d (%s) but got %d (%s)",
			currentIdx, currentStage, reqIdx, testType,
		)
	}

	return record, nil
}

func getTestFlowSteps(businessType string) []TestFlowStep {
	var stages []string
	var titles []string

	switch businessType {
	case TestTypeBasic:
		stages = testFlowForBasic
		titles = testFlowDescForBasic
	case TestTypePro:
		stages = testFlowForPro
		titles = testFlowDescForPro
	case TestTypeAdv:
		stages = testFlowForAdv
		titles = testFlowDescForAdv
	case TestTypeSchool:
		stages = testFlowForSchool
		titles = testFlowDescForSchool
	default:
		return nil
	}

	if len(stages) != len(titles) {
		// 理论上不会发生，如果发生了说明上面的常量维护出了问题
		// 为了安全起见，取两者较短长度
		n := len(stages)
		if len(titles) < n {
			n = len(titles)
		}
		steps := make([]TestFlowStep, 0, n)
		for i := 0; i < n; i++ {
			steps = append(steps, TestFlowStep{
				Stage: stages[i],
				Title: titles[i],
			})
		}
		return steps
	}

	steps := make([]TestFlowStep, 0, len(stages))
	for i := range stages {
		steps = append(steps, TestFlowStep{
			Stage: stages[i],
			Title: titles[i],
		})
	}
	return steps
}

func parseAITestTyp(testTyp, businessTyp string) ai_api.TestTyp {
	switch businessTyp {
	case TestTypeBasic:
		switch testTyp {
		case StageRiasec:
			return ai_api.TypRIASEC
		case StageAsc:
			return ai_api.TypASC
		}
	case TestTypePro:
		switch testTyp {
		case StageRiasec:
			return ai_api.TypRIASEC
		case StageAsc:
			return ai_api.TypASC
		}

	case TestTypeAdv:
		switch testTyp {
		case StageRiasec:
			return ai_api.TypRIASEC
		case StageOcean:
			return ai_api.TypOCEAN
		case StageAsc:
			return ai_api.TypASC
		}

	case TestTypeSchool:
		switch testTyp {
		case StageRiasec:
			return ai_api.TypRIASEC
		case StageOcean:
			return ai_api.TypOCEAN
		case StageAsc:
			return ai_api.TypASC
		}
	}

	return ai_api.TypUnknown
}
