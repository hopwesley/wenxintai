package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	Status        sessionStatus                `json:"status,omitempty"`
	InviteCode    string                       `json:"invite_code,omitempty"`
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

type inviteVerifyRequest struct {
	Code      string `json:"code"`
	SessionID string `json:"session_id,omitempty"`
}

type inviteVerifyResponse struct {
	SessionID     string    `json:"session_id"`
	Status        string    `json:"status"`
	ReservedUntil time.Time `json:"reserved_until"`
}

type inviteRedeemRequest struct {
	SessionID string `json:"session_id,omitempty"`
}

type sessionStatus string

const (
	sessionStatusAnonymous       sessionStatus = "anonymous"
	sessionStatusInvitedVerified sessionStatus = "invited_verified"
	sessionStatusActive          sessionStatus = "active"
	sessionStatusUser            sessionStatus = "user"
)

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
	mux.HandleFunc("/api/invites/verify", s.wrap(http.MethodPost, s.handleInviteVerify))
	mux.HandleFunc("/api/invites/redeem", s.wrap(http.MethodPost, s.handleInviteRedeem))
	return mux
}

func (s *pipelineServer) wrap(method string, handler func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		w.Header().Set("Content-Type", "application/json")
		if method != "" && r.Method != method {
			writeErrorJSON(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "只支持 "+method+" 请求")
			return
		}
		handler(w, r)
		log.Printf("[DONE] %s in %v", r.URL.Path, time.Since(start))
	}
}

func (s *pipelineServer) writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}

func writeErrorJSON(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, _ = io.WriteString(w, fmt.Sprintf(`{"error":%q,"code":%q}`, msg, code))
}

type sessionIDContextKey struct{}

func (s *pipelineServer) getSessionFromRequest(r *http.Request) (string, *sessionState, bool) {
	// Cookie first
	if cookie, err := r.Cookie("sid"); err == nil {
		if v := strings.TrimSpace(cookie.Value); v != "" {
			if sess, ok := s.getSession(v); ok {
				return v, sess, true
			}
		}
	}

	// Authorization header: Bearer token
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		candidate := strings.TrimSpace(authHeader[7:])
		if candidate != "" {
			if sess, ok := s.getSession(candidate); ok {
				return candidate, sess, true
			}
		}
	}

	if v := strings.TrimSpace(r.Header.Get("X-Session-ID")); v != "" {
		if sess, ok := s.getSession(v); ok {
			return v, sess, true
		}
	}

	if v := strings.TrimSpace(r.URL.Query().Get("session_id")); v != "" {
		if sess, ok := s.getSession(v); ok {
			return v, sess, true
		}
	}

	if ctxVal := r.Context().Value(sessionIDContextKey{}); ctxVal != nil {
		if v, ok := ctxVal.(string); ok && strings.TrimSpace(v) != "" {
			candidate := strings.TrimSpace(v)
			if sess, ok := s.getSession(candidate); ok {
				return candidate, sess, true
			}
		}
	}

	return "", nil, false
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
		writeErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "无效的请求体")
		return
	}
	if req.WeChatID == "" {
		writeErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "缺少微信标识")
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
		sess.Status = sessionStatusUser
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
		Status:    sessionStatusUser,
	}
	s.sessions[sessionID] = sess
	s.wechatIndex[req.WeChatID] = sessionID

	setSessionCookie(w, sessionID)

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
		writeErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "邀请码功能不可用")
		return
	}

	var req inviteVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "无效的请求体")
		return
	}

	code := strings.TrimSpace(req.Code)
	if code == "" {
		writeErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "请输入邀请码")
		return
	}

	requestWithSession := r
	if strings.TrimSpace(req.SessionID) != "" {
		requestWithSession = r.WithContext(context.WithValue(r.Context(), sessionIDContextKey{}, strings.TrimSpace(req.SessionID)))
	}

	sid, sess, hasSession := s.getSessionFromRequest(requestWithSession)
	sessionID := sid
	if !hasSession {
		sessionID = strings.TrimSpace(req.SessionID)
		if sessionID == "" {
			sessionID = uuid.NewString()
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	reservedUntil, apiErr, err := s.reserveInvite(ctx, code, sessionID)
	if err != nil {
		log.Printf("[INVITE] code=%s error=%v", maskInviteCode(code), err)
		writeErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "邀请码处理失败")
		return
	}
	if apiErr != nil {
		log.Printf("[INVITE] code=%s result=%s", maskInviteCode(code), apiErr.Code)
		writeErrorJSON(w, apiErr.Status, apiErr.Code, apiErr.Message)
		return
	}

	now := time.Now().UTC()
	s.mu.Lock()
	if hasSession {
		sess.InviteCode = code
		sess.Status = sessionStatusInvitedVerified
		if sess.CreatedAt.IsZero() {
			sess.CreatedAt = now
		}
		sess.UpdatedAt = now
	} else {
		s.sessions[sessionID] = &sessionState{
			ID:         sessionID,
			CreatedAt:  now,
			UpdatedAt:  now,
			Status:     sessionStatusInvitedVerified,
			InviteCode: code,
		}
	}
	s.mu.Unlock()

	setSessionCookie(w, sessionID)
	s.writeJSON(w, http.StatusOK, inviteVerifyResponse{
		SessionID:     sessionID,
		Status:        "reserved",
		ReservedUntil: reservedUntil,
	})
}

