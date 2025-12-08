package srv

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/hopwesley/wenxintai/server/dbSrv"
)

const (
	cookieKey         = "wx_user"
	redirectUrlBase   = "https://%s/api/wechat_signin"
	wxApiUserInfo     = "https://api.weixin.qq.com/sns/userinfo?"
	wxApiExchangeCode = "https://api.weixin.qq.com/sns/oauth2/access_token?"
	wxApiMiniAppToken = "https://api.weixin.qq.com/sns/jscode2session?"
)

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

type wxStatusResponse struct {
	Status      string `json:"status"`           // "pending" | "ok" | "expired"
	IsNew       *bool  `json:"is_new,omitempty"` // 只有 status == "ok" 时才会有
	Uid         string `json:"uid,omitempty"`
	NickName    string `json:"nick_name,omitempty"` // 登录后返回
	AvatarURL   string `json:"avatar_url,omitempty"`
	AppID       string `json:"appid,omitempty"`        // 微信扫码登录用的 appid
	RedirectURI string `json:"redirect_uri,omitempty"` // 微信扫码登录回调地址
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

	if token.UnionID == "" {
		s.log.Error().
			Str("openid", token.OpenID).
			Msg("empty unionid in wechat resp")
		http.Error(w, "wechat auth failed: empty unionid", http.StatusBadGateway)
		return
	}

	userInfo, err := s.fetchWeChatUserInfo(ctx, token.AccessToken, token.OpenID)
	if err != nil {
		s.log.Error().Err(err).Msg("fetch wechat userinfo failed")
	}

	var nickName, avatarURL string
	if userInfo != nil {
		nickName = userInfo.Nickname
		avatarURL = userInfo.HeadImgURL
	}

	s.log.Info().
		Str("openid", token.OpenID).
		Str("unionid", token.UnionID).
		Str("state", state).
		Str("nick", nickName).
		Str("avatar", avatarURL).
		Msg("WeChat oauth success")

	existing, err := dbSrv.Instance().QueryUserProfileUid(ctx, token.UnionID)
	if err != nil {
		s.log.Error().Err(err).Msg("QueryUserProfileUid failed")
		http.Error(w, "wechat auth failed", http.StatusBadGateway)
		return
	}
	isNew := existing == nil

	if err := dbSrv.Instance().InsertOrUpdateWeChatInfo(
		ctx,
		token.UnionID,
		nickName,
		avatarURL,
	); err != nil {
		s.log.Error().Err(err).Msg("InsertOrUpdateWeChatInfo failed")
		http.Error(w, "wechat auth failed", http.StatusBadGateway)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookieKey,
		Value:    token.UnionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	isNewVal := "0"
	if isNew {
		isNewVal = "1"
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "wx_is_new",
		Value:    isNewVal,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   12 * 3600,
	})

	http.Redirect(w, r, "/home", http.StatusFound)
}

