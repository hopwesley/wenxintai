package srv

import (
	"bytes"
	"context"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	rand2 "math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/hopwesley/wenxintai/server/dbSrv"
)

func (s *HttpSrv) apiWeChatPayCallBack(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if len(s.cfg.PaymentForward) > 0 {
		s.forwardCallback(w, r, s.cfg.PaymentForward)
		return
	}

	s.processWeChatPayment(w, r)
}

type wechatPayNotifyBody struct {
	ID           string `json:"id"`
	CreateTime   string `json:"create_time"`
	EventType    string `json:"event_type"`
	ResourceType string `json:"resource_type"`
	Summary      string `json:"summary"`
	Resource     struct {
		Algorithm      string `json:"algorithm"`
		Ciphertext     string `json:"ciphertext"`
		AssociatedData string `json:"associated_data"`
		Nonce          string `json:"nonce"`
		OriginalType   string `json:"original_type"`
	} `json:"resource"`
}

// 解密后的支付通知里最关键的是 out_trade_no / trade_state / amount
type wechatPayNotifyResource struct {
	AppID         string `json:"appid"`
	MchID         string `json:"mchid"`
	OutTradeNo    string `json:"out_trade_no"`
	TransactionID string `json:"transaction_id"`
	TradeType     string `json:"trade_type"`
	TradeState    string `json:"trade_state"` // SUCCESS/NOTPAY/CLOSED/REFUND...
	Payer         struct {
		OpenID string `json:"openid"`
	} `json:"payer"`
	Amount struct {
		Total      int64 `json:"total"`
		PayerTotal int64 `json:"payer_total"`
	} `json:"amount"`
	SuccessTime string `json:"success_time"`
}

func (s *HttpSrv) processWeChatPayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := s.log.With().Str("handler", "processWeChatPayment").Logger()

	// 1) 读 body
	bodyBytes, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		log.Error().Err(err).Msg("read notify body failed")
		http.Error(w, "read body failed", http.StatusBadRequest)
		return
	}
	_ = r.Body.Close()

	// 2) 验签：使用微信平台证书公钥验证 HTTP 头 + body
	if err := s.verifyWeChatNotifySignature(r.Header, bodyBytes); err != nil {
		log.Error().Err(err).Msg("verify signature failed")
		http.Error(w, "signature verification failed", http.StatusBadRequest)
		return
	}

	// 3) 解析外层 JSON
	var notify wechatPayNotifyBody
	if err := json.Unmarshal(bodyBytes, &notify); err != nil {
		log.Error().Err(err).Msg("unmarshal notify body failed")
		http.Error(w, "bad notify body", http.StatusBadRequest)
		return
	}

	if notify.ResourceType != "encrypt-resource" {
		log.Error().Str("type", notify.ResourceType).Msg("unexpected resource_type")
		http.Error(w, "bad resource_type", http.StatusBadRequest)
		return
	}

	// 4) 解密 resource → 得到真正的订单信息
	resource, err := s.decryptWeChatResource(&notify.Resource)
	if err != nil {
		log.Error().Err(err).Msg("decrypt resource failed")
		http.Error(w, "decrypt failed", http.StatusBadRequest)
		return
	}

	log.Info().
		Str("out_trade_no", resource.OutTradeNo).
		Str("trade_state", resource.TradeState).
		Str("transaction_id", resource.TransactionID).
		Msg("wechat payment notify")

	// 5) 根据 trade_state 更新本地订单
	switch resource.TradeState {
	case "SUCCESS":
		if err := s.updateWeChatOrderStatus(ctx, resource.OutTradeNo, resource.TradeState, resource, bodyBytes); err != nil {
			log.Error().Err(err).
				Str("out_trade_no", resource.OutTradeNo).
				Msg("updateWeChatOrderStatus failed")
		}

		// TODO：这里可以触发“开通测试权限”的业务逻辑，
		// 比如在 app.tests_record 里写一条已付款记录，关联 wx_user + plan + session 等

	case "REFUND", "CLOSED", "PAYERROR":
		if err := s.updateWeChatOrderStatus(ctx, resource.OutTradeNo, resource.TradeState, resource, bodyBytes); err != nil {
			log.Error().Err(err).
				Str("out_trade_no", resource.OutTradeNo).
				Msg("updateWeChatOrderStatus failed")
		}
	default:
		// NOTPAY 等状态，按需处理
	}

	// 6) 按微信要求返回 JSON：表示已成功接收
	type resp struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}

	writeJSON(w, http.StatusOK, &resp{
		Code:    "SUCCESS",
		Message: "成功",
	})
}
func (s *HttpSrv) verifyWeChatNotifySignature(header http.Header, body []byte) error {
	timestamp := header.Get("Wechatpay-Timestamp")
	nonce := header.Get("Wechatpay-Nonce")
	signature := header.Get("Wechatpay-Signature")
	serial := header.Get("Wechatpay-Serial")

	if timestamp == "" || nonce == "" || signature == "" || serial == "" {
		return fmt.Errorf("missing wechatpay headers")
	}

	// 1) 构造待验签串
	message := timestamp + "\n" + nonce + "\n" + string(body) + "\n"

	// 2) Base64 decode 签名
	sigBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("decode signature: %w", err)
	}

	// 3) 找到对应 serial 的平台证书
	cert := s.payment.PlatformCerts[serial]
	if cert == nil {
		return fmt.Errorf("unknown wechat platform cert serial: %s", serial)
	}

	pubKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("platform cert public key is not rsa")
	}

	// 4) 用平台证书公钥验签
	hash := sha256.Sum256([]byte(message))
	if err := rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hash[:], sigBytes); err != nil {
		return fmt.Errorf("verify signature failed: %w", err)
	}

	return nil
}

