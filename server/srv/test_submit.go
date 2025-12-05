package srv

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hopwesley/wenxintai/server/ai_api"
	"github.com/hopwesley/wenxintai/server/dbSrv"
)

type AnswerItem struct {
	ID int `json:"id"`

	// RIASEC / OCEAN 公共的维度字段：
	//  - RIASEC: R / I / A / S / E / C
	//  - OCEAN:  O / C / E / A / N
	Dimension string `json:"dimension,omitempty"`

	// ASC 专用：学科编码 & 标签
	Subject      string `json:"subject,omitempty"`       // "PHY"
	SubjectLabel string `json:"subject_label,omitempty"` // "物理"

	// ASC / OCEAN 专用：是否反向题
	Reverse bool `json:"reverse,omitempty"`

	// ASC 专用：题目子类型（Comparison / Efficacy / ...）
	Subtype string `json:"subtype,omitempty"`

	// 通用答案值：1 ~ 5
	Value int `json:"value"`
}

type tesSubmitRequest struct {
	TestPublicID string       `json:"public_id"`
	BusinessType string       `json:"business_type"`
	TestType     string       `json:"test_type"`
	Answers      []AnswerItem `json:"answers"`
}

func (req *tesSubmitRequest) parseObj(r *http.Request) *ApiErr {
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return ApiInvalidReq("invalid request body", err)
	}
	if !IsValidPublicID(req.TestPublicID) {
		return ApiInvalidReq("无效的问卷编号", nil)
	}
	if len(req.TestType) == 0 {
		return ApiInvalidReq("无效的测试类型", nil)
	}
	if len(req.BusinessType) == 0 {
		return ApiInvalidReq("无效的试卷类型", nil)
	}
	if len(req.Answers) == 0 {
		return ApiInvalidReq("无效的答案数据", nil)
	}
	return nil
}

func (s *HttpSrv) checkPreviousStageIfReady(ctx context.Context, publicId, bTyp string, stage ai_api.TestTyp) error {
	preStage := previousRoute(bTyp, stage)
	switch preStage {
	case StageBasic:
		record, err := dbSrv.Instance().QueryRecordByPid(ctx, publicId)
		if err != nil || !record.Mode.Valid {
			return fmt.Errorf("请先选择科目选择模式")
		}
		return nil
	case StageReport:
		return fmt.Errorf("没有找到前置流程")
	default:
		answers, err := dbSrv.Instance().FindQASession(ctx, string(preStage), publicId)
		if err != nil || answers == nil {
			return fmt.Errorf("请先完成【%s】阶段的问题", preStage)
		}
	}

	return nil
}
func (s *HttpSrv) handleTestSubmit(w http.ResponseWriter, r *http.Request) {
	var req tesSubmitRequest
	err := req.parseObj(r)
	if err != nil {
		s.log.Err(err).Msgf("invalid test submit request")
		writeError(w, err)
		return
	}
	sLog := s.log.With().Str("test_type", req.TestType).
		Str("business_type", req.BusinessType).
		Str("public_id", req.TestPublicID).
		Int("answer", len(req.Answers)).Logger()
	ctx := r.Context()

	sLog.Info().Msg("prepare to parse answers")

	cErr := s.checkPreviousStageIfReady(ctx, req.TestPublicID, req.BusinessType, ai_api.TestTyp(req.TestType))
	if cErr != nil {
		sLog.Err(cErr).Msg("previous stage check failed")
		writeError(w, ApiInvalidTestSequence(cErr))
		return
	}

	nextR, nextIdx, rErr := nextRoute(req.BusinessType, ai_api.TestTyp(req.TestType))
	if rErr != nil {
		s.log.Err(rErr).Msg("failed to find next route ")
		writeError(w, ApiInternalErr("未找到下一轮状态", rErr))
		return
	}

	uid := userIDFromContext(ctx)
	answersJSON, _ := json.Marshal(req.Answers)
	if err := dbSrv.Instance().SaveAnswer(ctx, req.TestType,
		req.TestPublicID, uid, answersJSON, nextIdx); err != nil {
		sLog.Err(err).Msg("保存答案失败")
		writeError(w, ApiInternalErr("无效的试卷类型", err))
		return
	}

	writeJSON(w, http.StatusOK,
		&CommonRes{Ok: true, Msg: "保存答案成功",
			NextRoute: string(nextR),
			NextRid:   nextIdx})

	sLog.Info().Int("next-route-id", nextIdx).Str("next-route", string(nextR)).Msg("save answers success")
}