func (s *pipelineServer) handleInviteRedeem(w http.ResponseWriter, r *http.Request) {
	if s.db == nil {
		log.Println("[INVITE] 数据库未初始化")
		writeErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "邀请码功能不可用")
		return
	}

	var req inviteRedeemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err == nil && strings.TrimSpace(req.SessionID) != "" {
		r = r.WithContext(context.WithValue(r.Context(), sessionIDContextKey{}, strings.TrimSpace(req.SessionID)))
	}

	sid, sess, ok := s.getSessionFromRequest(r)
	if !ok {
		writeErrorJSON(w, http.StatusUnauthorized, "NO_SESSION", "请先验证邀请码或登录")
		return
	}

	if sess.Status == sessionStatusActive || sess.Status == sessionStatusUser {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if sess.InviteCode == "" {
		writeErrorJSON(w, http.StatusForbidden, "INVITE_REQUIRED", "需要邀请码或登录后访问")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if apiErr, err := s.consumeInvite(ctx, sess); err != nil {
		log.Printf("[INVITE] redeem session=%s error=%v", sid, err)
		writeErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "邀请码处理失败")
		return
	} else if apiErr != nil {
		writeErrorJSON(w, apiErr.Status, apiErr.Code, apiErr.Message)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleQuestions generates a set of questions for the given session and
// assessment mode.
func (s *pipelineServer) handleQuestions(w http.ResponseWriter, r *http.Request) {
	var req questionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "无效的请求体")
		return
	}
	if strings.TrimSpace(req.SessionID) != "" {
		r = r.WithContext(context.WithValue(r.Context(), sessionIDContextKey{}, strings.TrimSpace(req.SessionID)))
	}

	sid, sess, ok := s.getSessionFromRequest(r)
	if !ok {
		writeErrorJSON(w, http.StatusUnauthorized, "NO_SESSION", "请先验证邀请码或登录")
		return
	}

	status := sess.Status
	if status != sessionStatusInvitedVerified && status != sessionStatusActive && status != sessionStatusUser {
		writeErrorJSON(w, http.StatusForbidden, "INVITE_REQUIRED", "需要邀请码或登录后访问")
		return
	}

	if status == sessionStatusInvitedVerified && sess.InviteCode != "" {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		if apiErr, err := s.consumeInvite(ctx, sess); err != nil {
			log.Printf("[INVITE] auto redeem session=%s error=%v", sid, err)
			writeErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "邀请码处理失败")
			return
		} else if apiErr != nil {
			writeErrorJSON(w, apiErr.Status, apiErr.Code, apiErr.Message)
			return
		}
	}

	// refresh session pointer after potential mutation
	sid, sess, _ = s.getSessionFromRequest(r)
	if sess == nil {
		writeErrorJSON(w, http.StatusUnauthorized, "NO_SESSION", "会话已失效")
		return
	}

	mode, valid := assessment.ParseMode(req.Mode)
	if !valid {
		writeErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "无效的模式")
		return
	}

	apiKey, ok := s.resolveAPIKey(req.APIKey)
	if !ok {
		writeErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "缺少 API 密钥")
		return
	}

	questions, err := assessment.GenerateQuestions(mode, apiKey, req.Gender, req.Grade, req.Hobby)
	if err != nil {
		writeErrorJSON(w, http.StatusBadGateway, "UPSTREAM_ERROR", err.Error())
		return
	}

	s.mu.Lock()
	sess.Mode = mode
	sess.Gender = req.Gender
	sess.Grade = req.Grade
	sess.Hobby = req.Hobby
	sess.Questions = questions
	if sess.Status == sessionStatusInvitedVerified {
		sess.Status = sessionStatusActive
	}
	sess.UpdatedAt = time.Now().UTC()
	s.mu.Unlock()

	s.writeJSON(w, http.StatusOK, questionsResponse{
		SessionID: sid,
		Questions: questions,
	})
}

