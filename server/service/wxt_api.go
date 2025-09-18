package service

import (
	"encoding/json"
	"net/http"
	"time"
)

// 定义一个响应结构
type ApiRes struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type ApiReq struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// helloHandler 基础测试接口
func helloHandler(w http.ResponseWriter, r *http.Request) {
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
