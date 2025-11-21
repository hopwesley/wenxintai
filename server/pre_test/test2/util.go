package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	core "github.com/hopwesley/wenxintai/server/ai_api"
	"github.com/sashabaranov/go-openai"
)

type StreamResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role    string `json:"role,omitempty"`
			Content string `json:"content,omitempty"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

func init() {
	core.DeepSeekCaller = callDeepSeek
}

func callDeepSeek(apiKey string, reqBody interface{}) (string, error) {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", "https://api.deepseek.com/v1/chat/completions", strings.NewReader(string(jsonData)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("deepseek status %d", resp.StatusCode)
	}

	reader := bufio.NewReader(resp.Body)
	var fullContent strings.Builder

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var streamResp StreamResponse
		if json.Unmarshal([]byte(data), &streamResp) != nil {
			continue
		}

		if len(streamResp.Choices) > 0 {
			content := streamResp.Choices[0].Delta.Content
			if content != "" {
				fmt.Print(content)
				fullContent.WriteString(content)
			}
		}
	}

	return fullContent.String(), nil
}

func CallQwen(apiKey, userPrompt, systemPrompt string) {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	client := openai.NewClientWithConfig(config)

	stream, err := client.CreateChatCompletionStream(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: "qwen-plus",
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
				{Role: openai.ChatMessageRoleUser, Content: userPrompt},
			},
			Stream: true,
		},
	)
	if err != nil {
		log.Fatal("Failed to create stream:", err)
	}
	defer stream.Close()

	var fullContent string // 用于累积完整回复

	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break // 流正常结束
		}
		if err != nil {
			log.Fatal("Error receiving from stream:", err)
		}

		// 可选：输出每个 chunk 的 JSON（按你之前需求）

		// 拼接完整内容
		if len(resp.Choices) > 0 {
			delta := resp.Choices[0].Delta
			if delta.Content != "" {
				fullContent += delta.Content
				fmt.Print(delta.Content)
			}
		}
	}

	// 流结束后，打印完整内容
	fmt.Println("\n--- Full Response ---")
	fmt.Println(fullContent)

	// 如果你最终需要返回这个完整内容，可以 return 它（需修改函数签名）
}
