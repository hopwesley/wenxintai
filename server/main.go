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

	"github.com/hopwesley/wenxintai/server/assessment"
	"github.com/hopwesley/wenxintai/server/internal/service"
	"github.com/hopwesley/wenxintai/server/internal/store"
	"github.com/hopwesley/wenxintai/server/internal/stream"
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
	loadHobbiesFromDB(context.Background(), repo)

	broker := stream.NewBroker(100, 2*time.Minute)
	defer broker.Stop()
	svc := service.NewSvc(repo, broker)
	inviteSvc := service.NewInviteService(repo)
	sseHandler := stream.NewSSEHandler(repo, broker)
	wsHandler := stream.NewWSHandler(repo, broker)
	api := newAPIHandler(svc, inviteSvc, sseHandler)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/assessments", api.handleAssessments)
	mux.HandleFunc("/api/assessments/", api.handleAssessmentDetail)
	mux.HandleFunc("/api/questions", api.creatingQuestionFromAI)
	mux.HandleFunc("/api/question_sets/", api.handleQuestionSetAnswers)
	mux.HandleFunc("/api/invites/verify", api.handleInviteVerify)
	mux.HandleFunc("/api/invites/redeem", api.handleInviteRedeem)
	mux.HandleFunc("/ws/assessments/", wsHandler)
	mux.HandleFunc("/api/hobbies", api.handleHobbies) // ← 新增
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/api/test_flow", api.handleTestFlow)

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

func loadHobbiesFromDB(ctx context.Context, repo store.Repo) {
	names, err := repo.ListHobbies(ctx)
	if err != nil {
		log.Printf("load hobbies from DB failed: %v; fallback to built-in (%d)", err, len(assessment.StudentHobbies))
		return
	}
	if len(names) == 0 {
		log.Printf("no hobbies found in DB; using built-in defaults (%d)", len(assessment.StudentHobbies))
		return
	}
	assessment.StudentHobbies = names
	log.Printf("loaded %d hobbies from DB", len(names))
}

type apiHandler struct {
	svc        *service.Svc
	invites    *service.InviteService
	sseHandler http.Handler
}

func newAPIHandler(svc *service.Svc, invites *service.InviteService, sse http.Handler) *apiHandler {
	return &apiHandler{svc: svc, invites: invites, sseHandler: sse}
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

// server/main.go 内，与其它 apiHandler 方法同级新增
func (h *apiHandler) handleHobbies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"hobbies": assessment.StudentHobbies,
	})
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
	AssessmentID string  `json:"assessment_id"`
	ReportID     *string `json:"report_id,omitempty"`
	Status       string  `json:"status"`
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

	nextStage, submitResult, err := h.svc.SubmitAnswersAndAdvance(r.Context(), questionSetID, req.Answers)
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

	if submitResult != nil {
		qs, err := h.svc.GetQuestionSet(r.Context(), questionSetID)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		status := "generating"
		if submitResult.ReportID != nil {
			status = "ready"
		}
		resp := submitS2Response{AssessmentID: qs.AssessmentID, ReportID: submitResult.ReportID, Status: status}
		writeJSON(w, http.StatusOK, resp)
		return
	}

	writeServiceError(w, newError(service.ErrorCodeInternal, "unknown workflow state", nil))
}

type inviteVerifyRequest struct {
	Code      string  `json:"code"`
	SessionID *string `json:"session_id,omitempty"`
}

type inviteVerifyResponse struct {
	SessionID     string `json:"session_id"`
	Status        string `json:"status"`
	ReservedUntil string `json:"reserved_until"`
}

func (h *apiHandler) handleInviteVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}
	if h.invites == nil {
		writeServiceError(w, newError(service.ErrorCodeInternal, "invite service unavailable", nil))
		return
	}
	var req inviteVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeServiceError(w, newError(service.ErrorCodeBadRequest, "invalid request body", err))
		return
	}
	result, err := h.invites.Verify(r.Context(), req.Code, req.SessionID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	resp := inviteVerifyResponse{
		SessionID:     result.SessionID,
		Status:        result.Status,
		ReservedUntil: result.ReservedUntil.UTC().Format(time.RFC3339),
	}
	writeJSON(w, http.StatusOK, resp)
}

type inviteRedeemRequest struct {
	SessionID string `json:"session_id"`
}

func (h *apiHandler) handleInviteRedeem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}
	if h.invites == nil {
		writeServiceError(w, newError(service.ErrorCodeInternal, "invite service unavailable", nil))
		return
	}
	var req inviteRedeemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeServiceError(w, newError(service.ErrorCodeBadRequest, "invalid request body", err))
		return
	}
	result, err := h.invites.Redeem(r.Context(), req.SessionID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "status": result.Status})
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
	case "events":
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
			return
		}
		h.sseHandler.ServeHTTP(w, r)
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

type createQuestionsRequest struct {
	SessionID string `json:"session_id"`
	Mode      string `json:"mode"`
	Grade     string `json:"grade"`
	Hobby     string `json:"hobby"`
}

type createQuestionsResponse struct {
	Status       string `json:"status"`        // "processing" / "ready" 等
	AssessmentID string `json:"assessment_id"` // 用于后续 SSE 订阅
	EventsURL    string `json:"events_url"`    // 前端直接拿来 new EventSource
}

