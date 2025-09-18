package service

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/hopwesley/wenxintai/server/deepseek"
)

type ApiRes struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	resp := ApiRes{
		Success: true,
		Message: "Hello from Go backend!",
		Data: map[string]string{
			"author":     "wenxintai",
			"lang":       "Go + TypeScript",
			"ai_service": "DeepSeek API",
			"timestamp":  time.Now().Format(time.RFC3339),
		},
	}

	json.NewEncoder(w).Encode(resp)
}

func StartReqHandler(w http.ResponseWriter, r *http.Request) {

	session, err := ParseUserInfo(w, r)
	if err != nil {
		resp := ApiRes{
			Success: false,
			Message: err.Error(),
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	_, err = deepseek.Instance().GetClient().CreateConversation(
		"deepseek-chat",
		"", // 空参数，等待下一步设置角色
		0.7,
	)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ApiRes{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	resp := ApiRes{
		Success: true,
		Message: "会话创建成功",
		Data: map[string]string{
			"session_id": session.ID,
			"expires_at": session.ExpiresAt.Format(time.RFC3339),
		},
	}

	json.NewEncoder(w).Encode(resp)
}
