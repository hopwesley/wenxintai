package srv

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hopwesley/wenxintai/server/ai_api"
	"github.com/hopwesley/wenxintai/server/dbSrv"
	"github.com/rs/zerolog"
)

type tesReportRequest struct {
	PublicID string `json:"public_id"`
}

type finishReportRequest struct {
	PublicID        string `json:"public_id"`
	RatingScore     *int   `json:"rating_score,omitempty"`
	FeedbackContent string `json:"feedback_content,omitempty"`
}

func (req *tesReportRequest) parseObj(r *http.Request) *ApiErr {
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return ApiInvalidReq("invalid request body", err)
	}
	if !IsValidPublicID(req.PublicID) {
		return ApiInvalidReq("无效的问卷编号", nil)
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
			Score:     r.Value,
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
			Score:   r.Value,
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
	Mode          string    `json:"mode"`
	GeneratedAt   time.Time `json:"generated_at"`
	ExpiredAt     time.Time `json:"expired_at"`
	PaidByInvite  bool      `json:"paid_by_invite"`
	*ai_api.EngineResult
	AIContent string `json:"ai_content,omitempty"`
}

func (s *HttpSrv) queryOrCreateReport(w http.ResponseWriter, r *http.Request) {

	var req tesReportRequest
	err := req.parseObj(r)
	if err != nil {
		s.log.Err(err).Msgf("invalid test report request")
		writeError(w, err)
		return
	}

	sLog := s.log.With().Str("public_id", req.PublicID).Logger()

	ctx := r.Context()
	uid := userIDFromContext(ctx)
	sLog.Debug().Msg("preparing report")

	record, cErr := dbSrv.Instance().QueryTestRecord(ctx, req.PublicID, uid)
	if cErr != nil {
		sLog.Err(cErr).Msg("no record found ")
		writeError(w, ApiInvalidNoTestRecord(cErr))
		return
	}

	if record.WeChatID.String != uid {
		sLog.Error().Msg("no right to find this test record")
		writeError(w, NewApiError(http.StatusForbidden, ErrorCodeForbidden, "无权查看", nil))
		return
	}

	if !record.PayOrderId.Valid || !record.PaidTime.Valid {
		sLog.Error().Msg(" record is not paid")
		writeError(w, ApiInternalErr("问卷尚未支付，请先支付再生产报告", nil))
		return
	}

	report, dbErr := dbSrv.Instance().QueryReportByPublicId(ctx, req.PublicID)
	if dbErr != nil {
		sLog.Err(dbErr).Msg(" report query error")
		writeError(w, ApiInternalErr("查询已经生成报告时异常", dbErr))
		return
	}

	var combinedResult *CombinedReport = nil
	if report == nil {
		combinedResult = s.newReport(ctx, w, req.PublicID, record.BusinessType, ai_api.Mode(record.Mode.String), sLog)

	} else {
		combinedResult = s.parseReport(w, report, sLog)
	}

	if combinedResult == nil {
		sLog.Error().Msg("report query or create failed")
		writeError(w, ApiInternalErr("生成或查询报告时异常", nil))
		return
	}

	user, pDBErr := dbSrv.Instance().QueryUserProfileUid(ctx, uid)
	if pDBErr != nil || user == nil {
		sLog.Err(pDBErr).Msg("failed to find user profile")
		writeError(w, ApiInternalErr("查找用户基本信息失败", pDBErr))
		return
	}
	combinedResult.UserProfile = user

	// 判断是否通过邀请码支付
	if record.PayOrderId.Valid {
		payOrderId := record.PayOrderId.String
		// 查询 invites 表，如果存在该邀请码，则是邀请码支付
		invite, invErr := dbSrv.Instance().GetInviteByCode(ctx, payOrderId)
		if invErr == nil && invite != nil {
			combinedResult.PaidByInvite = true
		}
	}

	writeJSON(w, http.StatusOK, combinedResult)

}