// handleAnswers scores the provided answers and stores the result on the
// session state. It returns the computed parameters necessary for generating
// the report.
func (s *pipelineServer) handleAnswers(w http.ResponseWriter, r *http.Request) {
	var req answersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "无效的请求体")
		return
	}
	if strings.TrimSpace(req.SessionID) != "" {
		r = r.WithContext(context.WithValue(r.Context(), sessionIDContextKey{}, strings.TrimSpace(req.SessionID)))
	}
	sessID, sess, ok := s.getSessionFromRequest(r)
	if !ok {
		writeErrorJSON(w, http.StatusNotFound, "SESSION_NOT_FOUND", "未找到会话")
		return
	}
	mode, valid := assessment.ParseMode(req.Mode)
	if !valid {
		writeErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "无效的模式")
		return
	}
	if len(req.RIASECAnswers) == 0 || len(req.ASCAnswers) == 0 {
		writeErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "答案不能为空")
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
		writeErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "评分失败")
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
	sess.UpdatedAt = time.Now().UTC()
	s.mu.Unlock()

	s.writeJSON(w, http.StatusOK, answersResponse{
		SessionID:     sessID,
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
		writeErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "无效的请求体")
		return
	}
	if strings.TrimSpace(req.SessionID) != "" {
		r = r.WithContext(context.WithValue(r.Context(), sessionIDContextKey{}, strings.TrimSpace(req.SessionID)))
	}
	sessID, sess, ok := s.getSessionFromRequest(r)
	if !ok {
		writeErrorJSON(w, http.StatusNotFound, "SESSION_NOT_FOUND", "未找到会话")
		return
	}
	mode, valid := assessment.ParseMode(req.Mode)
	if !valid {
		writeErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "无效的模式")
		return
	}
	if sess.Param == nil {
		writeErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "尚未提交答案")
		return
	}

	apiKey, ok := s.resolveAPIKey(req.APIKey)
	if !ok {
		writeErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "缺少 API 密钥")
		return
	}

	report, err := assessment.GenerateUnifiedReport(apiKey, *sess.Param, mode)
	if err != nil {
		writeErrorJSON(w, http.StatusBadGateway, "UPSTREAM_ERROR", err.Error())
		return
	}

	s.mu.Lock()
	sess.Report = report
	sess.UpdatedAt = time.Now().UTC()
	s.mu.Unlock()

	s.writeJSON(w, http.StatusOK, reportResponse{
		SessionID: sessID,
		Report:    report,
	})
}

const (
	inviteStatusUnused   = 0
	inviteStatusReserved = 1
	inviteStatusUsed     = 2
	inviteStatusExpired  = 3
	inviteStatusRevoked  = 4
)

type apiError struct {
	Status  int
	Code    string
	Message string
}

func (e *apiError) Error() string { return e.Message }

func newAPIError(status int, code, msg string) *apiError {
	return &apiError{Status: status, Code: code, Message: msg}
}

func (s *pipelineServer) reserveInvite(ctx context.Context, code, sessionID string) (time.Time, *apiError, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return time.Time{}, nil, err
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	var status int
	var expiresAt sql.NullTime
	var reservedBy sql.NullString
	var reservedUntil sql.NullTime
	err = tx.QueryRowContext(ctx,
		`SELECT status, expires_at, reserved_by_session_id, reserved_until
                   FROM app.invites
                  WHERE code = $1 FOR UPDATE`,
		code,
	).Scan(&status, &expiresAt, &reservedBy, &reservedUntil)
	if errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, newAPIError(http.StatusNotFound, "INVITE_NOT_FOUND", "邀请码不存在"), nil
	}
	if err != nil {
		return time.Time{}, nil, err
	}

	now := time.Now().UTC()
	if expiresAt.Valid && !expiresAt.Time.After(now) {
		return time.Time{}, newAPIError(http.StatusForbidden, "INVITE_EXPIRED", "邀请码已过期"), nil
	}

	switch status {
	case inviteStatusUsed:
		return time.Time{}, newAPIError(http.StatusConflict, "INVITE_ALREADY_USED", "邀请码已被使用"), nil
	case inviteStatusExpired:
		return time.Time{}, newAPIError(http.StatusForbidden, "INVITE_EXPIRED", "邀请码已过期"), nil
	case inviteStatusRevoked:
		return time.Time{}, newAPIError(http.StatusForbidden, "INVITE_REVOKED", "邀请码已失效"), nil
	case inviteStatusReserved:
		if reservedUntil.Valid && !reservedUntil.Time.After(now) {
			if _, err := tx.ExecContext(ctx,
				`UPDATE app.invites
                                    SET status = $2, reserved_by_session_id = NULL, reserved_until = NULL
                                  WHERE code = $1`,
				code, inviteStatusUnused,
			); err != nil {
				return time.Time{}, nil, err
			}
			status = inviteStatusUnused
		} else if reservedBy.Valid && reservedBy.String != "" && reservedBy.String != sessionID {
			return time.Time{}, newAPIError(http.StatusConflict, "INVITE_RESERVED", "邀请码已被占用"), nil
		}
	}

	expires := now.Add(15 * time.Minute)
	if _, err := tx.ExecContext(ctx,
		`UPDATE app.invites
                    SET status = $2,
                        reserved_by_session_id = $3,
                        reserved_until = $4
                  WHERE code = $1`,
		code, inviteStatusReserved, sessionID, expires,
	); err != nil {
		return time.Time{}, nil, err
	}

	if err := tx.Commit(); err != nil {
		return time.Time{}, nil, err
	}
	committed = true
	return expires, nil, nil
}

