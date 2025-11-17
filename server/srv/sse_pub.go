package srv

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hopwesley/wenxintai/server/ai_api"
	"github.com/hopwesley/wenxintai/server/dbSrv"
)

type AIChannelMsg struct {
	Typ string `json:"typ"`
	Msg string `json:"msg"`
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
	s.log.Info().
		Str("channel", channelID).
		Str("scaleKey", scaleKey).
		Str("testType", testType).
		Msg("SSE channel created")

	ctx := r.Context()

	msgCh := make(chan string, 16)

	go s.aiProcess(w, msgCh, channelID, aiTestType)

	for {
		select {
		case <-ctx.Done():
			s.log.Info().
				Str("channel", channelID).
				Msg("SSE channel closed by client")
			return

		case token, ok := <-msgCh:
			if !ok {
				// 后台 publish 那边关闭了 channel，说明本 channel 的任务结束
				s.log.Info().
					Str("channel", channelID).
					Msg("SSE channel closed: msgCh closed")
				return
			}

			msg := &AIChannelMsg{Msg: token, Typ: "data"}

			if _, err := fmt.Fprintf(w, "data: %s\n\n", msg.Str()); err != nil {
				s.log.Err(err).
					Str("channel", channelID).
					Msg("SSE write failed")
				return
			}

			flusher.Flush()
		}
	}
}

func (s *HttpSrv) aiProcess(w http.ResponseWriter, msgCh chan string, channelID string, aiTestType ai_api.TestTyp) {
	defer close(msgCh)

	bgCtx := context.Background()

	bi, dbErr := dbSrv.Instance().QueryBasicInfo(bgCtx, channelID)
	if dbErr != nil {
		s.log.Err(dbErr).Str("channel", channelID).Msg("Query basic info from SSE channel error")
		http.Error(w, "查询基本信息失败："+dbErr.Error(), http.StatusInternalServerError)
		return
	}

	callback := func(token string) error {
		return sendSafe(msgCh, token)
	}

	testContent, aiErr := ai_api.Instance().GenerateQuestion(bgCtx, bi, aiTestType, callback)
	if aiErr != nil {
		http.Error(w, "AI 生成 RIASEC 试卷失败："+aiErr.Error(), http.StatusInternalServerError)
		s.log.Err(aiErr).Str("channel", channelID).Msg("ai generate questions error")
		return
	}

	if err := s.saveAIContentByTyp(bgCtx, aiTestType, channelID, testContent); err != nil {
		http.Error(w, "保持 RIASEC 试卷失败："+err.Error(), http.StatusInternalServerError)
		return
	}

	msg := &AIChannelMsg{Msg: string(testContent), Typ: "done"}

	_, err := fmt.Fprintf(w, "data: %s\n\n", msg.Str())

	if err != nil {
		s.log.Err(err).Str("channel", channelID).Msg("maybe client is closed")
	} else {
		s.log.Info().
			Str("channel", channelID).
			Msg("GenerateQuestion finished and saved")
	}

}

func (s *HttpSrv) saveAIContentByTyp(bgCtx context.Context, typ ai_api.TestTyp, channelID string, content []byte) error {
	switch typ {
	case ai_api.TypRIASEC:
		dbErrR := dbSrv.Instance().SaveRiasecSession(bgCtx, channelID, content)
		if dbErrR != nil {
			s.log.Err(dbErrR).Str("channel", channelID).Msg("save questions error")
			return dbErrR
		}
	}

	s.log.Info().Msg("Invalid testType")
	return fmt.Errorf("invalid testType")
}

func sendSafe(ch chan string, token string) error {
	defer func() {
		_ = recover()
	}()

	ch <- token
	return nil
}
