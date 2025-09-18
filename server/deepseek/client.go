package deepseek

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

// Client DeepSeek API客户端单例
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

var (
	instance *Client
	once     sync.Once
)

// Message 聊天消息结构
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse 聊天响应结构
type ChatResponse struct {
	ID      string    `json:"id"`
	Object  string    `json:"object"`
	Created int64     `json:"created"`
	Choices []Choice  `json:"choices"`
	Usage   Usage     `json:"usage"`
	Error   *APIError `json:"error,omitempty"`
}

// Choice 选择项结构
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage token使用情况
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// APIError API错误响应
type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// Config 客户端配置
type Config struct {
	APIKey  string
	BaseURL string
	Timeout time.Duration
}

// InitFromEnv 从环境变量初始化客户端
func InitFromEnv() *Client {
	// 加载 .env 文件
	godotenv.Load()

	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		panic("DEEPSEEK_API_KEY environment variable is required")
	}

	baseURL := os.Getenv("DEEPSEEK_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.deepseek.com/v1"
	}

	timeoutStr := os.Getenv("DEEPSEEK_TIMEOUT")
	timeout := 30 * time.Second
	if timeoutStr != "" {
		if parsedTimeout, err := time.ParseDuration(timeoutStr); err == nil {
			timeout = parsedTimeout
		}
	}

	config := Config{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Timeout: timeout,
	}

	return NewClient(config)
}

// NewClient 创建单例客户端
func NewClient(config Config) *Client {
	once.Do(func() {
		instance = &Client{
			apiKey:  config.APIKey,
			baseURL: config.BaseURL,
			httpClient: &http.Client{
				Timeout: config.Timeout,
			},
		}

		if instance.baseURL == "" {
			instance.baseURL = "https://api.deepseek.com/v1"
		}

		if instance.httpClient.Timeout == 0 {
			instance.httpClient.Timeout = 30 * time.Second
		}
	})
	return instance
}

// GetInstance 获取客户端单例
func GetInstance() *Client {
	if instance == nil {
		panic("DeepSeek client not initialized. Call NewClient first.")
	}
	return instance
}

// Chat 发送聊天请求
func (c *Client) Chat(request ChatRequest) (*ChatResponse, error) {
	// 序列化请求体
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/chat/completions", c.baseURL),
		bytes.NewBuffer(requestBody),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		var apiError APIError
		if err := json.Unmarshal(body, &apiError); err != nil {
			return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("API error: %s (type: %s, code: %s)", apiError.Message, apiError.Type, apiError.Code)
	}

	// 解析响应
	var chatResponse ChatResponse
	if err := json.Unmarshal(body, &chatResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &chatResponse, nil
}

// CreateConversation 创建新对话的便捷方法
func (c *Client) CreateConversation(model string, message string, temperature float64) (*ChatResponse, error) {
	req := ChatRequest{
		Model: model,
		Messages: []Message{
			{
				Role:    "user",
				Content: message,
			},
		},
		Temperature: temperature,
		MaxTokens:   500,
	}

	return c.Chat(req)
}

// SetTimeout 设置请求超时时间
func (c *Client) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
}

// ChatStream 流式聊天请求
func (c *Client) ChatStream(request ChatRequest, callback func(*ChatResponse) error) error {
	// 设置流式请求
	request.Stream = true

	requestBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/chat/completions", c.baseURL),
		bytes.NewBuffer(requestBody),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// 处理流式响应
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read stream: %w", err)
		}

		// 处理SSE格式
		if bytes.HasPrefix(line, []byte("data: ")) {
			data := bytes.TrimPrefix(line, []byte("data: "))
			data = bytes.TrimSpace(data)

			if string(data) == "[DONE]" {
				break
			}

			var chunk ChatResponse
			if err := json.Unmarshal(data, &chunk); err != nil {
				return fmt.Errorf("failed to unmarshal chunk: %w", err)
			}

			if err := callback(&chunk); err != nil {
				return err
			}
		}
	}

	return nil
}
