package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/hopwesley/wenxintai/server/deepseek"
)

// 定义一个响应结构
type ApiResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// 聊天请求结构
type ChatRequest struct {
	Message     string  `json:"message"`
	Model       string  `json:"model,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

// 聊天响应结构
type ChatResponse struct {
	Reply       string `json:"reply"`
	TotalTokens int    `json:"total_tokens,omitempty"`
	RequestID   string `json:"request_id,omitempty"`
}

func main() {
	http.HandleFunc("/api/chat", chatHandler)
	http.HandleFunc("/api/hello", helloHandler)

	// 启动 HTTP 服务器，监听 80 端口
	log.Println("🚀 Server running on http://localhost:80")
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		log.Fatal(err)
	}
}

// helloHandler 基础测试接口
func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	resp := ApiResponse{
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

// chatHandler 与AI对话的接口
func chatHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 只允许 POST 请求
	if r.Method != http.MethodPost {
		resp := ApiResponse{
			Success: false,
			Message: "只支持 POST 请求",
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// 解析请求体
	var chatReq ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&chatReq); err != nil {
		resp := ApiResponse{
			Success: false,
			Message: "无效的请求格式",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// 验证消息内容
	if chatReq.Message == "" {
		resp := ApiResponse{
			Success: false,
			Message: "消息内容不能为空",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// 设置默认值
	if chatReq.Model == "" {
		chatReq.Model = "deepseek-chat"
	}
	if chatReq.Temperature == 0 {
		chatReq.Temperature = 0.7
	}

	// 调用 DeepSeek API
	log.Printf("🤖 AI请求: %s", chatReq.Message)

	response, err := deepseek.Instance().GetClient().CreateConversation(
		chatReq.Model,
		chatReq.Message,
		chatReq.Temperature,
	)

	if err != nil {
		log.Printf("❌ AI请求失败: %v", err)
		resp := ApiResponse{
			Success: false,
			Message: "AI服务暂时不可用: " + err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// 构建成功响应
	if len(response.Choices) > 0 {
		aiReply := response.Choices[0].Message.Content
		log.Printf("✅ AI回复: %s", aiReply)

		chatResp := ChatResponse{
			Reply:       aiReply,
			TotalTokens: response.Usage.TotalTokens,
			RequestID:   response.ID,
		}

		resp := ApiResponse{
			Success: true,
			Message: "AI回复成功",
			Data:    chatResp,
		}

		json.NewEncoder(w).Encode(resp)
	} else {
		resp := ApiResponse{
			Success: false,
			Message: "未收到AI回复",
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
	}
}
