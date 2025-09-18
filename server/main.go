package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/hopwesley/wenxintai/server/deepseek"
	"github.com/hopwesley/wenxintai/server/service"
)

func main() {
	http.HandleFunc("/api/start-session", startSessionHandler)
	http.HandleFunc("/api/hello", helloHandler)

	// 启动 HTTP 服务器，监听 80 端口
	log.Println("🚀 Server running on http://localhost:80")
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		log.Fatal(err)
	}
}

// 用户信息提交接口
func startSessionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		resp := ApiResponse{
			Success: false,
			Message: "只支持 POST 请求",
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// 创建会话
	session := createSession(userInfo)

	// 初始化AI角色（空参数请求）
	_, err := deepseek.Instance().GetClient().CreateConversation(
		"deepseek-chat",
		"", // 空参数，等待下一步设置角色
		0.7,
	)

	if err != nil {
		log.Printf("AI初始化失败: %v", err)
		// 不返回错误，继续创建会话
	}

	resp := ApiResponse{
		Success: true,
		Message: "会话创建成功",
		Data: map[string]string{
			"session_id": session.ID,
			"expires_at": session.ExpiresAt.Format(time.RFC3339),
		},
	}

	json.NewEncoder(w).Encode(resp)
}
