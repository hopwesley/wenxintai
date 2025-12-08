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

type WSMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
	Message string          `json:"message,omitempty"`
}

func parseTestIDFromWSPath(path string) (string, error) {
	if i := strings.Index(path, "?"); i >= 0 {
		path = path[:i]
	}

	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 4 {
		return "", fmt.Errorf("invalid path, want /api/ws/{question|report}/{id}, got: %s", path)
	}
	if parts[0] != "api" || parts[1] != "ws" {
		return "", fmt.Errorf("invalid ws path segments: %v", parts)
	}

	channel := parts[2]
	if channel != "question" && channel != "report" {
		return "", fmt.Errorf("invalid ws channel: %s", channel)
	}

	idStr := parts[3]
	if !IsValidPublicID(idStr) {
		return "", fmt.Errorf("无效的问卷编号: %s", idStr)
	}

	return idStr, nil
}

func (s *HttpSrv) handleQuestionWSEvent(w http.ResponseWriter, r *http.Request) {
	publicId, err := parseTestIDFromWSPath(r.URL.Path)
	if err != nil {
		s.log.Err(err).Msg("WebSocket channel parse failed")
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
		sLog.Err(rErr).Msg("WS channel query record failed")
		http.Error(w, "未找到测试问卷数据:"+rErr.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.checkPreviousStageIfReady(ctx, record, testType); err != nil {
		sLog.Err(err).Msg("WS previous stage check failed")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	conn, upErr := s.wsUpgrader.Upgrade(w, r)
	if upErr != nil {
		sLog.Err(upErr).Msg("upgrade websocket failed")
		return
	}
	defer conn.Close()

	msgCh := make(chan *SSEMessage, 64)
	go s.aiQuestionProcess(msgCh, publicId, testType)

	s.streamWS(ctx, publicId, msgCh, conn, sLog)
}

func (s *HttpSrv) handleReportWSEvent(w http.ResponseWriter, r *http.Request) {
	publicId, err := parseTestIDFromWSPath(r.URL.Path)
	if err != nil {
		s.log.Err(err).Msg("WebSocket channel parse failed")
		http.Error(w, "无效的问卷编号:"+err.Error(), http.StatusBadRequest)
		return
	}

	sLog := s.log.With().Str("public_id", publicId).Logger()

	conn, upErr := s.wsUpgrader.Upgrade(w, r)
	if upErr != nil {
		sLog.Err(upErr).Msg("upgrade websocket failed")
		return
	}
	defer conn.Close()

	msgCh := make(chan *SSEMessage, 64)
	go s.aiReportProcess(msgCh, publicId, sLog)

	s.streamWS(r.Context(), publicId, msgCh, conn, sLog)
}

func (s *HttpSrv) streamWS(
	ctx context.Context,
	channelID string,
	msgCh <-chan *SSEMessage,
	conn *wsConn,
	log zerolog.Logger,
) {
	heartbeat := time.NewTicker(time.Duration(s.miniCfg.HeartbeatInterval) * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-ctx.Done():
			_ = conn.WriteClose()
			return
		case msg, ok := <-msgCh:
			if !ok {
				_ = conn.WriteJSON(WSMessage{Type: "done"})
				log.Info().Str("channel", channelID).Msg("WebSocket channel closed: msgCh closed")
				return
			}

			wsMsg := convertSSEToWS(msg)
			if err := conn.WriteJSON(wsMsg); err != nil {
				log.Err(err).Str("channel", channelID).Msg("write websocket message failed")
				return
			}

			if wsMsg.Type == "done" || wsMsg.Type == "error" {
				return
			}
		case <-heartbeat.C:
			_ = conn.WritePing()
		}
	}
}

func convertSSEToWS(msg *SSEMessage) WSMessage {
	wsTyp := "data"
	switch msg.Typ {
	case SSE_MT_DONE:
		wsTyp = "done"
	case SSE_MT_ERROR:
		wsTyp = "error"
	}

	var payload json.RawMessage
	if json.Valid([]byte(msg.Msg)) {
		payload = json.RawMessage(msg.Msg)
	}

	return WSMessage{
		Type:    wsTyp,
		Payload: payload,
		Message: msg.Msg,
	}
}
