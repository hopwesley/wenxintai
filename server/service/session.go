package service

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Session 结构
type Session struct {
	ID        string       `json:"session_id,omitempty"`
	ReqType   PromptType   `json:"req_type"`
	UserInfo  *UserProfile `json:"user_info"`
	CreatedAt time.Time    `json:"created_at"`
	ExpiresAt time.Time    `json:"expires_at"`
}

// 会话存储
var sessions = make(map[string]*Session)
var sessionMutex sync.RWMutex

// 创建新会话
func CreateSession(userInfo *UserProfile) *Session {
	sessionID := uuid.New().String()

	session := &Session{
		ID:        sessionID,
		UserInfo:  userInfo,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24小时过期
	}

	sessionMutex.Lock()
	sessions[sessionID] = session
	sessionMutex.Unlock()

	return session
}

// 获取会话
func GetSession(sessionID string) (*Session, bool) {
	sessionMutex.RLock()
	defer sessionMutex.RUnlock()

	session, exists := sessions[sessionID]
	if !exists || time.Now().After(session.ExpiresAt) {
		return nil, false
	}

	return session, true
}

// 清理过期会话
func CleanupSessions() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		sessionMutex.Lock()
		for id, session := range sessions {
			if time.Now().After(session.ExpiresAt) {
				delete(sessions, id)
			}
		}
		sessionMutex.Unlock()
	}
}

func ParseUserInfo(_ http.ResponseWriter, r *http.Request) (*Session, error) {
	var rs Session
	if err := json.NewDecoder(r.Body).Decode(&rs); err != nil {
		return nil, errors.New("无效的用户信息格式")
	}

	if len(rs.ID) > 0 {
		session, ok := GetSession(rs.ID)
		if !ok {
			return nil, errors.New("请先登录")
		}
		return session, nil
	}

	if nil == rs.UserInfo {
		return nil, errors.New("需要用户基本信息")
	}

	if err := rs.UserInfo.CheckUserData(rs.ReqType); err != nil {
		return nil, err
	}

	return CreateSession(rs.UserInfo), nil
}
