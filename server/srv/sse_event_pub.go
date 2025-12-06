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

func (acm *SSEMessage) SSEMsg() string {
	if len(acm.Typ) == 0 {
		acm.Typ = SSE_MT_DATA
	}

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

type QuestionsPayload struct {
	Questions json.RawMessage `json:"questions"`
	Answers   json.RawMessage `json:"answers,omitempty"`
}

func (s *HttpSrv) initSSE() error {
	return nil
}

func parseTestIDFromPath(path string) (string, error) {
	// 去掉 query string
	if i := strings.Index(path, "?"); i >= 0 {
		path = path[:i]
	}

	parts := strings.Split(strings.Trim(path, "/"), "/")
	// 现在期望结构：/api/sub/{channel}/{id}
	// 例如：/api/sub/question/xxxxx 或 /api/sub/report/xxxxx
	if len(parts) < 4 {
		return "", fmt.Errorf("invalid path, want /api/sub/{question|report}/{id}, got: %s", path)
	}
	if parts[0] != "api" || parts[1] != "sub" {
		return "", fmt.Errorf("invalid path segments: %v", parts)
	}

	channel := parts[2]
	if channel != "question" && channel != "report" {
		return "", fmt.Errorf("invalid sse channel: %s", channel)
	}

	idStr := parts[3]
	if !IsValidPublicID(idStr) {
		return "", fmt.Errorf("无效的问卷编号: %s", idStr)
	}
	return idStr, nil
}

func (s *HttpSrv) handleQuestionSSEEvent(w http.ResponseWriter, r *http.Request) {
	publicId, err := parseTestIDFromPath(r.URL.Path)
	if err != nil {
		s.log.Err(err).Msg("SSE channel parse failed")
		http.Error(w, "无效的问卷编号:"+err.Error(), http.StatusBadRequest)
		return
	}

	q := r.URL.Query()
	businessTyp := q.Get("business_type")
	testType := ai_api.TestTyp(q.Get("test_type"))

	sLog := s.log.With().Str("public_id", publicId).
		Str("business_type", businessTyp).
		Str("test_type", string(testType)).Logger()

	ctx := r.Context()

	uid := userIDFromContext(ctx)
	record, rErr := dbSrv.Instance().QueryTestRecord(ctx, publicId, uid)
	if rErr != nil {
		s.log.Err(rErr).Msg("SSE channel query record failed")
		http.Error(w, "未找到测试问卷数据:"+rErr.Error(), http.StatusInternalServerError)
		return
	}

	err = s.checkPreviousStageIfReady(ctx, record, testType)
	if err != nil {
		sLog.Err(err).Msg("SSE previous stage check failed")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		sLog.Err(err).Msg("SSE channel created error")
		http.Error(w, "不支持流式数据传输", http.StatusInternalServerError)
		return
	}

	s.log.Info().
		Str("channel", publicId).
		Str("business_type", businessTyp).
		Str("testType", string(testType)).
		Msg("SSE channel created")

	msgCh := make(chan *SSEMessage, 64)

	go s.aiQuestionProcess(msgCh, publicId, testType)

	s.streamSSE(ctx, publicId, msgCh, w, flusher)
}

// streamSSE 会从 msgCh 读取消息，通过 writeSSE 持续写到客户端，直到：
// 1) ctx 取消；或 2) msgCh 关闭；或 3) 写入出错。
func (s *HttpSrv) streamSSE(
	ctx context.Context,
	channelID string,
	msgCh <-chan *SSEMessage,
	w http.ResponseWriter,
	flusher http.Flusher,
) {
	for {
		select {
		case <-ctx.Done():
			s.log.Info().
				Str("channel", channelID).
				Msg("SSE channel closed by client")
			return

		case msg, ok := <-msgCh:
			if !ok {
				s.log.Info().
					Str("channel", channelID).
					Msg("SSE channel closed: msgCh closed")
				return
			}
			if err := writeSSE(w, flusher, msg, &s.log); err != nil {
				s.log.Err(err).
					Str("channel", channelID).
					Msg("writeSSE failed, stop streaming")
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

func (s *HttpSrv) aiQuestionProcess(msgCh chan *SSEMessage, publicId string, aiTestType ai_api.TestTyp) {

	sLog := s.log.With().Str("channel", publicId).Str("ai_Type", string(aiTestType)).Logger()
	defer close(msgCh)

	bgCtx := context.Background()
	sLog.Info().Msg("start ai process")
	dbQuestion, err := dbSrv.Instance().FindQASession(bgCtx, string(aiTestType), publicId)
	if err != nil {
		sLog.Err(err).Msg("failed when find questions from database")
		msg := &SSEMessage{Msg: err.Error(), Typ: SSE_MT_ERROR}
		sendSafe(msgCh, msg, &s.log)
		return
	}

	if dbQuestion != nil {
		sLog.Info().Msg("found questions from database")
		payload := QuestionsPayload{
			Questions: dbQuestion.Questions,
			Answers:   dbQuestion.Answers,
		}
		buf, _ := json.Marshal(payload)
		msg := &SSEMessage{Msg: string(buf), Typ: SSE_MT_DONE}
		sendSafe(msgCh, msg, &s.log)
		return
	}

	bi, dbErr := dbSrv.Instance().QueryRecordBasicInfo(bgCtx, publicId)
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
		msg := &SSEMessage{Msg: "AI 生成 QA 试卷失败：" + aiErr.Error(), Typ: SSE_MT_ERROR}
		sendSafe(msgCh, msg, &s.log)
		return
	}

	s.log.Info().Msg("AI generate question success")

	if err := dbSrv.Instance().SaveQuestion(bgCtx, string(aiTestType), publicId, json.RawMessage(testContent)); err != nil {
		sLog.Err(err).Msg("保存 QA 试卷失败")
		msg := &SSEMessage{Msg: "保存 QA 试卷失败：" + err.Error(), Typ: SSE_MT_ERROR}
		sendSafe(msgCh, msg, &s.log)
		return
	}

	payload := QuestionsPayload{
		Questions: json.RawMessage(testContent),
	}

	buf, _ := json.Marshal(payload)
	msg := &SSEMessage{Msg: string(buf), Typ: SSE_MT_DONE}
	sendSafe(msgCh, msg, &s.log)

	sLog.Info().Msg("GenerateQuestion finished and saved")
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

func (s *HttpSrv) handleReportSSEEvent(w http.ResponseWriter, r *http.Request) {
	publicId, err := parseTestIDFromPath(r.URL.Path)
	if err != nil {
		s.log.Err(err).Msg("SSE channel parse failed")
		http.Error(w, "无效的问卷编号:"+err.Error(), http.StatusBadRequest)
		return
	}

	sLog := s.log.With().Str("channel", publicId).Logger()

	ctx := r.Context()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		sLog.Err(err).Msg("SSE channel created error")
		http.Error(w, "不支持流式数据传输", http.StatusInternalServerError)
		return
	}

	msgCh := make(chan *SSEMessage, 64)

	go s.aiReportProcess(msgCh, publicId, sLog)
	s.streamSSE(ctx, publicId, msgCh, w, flusher)
}

func (s *HttpSrv) aiReportProcess(msgCh chan *SSEMessage, publicId string, sLog zerolog.Logger) {
	bgCtx := context.Background()

	report, dbErr := dbSrv.Instance().QueryReportByPublicId(bgCtx, publicId)
	if dbErr != nil || report == nil {
		sLog.Err(dbErr).Msg("find finished report failed")
		sendSafe(msgCh, &SSEMessage{
			Typ: SSE_MT_ERROR,
			Msg: "查询报告数据库记录失败:",
		}, &s.log)
		return
	}

	if report.Status == dbSrv.ReportStatusSuccess {
		if report.AIContent == nil {
			sLog.Error().Msg("AIContent is nil")
			sendSafe(msgCh, &SSEMessage{Typ: SSE_MT_ERROR, Msg: "AI报告内容丢失:"}, &s.log)
			return
		}
		sendSafe(msgCh, &SSEMessage{Typ: SSE_MT_DONE, Msg: string(report.AIContent)}, &s.log)
		sLog.Info().Msg("got generated success")
		return
	}

	if report.ModeParam == nil {
		sLog.Error().Msg("ModeParam is nil")
		sendSafe(msgCh, &SSEMessage{Typ: SSE_MT_ERROR, Msg: "AI报告需要的选科参数缺失:"}, &s.log)
		return
	}

	var common ai_api.FullScoreResult
	if err := json.Unmarshal(report.CommonScore, &common); err != nil {
		sLog.Err(err).Msg("failed to unmarshal CommonScore")
		sendSafe(msgCh, &SSEMessage{Typ: SSE_MT_ERROR, Msg: "AI报告需要的基础参数缺失:"}, &s.log)
		return
	}

	callback := func(token string) error {
		msg := &SSEMessage{Msg: token, Typ: SSE_MT_DATA}
		sendSafe(msgCh, msg, &s.log)
		return nil
	}

	var paramMode interface{} = nil
	switch ai_api.Mode(report.Mode) {
	case ai_api.Mode33:
		var param ai_api.Mode33Section
		if jErr := json.Unmarshal(report.ModeParam, &param); jErr != nil {
			sLog.Err(jErr).Msg("failed to unmarshal ModeParam 3+3 data")
			sendSafe(msgCh, &SSEMessage{Typ: SSE_MT_ERROR, Msg: "解析AI报告(3+3)需要的参数失败:" + jErr.Error()}, &s.log)
			return
		}
		paramMode = &param
	case ai_api.Mode312:
		var param ai_api.Mode312Section
		if jErr := json.Unmarshal(report.ModeParam, &param); jErr != nil {
			sLog.Err(jErr).Msg("failed to unmarshal ModeParam 3+1+2 data")
			sendSafe(msgCh, &SSEMessage{Typ: SSE_MT_ERROR, Msg: "解析AI报告(3+1+2)需要的参数失败:" + jErr.Error()}, &s.log)
			return
		}
		paramMode = &param
	default:
		sLog.Error().Msg("param mode is invalid:" + report.Mode)
		sendSafe(msgCh, &SSEMessage{Typ: SSE_MT_ERROR, Msg: "无效的选科模式参数:" + report.Mode}, &s.log)
		return
	}

	aiContent, err := ai_api.Instance().GenerateUnifiedReport(bgCtx, common.Common, paramMode, ai_api.Mode(report.Mode), callback)
	if err != nil {
		sLog.Err(err).Msg("GenerateReportMod33 failed")
		sendSafe(msgCh, &SSEMessage{Typ: SSE_MT_ERROR, Msg: "生成报告(3+3)失败:" + err.Error()}, &s.log)
		return
	}

	dbErr = dbSrv.Instance().UpdateReportAIContent(bgCtx, publicId, []byte(aiContent))
	if dbErr != nil {
		sLog.Err(dbErr).Msg("UpdateReportAIContent failed")
		sendSafe(msgCh, &SSEMessage{Typ: SSE_MT_ERROR, Msg: "保存报告数据失败:" + dbErr.Error()}, &s.log)
		return
	}

	sendSafe(msgCh, &SSEMessage{Typ: SSE_MT_DONE, Msg: aiContent}, &s.log)

	sLog.Info().Msg("get ai-generated report success")
}
