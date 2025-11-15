package srv

import "fmt"

type Config struct {
	Host           string `json:"host"`
	Port           string `json:"port"`
	StaticDir      string `json:"static_dir"`
	DefaultAPIKey  string `json:"default_api_key"`
	studentHobbies []string
	ReadTimeout    int64 `json:"read_timeout,omitempty"`
	WriteTimeout   int64 `json:"write_timeout,omitempty"`
	IdleTimeout    int64 `json:"idle_timeout,omitempty"`
}

func (cfg *Config) srvAddr() string {
	return cfg.Host + ":" + cfg.Port
}

func (cfg *Config) Validate() error {

	if len(cfg.DefaultAPIKey) < 4 {
		return fmt.Errorf("invalid ai key:%s", cfg.DefaultAPIKey)
	}

	if cfg.Host == "" {
		cfg.Host = "localhost"
	}

	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.StaticDir == "" {
		cfg.StaticDir = "./static"
	}

	if cfg.ReadTimeout <= 0 {
		cfg.ReadTimeout = 15
	}

	if cfg.WriteTimeout <= 0 {
		cfg.WriteTimeout = 15
	}

	if cfg.IdleTimeout <= 0 {
		cfg.IdleTimeout = 60
	}

	return nil
}

var defaultHobbies = []string{
	// 体育类
	"篮球",
	"足球",
	"羽毛球",
	"跑步",
	"游泳",
	"乒乓球",
	"健身",

	// 艺术类
	"音乐",
	"绘画",
	"舞蹈",
	"摄影",
	"书法",
	"写作",

	// 科技类
	"编程",
	"机器人",
	"科学实验",
	"电子制作",
	"下棋（象棋/围棋/国际象棋）",

	// 生活方式类
	"旅行",
	"美食",
	"志愿活动",
	"阅读",
	"看电影",
	"园艺",
}
