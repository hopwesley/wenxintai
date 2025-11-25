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

func (dai *DeepSeekApi) GenerateQuestion(ctx context.Context, bi *BasicInfo, tt TestTyp, callback TokenHandler) (string, error) {
	sLog := dai.log.With().
		Str("ai-test-type", string(tt)).
		Str("public-id", bi.PublicId).
		Logger()

	systemPrompt, err := composeSystemPrompt(tt)
	if err != nil {
		sLog.Err(err).Msg("composeSystemPrompt failed")
		return "", err
	}

	temperature := getTemperature(tt)

	userPrompt := genUserPrompt(bi)

	reqBody := map[string]interface{}{
		"model":           "deepseek-chat",
		"temperature":     temperature,
		"max_tokens":      dai.cfg.QMaxToken,
		"stream":          true,
		"response_format": map[string]string{"type": "json_object"},
		"messages": []map[string]string{
			{"role": "system", "content": strings.TrimSpace(systemPrompt)},
			{"role": "user", "content": userPrompt},
		},
	}

	return dai.validResult(ctx, reqBody, callback, sLog)
}

func (dai *DeepSeekApi) validResult(ctx context.Context, reqBody interface{}, callback TokenHandler, sLog zerolog.Logger) (string, error) {
	content, sErr := dai.streamChat(ctx, reqBody, callback)
	if sErr != nil {
		sLog.Err(sErr).Msg("streamChat failed")
		return "", sErr
	}

	raw := strings.TrimSpace(content)
	if raw == "" {
		sLog.Warn().Msg("test content from ai is empty")
		return "", fmt.Errorf("模型返回空内容 for ")
	}

	var tmp any
	if err := json.Unmarshal([]byte(raw), &tmp); err != nil {
		sLog.Err(err).Msg("test content is not json")
		return "", fmt.Errorf("返回内容非合法 JSON: %w", err)
	}

	sLog.Info().Msg("generate ai test success")
	return raw, nil
}

func (dai *DeepSeekApi) streamChat(ctx context.Context, reqBody interface{}, onToken TokenHandler) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		dai.log.Err(err).Msg("invalid ai request body failed")
		return "", fmt.Errorf("marshal request: %w", err)
	}

	endpoint := strings.TrimRight(dai.cfg.BaseUrl, "/") + "/v1/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(jsonData))
	if err != nil {
		dai.log.Err(err).Msg("build ai request body failed")
		return "", fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if dai.cfg.ApiKey != "" {
		req.Header.Set("Authorization", "Bearer "+dai.cfg.ApiKey)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		dai.log.Err(err).Msg("request deepseek failed")
		return "", fmt.Errorf("request deepseek: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		dai.log.Warn().Str("err-response", strings.TrimSpace(string(body))).Msg(" deepseek status is not ok")
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
				dai.log.Err(err).Str("content", content).Msg("onToken failed, Maybe client is closed.")
			}
		}
	}

	dai.log.Info().Msg("DeepSeek stream finished")
	return fullContent.String(), nil
}

func (dai *DeepSeekApi) GenerateUnifiedReport(ctx context.Context, common *CommonSection, modeParam interface{}, mode Mode, callback TokenHandler) (string, error) {
	sLog := dai.log.With().Str("mode", string(mode)).Logger()

	systemPrompt := systemPromptUnified() + "\n" + systemPromptCommon()
	if mode == Mode33 {
		systemPrompt += "\n" + systemPromptMode33()
	} else {
		systemPrompt += "\n" + systemPromptMode312()
	}
	systemPrompt += "\n" + systemPromptFinal(mode)

	userPrompt := userPromptUnified(common, modeParam, mode)

	reqBody := map[string]interface{}{
		"model":       "deepseek-chat",
		"temperature": dai.cfg.ReportTemperature,
		"max_tokens":  dai.cfg.RMaxToken,
		"stream":      true,
		"response_format": map[string]string{
			"type": "json_object",
		},
		"messages": []map[string]string{
			{"role": "system", "content": strings.TrimSpace(systemPrompt)},
			{"role": "user", "content": strings.TrimSpace(userPrompt)},
		},
	}

	fmt.Println(systemPrompt)
	fmt.Println(userPrompt)

	return dai.validResult(ctx, reqBody, callback, sLog)
}
