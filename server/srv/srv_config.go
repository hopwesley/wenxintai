package srv

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

type Config struct {
	Host                 string `json:"host"`
	Port                 string `json:"port"`
	StaticDir            string `json:"static_dir"`
	studentHobbies       []string
	ReadTimeout          int64  `json:"read_timeout,omitempty"`
	WeChatAppID          string `json:"we_chat_app_id"`
	WeChatAppSecret      string `json:"we_chat_app_sec"`
	WeChatRedirectDomain string `json:"we_chat_redirect_domain"`
	PaymentForward       string `json:"payment_forward,omitempty"`
	WeChatAPIV3Key       string `json:"we_chat_api_v3_key"`
}

type WeChatPayConfig struct {
	MchID            string // 商户号
	AppID            string // 公众号/网站对应的 appid
	APIV3Key         string // v3 密钥（32 字节）
	MchSerial        string // 商户证书序列号
	MchPrivateKeyPEM string // 商户私钥 PEM（pkcs1 或 pkcs8）

	// 平台证书：key 为 serial 号，值为证书对象，用于验签
	PlatformCerts map[string]*x509.Certificate

	NotifyURL string // 支付结果通知地址：https://你的域名/api/pay/wechat/callback
}

func (c *WeChatPayConfig) Validate() error {

	wxPlatformCert, _ := loadWeChatPlatformCert(os.Getenv("WX_PLATFORM_CERT_PEM"))
	c.PlatformCerts = map[string]*x509.Certificate{
		"微信平台证书序列号": wxPlatformCert,
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

func loadWeChatPlatformCert(pemStr string) (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("invalid platform cert pem")
	}
	return x509.ParseCertificate(block.Bytes)
}
