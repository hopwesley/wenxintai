package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

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
	Pair string `json:"pair,omitempty"` // 仅家长对应题出现
}

type Question struct {
	RequestID        string `json:"request_id"`
	StudentID        string `json:"student_id"`
	StudentQuestions []Item `json:"student_questions"`
	ParentQuestions  []Item `json:"parent_questions"`
}

func callDeepSeek2(apiKey string, reqBody Request) string {
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

func callDeepSeek(apiKey string, reqBody interface{}) string {
	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "https://api.deepseek.com/v1/chat/completions", strings.NewReader(string(jsonData)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("请求错误: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("API错误: %d\n", resp.StatusCode)
		return ""
	}

	reader := bufio.NewReader(resp.Body)
	var fullContent strings.Builder

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("读取错误: %v\n", err)
			break
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
			fmt.Println("\n--- 流式传输结束 [DONE] ---")
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
				//fmt.Printf("\u001B[2J\u001B[H %s", content)
				fullContent.WriteString(content)
			}
		}
	}

	fmt.Println("\n=== 最终完整结果 ===")
	fmt.Println(fullContent.String())
	fmt.Println("====================")

	return fullContent.String()
}