func (h *apiHandler) creatingQuestionFromAI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	var req createQuestionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeServiceError(w, newError(service.ErrorCodeBadRequest, "invalid request body", err))
		return
	}

	err := h.svc.StartQuestionGeneration(
		r.Context(),
		req.SessionID,
		req.Mode,
		req.Grade,
		req.Hobby,
	)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	resp := createQuestionsResponse{
		Status:       "processing",
		AssessmentID: req.SessionID,
		EventsURL:    "/api/assessments/" + req.SessionID + "/events",
	}
	// 可以用 200 或者 202，这里更语义化一点用 202 Accepted
	writeJSON(w, http.StatusAccepted, resp)
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
		case service.ErrorCodeInviteReserved:
			status = http.StatusConflict
		case service.ErrorCodeInviteDisabled, service.ErrorCodeInviteRedeemed:
			status = http.StatusGone
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

// ----- /api/test_flow 请求 & 响应 -----

// 前端发来的请求体
type testFlowRequest struct {
	TestType     string  `json:"test_type"`
	InviteCode   *string `json:"invite_code,omitempty"`
	WechatOpenID *string `json:"wechat_openid,omitempty"`
}

// 单个测试步骤
type testRouteDef struct {
	Router string `json:"router"` // 英文路由名，例如 basic-info / riasec / asc / report
	Desc   string `json:"desc"`   // 中文描述，例如 基本信息 / 兴趣测试 / 能力测试 / 测试报告
}

// 下一步题目从哪里来
type questionSource string

const (
	questionSourceDB questionSource = "db" // 从数据库读取题目
	questionSourceAI questionSource = "ai" // 需要 AI 生成题目
)

// 下一步路由信息
type nextRouteInfo struct {
	Router string         `json:"router"` // 英文路由名
	Source questionSource `json:"source"` // 'db' or 'ai'
}

// /api/test_flow 响应体
type testFlowResponse struct {
	TestType  string         `json:"test_type"`
	Routes    []testRouteDef `json:"routes"`
	NextRoute *nextRouteInfo `json:"nextRoute,omitempty"`
}

// /api/test_flow: 根据 test_type + 身份信息，返回整套测试流程 + 下一步要进入的路由
func (h *apiHandler) handleTestFlow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	var req testFlowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeServiceError(w, newError(service.ErrorCodeBadRequest, "invalid request body", err))
		return
	}

	// test_type 必填
	if strings.TrimSpace(req.TestType) == "" {
		writeServiceError(w, newError(service.ErrorCodeBadRequest, "test_type is required", nil))
		return
	}

	// 身份：邀请码或微信至少有一个（后面我们用它去查 tests_record）
	if (req.InviteCode == nil || strings.TrimSpace(*req.InviteCode) == "") &&
		(req.WechatOpenID == nil || strings.TrimSpace(*req.WechatOpenID) == "") {
		writeServiceError(w, newError(service.ErrorCodeBadRequest, "invite_code or wechat_openid is required", nil))
		return
	}

	// 1) 根据 test_type 构建完整的测试流程 routes
	routes := buildTestRoutes(req.TestType)

	// 2) 计算 nextRoute：当前先不查数据库，统一返回 nil
	//    等 tests_record + 各量表运行记录表建好后，在这里接 service 计算实际进度。
	var next *nextRouteInfo

	// 示例：将来可以大致这样写（这里只是结构示意，暂不启用）：
	// next, err := h.svc.ComputeNextRoute(r.Context(), req.TestType, req.InviteCode, req.WechatOpenID)
	// if err != nil { ... }

	resp := testFlowResponse{
		TestType:  req.TestType,
		Routes:    routes,
		NextRoute: next,
	}
	writeJSON(w, http.StatusOK, resp)
}

// 根据测试类型构建完整的测试流程。
// 规则：basic-info 一定在最前，report 一定在最后，中间根据 test_type 不同插入不同的量表。
func buildTestRoutes(testType string) []testRouteDef {
	// 基本信息 & 报告，这两个是固定的
	basic := testRouteDef{Router: "basic-info", Desc: "基本信息"}
	report := testRouteDef{Router: "report", Desc: "测试报告"}

	var middle []testRouteDef

	switch testType {
	case "basic":
		// 基础版：RIASEC + ASC
		middle = []testRouteDef{
			{Router: "riasec", Desc: "兴趣测试"},
			{Router: "asc", Desc: "能力测试"},
		}
	case "pro":
		// 举例：专业版增加更多量表，后面可以按实际再调整
		middle = []testRouteDef{
			{Router: "riasec", Desc: "兴趣测试"},
			{Router: "asc", Desc: "能力测试"},
			{Router: "big5", Desc: "人格测试"},
			{Router: "motivation", Desc: "学习动机测试"},
		}
	default:
		// 默认：至少给出基础流程，避免前端拿不到任何路由
		middle = []testRouteDef{
			{Router: "riasec", Desc: "兴趣测试"},
			{Router: "asc", Desc: "能力测试"},
		}
	}

	routes := make([]testRouteDef, 0, len(middle)+2)
	routes = append(routes, basic)
	routes = append(routes, middle...)
	routes = append(routes, report)
	return routes
}
