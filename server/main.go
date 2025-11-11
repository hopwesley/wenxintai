package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/hopwesley/wenxintai/server/internal/service"
	"github.com/hopwesley/wenxintai/server/internal/store"
)

func main() {
	cfg, err := loadAppConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := connectDatabase(cfg.Database)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer db.Close()

	repo := store.NewSQLRepo(db)
	svc := service.NewSvc(repo)
	api := newAPIHandler(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/assessments", api.handleAssessments)
	mux.HandleFunc("/api/assessments/", api.handleAssessmentDetail)
	mux.HandleFunc("/api/question_sets/", api.handleQuestionSetAnswers)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	if stat, err := os.Stat(cfg.Server.StaticDir); err == nil && stat.IsDir() {
		fileServer := http.FileServer(http.Dir(cfg.Server.StaticDir))
		mux.Handle("/", fileServer)
	}

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      loggingMiddleware(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("HTTP server listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
	log.Println("server stopped")
}

type apiHandler struct {
	svc *service.Svc
}

func newAPIHandler(svc *service.Svc) *apiHandler {
	return &apiHandler{svc: svc}
}

type createAssessmentRequest struct {
	Mode         string  `json:"mode"`
	InviteCode   *string `json:"invite_code"`
	WechatOpenID *string `json:"wechat_openid"`
}

type createAssessmentResponse struct {
	AssessmentID  string          `json:"assessment_id"`
	QuestionSetID string          `json:"question_set_id"`
	Stage         string          `json:"stage"`
	Questions     json.RawMessage `json:"questions"`
}

func (h *apiHandler) handleAssessments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	var req createAssessmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeServiceError(w, newError(service.ErrorCodeBadRequest, "invalid request body", err))
		return
	}

	assessmentID, questionSetID, questions, err := h.svc.CreateAssessment(r.Context(), req.Mode, req.InviteCode, req.WechatOpenID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	resp := createAssessmentResponse{
		AssessmentID:  assessmentID,
		QuestionSetID: questionSetID,
		Stage:         "S1",
		Questions:     questions,
	}
	writeJSON(w, http.StatusOK, resp)
}

type submitAnswersRequest struct {
	Answers json.RawMessage `json:"answers"`
}

type submitS1Response struct {
	NextQuestionSetID string          `json:"next_question_set_id"`
	Stage             string          `json:"stage"`
	Questions         json.RawMessage `json:"questions"`
}

type submitS2Response struct {
	AssessmentID string `json:"assessment_id"`
	ReportID     string `json:"report_id"`
}

type reportResponse struct {
	ReportID string          `json:"report_id"`
	Summary  *string         `json:"summary,omitempty"`
	Full     json.RawMessage `json:"full"`
}

type progressResponse struct {
	Status int16  `json:"status"`
	Label  string `json:"label"`
}

func (h *apiHandler) handleQuestionSetAnswers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}
	trimmed := strings.TrimPrefix(r.URL.Path, "/api/question_sets/")
	parts := strings.Split(trimmed, "/")
	if len(parts) != 2 || parts[1] != "answers" || parts[0] == "" {
		writeServiceError(w, newError(service.ErrorCodeNotFound, "resource not found", nil))
		return
	}
	questionSetID := parts[0]

	var req submitAnswersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeServiceError(w, newError(service.ErrorCodeBadRequest, "invalid request body", err))
		return
	}
	if len(req.Answers) == 0 {
		writeServiceError(w, newError(service.ErrorCodeBadRequest, "answers is required", nil))
		return
	}

	nextStage, reportID, err := h.svc.SubmitAnswersAndAdvance(r.Context(), questionSetID, req.Answers)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	if nextStage != nil {
		resp := submitS1Response{
			NextQuestionSetID: nextStage.QuestionSetID,
			Stage:             nextStage.Stage,
			Questions:         nextStage.Questions,
		}
		writeJSON(w, http.StatusOK, resp)
		return
	}

	if reportID != nil {
		qs, err := h.svc.GetQuestionSet(r.Context(), questionSetID)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		resp := submitS2Response{AssessmentID: qs.AssessmentID, ReportID: *reportID}
		writeJSON(w, http.StatusOK, resp)
		return
	}

	writeServiceError(w, newError(service.ErrorCodeInternal, "unknown workflow state", nil))
}

func (h *apiHandler) handleAssessmentDetail(w http.ResponseWriter, r *http.Request) {
	trimmed := strings.TrimPrefix(r.URL.Path, "/api/assessments/")
	parts := strings.Split(trimmed, "/")
	if len(parts) < 2 || parts[0] == "" {
		writeServiceError(w, newError(service.ErrorCodeNotFound, "resource not found", nil))
		return
	}
	assessmentID := parts[0]
	switch parts[1] {
	case "report":
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
			return
		}
		report, err := h.svc.GetReport(r.Context(), assessmentID)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		resp := reportResponse{ReportID: report.ID, Summary: report.Summary, Full: report.FullJSON}
		writeJSON(w, http.StatusOK, resp)
	case "progress":
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
			return
		}
		progress, err := h.svc.GetProgress(r.Context(), assessmentID)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, progressResponse{Status: progress.Status, Label: progress.Label})
	default:
		writeServiceError(w, newError(service.ErrorCodeNotFound, "resource not found", nil))
	}
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, map[string]string{
		"code":    code,
		"message": message,
	})
}

func writeServiceError(w http.ResponseWriter, err error) {
	if err == nil {
		writeError(w, http.StatusInternalServerError, string(service.ErrorCodeInternal), "internal error")
		return
	}
	var svcErr *service.Error
	status := http.StatusInternalServerError
	code := string(service.ErrorCodeInternal)
	message := err.Error()
	if errors.As(err, &svcErr) {
		code = string(svcErr.Code)
		message = svcErr.Message
		switch svcErr.Code {
		case service.ErrorCodeBadRequest:
			status = http.StatusBadRequest
		case service.ErrorCodeNotFound:
			status = http.StatusNotFound
		case service.ErrorCodeConflict:
			status = http.StatusConflict
		default:
			status = http.StatusInternalServerError
		}
	}
	writeError(w, status, code, message)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func newError(code service.ErrorCode, message string, err error) error {
	return &service.Error{Code: code, Message: message, Err: err}
}
