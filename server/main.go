package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/hopwesley/wenxintai/server/assessment"
)

type pipelineServer struct {
	mu            sync.RWMutex
	sessions      map[string]*sessionState
	wechatIndex   map[string]string
	defaultAPIKey string
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

func newPipelineServer(defaultKey string) *pipelineServer {
	return &pipelineServer{
		sessions:      make(map[string]*sessionState),
		wechatIndex:   make(map[string]string),
		defaultAPIKey: defaultKey,
	}
}

func (s *pipelineServer) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", s.wrap(http.MethodPost, s.handleLogin))
	mux.HandleFunc("/api/hobbies", s.wrap(http.MethodGet, s.handleHobbies))
	mux.HandleFunc("/api/questions", s.wrap(http.MethodPost, s.handleQuestions))
	mux.HandleFunc("/api/answers", s.wrap(http.MethodPost, s.handleAnswers))
	mux.HandleFunc("/api/report", s.wrap(http.MethodPost, s.handleReport))
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
	now := time.Now()
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

func (s *pipelineServer) handleHobbies(w http.ResponseWriter, _ *http.Request) {
	s.writeJSON(w, http.StatusOK, hobbiesResponse{Hobbies: assessment.StudentHobbies})
}

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

func main() {
	defaultKey := os.Getenv("DEEPSEEK_API_KEY")
	srv := newPipelineServer(defaultKey)
	handler := srv.routes()

	log.Println("🚀 Server running on http://localhost:80")
	if err := http.ListenAndServe(":80", handler); err != nil {
		log.Fatal(err)
	}
}