func (s *HttpSrv) decryptWeChatResource(res *struct {
	Algorithm      string `json:"algorithm"`
	Ciphertext     string `json:"ciphertext"`
	AssociatedData string `json:"associated_data"`
	Nonce          string `json:"nonce"`
	OriginalType   string `json:"original_type"`
}) (*wechatPayNotifyResource, error) {
	if res.Algorithm != "AEAD_AES_256_GCM" {
		return nil, fmt.Errorf("unexpected algorithm: %s", res.Algorithm)
	}

	// 1) Base64 解码 ciphertext
	cipherBytes, err := base64.StdEncoding.DecodeString(res.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("decode ciphertext: %w", err)
	}

	// 2) 用 APIv3Key 做 AES-GCM 解密
	// apiV3Key 通常从配置加载：32 字节
	apiV3Key := []byte(s.cfg.WeChatAPIV3Key)
	block, err := aes.NewCipher(apiV3Key)
	if err != nil {
		return nil, fmt.Errorf("new cipher: %w", err)
	}
	aesGcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("new gcm: %w", err)
	}

	nonceBytes := []byte(res.Nonce)
	ad := []byte(res.AssociatedData)

	plain, err := aesGcm.Open(nil, nonceBytes, cipherBytes, ad)
	if err != nil {
		return nil, fmt.Errorf("aes gcm open: %w", err)
	}

	var result wechatPayNotifyResource
	if err := json.Unmarshal(plain, &result); err != nil {
		return nil, fmt.Errorf("unmarshal resource: %w", err)
	}

	return &result, nil
}

type WeChatNativeCreateReq struct {
	BusinessType string `json:"business_type"` // e.g. "riasec_basic"
	PlanKey      string `json:"plan_key"`      // e.g. "basic", "pro"，前端传的 TestTypeBasic 之类
}

type WeChatNativeCreateRes struct {
	Ok          bool   `json:"ok"`
	OrderID     string `json:"order_id"`              // 你自己的订单号（建议用 out_trade_no）
	CodeURL     string `json:"code_url"`              // 微信返回，用来生成二维码
	Amount      int64  `json:"amount"`                // 单位：分
	Description string `json:"description"`           // 商品描述
	ErrMessage  string `json:"err_message,omitempty"` // 失败时的错误信息
}

