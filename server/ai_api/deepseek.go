package ai_api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/hopwesley/wenxintai/server/comm"
	"github.com/rs/zerolog"
)

var (
	_aiOnce = sync.Once{}

	_aiIns AIApi = nil
)

type DeepSeekApi struct {
	log zerolog.Logger
	cfg *Cfg
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

func newDeepSeek() *DeepSeekApi {
	return &DeepSeekApi{
		log: comm.LogInst().With().Str("model", "DeepSeek").Logger(),
	}
}

func Instance() AIApi {
	_aiOnce.Do(func() {
		_aiIns = newDeepSeek()
	})
	return _aiIns
}

func (dai *DeepSeekApi) Init(cfg *Cfg) error {
	dai.cfg = cfg
	return nil
}

func (dai *DeepSeekApi) GenerateQuestion(ctx context.Context, basicInfo *BasicInfo, tt TestTyp, callback TokenHandler) (json.RawMessage, error) {
	systemPrompt, err := composeSystemPrompt(tt)
	if err != nil {
		dai.log.Err(err).Str("test_type", string(tt)).Msgf("composeSystemPrompt failed:%v", basicInfo)
		return nil, err
	}

	temperature := getTemperature(tt)

	userPrompt := genUserPrompt(basicInfo)

	reqBody := map[string]interface{}{
		"model":           "deepseek-chat",
		"temperature":     temperature,
		"max_tokens":      dai.cfg.QuestionMaxToken,
		"stream":          true,
		"response_format": map[string]string{"type": "json_object"},
		"messages": []map[string]string{
			{"role": "system", "content": strings.TrimSpace(systemPrompt)},
			{"role": "user", "content": userPrompt},
		},
	}

	content, sErr := dai.streamChat(ctx, reqBody, callback)
	if sErr != nil {
		dai.log.Err(sErr).Str("test_type", string(tt)).Msgf("streamChat failed:%v", basicInfo)
		return nil, sErr
	}
	raw := strings.TrimSpace(content)
	if raw == "" {
		return nil, fmt.Errorf("模型返回空内容 for %s", tt)
	}

	var tmp any
	if err := json.Unmarshal([]byte(raw), &tmp); err != nil {
		return nil, fmt.Errorf("%s 返回内容非合法 JSON: %w", tt, err)
	}

	return json.RawMessage(raw), nil

}

func (dai *DeepSeekApi) streamChat(ctx context.Context, reqBody interface{}, onToken TokenHandler) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	endpoint := strings.TrimRight(dai.cfg.BaseUrl, "/") + "/v1/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if dai.cfg.ApiKey != "" {
		req.Header.Set("Authorization", "Bearer "+dai.cfg.ApiKey)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request deepseek: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("deepseek status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	reader := bufio.NewReader(resp.Body)
	var fullContent strings.Builder

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("read stream: %w", err)
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
		if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
			continue
		}

		if len(streamResp.Choices) == 0 {
			continue
		}
		content := streamResp.Choices[0].Delta.Content
		if len(content) == 0 {
			continue
		}

		fullContent.WriteString(content)
		if onToken != nil {
			if err := onToken(content); err != nil {
				dai.log.Error().Err(err).Str("content", content).Msg("onToken failed, Maybe client is closed.")
			}
		}
	}

	return fullContent.String(), nil
}
