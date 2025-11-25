package ai_api

import "fmt"

const (
	DefaultModel  = "deepseek-chat"
	DefaultApiUrl = "https://api.deepseek.com"

	DefaultMaxQToken         = 8000
	DefaultMaxRToken         = 4000
	DefaultReportTemperature = 0.4
)

type Cfg struct {
	ApiKey    string `json:"api_key"`
	Model     string `json:"model"`
	QMaxToken int    `json:"q_max_token"`
	RMaxToken int    `json:"r_max_token"`
	BaseUrl   string `json:"base_url"`

	ReportTemperature float64 `json:"report_temperature"`
}

func (cfg *Cfg) Validate() error {
	if len(cfg.ApiKey) < 4 {
		return fmt.Errorf("invalid ai key:%s", cfg.ApiKey)
	}
	if len(cfg.Model) == 0 {
		cfg.Model = DefaultModel
	}
	if cfg.QMaxToken < 100 {
		cfg.QMaxToken = DefaultMaxQToken
	}
	if len(cfg.BaseUrl) < 4 {
		cfg.BaseUrl = DefaultApiUrl
	}
	if cfg.RMaxToken < 100 {
		cfg.RMaxToken = DefaultMaxRToken
	}

	if cfg.ReportTemperature <= 0.0 {
		cfg.ReportTemperature = DefaultReportTemperature
	}

	return nil
}