func (s *HttpSrv) apiWeChatCreateNativeOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	log := s.log.With().Str("handler", "apiWeChatCreateNativeOrder").Logger()

	var req WeChatNativeCreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("decode create native order req failed")
		writeJSON(w, http.StatusBadRequest, &WeChatNativeCreateRes{
			Ok:         false,
			ErrMessage: "invalid request body",
		})
		return
	}

	req.BusinessType = strings.TrimSpace(req.BusinessType)
	req.PlanKey = strings.TrimSpace(req.PlanKey)
	if req.BusinessType == "" || req.PlanKey == "" {
		writeJSON(w, http.StatusBadRequest, &WeChatNativeCreateRes{
			Ok:         false,
			ErrMessage: "missing business_type or plan_key",
		})
		return
	}

	// 1) 通过业务类型 + planKey 查价格 & 描述（你自己的配置）
	amount, description, err := s.lookupWeChatPlan(ctx, req.BusinessType, req.PlanKey)
	if err != nil {
		log.Error().Err(err).
			Str("business_type", req.BusinessType).
			Str("plan_key", req.PlanKey).
			Msg("lookup plan failed")
		writeJSON(w, http.StatusBadRequest, &WeChatNativeCreateRes{
			Ok:         false,
			ErrMessage: "unknown plan",
		})
		return
	}

	// 2) 生成商户订单号（建议全局唯一）
	outTradeNo := s.generateOutTradeNo()

	// 3) 调微信 v3 Native 下单，拿 code_url
	codeURL, err := s.wechatCreateNativeOrder(ctx, outTradeNo, amount, description)
	if err != nil {
		log.Error().Err(err).
			Str("out_trade_no", outTradeNo).
			Msg("wechat native create order failed")
		writeJSON(w, http.StatusBadGateway, &WeChatNativeCreateRes{
			Ok:         false,
			ErrMessage: "create wechat order failed",
		})
		return
	}

	// 4) 落库：记录订单，初始状态 NOTPAY
	var wxUnionID string
	if c, err := r.Cookie("wx_user"); err == nil && c.Value != "" {
		wxUnionID = c.Value
	}

	if err := s.saveWeChatOrder(ctx, outTradeNo, req.BusinessType, req.PlanKey, amount, description, wxUnionID); err != nil {
		log.Error().Err(err).
			Str("out_trade_no", outTradeNo).
			Msg("save order failed")
		writeJSON(w, http.StatusInternalServerError, &WeChatNativeCreateRes{
			Ok:         false,
			ErrMessage: "save order failed",
		})
		return
	}

	log.Info().
		Str("out_trade_no", outTradeNo).
		Str("business_type", req.BusinessType).
		Str("plan_key", req.PlanKey).
		Int64("amount", amount).
		Msg("wechat native order created")

	writeJSON(w, http.StatusOK, &WeChatNativeCreateRes{
		Ok:          true,
		OrderID:     outTradeNo,
		CodeURL:     codeURL,
		Amount:      amount,
		Description: description,
	})
}

type WeChatOrderStatusRes struct {
	Ok         bool   `json:"ok"`
	OrderID    string `json:"order_id"`
	Paid       bool   `json:"paid"`
	Status     string `json:"status"`                // e.g. "NOTPAY", "SUCCESS", "REFUND" ...
	ErrMessage string `json:"err_message,omitempty"` // 如果查询失败
}

