package stream

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/hopwesley/wenxintai/server/comm"
	"github.com/hopwesley/wenxintai/server/internal/store"
)

func NewSSEHandler(repo store.Repo, broker *Broker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		assessmentID, ok := parseAssessmentPath(r.URL.Path, "events")
		if !ok {
			http.NotFound(w, r)
			return
		}
		if _, err := repo.GetAssessmentByID(r.Context(), assessmentID); err != nil {
			if errors.Is(err, comm.ErrNotFound) {
				http.NotFound(w, r)
				return
			}
			http.Error(w, "failed to load assessment", http.StatusInternalServerError)
			return
		}
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "stream unsupported", http.StatusInternalServerError)
			return
		}

		lastID := r.Header.Get("Last-Event-ID")
		if fromQuery := r.URL.Query().Get("last_event_id"); fromQuery != "" {
			lastID = fromQuery
		}

		events, cancel := broker.Subscribe(assessmentID, lastID)
		defer cancel()

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Connection", "keep-alive")
		flusher.Flush()

		ticker := time.NewTicker(20 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case evt, ok := <-events:
				if !ok {
					return
				}
				payload := struct {
					Type string          `json:"type"`
					Data json.RawMessage `json:"data"`
				}{Type: evt.Type, Data: evt.Data}
				encoded, err := json.Marshal(payload)
				if err != nil {
					continue
				}
				if evt.ID != "" {
					_, _ = w.Write([]byte("id: " + evt.ID + "\n"))
				}
				_, _ = w.Write([]byte("data: " + string(encoded) + "\n\n"))
				flusher.Flush()
			case <-ticker.C:
				_, _ = w.Write([]byte(": ping\n\n"))
				flusher.Flush()
			}
		}
	}
}

func parseAssessmentPath(path, suffix string) (string, bool) {
	trimmed := strings.TrimPrefix(path, "/api/assessments/")
	parts := strings.Split(trimmed, "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] != suffix {
		return "", false
	}
	return parts[0], true
}
