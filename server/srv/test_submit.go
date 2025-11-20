package srv

import (
	"encoding/json"
	"net/http"

	"github.com/hopwesley/wenxintai/server/ai_api"
	"github.com/hopwesley/wenxintai/server/dbSrv"
)

type AnswerTriple struct {
	ID        int    `json:"id"`
	Dimension string `json:"dimension"`
	Value     int    `json:"value"`
}
type tesSubmitRequest struct {
	TestPublicID string         `json:"public_id"`
	BusinessType string         `json:"business_type"`
	TestType     string         `json:"test_type"`
	Answers      []AnswerTriple `json:"answers"`
}

func (req *tesSubmitRequest) parseObj(r *http.Request) *ApiErr {
	if r.Method != http.MethodPost {
		return ApiMethodInvalid
	}
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
		Str("public_id", req.TestPublicID).Logger()

	sLog.Info().Msg("prepare parse answers")

	ctx := r.Context()

	aiTestType := parseAITestTyp(req.TestType, req.BusinessType)
	if len(aiTestType) == 0 || aiTestType == ai_api.TypUnknown {
		s.log.Error().Msg("Invalid TestType or BusinessType")
		writeError(w, ApiInvalidReq("无效的试卷类型", nil))
		return
	}

	answersJSON, _ := json.Marshal(req.Answers)
	if err := dbSrv.Instance().SaveAnswer(ctx, req.BusinessType, string(aiTestType), req.TestPublicID, answersJSON); err != nil {
		sLog.Err(err).Msg("保存答案失败")
		writeError(w, ApiInternalErr("无效的试卷类型", err))
		return
	}

	sLog.Info().Msg("save answers success")
	writeJSON(w, http.StatusOK, &CommonRes{Ok: true, Msg: "保存答案成功"})
}
