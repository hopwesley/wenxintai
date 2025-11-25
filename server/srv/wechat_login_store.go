package srv

import (
	"sync"
	"time"
)

// WxLoginEntry 代表一次扫码登录（按 state 维度）
type WxLoginEntry struct {
	OpenID    string
	UnionID   string
	IsNew     bool
	Status    string    // "ok" | "error"
	CreatedAt time.Time // 用于简单判断是否过期
}

// WxLoginStore 是一个线程安全的内存缓存
type WxLoginStore struct {
	mu         sync.Mutex
	byState    map[string]WxLoginEntry
	knownUsers map[string]struct{} // 用 unionid/openid 判断是不是“老用户”
}

func NewWxLoginStore() *WxLoginStore {
	return &WxLoginStore{
		byState:    make(map[string]WxLoginEntry),
		knownUsers: make(map[string]struct{}),
	}
}

// MarkLogin 在用户扫码并授权成功后写入结果，并返回这次是否是“新用户”
func (s *WxLoginStore) MarkLogin(state, openID, unionID string) WxLoginEntry {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := unionID
	if key == "" {
		key = openID
	}

	isNew := false
	if key != "" {
		if _, ok := s.knownUsers[key]; !ok {
			isNew = true
			s.knownUsers[key] = struct{}{}
		}
	} else {
		// 正常情况下不会走到这里，这里保守当作新用户
		isNew = true
	}

	entry := WxLoginEntry{
		OpenID:    openID,
		UnionID:   unionID,
		IsNew:     isNew,
		Status:    "ok",
		CreatedAt: time.Now(),
	}

	s.byState[state] = entry
	return entry
}

// Get 根据 state 取出这次登录结果
func (s *WxLoginStore) Get(state string) (WxLoginEntry, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.byState[state]
	return entry, ok
}

// 可选：简单清理函数（目前没地方调用，你后面可以按需加定时清理）
func (s *WxLoginStore) CleanupExpired(maxAge time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	for k, v := range s.byState {
		if v.CreatedAt.Before(cutoff) {
			delete(s.byState, k)
		}
	}
}

// 全局实例：简单一点，不改 HttpSrv 结构
var wxLoginStore = NewWxLoginStore()
