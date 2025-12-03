package srv

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

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

	BusinessTypeBasic  = "basic"
	BusinessTypePro    = "pro"
	BusinessTypeAdv    = "adv"
	BusinessTypeSchool = "school"
)

var testFlowForBasic = []string{StageBasic, StageRiasec, StageAsc, StageReport}
var testFlowForPro = []string{StageBasic, StageRiasec, StageAsc, StageReport}
var testFlowForAdv = []string{StageBasic, StageRiasec, StageAsc, StageOcean, StageMotivation, StageReport}
var testFlowForSchool = []string{StageBasic, StageRiasec, StageAsc, StageOcean, StageMotivation, StageReport}

func isValidBusinessType(businessTyp string) bool {
	switch businessTyp {
	case BusinessTypeBasic, BusinessTypePro, BusinessTypeAdv, BusinessTypeSchool:
		return true
	default:
		return false
	}
}

func nextRoute(businessTyp, curStage string) (int, string, error) {
	var flow []string

	switch businessTyp {
	case BusinessTypeBasic:
		flow = testFlowForBasic
	case BusinessTypePro:
		flow = testFlowForPro
	case BusinessTypeAdv:
		flow = testFlowForAdv
	case BusinessTypeSchool:
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
	case BusinessTypeBasic:
		return testFlowForBasic
	case BusinessTypePro:
		return testFlowForPro
	case BusinessTypeAdv:
		return testFlowForAdv
	case BusinessTypeSchool:
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
	case BusinessTypeBasic:
		return testFlowDescForBasic
	case BusinessTypePro:
		return testFlowDescForPro
	case BusinessTypeAdv:
		return testFlowDescForAdv
	case BusinessTypeSchool:
		return testFlowDescForSchool
	default:
		return nil
	}
}

func parseStatusToRoute(status int, routes []string) (string, int) {
	if status >= len(routes) {
		return StageReport, len(routes) - 1
	}
	switch {
	case status == RecordStatusInit:
		return StageBasic, RecordStatusInit
	case status >= RecordStatusInTest && status < RecordStatusInReport:
		return routes[status], status
	default:
		return StageBasic, RecordStatusInit
	}
}

// TODO::check this method again
func (s *HttpSrv) checkTestSequence(ctx context.Context, publicID, testType string) (*dbSrv.TestRecord, error) {
	record, dbErr := dbSrv.Instance().QueryUnfinishedTest(ctx, publicID)
	if dbErr != nil {
		return nil, dbErr
	}

	if record == nil {
		return nil, fmt.Errorf("没有找到当前测试问卷")
	}

	flow := getTestRoutes(record.BusinessType)
	if len(flow) == 0 {
		return nil, fmt.Errorf("no test flow configured for business type %s", record.BusinessType)
	}

	currentStage, currentIdx := parseStatusToRoute(int(record.Status), flow)

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
	case BusinessTypeBasic:
		stages = testFlowForBasic
		titles = testFlowDescForBasic
	case BusinessTypePro:
		stages = testFlowForPro
		titles = testFlowDescForPro
	case BusinessTypeAdv:
		stages = testFlowForAdv
		titles = testFlowDescForAdv
	case BusinessTypeSchool:
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
	case BusinessTypeBasic:
		switch testTyp {
		case StageRiasec:
			return ai_api.TypRIASEC
		case StageAsc:
			return ai_api.TypASC
		}
	case BusinessTypePro:
		switch testTyp {
		case StageRiasec:
			return ai_api.TypRIASEC
		case StageAsc:
			return ai_api.TypASC
		}

	case BusinessTypeAdv:
		switch testTyp {
		case StageRiasec:
			return ai_api.TypRIASEC
		case StageOcean:
			return ai_api.TypOCEAN
		case StageAsc:
			return ai_api.TypASC
		}

	case BusinessTypeSchool:
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

// 通用转发：保留 method + query + headers + body
func (s *HttpSrv) forwardCallback(w http.ResponseWriter, r *http.Request, target string) {
	log := s.log.With().
		Str("handler", "forwardCallback").
		Str("target", target).
		Logger()

	// 1. 读 body（有就读，没有也没事）
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1MB 限制
	if err != nil {
		log.Error().Err(err).Msg("read body failed")
		http.Error(w, "read body failed", http.StatusInternalServerError)
		return
	}
	_ = r.Body.Close()

	// 2. 拼接 query 到目标 URL 上
	u, err := url.Parse(target)
	if err != nil {
		log.Error().Err(err).Msg("parse target url failed")
		http.Error(w, "bad target", http.StatusInternalServerError)
		return
	}
	u.RawQuery = r.URL.RawQuery

	// 3. 用原始 method 转发（GET/POST 都适用）
	forwardReq, err := http.NewRequestWithContext(
		r.Context(),
		r.Method,
		u.String(),
		bytes.NewReader(body),
	)
	if err != nil {
		log.Error().Err(err).Msg("create forward request failed")
		http.Error(w, "forward init failed", http.StatusInternalServerError)
		return
	}

	// 4. 拷贝 header（保留 Wechatpay-* 等签名头）
	for name, values := range r.Header {
		if strings.EqualFold(name, "Host") ||
			strings.EqualFold(name, "Content-Length") ||
			strings.EqualFold(name, "Content-Encoding") {
			continue
		}
		for _, v := range values {
			forwardReq.Header.Add(name, v)
		}
	}

	resp, err := http.DefaultClient.Do(forwardReq)
	if err != nil {
		log.Error().Err(err).Msg("forward request failed")
		http.Error(w, "forward failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("read forward response body failed")
		http.Error(w, "read forward response failed", http.StatusInternalServerError)
		return
	}

	for name, values := range resp.Header {
		if strings.EqualFold(name, "Content-Length") ||
			strings.EqualFold(name, "Transfer-Encoding") {
			continue
		}
		for _, v := range values {
			w.Header().Add(name, v)
		}
	}

	w.WriteHeader(resp.StatusCode)
	if _, err := w.Write(respBody); err != nil {
		log.Error().Err(err).Msg("write response to client failed")
		return
	}
}

func nullToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}