func (s *HttpSrv) exchangeWeChatCode(ctx context.Context, code string) (*wechatTokenResp, error) {
	v := url.Values{}
	v.Set("appid", s.cfg.WeChatAppID)
	v.Set("secret", s.cfg.WeChatAppSecret)
	v.Set("code", code)
	v.Set("grant_type", "authorization_code")

	u := wxApiExchangeCode + v.Encode()

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

func (s *HttpSrv) wechatSignStatus(w http.ResponseWriter, r *http.Request) {

	uid, err := s.currentUserFromCookie(r)
	if err != nil {
		s.log.Err(err).Msg("wechatSignStatus: currentUserFromCookie failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	redirectUrl := fmt.Sprintf(redirectUrlBase, s.cfg.WeChatRedirectDomain)
	resp := wxStatusResponse{
		Status:      "pending",
		AppID:       s.cfg.WeChatAppID,
		RedirectURI: redirectUrl,
	}

	if len(uid) == 0 {
		writeJSON(w, http.StatusOK, resp)
		return
	}

	user, dbErr := dbSrv.Instance().QueryUserProfileUid(r.Context(), uid)
	if dbErr != nil || user == nil {
		resp.Status = "expired"
	} else {
		resp.Status = "ok"
		if c, err := r.Cookie("wx_is_new"); err == nil {
			v := c.Value == "1"
			resp.IsNew = &v
		}
		resp.NickName = user.NickName
		resp.AvatarURL = user.AvatarUrl
		resp.Uid = user.Uid
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *HttpSrv) fetchWeChatUserInfo(ctx context.Context, accessToken, openID string) (*wechatUserInfoResp, error) {
	v := url.Values{}
	v.Set("access_token", accessToken)
	v.Set("openid", openID)
	v.Set("lang", "zh_CN")

	u := wxApiUserInfo + v.Encode()

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

func (s *HttpSrv) currentUserFromCookie(r *http.Request) (string, error) {
	c, err := r.Cookie(cookieKey)
	if err != nil || c.Value == "" {
		return "", nil
	}
	uid := c.Value
	return uid, nil
}

func (s *HttpSrv) wechatLogout(w http.ResponseWriter, _ *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieKey,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:   "wx_is_new",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	w.WriteHeader(http.StatusNoContent)
}

func (s *HttpSrv) apiWeChatUpdateProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := userIDFromContext(ctx)

	var extraData dbSrv.UsrProfileExtra
	if err := json.NewDecoder(r.Body).Decode(&extraData); err != nil {
		s.log.Err(err).Msg("apiWeChatUpdateProfile: parameter of request invalid")
		writeError(w, ApiInvalidReq("invalid request body", err))
		return
	}

	if err := dbSrv.Instance().UpdateUserProfileExtra(ctx, uid, extraData); err != nil {
		s.log.Err(err).Msg("apiWeChatUpdateProfile: update user profile failed")
		writeError(w, ApiInternalErr("更新基本信息失败", err))
		return
	}

	writeJSON(w, http.StatusOK, &CommonRes{Ok: true, Msg: "更新用户基本信息成功"})
	s.log.Info().Msg("apiWeChatUpdateProfile: update user profile success")
}

type TestResponse struct {
	Profile *dbSrv.UserProfile `json:"profile"`
	Tests   []*dbSrv.TestItem  `json:"tests,omitempty"`
}

func (s *HttpSrv) apiWeChatMyProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	uid := userIDFromContext(ctx)

	tests, dbErr := dbSrv.Instance().QueryTestInfos(ctx, uid)
	if dbErr != nil {
		s.log.Err(dbErr).Msg("apiWeChatMyProfile: query test records and reports failed")
		writeError(w, ApiInternalErr("查询问卷数据失败", dbErr))
		return
	}
	user, pDBErr := dbSrv.Instance().QueryUserProfileUid(ctx, uid)
	if pDBErr != nil {
		s.log.Err(pDBErr).Msg("apiWeChatMyProfile: query user profile failed")
		writeError(w, ApiInternalErr("查询用户数据失败", dbErr))
		return
	}

	var resp = &TestResponse{
		Profile: user,
		Tests:   tests,
	}
	writeJSON(w, http.StatusOK, resp)
	s.log.Info().Str("wechat_id", uid).Msg("apiWeChatMyProfile: query user profile success")
}

// 小程序登录：根据 wx.login 的 code 换取 unionid，并可同时更新头像昵称
func (s *HttpSrv) apiMiniAppSignIn(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 解析请求体 { "code": "xxx", "nick_name": "...", "avatar_url": "..." }
	var req struct {
		Code      string `json:"code"`
		NickName  string `json:"nick_name"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	req.Code = strings.TrimSpace(req.Code)
	if req.Code == "" {
		http.Error(w, "code is required", http.StatusBadRequest)
		return
	}

	// 2. 用 code 向微信小程序服务端换取 openid / unionid / session_key
	openid, unionid, sessionKey, err := s.miniAppCode2Session(ctx, req.Code)
	if err != nil {
		s.log.Err(err).Msg("miniAppSignIn: jscode2session failed")
		http.Error(w, "wechat miniapp signin failed", http.StatusBadGateway)
		return
	}

	// TODO: 如果以后要做手机号解密或其它加密数据解密，可以在这里把 sessionKey 存到 session 表里
	_ = sessionKey

	if unionid == "" {
		s.log.Error().
			Str("openid", openid).
			Msg("miniapp signin: empty unionid in wechat resp")
		http.Error(w, "wechat miniapp signin failed: empty unionid", http.StatusBadGateway)
		return
	}

	// 3. 根据 unionid 查/建用户，并可同时更新头像昵称
	existing, err := dbSrv.Instance().QueryUserProfileUid(ctx, unionid)
	if err != nil {
		s.log.Error().Err(err).Msg("miniapp signin: QueryUserProfileUid failed")
		http.Error(w, "wechat miniapp signin failed", http.StatusBadGateway)
		return
	}
	isNew := existing == nil

	// 清理一下字符串
	req.NickName = strings.TrimSpace(req.NickName)
	req.AvatarURL = strings.TrimSpace(req.AvatarURL)

	// 如果前端传了头像昵称，就同步更新到你现在的 wechat info 表里
	if req.NickName != "" || req.AvatarURL != "" {
		if err := dbSrv.Instance().InsertOrUpdateWeChatInfo(
			ctx,
			unionid,
			req.NickName,
			req.AvatarURL,
		); err != nil {
			s.log.Error().Err(err).Msg("miniapp signin: InsertOrUpdateWeChatInfo failed")
			http.Error(w, "wechat miniapp signin failed", http.StatusBadGateway)
			return
		}
	} else if isNew {
		// 可选：如果是新用户但没传头像昵称，可以先插一条空记录，和网站逻辑对齐
		if err := dbSrv.Instance().InsertOrUpdateWeChatInfo(
			ctx,
			unionid,
			"",
			"",
		); err != nil {
			s.log.Error().Err(err).Msg("miniapp signin: InsertOrUpdateWeChatInfo (empty) failed")
			http.Error(w, "wechat miniapp signin failed", http.StatusBadGateway)
			return
		}
	}

	// 4. 写 cookie，和网站保持一致
	http.SetCookie(w, &http.Cookie{
		Name:     cookieKey, // "wx_user"
		Value:    unionid,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	// 可选：写 wx_is_new
	isNewVal := "0"
	if isNew {
		isNewVal = "1"
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "wx_is_new",
		Value:    isNewVal,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   12 * 3600,
	})

	// 5. 返回 token（现在就是 unionid）
	resp := map[string]string{
		"token": unionid,
	}

	writeJSON(w, http.StatusOK, resp)

	s.log.Info().
		Str("openid", openid).
		Str("unionid", unionid).
		Msg("miniapp signin success")
}

type miniAppCode2SessionResp struct {
	OpenID     string `json:"openid"`
	UnionID    string `json:"unionid"`
	SessionKey string `json:"session_key"`

	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (s *HttpSrv) miniAppCode2Session(ctx context.Context, code string) (openid, unionid, sessionKey string, err error) {
	v := url.Values{}
	v.Set("appid", s.miniCfg.MiniAppAppID)
	v.Set("secret", s.miniCfg.MiniAppAppSecret)
	v.Set("js_code", code)
	v.Set("grant_type", "authorization_code")

	u := wxApiMiniAppToken + v.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", "", "", fmt.Errorf("new request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", "", fmt.Errorf("http do: %w", err)
	}
	defer resp.Body.Close()

	var body miniAppCode2SessionResp
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", "", "", fmt.Errorf("decode jscode2session resp: %w", err)
	}

	if body.ErrCode != 0 {
		return "", "", "", fmt.Errorf("jscode2session err: %d %s", body.ErrCode, body.ErrMsg)
	}

	if body.OpenID == "" {
		return "", "", "", fmt.Errorf("empty openid in jscode2session resp")
	}

	return body.OpenID, body.UnionID, body.SessionKey, nil
}