func (s *HttpSrv) apiWeChatOrderStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	log := s.log.With().Str("handler", "apiWeChatOrderStatus").Logger()

	orderID := strings.TrimSpace(r.URL.Query().Get("order_id"))
	if orderID == "" {
		writeJSON(w, http.StatusBadRequest, &WeChatOrderStatusRes{
			Ok:         false,
			ErrMessage: "missing order_id",
		})
		return
	}

	order, err := s.getWeChatOrder(ctx, orderID)
	if err != nil {
		log.Error().Err(err).
			Str("order_id", orderID).
			Msg("getWeChatOrder failed")
		writeJSON(w, http.StatusInternalServerError, &WeChatOrderStatusRes{
			Ok:         false,
			OrderID:    orderID,
			ErrMessage: "query order failed",
		})
		return
	}
	if order == nil {
		writeJSON(w, http.StatusNotFound, &WeChatOrderStatusRes{
			Ok:         false,
			OrderID:    orderID,
			ErrMessage: "order not found",
		})
		return
	}

	writeJSON(w, http.StatusOK, &WeChatOrderStatusRes{
		Ok:      true,
		OrderID: orderID,
		Paid:    order.TradeState == "SUCCESS",
		Status:  order.TradeState,
	})
}

// 根据业务类型 + planKey 查金额（分）和商品描述
func (s *HttpSrv) lookupWeChatPlan(ctx context.Context, businessType, planKey string) (amount int64, desc string, err error) {
	// 最简单：先用硬编码，后面你可以改到 DB/配置文件
	switch businessType {
	case "riasec":
		switch planKey {
		case "basic":
			return 9900, "智择未来 · 基础版测评", nil
		case "pro":
			return 19900, "智择未来 · 专业版测评", nil
		}
	}
	return 0, "", fmt.Errorf("unknown plan: business_type=%s plan_key=%s", businessType, planKey)
}

func (s *HttpSrv) generateOutTradeNo() string {
	// 你可以换成更严格的生成方式（数据库自增 ID + 时间戳等）
	return fmt.Sprintf("WXT%s%06d", time.Now().Format("20060102150405"), rand2.Intn(1000000))
}

func (s *HttpSrv) saveWeChatOrder(
	ctx context.Context,
	orderID, businessType, planKey string,
	amount int64,
	desc string,
	wxUnionID string, // 从当前登录态里拿 wx_user
) error {
	po := &dbSrv.WeChatOrder{
		OrderID:      orderID,
		BusinessType: businessType,
		PlanKey:      planKey,
		AmountTotal:  amount,
		Currency:     "CNY",
		Description:  desc,
	}
	if wxUnionID != "" {
		po.WeChatUnionID = sql.NullString{String: wxUnionID, Valid: true}
	}
	return dbSrv.Instance().InsertWeChatOrder(ctx, po)
}

func (s *HttpSrv) getWeChatOrder(ctx context.Context, orderID string) (*dbSrv.WeChatOrder, error) {
	return dbSrv.Instance().FindWeChatOrderByID(ctx, orderID)
}

func (s *HttpSrv) updateWeChatOrderStatus(
	ctx context.Context,
	orderID string,
	tradeState string,
	resource *wechatPayNotifyResource,
	rawBody []byte,
) error {
	var transID *string
	var openID *string
	var paidAt *time.Time

	if resource != nil {
		if resource.TransactionID != "" {
			transID = &resource.TransactionID
		}
		if resource.Payer.OpenID != "" {
			openID = &resource.Payer.OpenID
		}
		if resource.TradeState == "SUCCESS" {
			// success_time 是 ISO8601 字符串，形如 2020-06-08T10:34:56+08:00
			if resource.SuccessTime != "" {
				if t, err := time.Parse(time.RFC3339, resource.SuccessTime); err == nil {
					paidAt = &t
				}
			}
		}
	}

	return dbSrv.Instance().UpdateWeChatOrderStatus(
		ctx,
		orderID,
		tradeState,
		transID,
		openID,
		paidAt,
		rawBody, // 把原始回调存一下
	)
}

// 解析商户私钥 PEM
func (s *HttpSrv) wechatMerchantPrivateKey() (*rsa.PrivateKey, error) {
	pemStr := s.payment.MchPrivateKeyPEM
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("invalid merchant private key pem")
	}

	var key any
	var err error

	if strings.Contains(block.Type, "PRIVATE KEY") {
		key, err = x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			// 有些是 pkcs1
			rsaKey, err2 := x509.ParsePKCS1PrivateKey(block.Bytes)
			if err2 != nil {
				return nil, fmt.Errorf("parse private key failed: %w, %v", err, err2)
			}
			return rsaKey, nil
		}
	} else {
		return nil, fmt.Errorf("unexpected pem type: %s", block.Type)
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not rsa")
	}
	return rsaKey, nil
}

