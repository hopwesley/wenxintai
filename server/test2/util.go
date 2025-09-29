package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Mode string

const (
	Mode33  Mode = "3+3"
	Mode312 Mode = "3+1+2"
)

func ParseMode(s string) (Mode, bool) {
	switch s {
	case string(Mode33):
		return Mode33, true
	case string(Mode312):
		return Mode312, true
	default:
		return "", false
	}
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
