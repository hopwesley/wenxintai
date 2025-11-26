package srv

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type UserProfile struct {
	Uid        string `json:"uid"`
	Location   string `json:"location,omitempty"`
	StudyId    string `json:"study_id,omitempty"`
	SchoolName string `json:"school_name,omitempty"`
	NickName   string `json:"nick_name,omitempty"`
	AvatarUrl  string `json:"avatar_url,omitempty"`
}

type wechatTokenResp struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	Scope        string `json:"scope"`
	UnionID      string `json:"unionid"`

	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// wxStatusResponse 是返回给前端轮询接口的响应结构
type wxStatusResponse struct {
	Status string `json:"status"`           // "pending" | "ok" | "expired"
	IsNew  *bool  `json:"is_new,omitempty"` // 只有 status == "ok" 时才会有
}

// HandleWeChatCallback 处理微信扫码登录回调
// GET /api/auth/wx/callback?code=xxx&state=xxx
func (s *HttpSrv) wechatSignInCallBack(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" || state == "" {
		http.Error(w, "missing code or state", http.StatusBadRequest)
		return
	}

	s.log.Info().
		Str("code", code).
		Str("state", state).
		Msg("WeChat callback start")

	token, err := s.exchangeWeChatCode(ctx, code)
	if err != nil {
		s.log.Error().
			Err(err).
			Str("code", code).
			Msg("exchange wechat code failed")
		http.Error(w, "wechat auth failed", http.StatusBadGateway)
		return
	}

	userInfo, err := s.fetchWeChatUserInfo(ctx, token.AccessToken, token.OpenID)
	if err != nil {
		s.log.Error().Err(err).Msg("fetch wechat userinfo failed")
		// 一般来说，这个失败不应该中断整个登录，可以继续走，只是没头像/昵称
	}

	s.log.Info().
		Str("openid", token.OpenID).
		Str("unionid", token.UnionID).
		Str("state", state).
		Interface("profile", userInfo).
		Msg("WeChat oauth success")

	// TODO: 这里就是你之后要接数据库的地方：
	// 1. 用 token.UnionID / token.OpenID 去查用户表
	// 2. 如果找不到 -> 创建新用户（isNew = true）
	// 3. 如果能找到 -> isNew = false
	//
	// 现在先用 WxLoginStore 模拟一下“新老用户”：
	entry := wxLoginStore.MarkLogin(state, token.OpenID, token.UnionID)

	// TODO: 真正的登录态（session / JWT）请在这里设置。
	// 下面这行只是 demo：把 openid 写到一个 cookie 里，方便你调试。
	http.SetCookie(w, &http.Cookie{
		Name:     "wx_user",
		Value:    entry.OpenID,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	// 注意：这个 callback 是跑在 <iframe> 里，用户基本看不到页面内容，
	// 所以返回一段简单 HTML 提示就够了。
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<html><body>微信登录成功，可以返回原页面。</body></html>`))
}

// exchangeWeChatCode 向微信服务器用 code 换取 access_token / openid / unionid
func (s *HttpSrv) exchangeWeChatCode(ctx context.Context, code string) (*wechatTokenResp, error) {
	v := url.Values{}
	v.Set("appid", s.cfg.WeChatAppID)
	v.Set("secret", s.cfg.WeChatAppSecret)
	v.Set("code", code)
	v.Set("grant_type", "authorization_code")

	u := "https://api.weixin.qq.com/sns/oauth2/access_token?" + v.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http do: %w", err)
	}
	defer resp.Body.Close()

	var token wechatTokenResp
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, fmt.Errorf("decode wechat token resp: %w", err)
	}

	if token.ErrCode != 0 {
		return nil, fmt.Errorf("wechat err: %d %s", token.ErrCode, token.ErrMsg)
	}

	if token.OpenID == "" {
		return nil, fmt.Errorf("empty openid in wechat resp")
	}

	return &token, nil
}

// HandleWeChatStatus 供前端轮询扫码登录状态
// GET /api/auth/wx/status?state=xxx
func (s *HttpSrv) wechatSignStatus(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	if state == "" {
		http.Error(w, "missing state", http.StatusBadRequest)
		return
	}

	entry, ok := wxLoginStore.Get(state)

	resp := wxStatusResponse{
		Status: "pending",
	}

	if ok {
		// 简单做一个过期判断，超过 10 分钟就认为过期
		if time.Since(entry.CreatedAt) > 10*time.Minute {
			resp.Status = "expired"
		} else {
			resp.Status = entry.Status
			if entry.Status == "ok" {
				resp.IsNew = &entry.IsNew
			}
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(resp)
}

type wechatUserInfoResp struct {
	OpenID     string `json:"openid"`
	Nickname   string `json:"nickname"`
	Sex        int    `json:"sex"`
	Province   string `json:"province"`
	City       string `json:"city"`
	Country    string `json:"country"`
	HeadImgURL string `json:"headimgurl"`
	UnionID    string `json:"unionid"`

	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (s *HttpSrv) fetchWeChatUserInfo(ctx context.Context, accessToken, openID string) (*wechatUserInfoResp, error) {
	v := url.Values{}
	v.Set("access_token", accessToken)
	v.Set("openid", openID)
	v.Set("lang", "zh_CN")

	u := "https://api.weixin.qq.com/sns/userinfo?" + v.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http do: %w", err)
	}
	defer resp.Body.Close()

	var ui wechatUserInfoResp
	if err := json.NewDecoder(resp.Body).Decode(&ui); err != nil {
		return nil, fmt.Errorf("decode wechat userinfo resp: %w", err)
	}

	if ui.ErrCode != 0 {
		return nil, fmt.Errorf("wechat userinfo err: %d %s", ui.ErrCode, ui.ErrMsg)
	}

	return &ui, nil
}
