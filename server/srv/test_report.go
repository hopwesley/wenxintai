package srv

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hopwesley/wenxintai/server/ai_api"
	"github.com/hopwesley/wenxintai/server/dbSrv"
)

type tesReportRequest struct {
	PublicID     string `json:"public_id"`
	BusinessType string `json:"business_type"`
}

func (req *tesReportRequest) parseObj(r *http.Request) *ApiErr {
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return ApiInvalidReq("invalid request body", err)
	}
	if !IsValidPublicID(req.PublicID) {
		return ApiInvalidReq("无效的问卷编号", nil)
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

type rawOcean struct {
	ID        int    `json:"id"`
	Value     int    `json:"value"`
	Dimension string `json:"dimension"`
	Reverse   bool   `json:"reverse"`
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

func convertOcean(rawJSON []byte) ([]ai_api.OCEANCAnswer, error) {
	if rawJSON == nil {
		return nil, nil
	}
	var raws []rawOcean
	if err := json.Unmarshal(rawJSON, &raws); err != nil {
		return nil, err
	}

	out := make([]ai_api.OCEANCAnswer, 0, len(raws))
	for _, r := range raws {
		out = append(out, ai_api.OCEANCAnswer{
			ID:        r.ID,
			Score:     r.Value, // 👈 关键：value -> Score
			Dimension: r.Dimension,
			Reverse:   r.Reverse,
		})
	}
	return out, nil
}

const ReportInvalidDuration = 6 * 30 * 24 * time.Hour

type CombinedReport struct {
	*dbSrv.UserProfile
	Mode        string    `json:"mode"`
	GeneratedAt time.Time `json:"generate_at"`
	ExpiredAt   time.Time `json:"expired_at"`
	*ai_api.EngineResult
}

func (s *HttpSrv) generateReport(w http.ResponseWriter, r *http.Request) {

	var req tesReportRequest
	err := req.parseObj(r)
	if err != nil {
		s.log.Err(err).Msgf("invalid test report request")
		writeError(w, err)
		return
	}

	sLog := s.log.With().
		Str("business_type", req.BusinessType).
		Str("public_id", req.PublicID).Logger()

	ctx := r.Context()

	record, cErr := dbSrv.Instance().QueryRecordById(ctx, req.PublicID)
	if cErr != nil {
		sLog.Err(cErr).Msg("no record found ")
		writeError(w, ApiInvalidNoTestRecord(cErr))
		return
	}

	if !record.PayOrderId.Valid || !record.PaidTime.Valid {
		sLog.Error().Msg(" record is not paid")
		writeError(w, ApiInternalErr("问卷尚未支付，请先支付再生产报告", cErr))
		return
	}

	sessions, dbErr := dbSrv.Instance().FindQASessionsForReport(ctx, req.PublicID)
	if dbErr != nil || len(sessions) == 0 {
		sLog.Err(dbErr).Msg("FindQASessionsForReport failed")
		writeError(w, ApiInternalErr("未找到问卷测试的题目与答案", dbErr))
		return
	}

	var riasecJSON, ascJSON, oceanJSON []byte
	for _, s := range sessions {
		if len(s.Answers) == 0 {
			sLog.Err(dbErr).Msg("no valid answer data for:" + s.TestType)
			writeError(w, ApiInternalErr("问卷没有有效答案", nil))
			return
		}
		switch ai_api.TestTyp(s.TestType) {
		case ai_api.TypRIASEC:
			riasecJSON = s.Answers
		case ai_api.TypASC:
			ascJSON = s.Answers
		case ai_api.TypOCEAN:
			oceanJSON = s.Answers
		}
	}

	riaAnswers, rErr := convertRIASEC(riasecJSON)
	ascAnswers, aErr := convertASC(ascJSON)
	oceanAnswers, oErr := convertOcean(oceanJSON)
	if rErr != nil || aErr != nil || oErr != nil {
		cErr := fmt.Errorf(" riasec"+
			" err:%s asc err:%s ocean err:%s", rErr, aErr, oErr)
		sLog.Err(cErr).Msg("parse answer to ai param failed")
		writeError(w, ApiInternalErr("解析问卷答案为 AI 参数失败", cErr))
		return
	}

	answersMap := map[ai_api.TestTyp]any{
		ai_api.TypRIASEC: riaAnswers,
		ai_api.TypASC:    ascAnswers,
		ai_api.TypOCEAN:  oceanAnswers,
	}

	var resp *ai_api.EngineResult
	var aiErr error
	switch strings.ToLower(record.BusinessType) {
	case BusinessTypeBasic:
		resp, aiErr = ai_api.BasicBuildReportParam(ai_api.Mode(record.Mode.String), answersMap)
	case BusinessTypePro:
		resp, aiErr = ai_api.ProBuildReportParam(ai_api.Mode(record.Mode.String), answersMap)
	case BusinessTypeAdv:
		resp, aiErr = ai_api.ProBuildReportParam(ai_api.Mode(record.Mode.String), answersMap)
	case BusinessTypeSchool:
		resp, aiErr = ai_api.SchoolBuildReportParam(ai_api.Mode(record.Mode.String), answersMap)
	default:
		sLog.Warn().Msg("unknown business type when building report param")
		writeError(w, ApiInternalErr("未知的测试类型", aiErr))
		return
	}

	if aiErr != nil || resp == nil {
		sLog.Err(dbErr).Msg("failed to build report param")
		writeError(w, ApiInternalErr("生成 AI 报告需要的参数失败", aiErr))
		return
	}

	var aiParamForMode []byte
	commonScore, _ := json.Marshal(resp.CommonScore)
	if resp.Recommend33 != nil {
		aiParamForMode, _ = json.Marshal(resp.Recommend33)
	} else {
		aiParamForMode, _ = json.Marshal(resp.Recommend312)
	}

	dbErr = dbSrv.Instance().SaveTestReportCore(ctx, req.PublicID, record.Mode.String, commonScore, aiParamForMode)
	if dbErr != nil {
		sLog.Err(dbErr).Msg("failed to save report param")
		writeError(w, ApiInternalErr("保存 AI 报告需要的参数失败", dbErr))
		return
	}

	sLog.Info().Msg("build param of report success")
	user, pDBErr := dbSrv.Instance().FindUserProfileByUid(ctx, record.WeChatID.String)
	if pDBErr != nil || user == nil {
		sLog.Err(pDBErr).Msg("failed to find user profile")
		writeError(w, ApiInternalErr("查找用户基本信息失败", pDBErr))
		return
	}

	combinedResult := &CombinedReport{
		UserProfile:  user,
		Mode:         record.Mode.String,
		GeneratedAt:  time.Now(),
		EngineResult: resp,
	}

	combinedResult.ExpiredAt = combinedResult.GeneratedAt.Add(ReportInvalidDuration)

	writeJSON(w, http.StatusOK, combinedResult)
}

func (s *HttpSrv) finishReport(w http.ResponseWriter, r *http.Request) {

	var req tesReportRequest
	err := req.parseObj(r)
	if err != nil {
		s.log.Err(err).Msgf("invalid test report request")
		writeError(w, err)
		return
	}

	sLog := s.log.With().
		Str("business_type", req.BusinessType).
		Str("public_id", req.PublicID).Logger()

	ctx := r.Context()

	var dbErr = dbSrv.Instance().FinalizedTest(ctx, req.PublicID, req.BusinessType)
	if dbErr != nil {
		sLog.Err(dbErr).Msg("failed to finalized test")
		writeError(w, ApiInternalErr("保存 AI 报告需要的参数失败", dbErr))
		return
	}

	writeJSON(w, http.StatusOK,
		&CommonRes{Ok: true, Msg: "完成报告设计"})
}
