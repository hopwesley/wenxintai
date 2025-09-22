package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// 在 main 包中定义这些结构体，而不是使用 test 包
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type ResponseFormat struct {
	Type       string                 `json:"type"`
	JSONSchema map[string]interface{} `json:"json_schema,omitempty"`
}
type Request struct {
	Model          string          `json:"model"`
	Temperature    float32         `json:"temperature"`
	MaxTokens      int             `json:"max_tokens"`
	Stream         bool            `json:"stream"`
	Messages       []Message       `json:"messages"`
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`
}

type ChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type Item struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
	Type string `json:"type"`
	Rev  bool   `json:"rev"`
}

type Question struct {
	RequestID        string `json:"request_id"`
	StudentID        string `json:"student_id"`
	StudentQuestions []Item `json:"student_questions"`
	ParentQuestions  []Item `json:"parent_questions"`
}

func callDeepSeek(apiKey string, reqBody Request) string {
	bs, _ := json.Marshal(reqBody)
	client := &http.Client{Timeout: 120 * time.Second}
	httpReq, _ := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewReader(bs))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Accept", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		fmt.Println("请求错误:", err)
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var cr ChatResponse
	if err := json.Unmarshal(body, &cr); err == nil && len(cr.Choices) > 0 {
		return cr.Choices[0].Message.Content
	}
	return string(body)
}

func uuidLike() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}
