package ai_api

import "fmt"

type Cfg struct {
	ApiKey           string `json:"api_key"`
	Model            string `json:"model"`
	QuestionMaxToken int    `json:"question_max_token"`
	BaseUrl          string `json:"base_url"`
}

func (cfg *Cfg) Validate() error {
	if len(cfg.ApiKey) < 4 {
		return fmt.Errorf("invalid ai key:%s", cfg.ApiKey)
	}
	if len(cfg.Model) == 0 {
		cfg.Model = DefaultModel
	}
	if cfg.QuestionMaxToken < 100 {
		cfg.QuestionMaxToken = DefaultMaxQToken
	}
	if len(cfg.BaseUrl) < 4 {
		cfg.BaseUrl = DefaultApiUrl
	}
	return nil
}
