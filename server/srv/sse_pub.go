package srv

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
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

func (s *HttpSrv) handleSSEEvent(w http.ResponseWriter, r *http.Request) {
	channelID, err := parseTestIDFromPath(r.URL.Path)
	if err != nil {
		s.log.Err(err).Msg("SSE channel parse failed")
		writeError(w, ApiInvalidReq("invalid test id", err))
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		s.log.Err(err).Str("channel", channelID).Msg("SSE channel created error")
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	s.log.Info().Str("channel", channelID).Msg("SSE channel created")

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
