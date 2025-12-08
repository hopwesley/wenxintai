package srv

import (
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Host                 string `json:"host"`
	Port                 string `json:"port"`
	StaticDir            string `json:"static_dir"`
	studentHobbies       []string
	ReadTimeout          int64           `json:"read_timeout,omitempty"`
	WeChatAppID          string          `json:"we_chat_app_id"`
	WeChatAppSecret      string          `json:"we_chat_app_sec"`
	WeChatRedirectDomain string          `json:"we_chat_redirect_domain"`
	PaymentForward       string          `json:"payment_forward,omitempty"`
	WeChatAPIV3Key       string          `json:"we_chat_api_v3_key"`
	WxPaymentTimeout     int             `json:"wx_payment_timeout"`
	Websocket            WebsocketConfig `json:"websocket,omitempty"`
}

type WebsocketConfig struct {
	AllowedOrigins    []string `json:"allowed_origins,omitempty"`
	ReadBufferSize    int      `json:"read_buffer_size,omitempty"`
	WriteBufferSize   int      `json:"write_buffer_size,omitempty"`
	HandshakeTimeout  int      `json:"handshake_timeout,omitempty"`
	HeartbeatInterval int      `json:"heartbeat_interval,omitempty"`
}

type WeChatPayConfig struct {
	MchID       string `json:"mch_id"`        // 商户号
	AppID       string `json:"app_id"`        // 公众号/小程序 AppID
	APIV3Key    string `json:"apiv_3_key"`    // API v3 密钥（32 字节）
	MchSerial   string `json:"mch_serial"`    // 商户证书序列号
	PublicKeyID string `json:"public_key_id"` // 你截图里的 PUB_KEY_ID_...
	NotifyURL   string `json:"notify_url"`    // 回调地址：https://xxx/api/pay/wechat/callback

	privateKeyPEM string
	publicKeyPEM  string
}

func (c *WeChatPayConfig) Validate(configDir, WechatPayPubKeyFile, MchPrivateKeyFile string) error {

	// --------------------------
	// 1. 读取商户私钥 PEM（apiclient_key.pem）
	// --------------------------
	if MchPrivateKeyFile != "" {
		fullPath := filepath.Join(configDir, MchPrivateKeyFile)
		keyBytes, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("读取商户私钥文件失败: %s, %w", fullPath, err)
		}
		c.privateKeyPEM = string(keyBytes)
	}

	// --------------------------
	// 2. 读取微信支付公钥 PEM（wechatpay_public.pem）
	// --------------------------
	if WechatPayPubKeyFile != "" {
		fullPath := filepath.Join(configDir, WechatPayPubKeyFile)
		pubBytes, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("读取微信支付公钥文件失败: %s, %w", fullPath, err)
		}
		c.publicKeyPEM = string(pubBytes)
	}

	if c.MchID == "" {
		return fmt.Errorf("wechat pay: mch_id empty")
	}
	if c.AppID == "" {
		return fmt.Errorf("wechat pay: app_id empty")
	}
	if len(c.APIV3Key) != 32 {
		return fmt.Errorf("wechat pay: apiv_3_key must be 32 bytes")
	}
	if c.MchSerial == "" {
		return fmt.Errorf("wechat pay: mch_serial empty")
	}
	if c.privateKeyPEM == "" {
		return fmt.Errorf("wechat pay: mch_private_key_pem empty")
	}
	if c.NotifyURL == "" {
		return fmt.Errorf("wechat pay: notify_url empty")
	}

	return nil
}

func (cfg *Config) srvAddr() string {
	return cfg.Host + ":" + cfg.Port
}

func (cfg *Config) Validate() error {

	if cfg.Host == "" {
		cfg.Host = "localhost"
	}

	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	if cfg.ReadTimeout <= 0 {
		cfg.ReadTimeout = 10
	}
	if cfg.WxPaymentTimeout <= 0 {
		cfg.WxPaymentTimeout = 30
	}

	if cfg.Websocket.HandshakeTimeout <= 0 {
		cfg.Websocket.HandshakeTimeout = 10
	}
	if cfg.Websocket.HeartbeatInterval <= 0 {
		cfg.Websocket.HeartbeatInterval = 30
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
	"下棋",

	// 生活方式类
	"旅行",
	"美食",
	"志愿活动",
	"阅读",
	"看电影",
	"园艺",
}
