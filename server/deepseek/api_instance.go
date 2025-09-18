package deepseek

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// DeepSeekSrv DeepSeek 服务单例
type DeepSeekSrv struct {
	client *Client
}

var (
	serviceInstance *DeepSeekSrv
	serviceOnce     sync.Once
)

// Instance 获取服务单例实例，如果未初始化则从环境变量自动初始化
func Instance() *DeepSeekSrv {
	if serviceInstance == nil {
		// 使用 sync.Once 确保线程安全的延迟初始化
		serviceOnce.Do(func() {
			client := InitFromEnv()
			serviceInstance = &DeepSeekSrv{
				client: client,
			}
			fmt.Println("⚠️  DeepSeek服务未初始化，已自动从环境变量初始化")
		})
	}
	return serviceInstance
}

// TestHello 测试服务连通性
func (s *DeepSeekSrv) TestHello() (string, error) {
	response, err := s.client.CreateConversation(
		"deepseek-chat",
		"你好，请简单介绍一下你自己，只用一句话回答",
		0.7,
	)

	if err != nil {
		return "", fmt.Errorf("服务测试失败: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("未收到有效回复")
	}

	return response.Choices[0].Message.Content, nil
}

// GetClient 获取底层客户端实例
func (s *DeepSeekSrv) GetClient() *Client {
	return s.client
}

// SetClientTimeout 设置客户端超时时间
func (s *DeepSeekSrv) SetClientTimeout(timeout time.Duration) {
	s.client.SetTimeout(timeout)
}

// HealthCheck 健康检查
func (s *DeepSeekSrv) HealthCheck() (bool, error) {
	_, err := s.TestHello()
	if err != nil {
		return false, err
	}
	return true, nil
}

// ResetService 重置服务单例（主要用于测试）
func ResetService() {
	serviceInstance = nil
}

// Session 会话上下文，包含用户特定的数据
type Session struct {
	SessionID  string
	UserID     string
	Messages   []Message // 会话历史
	CreatedAt  time.Time
	LastActive time.Time
}

// ChatRequestWithContext 带上下文的聊天请求
type ChatRequestWithContext struct {
	Model       string
	Temperature float64
	MaxTokens   int
	Stream      bool
	SessionID   string // 可选：会话ID用于保持上下文
	UserID      string // 可选：用户ID用于隔离数据
}

// CreateConversationWithContext 带上下文的对话创建
func (s *DeepSeekSrv) CreateConversationWithContext(ctx context.Context, req ChatRequestWithContext, userMessage string) (*ChatResponse, error) {
	// 这里可以根据sessionID获取历史消息
	messages := s.getSessionMessages(req.SessionID)

	// 添加用户新消息
	messages = append(messages, Message{
		Role:    "user",
		Content: userMessage,
	})

	chatReq := ChatRequest{
		Model:       req.Model,
		Messages:    messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		Stream:      req.Stream,
	}

	// 使用底层客户端发送请求
	response, err := s.client.Chat(chatReq)
	if err != nil {
		return nil, err
	}

	// 保存会话历史（如果需要）
	if req.SessionID != "" {
		s.saveSessionMessages(req.SessionID, messages, response)
	}

	return response, nil
}

// 简单的会话管理（实际项目中可以用Redis等）
func (s *DeepSeekSrv) getSessionMessages(sessionID string) []Message {
	// 这里实现从缓存或数据库获取历史消息
	// 简化示例：返回空数组
	return []Message{}
}

func (s *DeepSeekSrv) saveSessionMessages(sessionID string, messages []Message, response *ChatResponse) {
	// 这里实现保存会话历史到缓存或数据库
	// 注意：实际项目中要控制历史消息长度，避免token超限
}

func ProcessUserRequest(service *DeepSeekSrv, userID string) {
	// 为每个用户创建独立的请求上下文
	ctx := context.WithValue(context.Background(), "user_id", userID)

	req := ChatRequestWithContext{
		Model:       "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   500,
		SessionID:   fmt.Sprintf("session_%s_%d", userID, time.Now().Unix()),
		UserID:      userID,
	}

	// 每个用户有自己独立的消息
	userMessage := fmt.Sprintf("你好，我是用户%s，请为我推荐一些%s相关的内容", userID, getUserInterest(userID))

	response, err := service.CreateConversationWithContext(ctx, req, userMessage)
	if err != nil {
		log.Printf("用户 %s 请求失败: %v", userID, err)
		return
	}

	if len(response.Choices) > 0 {
		fmt.Printf("用户 %s 收到回复: %s\n", userID, response.Choices[0].Message.Content)
	}
}

func getUserInterest(userID string) string {
	// 模拟不同用户的兴趣
	interests := map[string]string{
		"user1": "技术",
		"user2": "音乐",
		"user3": "体育",
		"user4": "美食",
		"user5": "旅行",
	}
	return interests[userID]
}
