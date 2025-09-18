package service

import (
	"github.com/google/uuid"
	"sync"
	"time"
)

// Session 结构
type Session struct {
	ID        string    `json:"session_id"`
	UserInfo  UserInfo  `json:"user_info"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// 会话存储
var sessions = make(map[string]*Session)
var sessionMutex sync.RWMutex

// 创建新会话
func CreateSession(userInfo UserInfo) *Session {
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
