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
	"time"

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
	StageBasic    ai_api.TestTyp = "basic-info"
	StageBasicDes                = "基础信息"

	StageReport    ai_api.TestTyp = "report"
	StageReportDes                = "测评报告"

	StageRiasec    = ai_api.TypRIASEC
	StageRiasecDes = "兴趣测试"

	StageAsc    = ai_api.TypASC
	StageAscDes = "能力测试"

	StageOcean    = ai_api.TypOCEAN
	StageOceanDes = "性格测试"

	StageMotivation    ai_api.TestTyp = "MOTIVATION"
	StageMotivationDes                = "价值观测试"

	BusinessTypeBasic  = "basic"
	BusinessTypePro    = "pro"
	BusinessTypeAdv    = "adv"
	BusinessTypeSchool = "school"
)

func isValidBusinessType(businessTyp string) bool {
	switch businessTyp {
	case BusinessTypeBasic, BusinessTypePro, BusinessTypeAdv, BusinessTypeSchool:
		return true
	default:
		return false
	}
}

type TestFlowStep struct {
	Stage ai_api.TestTyp `json:"stage"`
	Title string         `json:"title"`
}

var basicTestFlow = []TestFlowStep{
	{
		Stage: StageBasic,
		Title: StageBasicDes,
	},
	{
		Stage: StageRiasec,
		Title: StageRiasecDes,
	},
	{
		Stage: StageAsc,
		Title: StageAscDes,
	},
	{
		Stage: StageReport,
		Title: StageReportDes,
	},
}
var proTestFlow = []TestFlowStep{
	{
		Stage: StageBasic,
		Title: StageBasicDes,
	},
	{
		Stage: StageRiasec,
		Title: StageRiasecDes,
	},
	{
		Stage: StageAsc,
		Title: StageAscDes,
	},
	{
		Stage: StageReport,
		Title: StageReportDes,
	},
}
var advTestFlow = []TestFlowStep{
	{
		Stage: StageBasic,
		Title: StageBasicDes,
	},
	{
		Stage: StageRiasec,
		Title: StageRiasecDes,
	},
	{
		Stage: StageAsc,
		Title: StageAscDes,
	},
	{
		Stage: StageOcean,
		Title: StageOceanDes,
	},
	{
		Stage: StageMotivation,
		Title: StageMotivationDes,
	},
	{
		Stage: StageReport,
		Title: StageReportDes,
	},
}
var schoolTestFlow = []TestFlowStep{
	{
		Stage: StageBasic,
		Title: StageBasicDes,
	},
	{
		Stage: StageRiasec,
		Title: StageRiasecDes,
	},
	{
		Stage: StageAsc,
		Title: StageAscDes,
	},
	{
		Stage: StageOcean,
		Title: StageOceanDes,
	},
	{
		Stage: StageMotivation,
		Title: StageMotivationDes,
	},
	{
		Stage: StageReport,
		Title: StageReportDes,
	},
}

var testFlowMap = map[string][]TestFlowStep{
	BusinessTypeBasic:  basicTestFlow,
	BusinessTypePro:    proTestFlow,
	BusinessTypeAdv:    advTestFlow,
	BusinessTypeSchool: schoolTestFlow,
}

func getTestFlowSteps(businessType string) []TestFlowStep {
	return testFlowMap[businessType]
}

func getStageIndex(flow []TestFlowStep, stage int16) ai_api.TestTyp {
	if int(stage) >= len(flow) {
		return StageReport
	}
	return flow[stage].Stage
}

func nextRoute(businessType string, stage ai_api.TestTyp) (ai_api.TestTyp, int, error) {
	flow := testFlowMap[businessType]
	idx := -1
	if len(flow) == 0 {
		return "", idx, fmt.Errorf("no flow data for type:[%s]", businessType)
	}

	for i := 0; i < len(flow); i++ {
		if flow[i].Stage == stage {
			idx = i
			break
		}
	}

	if idx == -1 || idx+1 >= len(flow) {
		return "", -1, fmt.Errorf("no next route stage")
	}
	return flow[idx+1].Stage, idx + 1, nil
}

func previousRoute(businessType string, stage ai_api.TestTyp) ai_api.TestTyp {
	flow := testFlowMap[businessType]
	if len(flow) == 0 {
		return StageReport
	}

	idx := -1
	for i := 0; i < len(flow); i++ {
		if flow[i].Stage == stage {
			idx = i
			break
		}
	}
	if idx < 1 {
		return StageReport
	}

	return flow[idx-1].Stage
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// 3. 用原始 method 转发（GET/POST 都适用）
	forwardReq, err := http.NewRequestWithContext(
		ctx,
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

func safeStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

type ctxKey string

const ctxKeyUserID ctxKey = "context_wx_user_id"

// 一个小工具函数，方便 handler 里直接用
func userIDFromContext(ctx context.Context) string {
	v := ctx.Value(ctxKeyUserID)
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// 或者再包一层，从 *http.Request 里取，handler 用起来更顺手
func userIDFromRequest(r *http.Request) string {
	return userIDFromContext(r.Context())
}
