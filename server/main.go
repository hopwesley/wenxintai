package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/hopwesley/wenxintai/server/assessment"
)

// pipelineServer and related types define the in-memory state machine used by
// the RIASEC/ASC assessment API. These definitions mirror the original
// implementation in the upstream repository.
type pipelineServer struct {
	mu            sync.RWMutex
	sessions      map[string]*sessionState
	wechatIndex   map[string]string
	defaultAPIKey string
	db            *sql.DB
}

type sessionState struct {
	ID            string                       `json:"session_id"`
	WeChatID      string                       `json:"wechat_id"`
	Nickname      string                       `json:"nickname"`
	AvatarURL     string                       `json:"avatar_url"`
	CreatedAt     time.Time                    `json:"created_at"`
	UpdatedAt     time.Time                    `json:"updated_at"`
	Mode          assessment.Mode              `json:"mode,omitempty"`
	Gender        string                       `json:"gender,omitempty"`
	Grade         string                       `json:"grade,omitempty"`
	Hobby         string                       `json:"hobby,omitempty"`
	Questions     *assessment.QuestionsResult  `json:"questions,omitempty"`
	Answers       *answersPayload              `json:"answers,omitempty"`
	Param         *assessment.ParamForAIPrompt `json:"param,omitempty"`
	Radar         *assessment.RadarData        `json:"radar,omitempty"`
	SubjectScores []assessment.SubjectScores   `json:"subject_scores,omitempty"`
	Report        json.RawMessage              `json:"report,omitempty"`
}

type answersPayload struct {
	RIASEC []assessment.RIASECAnswer `json:"riasec_answers"`
	ASC    []assessment.ASCAnswer    `json:"asc_answers"`
	Alpha  float64                   `json:"alpha"`
	Beta   float64                   `json:"beta"`
	Gamma  float64                   `json:"gamma"`
}