func (s *pipelineServer) consumeInvite(ctx context.Context, sess *sessionState) (*apiError, error) {
	if sess.InviteCode == "" {
		return newAPIError(http.StatusForbidden, "INVITE_REQUIRED", "需要邀请码或登录后访问"), nil
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, err
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	var status int
	var reservedBy sql.NullString
	var reservedUntil sql.NullTime
	var expiresAt sql.NullTime
	var usedBy sql.NullString
	err = tx.QueryRowContext(ctx,
		`SELECT status, reserved_by_session_id, reserved_until, expires_at, used_by_session_id
                   FROM app.invites
                  WHERE code = $1 FOR UPDATE`,
		sess.InviteCode,
	).Scan(&status, &reservedBy, &reservedUntil, &expiresAt, &usedBy)
	if errors.Is(err, sql.ErrNoRows) {
		return newAPIError(http.StatusNotFound, "INVITE_NOT_FOUND", "邀请码不存在"), nil
	}
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	if expiresAt.Valid && !expiresAt.Time.After(now) {
		return newAPIError(http.StatusForbidden, "INVITE_EXPIRED", "邀请码已过期"), nil
	}

	switch status {
	case inviteStatusUsed:
		if usedBy.Valid && usedBy.String == sess.ID {
			committed = true
			_ = tx.Commit()
			s.mu.Lock()
			sess.Status = sessionStatusActive
			sess.UpdatedAt = now
			s.mu.Unlock()
			return nil, nil
		}
		return newAPIError(http.StatusConflict, "INVITE_ALREADY_USED", "邀请码已被使用"), nil
	case inviteStatusReserved:
		if reservedBy.Valid && reservedBy.String != sess.ID {
			return newAPIError(http.StatusForbidden, "INVITE_REQUIRED", "邀请码已被其他访客占用"), nil
		}
		if reservedUntil.Valid && !reservedUntil.Time.After(now) {
			return newAPIError(http.StatusForbidden, "INVITE_EXPIRED", "邀请码已过期"), nil
		}
	case inviteStatusUnused:
		// 未预留直接核销允许（例如直接请求题目）
	default:
		return newAPIError(http.StatusForbidden, "INVITE_REQUIRED", "邀请码状态异常"), nil
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE app.invites
                    SET status = $2,
                        used_by_session_id = $3,
                        used_at = NOW(),
                        reserved_by_session_id = NULL,
                        reserved_until = NULL
                  WHERE code = $1`,
		sess.InviteCode, inviteStatusUsed, sess.ID,
	); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	committed = true

	s.mu.Lock()
	sess.Status = sessionStatusActive
	sess.UpdatedAt = now
	s.mu.Unlock()

	return nil, nil
}

func setSessionCookie(w http.ResponseWriter, sid string) {
	if sid == "" {
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "sid",
		Value:    sid,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(30 * 24 * time.Hour),
	})
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

func main() {
	cfg, err := loadAppConfig()
	if err != nil {
		log.Printf("数据库配置错误: %v", err)
		os.Exit(1)
	}

	db, err := connectDatabase(cfg.Database)
	if err != nil {
		log.Printf("数据库连接失败: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	srv := newPipelineServer(cfg.Server.DefaultAPIKey, db)

	mux := http.NewServeMux()
	mux.Handle("/api/", srv.routes())
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		// 尝试物理文件
		tryPath := cfg.Server.StaticDir + r.URL.Path
		if fi, err := os.Stat(tryPath); err == nil && !fi.IsDir() {
			http.ServeFile(w, r, tryPath)
			return
		}
		// 回退到 index.html
		http.ServeFile(w, r, cfg.Server.StaticDir+"/index.html")
	})

	// 端口支持环境变量 PORT，默认 8080（开发友好；线上由 Nginx 反代 80/443）
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("🚀 Server running on http://localhost:" + port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