func (s *HttpSrv) newReport(ctx context.Context, w http.ResponseWriter, publicID, businessTyp string, mode ai_api.Mode, sLog zerolog.Logger) *CombinedReport {
	sLog.Debug().Str("business_type", businessTyp).Str("mode", string(mode)).Msg("creating new report")
	sessions, dbErr := dbSrv.Instance().FindQASessionsForReport(ctx, publicID)
	if dbErr != nil || len(sessions) == 0 {
		sLog.Err(dbErr).Msg("FindQASessionsForReport failed")
		writeError(w, ApiInternalErr("未找到问卷测试的题目与答案", dbErr))
		return nil
	}

	var riasecJSON, ascJSON, oceanJSON []byte
	for _, s := range sessions {
		if len(s.Answers) == 0 {
			sLog.Error().Msg("no valid answer data for:" + s.TestType)
			writeError(w, ApiInternalErr("问卷没有有效答案", nil))
			return nil

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
		return nil

	}

	answersMap := map[ai_api.TestTyp]any{
		ai_api.TypRIASEC: riaAnswers,
		ai_api.TypASC:    ascAnswers,
		ai_api.TypOCEAN:  oceanAnswers,
	}

	var resp *ai_api.EngineResult
	var aiErr error
	switch strings.ToLower(businessTyp) {
	case BusinessTypeBasic:
		resp, aiErr = ai_api.BasicBuildReportParam(mode, answersMap)
	case BusinessTypePro:
		resp, aiErr = ai_api.ProBuildReportParam(mode, answersMap)
	case BusinessTypeAdv:
		resp, aiErr = ai_api.ProBuildReportParam(mode, answersMap)
	case BusinessTypeSchool:
		resp, aiErr = ai_api.SchoolBuildReportParam(mode, answersMap)
	default:
		sLog.Warn().Msg("unknown business type when building report param")
		writeError(w, ApiInternalErr("未知的测试类型", aiErr))
		return nil

	}

	if aiErr != nil || resp == nil {
		sLog.Err(aiErr).Msg("failed to build report param")
		writeError(w, ApiInternalErr("生成 AI 报告需要的参数失败", aiErr))
		return nil
	}

	var aiParamForMode []byte
	commonScore, _ := json.Marshal(resp.CommonScore)
	if resp.Recommend33 != nil {
		aiParamForMode, _ = json.Marshal(resp.Recommend33)
	} else {
		aiParamForMode, _ = json.Marshal(resp.Recommend312)
	}

	dbErr = dbSrv.Instance().SaveReportCore(ctx, publicID, string(mode), commonScore, aiParamForMode)
	if dbErr != nil {
		sLog.Err(dbErr).Msg("failed to save report param")
		writeError(w, ApiInternalErr("保存 AI 报告需要的参数失败", dbErr))
		return nil
	}

	sLog.Info().Msg("build param of report success")

	now := time.Now()
	combinedResult := &CombinedReport{
		Mode:         string(mode),
		GeneratedAt:  now,
		EngineResult: resp,
	}

	combinedResult.ExpiredAt = now.Add(ReportInvalidDuration)
	return combinedResult
}

func (s *HttpSrv) parseReport(w http.ResponseWriter, report *dbSrv.TestReport, sLog zerolog.Logger) *CombinedReport {

	var cs ai_api.FullScoreResult
	if err := json.Unmarshal(report.CommonScore, &cs); err != nil {
		sLog.Err(err).Msg("failed to Unmarshal common score")
		writeError(w, ApiInternalErr("解析学科json 数据失败", err))
		return nil
	}

	var resp = &ai_api.EngineResult{
		CommonScore: &cs,
	}

	switch ai_api.Mode(report.Mode) {
	case ai_api.Mode33:
		{
			var aiParamForMode ai_api.Mode33Section
			if err := json.Unmarshal(report.ModeParam, &aiParamForMode); err != nil {
				sLog.Err(err).Msg("failed to Unmarshal mode 33 param")
				writeError(w, ApiInternalErr("解析3+3组合json 数据失败", err))
				return nil
			}
			resp.Recommend33 = &aiParamForMode
			break
		}
	case ai_api.Mode312:
		{
			var aiParamForMode ai_api.Mode312Section
			if err := json.Unmarshal(report.ModeParam, &aiParamForMode); err != nil {
				sLog.Err(err).Msg("failed to Unmarshal mode 312 param")
				writeError(w, ApiInternalErr("解析3+1+2组合json 数据失败", err))
				return nil
			}
			resp.Recommend312 = &aiParamForMode
			break
		}
	default:
		sLog.Error().Str("mode-indb", report.Mode).Msg("unknown mode")
		writeError(w, ApiInternalErr("未知的科目模式", nil))
		return nil
	}

	combinedResult := &CombinedReport{
		Mode:         report.Mode,
		GeneratedAt:  report.GeneratedAt,
		EngineResult: resp,
	}

	combinedResult.ExpiredAt = combinedResult.GeneratedAt.Add(ReportInvalidDuration)

	if report.AIContent != nil {
		combinedResult.AIContent = string(report.AIContent)
	}

	sLog.Info().Msg("parse report success")
	return combinedResult
}

func (s *HttpSrv) finalizedReport(w http.ResponseWriter, r *http.Request) {
	var req finishReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.log.Err(err).Msg("invalid finish report request")
		writeError(w, ApiInvalidReq("invalid request body", err))
		return
	}

	if !IsValidPublicID(req.PublicID) {
		writeError(w, ApiInvalidReq("无效的问卷编号", nil))
		return
	}

	ctx := r.Context()
	uid := userIDFromContext(ctx)
	sLog := s.log.With().Str("public_id", req.PublicID).Str("uid", uid).Logger()

	// 如果提供了反馈数据，则保存反馈
	if req.RatingScore != nil {
		// 验证评分范围
		if *req.RatingScore < 0 || *req.RatingScore > 10 {
			writeError(w, ApiInvalidReq("评分必须在 0-10 之间", nil))
			return
		}

		// 查询支付信息，获取邀请码
		payOrderId, _, err := dbSrv.Instance().QueryTestRecordPaymentInfo(ctx, req.PublicID)
		if err != nil {
			sLog.Err(err).Msg("query payment info failed")
			writeError(w, ApiInternalErr("查询支付信息失败", err))
			return
		}

		if payOrderId == "" {
			sLog.Warn().Msg("no payment info found")
			writeError(w, ApiInvalidReq("未找到支付信息", nil))
			return
		}

		// 验证是否是邀请码支付
		invite, invErr := dbSrv.Instance().GetInviteByCode(ctx, payOrderId)
		if invErr != nil {
			sLog.Err(invErr).Msg("query invite code failed")
			writeError(w, ApiInternalErr("查询邀请码失败", invErr))
			return
		}

		if invite == nil {
			sLog.Warn().Str("pay_order_id", payOrderId).Msg("not paid by invite code")
			writeError(w, ApiInvalidReq("该报告不是通过邀请码支付", nil))
			return
		}

		// 保存反馈
		feedback := &dbSrv.ReportFeedback{
			PublicID:        req.PublicID,
			Uid:             uid,
			InviteCode:      payOrderId,
			RatingScore:     *req.RatingScore,
			FeedbackContent: req.FeedbackContent,
		}

		if err := dbSrv.Instance().InsertReportFeedback(ctx, feedback); err != nil {
			sLog.Err(err).Msg("insert feedback failed")
			writeError(w, ApiInternalErr("保存反馈失败", err))
			return
		}

		sLog.Info().Int("rating", *req.RatingScore).Msg("feedback saved successfully")
	}

	writeJSON(w, http.StatusOK, CommonRes{Ok: true, Msg: "感谢您的反馈"})
	sLog.Debug().Msg("finish report success")
}
