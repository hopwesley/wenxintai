package srv

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hopwesley/wenxintai/server/ai_api"
	"github.com/hopwesley/wenxintai/server/dbSrv"
	"github.com/rs/zerolog"
)

type SSEMsgTyp string

const (
	SSE_MT_DATA  SSEMsgTyp = "message"
	SSE_MT_ERROR SSEMsgTyp = "app-error"
	SSE_MT_DONE  SSEMsgTyp = "done"
)

type SSEMessage struct {
	Typ SSEMsgTyp
	Msg string
}

//
//func (acm *SSEMessage) SSEMsg() string {
//	if len(acm.Typ) == 0 {
//		acm.Typ = SSE_MT_DATA
//	}
//	return fmt.Sprintf("event: %s\ndata: %s\n\n", acm.Typ, acm.Msg)
//}

func (acm *SSEMessage) SSEMsg() string {
	if len(acm.Typ) == 0 {
		acm.Typ = SSE_MT_DATA
	}

	// 统一换行符，先把 \r\n 转成 \n
	normalized := strings.ReplaceAll(acm.Msg, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")

	var b strings.Builder
	b.WriteString("event: ")
	b.WriteString(string(acm.Typ))
	b.WriteByte('\n')

	for _, line := range lines {
		b.WriteString("data: ")
		b.WriteString(line)
		b.WriteByte('\n')
	}

	b.WriteByte('\n') // 结束整个 event
	return b.String()
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
	if !IsValidPublicID(idStr) {
		return "", fmt.Errorf("无效的问卷编号: %s", path)
	}
	return idStr, nil
}

func parseAITestTyp(testTyp, businessTyp string) ai_api.TestTyp {
	switch businessTyp {
	case TestTypeBasic:
		switch testTyp {
		case StageRiasec:
			return ai_api.TypRIASEC
		case StageAsc:
			return ai_api.TypSEC
		}
	case TestTypePro:
		switch testTyp {
		case StageRiasec:
			return ai_api.TypRIASEC
		case StageOcean:
			return ai_api.TypOCEAN
		case StageAsc:
			return ai_api.TypSEC
		}

	case TestTypeSchool:
		switch testTyp {
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
	publicId, err := parseTestIDFromPath(r.URL.Path)
	if err != nil {
		s.log.Err(err).Msg("SSE channel parse failed")
		http.Error(w, "无效的问卷编号:"+err.Error(), http.StatusBadRequest)
		return
	}

	q := r.URL.Query()
	businessTyp := q.Get("business_type")
	testType := q.Get("test_type")

	aiTestType := parseAITestTyp(testType, businessTyp)
	if len(aiTestType) == 0 || aiTestType == ai_api.TypUnknown {
		s.log.Error().Str("channel", publicId).Msg("Invalid scaleKey or testType")
		http.Error(w, "需要参数正确的测试类型和测试阶段参数：scaleKey testType", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		s.log.Err(err).Str("channel", publicId).Msg("SSE channel created error")
		http.Error(w, "不支持流式数据传输", http.StatusInternalServerError)
		return
	}

	s.log.Info().
		Str("channel", publicId).
		Str("business_type", businessTyp).
		Str("testType", testType).
		Msg("SSE channel created")

	ctx := r.Context()

	msgCh := make(chan *SSEMessage, 64)

	go s.aiProcess(msgCh, publicId, businessTyp, aiTestType)

	for {
		select {
		case <-ctx.Done():
			s.log.Info().
				Str("channel", publicId).
				Msg("SSE channel closed by client")
			return

		case msg, ok := <-msgCh:
			if !ok {
				s.log.Info().
					Str("channel", publicId).
					Msg("SSE channel closed: msgCh closed")
				return
			}
			if err := writeSSE(w, flusher, msg, &s.log); err != nil {
				return
			}
		}
	}
}

func writeSSE(w http.ResponseWriter, flusher http.Flusher, msg *SSEMessage, log *zerolog.Logger) error {
	if _, err := fmt.Fprint(w, msg.SSEMsg()); err != nil {
		log.Err(err).Str("typ", string(msg.Typ)).Msg("SSE write failed")
		return err
	}
	flusher.Flush()
	return nil
}

func (s *HttpSrv) querySavedQuestionsFirst(bgCtx context.Context, publicId, businessTyp string, aiTestType ai_api.TestTyp) (dbSrv.QuestionRecord, error) {
	switch aiTestType {
	case ai_api.TypRIASEC:
		riasecRecord, err := dbSrv.Instance().FindRiasecSession(bgCtx, businessTyp, publicId)
		if err != nil {
			return nil, fmt.Errorf("查询 RIASEC 试题失败：" + err.Error())
		}
		if riasecRecord == nil {
			return nil, nil
		}
		return riasecRecord, nil

	case ai_api.TypSEC:
		return nil, nil

	case ai_api.TypOCEAN:
		return nil, nil

	default:
		return nil, nil
	}
}

func (s *HttpSrv) aiProcess(msgCh chan *SSEMessage, publicId, businessTyp string, aiTestType ai_api.TestTyp) {

	sLog := s.log.With().Str("channel", publicId).Str("business_type", businessTyp).Str("ai_Type", string(aiTestType)).Logger()
	//defer close(msgCh)

	bgCtx := context.Background()

	dbQuestion, err := s.querySavedQuestionsFirst(bgCtx, publicId, businessTyp, aiTestType)
	if err != nil {
		sLog.Err(err).Msg("failed when find questions from database")
		msg := &SSEMessage{Msg: err.Error(), Typ: SSE_MT_ERROR}
		sendSafe(msgCh, msg, &s.log)
		return
	}

	if dbQuestion != nil {
		sLog.Info().Msg("found questions from database")
		msg := &SSEMessage{Msg: string(dbQuestion.GetQuestions()), Typ: SSE_MT_DONE}
		sendSafe(msgCh, msg, &s.log)
		return
	}

	bi, dbErr := dbSrv.Instance().QueryBasicInfo(bgCtx, publicId)
	if dbErr != nil {
		sLog.Err(dbErr).Msg("Query basic info from SSE channel error")
		msg := &SSEMessage{Msg: "查询基本信息失败：" + dbErr.Error(), Typ: SSE_MT_ERROR}
		sendSafe(msgCh, msg, &s.log)
		return
	}

	callback := func(token string) error {
		msg := &SSEMessage{Msg: token, Typ: SSE_MT_DATA}
		sendSafe(msgCh, msg, &s.log)
		return nil
	}

	testContent, aiErr := ai_api.Instance().GenerateQuestion(bgCtx, bi, aiTestType, callback)
	if aiErr != nil {
		sLog.Err(aiErr).Msg("ai generate questions error")
		msg := &SSEMessage{Msg: "AI 生成 RIASEC 试卷失败：" + aiErr.Error(), Typ: SSE_MT_ERROR}
		sendSafe(msgCh, msg, &s.log)
		return
	}

	if err := s.saveAIContentByTyp(bgCtx, aiTestType, publicId, businessTyp, json.RawMessage(testContent)); err != nil {
		sLog.Err(err).Msg("保存 RIASEC 试卷失败")
		msg := &SSEMessage{Msg: "保存 RIASEC 试卷失败：" + err.Error(), Typ: SSE_MT_ERROR}
		sendSafe(msgCh, msg, &s.log)
		return
	}

	msg := &SSEMessage{Msg: testContent, Typ: SSE_MT_DONE}
	sendSafe(msgCh, msg, &s.log)

	sLog.Info().Msg("GenerateQuestion finished and saved")
}

func (s *HttpSrv) saveAIContentByTyp(bgCtx context.Context, typ ai_api.TestTyp, publicId, businessTyp string, content []byte) error {
	switch typ {
	case ai_api.TypRIASEC:
		dbErrR := dbSrv.Instance().SaveRiasecSession(bgCtx, publicId, businessTyp, content)
		if dbErrR != nil {
			s.log.Err(dbErrR).Str("channel", publicId).Msg("save questions error")
			return dbErrR
		}
		return nil
	default:
		s.log.Info().Str("ai-testType", string(typ)).Msg("Invalid testType")
		return fmt.Errorf("invalid testType")
	}
}

func sendSafe(ch chan *SSEMessage, msg *SSEMessage, log *zerolog.Logger) {
	defer func() { _ = recover() }()
	select {
	case ch <- msg:
		return
	default:
		log.Debug().Msg("client is close when generating ai questions")
	}
}
