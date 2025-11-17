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

type AIChannelMsg struct {
	ChID string `json:"ch_id"`
	Msg  string `json:"msg"`
}

func (acm *AIChannelMsg) Str() string {
	bts, _ := json.Marshal(acm)
	return string(bts)
}

func (s *HttpSrv) initSSE() error {

	return nil
}

func parseTestIDFromPath(path string) (string, error) {
	if i := strings.Index(path, "?"); i >= 0 {
		path = path[:i]
	}

	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 3 {
		return "", fmt.Errorf("invalid path, want /api/sub/{id}, got: %s", path)
	}
	if parts[0] != "api" || parts[1] != "sub" {
		return "", fmt.Errorf("invalid path segments: %v", parts)
	}

	idStr := parts[2]
	if idStr == "" {
		return "", fmt.Errorf("empty id in path: %s", path)
	}
	return idStr, nil
}

func parseAITestTyp(scaleKey, testType string) ai_api.TestTyp {
	switch testType {
	case TestTypeBasic:
		switch scaleKey {
		case StageRiasec:
			return ai_api.TypRIASEC
		case StageAsc:
			return ai_api.TypSEC
		}
	case TestTypePro:
		switch scaleKey {
		case StageRiasec:
			return ai_api.TypRIASEC
		case StageOcean:
			return ai_api.TypOCEAN
		case StageAsc:
			return ai_api.TypSEC
		}

	case TestTypeSchool:
		switch scaleKey {
		case StageRiasec:
			return ai_api.TypRIASEC
		case StageOcean:
			return ai_api.TypOCEAN
		case StageAsc:
			return ai_api.TypSEC
		}
	}

	return ai_api.TypUnknown
}

func (s *HttpSrv) handleSSEEvent(w http.ResponseWriter, r *http.Request) {
	channelID, err := parseTestIDFromPath(r.URL.Path)
	if err != nil {
		s.log.Err(err).Msg("SSE channel parse failed")
		http.Error(w, "无效的问卷编号:"+err.Error(), http.StatusBadRequest)

		return
	}

	q := r.URL.Query()
	scaleKey := q.Get("scaleKey")
	testType := q.Get("testType")

	s.log.Info().
		Str("channel", channelID).
		Str("scaleKey", scaleKey).
		Str("testType", testType).
		Msg("SSE channel created")

	aiTestType := parseAITestTyp(scaleKey, testType)
	if len(aiTestType) == 0 || aiTestType == ai_api.TypUnknown {
		s.log.Error().Str("channel", channelID).Msg("Invalid scaleKey or testType")
		http.Error(w, "需要参数正确的测试类型和测试阶段参数：scaleKey testType", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		s.log.Err(err).Str("channel", channelID).Msg("SSE channel created error")
		http.Error(w, "不支持流式数据传输", http.StatusInternalServerError)
		return
	}
	s.log.Info().Str("channel", channelID).Msg("SSE channel created")

	ctx := r.Context()

	bi, dbErr := dbSrv.Instance().QueryBasicInfo(ctx, channelID)
	if dbErr != nil {
		s.log.Err(dbErr).Str("channel", channelID).Msg("Query basic info from SSE channel error")
		http.Error(w, "查询用户基本信息失败:"+dbErr.Error(), http.StatusInternalServerError)
		return
	}

	testContent, aiErr := ai_api.Instance().GenerateQuestion(ctx, bi, aiTestType, s.aiProcessInfo)
	if aiErr != nil {
		s.log.Err(dbErr).Str("channel", channelID).Msg("ai generate questions error")
		http.Error(w, "AI 生成测试题目失败:"+aiErr.Error(), http.StatusInternalServerError)
		return
	}

	dbErrR := dbSrv.Instance().SaveRiasecSession(ctx, channelID, testContent)
	if dbErrR != nil {
		s.log.Err(dbErr).Str("channel", channelID).Msg("save questions error")
		http.Error(w, "存储AI 测试题目失败:"+aiErr.Error(), http.StatusInternalServerError)
		return
	}

	for i := 1; i <= 200; i++ {
		select {
		case <-ctx.Done():
			s.log.Info().Str("channel", channelID).Msg("SSE channel closed by client")
			return
		case <-time.After(1 * time.Second):

			msg := &AIChannelMsg{Msg: "this message is from ai", ChID: channelID}

			n, err := fmt.Fprintf(w, "data: %s\n\n", msg.Str())
			if n == 0 || err != nil {
				s.log.Err(err).Str("channel", channelID).Msg("SSE channel failed")
				return
			}

			flusher.Flush()
		}
	}
}

func (s *HttpSrv) aiProcessInfo(token string) error {

}