type loginRequest struct {
	WeChatID  string `json:"wechat_id"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}

type loginResponse struct {
	SessionID string    `json:"session_id"`
	WeChatID  string    `json:"wechat_id"`
	Nickname  string    `json:"nickname"`
	AvatarURL string    `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type questionsRequest struct {
	SessionID string `json:"session_id"`
	Mode      string `json:"mode"`
	Gender    string `json:"gender"`
	Grade     string `json:"grade"`
	Hobby     string `json:"hobby"`
	APIKey    string `json:"api_key,omitempty"`
}

type questionsResponse struct {
	SessionID string                      `json:"session_id"`
	Questions *assessment.QuestionsResult `json:"questions"`
}

type answersRequest struct {
	SessionID     string                    `json:"session_id"`
	Mode          string                    `json:"mode"`
	RIASECAnswers []assessment.RIASECAnswer `json:"riasec_answers"`
	ASCAnswers    []assessment.ASCAnswer    `json:"asc_answers"`
	Alpha         float64                   `json:"alpha,omitempty"`
	Beta          float64                   `json:"beta,omitempty"`
	Gamma         float64                   `json:"gamma,omitempty"`
}

type answersResponse struct {
	SessionID     string                       `json:"session_id"`
	Param         *assessment.ParamForAIPrompt `json:"param"`
	Radar         *assessment.RadarData        `json:"radar"`
	SubjectScores []assessment.SubjectScores   `json:"subject_scores"`
}

type reportRequest struct {
	SessionID string `json:"session_id"`
	Mode      string `json:"mode"`
	APIKey    string `json:"api_key,omitempty"`
}

type reportResponse struct {
	SessionID string          `json:"session_id"`
	Report    json.RawMessage `json:"report"`
}

type hobbiesResponse struct {
	Hobbies []string `json:"hobbies"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type inviteVerifyRequest struct {
	Code string `json:"code"`
}

type inviteVerifyResponse struct {
	OK     bool   `json:"ok"`
	Reason string `json:"reason,omitempty"`
}

func newPipelineServer(defaultKey string, db *sql.DB) *pipelineServer {
	return &pipelineServer{
		sessions:      make(map[string]*sessionState),
		wechatIndex:   make(map[string]string),
		defaultAPIKey: defaultKey,
		db:            db,
	}
}

// routes sets up API endpoints on a new ServeMux. The handler returned here
// responds only to paths beginning with /api/.
func (s *pipelineServer) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", s.wrap(http.MethodPost, s.handleLogin))
	mux.HandleFunc("/api/hobbies", s.wrap(http.MethodGet, s.handleHobbies))
	mux.HandleFunc("/api/questions", s.wrap(http.MethodPost, s.handleQuestions))
	mux.HandleFunc("/api/answers", s.wrap(http.MethodPost, s.handleAnswers))
	mux.HandleFunc("/api/report", s.wrap(http.MethodPost, s.handleReport))
	mux.HandleFunc("/api/invites/verify-and-redeem", s.wrap(http.MethodPost, s.handleInviteVerify))
	return mux
}

func (s *pipelineServer) wrap(method string, handler func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		w.Header().Set("Content-Type", "application/json")
		if method != "" && r.Method != method {
			s.writeError(w, http.StatusMethodNotAllowed, "只支持 "+method+" 请求")
			return
		}
		handler(w, r)
		log.Printf("[DONE] %s in %v", r.URL.Path, time.Since(start))
	}
}

func (s *pipelineServer) writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.WriteHeader(status)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}

func (s *pipelineServer) writeError(w http.ResponseWriter, status int, msg string) {
	s.writeJSON(w, status, errorResponse{Error: msg})
}

func (s *pipelineServer) resolveAPIKey(candidate string) (string, bool) {
	if candidate != "" {
		return candidate, true
	}
	if s.defaultAPIKey != "" {
		return s.defaultAPIKey, true
	}
	return "", false
}

func (s *pipelineServer) getSession(id string) (*sessionState, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[id]
	return sess, ok
}

// handleLogin manages user login. It will create a new session if the given
// WeChat ID has not been seen before, otherwise it updates the existing session.
func (s *pipelineServer) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "无效的请求体")
		return
	}
	if req.WeChatID == "" {
		s.writeError(w, http.StatusBadRequest, "缺少微信标识")
		return
	}
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()

	if existingID, ok := s.wechatIndex[req.WeChatID]; ok {
		sess := s.sessions[existingID]
		sess.Nickname = req.Nickname
		sess.AvatarURL = req.AvatarURL
		sess.UpdatedAt = now
		resp := loginResponse{
			SessionID: sess.ID,
			WeChatID:  sess.WeChatID,
			Nickname:  sess.Nickname,
			AvatarURL: sess.AvatarURL,
			CreatedAt: sess.CreatedAt,
			UpdatedAt: sess.UpdatedAt,
		}
		s.writeJSON(w, http.StatusOK, resp)
		return
	}

	sessionID := uuid.NewString()
	sess := &sessionState{
		ID:        sessionID,
		WeChatID:  req.WeChatID,
		Nickname:  req.Nickname,
		AvatarURL: req.AvatarURL,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.sessions[sessionID] = sess
	s.wechatIndex[req.WeChatID] = sessionID

	resp := loginResponse{
		SessionID: sessionID,
		WeChatID:  req.WeChatID,
		Nickname:  req.Nickname,
		AvatarURL: req.AvatarURL,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.writeJSON(w, http.StatusOK, resp)
}

// handleHobbies returns a list of hobbies. These values originate from
// assessment.StudentHobbies.
func (s *pipelineServer) handleHobbies(w http.ResponseWriter, _ *http.Request) {
	s.writeJSON(w, http.StatusOK, hobbiesResponse{Hobbies: assessment.StudentHobbies})
}

func (s *pipelineServer) handleInviteVerify(w http.ResponseWriter, r *http.Request) {
	if s.db == nil {
		log.Println("[INVITE] 数据库未初始化")
		s.writeJSON(w, http.StatusInternalServerError, inviteVerifyResponse{OK: false})
		return
	}

	var req inviteVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	code := strings.TrimSpace(req.Code)
	if code == "" {
		s.writeJSON(w, http.StatusOK, inviteVerifyResponse{OK: false, Reason: "not_found"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	reason, err := s.redeemInvite(ctx, code)
	if err != nil {
		log.Printf("[INVITE] code=%s error=%v", maskInviteCode(code), err)
		s.writeJSON(w, http.StatusInternalServerError, inviteVerifyResponse{OK: false})
		return
	}

	if reason != "" {
		log.Printf("[INVITE] code=%s result=%s", maskInviteCode(code), reason)
		s.writeJSON(w, http.StatusOK, inviteVerifyResponse{OK: false, Reason: reason})
		return
	}

	log.Printf("[INVITE] code=%s result=ok", maskInviteCode(code))
	s.writeJSON(w, http.StatusOK, inviteVerifyResponse{OK: true})
}

// handleQuestions generates a set of questions for the given session and
// assessment mode.
func (s *pipelineServer) handleQuestions(w http.ResponseWriter, r *http.Request) {
	var req questionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "无效的请求体")
		return
	}
	sess, ok := s.getSession(req.SessionID)
	if !ok {
		s.writeError(w, http.StatusNotFound, "未找到会话")
		return
	}

	mode, valid := assessment.ParseMode(req.Mode)
	if !valid {
		s.writeError(w, http.StatusBadRequest, "无效的模式")
		return
	}

	apiKey, ok := s.resolveAPIKey(req.APIKey)
	if !ok {
		s.writeError(w, http.StatusBadRequest, "缺少 API 密钥")
		return
	}

	questions, err := assessment.GenerateQuestions(mode, apiKey, req.Gender, req.Grade, req.Hobby)
	if err != nil {
		s.writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	s.mu.Lock()
	sess.Mode = mode
	sess.Gender = req.Gender
	sess.Grade = req.Grade
	sess.Hobby = req.Hobby
	sess.Questions = questions
	sess.UpdatedAt = time.Now()
	s.mu.Unlock()

	s.writeJSON(w, http.StatusOK, questionsResponse{
		SessionID: sess.ID,
		Questions: questions,
	})
}

// handleAnswers scores the provided answers and stores the result on the
// session state. It returns the computed parameters necessary for generating
// the report.
func (s *pipelineServer) handleAnswers(w http.ResponseWriter, r *http.Request) {
	var req answersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "无效的请求体")
		return
	}
	sess, ok := s.getSession(req.SessionID)
	if !ok {
		s.writeError(w, http.StatusNotFound, "未找到会话")
		return
	}
	mode, valid := assessment.ParseMode(req.Mode)
	if !valid {
		s.writeError(w, http.StatusBadRequest, "无效的模式")
		return
	}
	if len(req.RIASECAnswers) == 0 || len(req.ASCAnswers) == 0 {
		s.writeError(w, http.StatusBadRequest, "答案不能为空")
		return
	}

	out, err := assessment.Run(assessment.Input{
		RIASECAnswers: req.RIASECAnswers,
		ASCAnswers:    req.ASCAnswers,
		Alpha:         req.Alpha,
		Beta:          req.Beta,
		Gamma:         req.Gamma,
	}, mode)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "评分失败")
		return
	}

	param := out.Param
	result := out.Result
	scores := out.Scores

	s.mu.Lock()
	sess.Mode = mode
	sess.Answers = &answersPayload{
		RIASEC: req.RIASECAnswers,
		ASC:    req.ASCAnswers,
		Alpha:  req.Alpha,
		Beta:   req.Beta,
		Gamma:  req.Gamma,
	}
	sess.Param = param
	if result != nil {
		sess.Radar = result.Radar
	} else {
		sess.Radar = nil
	}
	sess.SubjectScores = scores
	sess.UpdatedAt = time.Now()
	s.mu.Unlock()

	s.writeJSON(w, http.StatusOK, answersResponse{
		SessionID:     sess.ID,
		Param:         param,
		Radar:         sess.Radar,
		SubjectScores: scores,
	})
}

// handleReport generates a unified AI report for the session using the stored
// parameters. It persists the report on the session.
func (s *pipelineServer) handleReport(w http.ResponseWriter, r *http.Request) {
	var req reportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "无效的请求体")
		return
	}
	sess, ok := s.getSession(req.SessionID)
	if !ok {
		s.writeError(w, http.StatusNotFound, "未找到会话")
		return
	}
	mode, valid := assessment.ParseMode(req.Mode)
	if !valid {
		s.writeError(w, http.StatusBadRequest, "无效的模式")
		return
	}
	if sess.Param == nil {
		s.writeError(w, http.StatusBadRequest, "尚未提交答案")
		return
	}

	apiKey, ok := s.resolveAPIKey(req.APIKey)
	if !ok {
		s.writeError(w, http.StatusBadRequest, "缺少 API 密钥")
		return
	}

	report, err := assessment.GenerateUnifiedReport(apiKey, *sess.Param, mode)
	if err != nil {
		s.writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	s.mu.Lock()
	sess.Report = report
	sess.UpdatedAt = time.Now()
	s.mu.Unlock()

	s.writeJSON(w, http.StatusOK, reportResponse{
		SessionID: sess.ID,
		Report:    report,
	})
}

func (s *pipelineServer) redeemInvite(ctx context.Context, code string) (string, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return "", err
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	var status string
	var expiresAt sql.NullTime
	err = tx.QueryRowContext(ctx, `SELECT status, expires_at FROM app.invites WHERE code = $1 FOR UPDATE`, code).Scan(&status, &expiresAt)
	if err == sql.ErrNoRows {
		return "not_found", nil
	}
	if err != nil {
		return "", err
	}

	now := time.Now()
	if status != "unused" {
		return "used", nil
	}
	if expiresAt.Valid && !expiresAt.Time.After(now) {
		return "expired", nil
	}

	if _, err := tx.ExecContext(ctx, `UPDATE app.invites SET status = 'used', used_by = $2, used_at = NOW() WHERE code = $1 AND status = 'unused'`, code, "public_portal"); err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}
	committed = true
	return "", nil
}

func maskInviteCode(code string) string {
	trimmed := strings.TrimSpace(code)
	length := len(trimmed)
	if length == 0 {
		return "(empty)"
	}
	if length <= 4 {
		return strings.Repeat("*", length)
	}
	return strings.Repeat("*", length-4) + trimmed[length-4:]
}

// main configures and starts the HTTP server. It mounts both the API handlers
// (under /api/) and a static file server for the compiled Vue SPA. When the
// frontend is built using `npm run build` the output should be placed in
// frontend/dist. Requests that do not begin with /api/ will be served by
// the static file server.
func main() {
	cfg, err := loadDatabaseConfig()
	if err != nil {
		log.Printf("数据库配置错误: %v", err)
		os.Exit(1)
	}

	db, err := connectDatabase(cfg)
	if err != nil {
		log.Printf("数据库连接失败: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	defaultKey := os.Getenv("DEEPSEEK_API_KEY")
	srv := newPipelineServer(defaultKey, db)

	mux := http.NewServeMux()
	// API 还是 /api/* 前缀
	mux.Handle("/api/", srv.routes())

	// ---- 新增：静态目录与 SPA fallback ----
	// 运行时可通过 STATIC_DIR 指定静态目录，默认 ./static
	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "./static"
	}

	// 简单的静态文件+SPA回退：
	// 1) 如果请求的物理文件存在且不是目录，直接返回该文件
	// 2) 否则返回 index.html，让前端路由（vue-router）处理
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		// 尝试物理文件
		tryPath := staticDir + r.URL.Path
		if fi, err := os.Stat(tryPath); err == nil && !fi.IsDir() {
			http.ServeFile(w, r, tryPath)
			return
		}
		// 回退到 index.html
		http.ServeFile(w, r, staticDir+"/index.html")
	})

	// 端口支持环境变量 PORT，默认 8080（开发友好；线上由 Nginx 反代 80/443）
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("🚀 Server running on http://localhost:" + port)
	log.Println("    static from:", staticDir)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