// wechatCreateNativeOrder 调用微信 v3 Native 下单，返回 code_url
func (s *HttpSrv) wechatCreateNativeOrder(
	ctx context.Context,
	outTradeNo string,
	amount int64,
	description string,
) (string, error) {
	wxCfg := s.payment

	// 1) 组装请求体
	type amountReq struct {
		Total    int64  `json:"total"`
		Currency string `json:"currency"`
	}
	type nativeReq struct {
		AppID       string    `json:"appid"`
		MchID       string    `json:"mchid"`
		Description string    `json:"description"`
		OutTradeNo  string    `json:"out_trade_no"`
		NotifyURL   string    `json:"notify_url"`
		Amount      amountReq `json:"amount"`
	}

	reqBody := nativeReq{
		AppID:       wxCfg.AppID,
		MchID:       wxCfg.MchID,
		Description: description,
		OutTradeNo:  outTradeNo,
		NotifyURL:   wxCfg.NotifyURL, // eg. https://xxx/api/pay/wechat/callback
		Amount: amountReq{
			Total:    amount,
			Currency: "CNY",
		},
	}

	bodyBytes, err := json.Marshal(&reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal wechat native req: %w", err)
	}

	// 2) 生成签名相关参数
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	nonceStr := randomString(16) // 自己实现一个简单随机串

	// v3 签名串 = HTTP方法 + "\n" + URL路径(含query) + "\n" + timestamp + "\n" + nonceStr + "\n" + body + "\n"
	method := "POST"
	canonicalURL := "/v3/pay/transactions/native"
	signMessage := method + "\n" + canonicalURL + "\n" + timestamp + "\n" + nonceStr + "\n" + string(bodyBytes) + "\n"

	privKey, err := s.wechatMerchantPrivateKey()
	if err != nil {
		return "", fmt.Errorf("get merchant private key: %w", err)
	}

	hash := sha256.Sum256([]byte(signMessage))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", fmt.Errorf("sign message: %w", err)
	}

	sigBase64 := base64.StdEncoding.EncodeToString(signature)

	// 3) 组装 Authorization 头
	// 参考格式：
	// Authorization: WECHATPAY2-SHA256-RSA2048 mchid="...",nonce_str="...",signature="...",timestamp="...",serial_no="..."
	authHeader := fmt.Sprintf(
		`WECHATPAY2-SHA256-RSA2048 mchid="%s",nonce_str="%s",signature="%s",timestamp="%s",serial_no="%s"`,
		wxCfg.MchID,
		nonceStr,
		sigBase64,
		timestamp,
		wxCfg.MchSerial,
	)

	// 4) 发起 HTTP 请求
	url := "https://api.mch.weixin.qq.com" + canonicalURL
	httpReq, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("new http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json; charset=utf-8")
	httpReq.Header.Set("Authorization", authHeader)
	httpReq.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("do http request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("wechat native resp status=%d body=%s", resp.StatusCode, string(respBytes))
	}

	// 5) 解析响应，拿 code_url
	type nativeResp struct {
		CodeURL string `json:"code_url"`
		// 还有别的字段用得上可以一起解析
	}

	var nr nativeResp
	if err := json.Unmarshal(respBytes, &nr); err != nil {
		return "", fmt.Errorf("unmarshal native resp: %w", err)
	}
	if nr.CodeURL == "" {
		return "", fmt.Errorf("empty code_url in native resp")
	}

	return nr.CodeURL, nil
}

// 简单随机串
func randomString(n int) string {
	const letters = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		// fallback
		for i := range b {
			b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		}
		return string(b)
	}
	for i := range b {
		b[i] = letters[int(b[i])%len(letters)]
	}
	return string(b)
}
