package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// 定义一个响应结构
type ApiResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func main() {
	// 设置一个简单的路由
	http.HandleFunc("/api/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := ApiResponse{
			Success: true,
			Message: "Hello from Go backend!",
			Data: map[string]string{
				"author": "wenxintai",
				"lang":   "Go + TypeScript",
			},
		}
		json.NewEncoder(w).Encode(resp)
	})

	// 启动 HTTP 服务器，监听 80 端口
	log.Println("🚀 Server running on http://localhost:80")
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		log.Fatal(err)
	}
}
