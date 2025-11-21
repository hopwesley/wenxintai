package srv

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/hopwesley/wenxintai/server/ai_api"
	"github.com/hopwesley/wenxintai/server/dbSrv"
)

type tesReportRequest struct {
	TestPublicID string `json:"public_id"`
	BusinessType string `json:"business_type"`
	TestType     string `json:"test_type"`
}

func (req *tesReportRequest) parseObj(r *http.Request) *ApiErr {
	if r.Method != http.MethodPost {
		return ApiMethodInvalid
	}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return ApiInvalidReq("invalid request body", err)
	}
	if !IsValidPublicID(req.TestPublicID) {
		return ApiInvalidReq("无效的问卷编号", nil)
	}
	if (req.TestType) != StageReport {
		return ApiInvalidReq("当前请求不是测试报告阶段", nil)
	}
	if len(req.BusinessType) == 0 {
		return ApiInvalidReq("无效的试卷类型", nil)
	}
	return nil
}

// 先定义与前端 JSON 对应的 raw 结构
type rawRiasec struct {
	ID        int    `json:"id"`
	Dimension string `json:"dimension"`
	Value     int    `json:"value"`
}

type rawAsc struct {
	ID           int    `json:"id"`
	Subject      string `json:"subject"`
	SubjectLabel string `json:"subject_label"`
	Value        int    `json:"value"`
	Reverse      bool   `json:"reverse"`
	Subtype      string `json:"subtype"`
}

// 从 QASession.Answers 解析并转换
func convertRIASEC(rawJSON []byte) ([]ai_api.RIASECAnswer, error) {
	var raws []rawRiasec
	if err := json.Unmarshal(rawJSON, &raws); err != nil {
		return nil, err
	}

	out := make([]ai_api.RIASECAnswer, 0, len(raws))
	for _, r := range raws {
		// 这里假设 Dimension 已经是 "R"/"I"/...，否则要从 questions 里补
		out = append(out, ai_api.RIASECAnswer{
			ID:        r.ID,
			Dimension: r.Dimension,
			Score:     r.Value, // 👈 关键：value -> Score
		})
	}
	return out, nil
}

func convertASC(rawJSON []byte) ([]ai_api.ASCAnswer, error) {
	var raws []rawAsc
	if err := json.Unmarshal(rawJSON, &raws); err != nil {
		return nil, err
	}

	out := make([]ai_api.ASCAnswer, 0, len(raws))
	for _, r := range raws {
		out = append(out, ai_api.ASCAnswer{
			ID:      r.ID,
			Subject: r.Subject,
			Score:   r.Value, // 👈 关键：value -> Score
			Reverse: r.Reverse,
			Subtype: r.Subtype,
		})
	}
	return out, nil
}

func (s *HttpSrv) handleTestReport(w http.ResponseWriter, r *http.Request) {

	var req tesReportRequest
	err := req.parseObj(r)
	if err != nil {
		s.log.Err(err).Msgf("invalid test report request")
		writeError(w, err)
		return
	}

	sLog := s.log.With().Str("test_type", req.TestType).
		Str("business_type", req.BusinessType).
		Str("public_id", req.TestPublicID).Logger()

	ctx := r.Context()

	record, cErr := s.checkTestSequence(ctx, req.TestPublicID, req.TestType)
	if cErr != nil {
		sLog.Err(cErr).Msg("invalid test sequence request")
		writeError(w, ApiInvalidTestSequence(cErr))
		return
	}

	// 2️⃣ 获取所有阶段问卷答案
	sessions, dbErr := dbSrv.Instance().FindQASessionsForReport(ctx, record.BusinessType, req.TestPublicID)
	if dbErr != nil {
		sLog.Err(dbErr).Msg("FindQASessionsForReport failed")
		writeError(w, ApiInternalErr("未找到问卷测试的题目与答案", dbErr))
		return
	}
	if len(sessions) == 0 {
		sLog.Err(dbErr).Msg("no question_answers found for this test")
		writeError(w, ApiInternalErr("未找到问卷测试的题目与答案", nil))
		return
	}

	// 3️⃣ 取出不同阶段的数据
	var riasecJSON, ascJSON []byte
	for _, s := range sessions {
		switch strings.ToUpper(s.TestType) {
		case "RIASEC":
			riasecJSON = s.Answers
		case "ASC":
			ascJSON = s.Answers
		}
	}

	if len(riasecJSON) == 0 || len(ascJSON) == 0 {
		sLog.Err(dbErr).Msg("missing required test stages")
		writeError(w, ApiInternalErr("未找到兴趣问卷或者能力问卷", nil))
		return
	}

	// 4️⃣ 解析并转换为算法输入结构
	riaAnswers, _ := convertRIASEC(riasecJSON)
	ascAnswers, _ := convertASC(ascJSON)

	// 5️⃣ 按业务类型 & 模式调用不同算法逻辑
	var (
		param  *ai_api.ParamForAIPrompt
		result *ai_api.FullScoreResult
		scores []ai_api.SubjectScores
	)

	switch strings.ToLower(record.BusinessType) {
	case "basic":
		// 默认教育测评逻辑：兴趣+能力 -> 3+1+2 / 3+3 双模分析
		param, result, scores = ai_api.BuildFullParam(riaAnswers, ascAnswers, 0.4, 0.4, 0.2)

	case "pro":
		// 专业版逻辑，可加入额外计算或权重调整
		param, result, scores = ai_api.BuildFullParam(riaAnswers, ascAnswers, 0.5, 0.3, 0.2)

	case "school":
		// 校园版：可能只输出 AnchorPHY/HIS，不生成组合
		param, result, scores = ai_api.BuildFullParam(riaAnswers, ascAnswers, 0.3, 0.4, 0.3)

	default:
		param, result, scores = ai_api.BuildFullParam(riaAnswers, ascAnswers, 0.4, 0.4, 0.2)
	}

	// 6️⃣ 按 Mode 输出不同内容
	switch strings.ToLower(record.Mode.String) {
	case "3+3":
		param.Mode312 = nil
	case "3+1+2":
		param.Mode33 = nil
	}

	resp := map[string]any{
		"report": param,
		"scores": scores,
		"common": result.Common,
	}

	writeJSON(w, http.StatusOK, resp)
}
