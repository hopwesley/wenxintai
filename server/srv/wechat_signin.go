package srv

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/hopwesley/wenxintai/server/dbSrv"
)

const (
	redirectUrlBase   = "https://%s/api/wechat_signin"
	wxApiUserInfo     = "https://api.weixin.qq.com/sns/userinfo?"
	wxApiExchangeCode = "https://api.weixin.qq.com/sns/oauth2/access_token?"
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

	existing, err := dbSrv.Instance().FindUserProfileByUid(ctx, token.UnionID)
	if err != nil {
		s.log.Error().Err(err).Msg("FindUserProfileByUid failed")
		http.Error(w, "wechat auth failed", http.StatusBadGateway)
		return
	}
	isNew := existing == nil

	if err := dbSrv.Instance().InsertOrUpdateUserProfileBasic(
		ctx,
		token.UnionID,
		nickName,
		avatarURL,
	); err != nil {
		s.log.Error().Err(err).Msg("InsertOrUpdateUserProfileBasic failed")
		http.Error(w, "wechat auth failed", http.StatusBadGateway)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "wx_user",
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
	ctx := r.Context()

	user, err := s.currentUserFromCookie(ctx, r)
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

	if user != nil {
		resp.Status = "ok"

		if c, err := r.Cookie("wx_is_new"); err == nil {
			v := c.Value == "1"
			resp.IsNew = &v
		}
		resp.NickName = user.NickName
		resp.AvatarURL = user.AvatarUrl
		resp.Uid = user.Uid
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(resp)
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

func (s *HttpSrv) currentUserFromCookie(ctx context.Context, r *http.Request) (*dbSrv.UserProfile, error) {
	c, err := r.Cookie("wx_user")
	if err != nil || c.Value == "" {
		return nil, nil
	}
	uid := c.Value

	profile, err := dbSrv.Instance().FindUserProfileByUid(ctx, uid)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, nil
	}
	return profile, nil
}

func (s *HttpSrv) wechatLogout(w http.ResponseWriter, _ *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "wx_user",
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

type UsrProfileExtra struct {
	City        string `json:"city"`
	Province    string `json:"province"`
	ParentPhone string `json:"parent_phone,omitempty"`
	StudyId     string `json:"study_id,omitempty"`
	SchoolName  string `json:"school_name"`
}

func (upe *UsrProfileExtra) parseObj(r *http.Request) *ApiErr {
	if r.Method != http.MethodPost {
		return ApiMethodInvalid
	}
	if err := json.NewDecoder(r.Body).Decode(upe); err != nil {
		return ApiInvalidReq("invalid request body", err)
	}
	if len(upe.Province) == 0 {
		return ApiInvalidReq("无效的省信息", nil)
	}
	if len(upe.City) == 0 {
		return ApiInvalidReq("无效的市信息", nil)
	}
	return nil
}

func (s *HttpSrv) apiWeChatUpdateProfile(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	user, err := s.currentUserFromCookie(ctx, r)
	if err != nil || user == nil {
		s.log.Err(err).Msg("apiWeChatUpdateProfile: no uid in cookie or no such user")
		writeError(w, ApiInvalidReq("请先登录", err))
		return
	}

	var extraData UsrProfileExtra
	if err := extraData.parseObj(r); err != nil {
		s.log.Err(err).Msg("apiWeChatUpdateProfile: parseObj failed")
		writeError(w, err)
		return
	}

	err = dbSrv.Instance().UpdateUserProfileExtra(ctx, user.Uid, extraData.ParentPhone, extraData.StudyId,
		extraData.SchoolName, extraData.Province, extraData.City)
	if err != nil {
		s.log.Err(err).Msg("apiWeChatUpdateProfile: update user profile failed")
		writeError(w, ApiInternalErr("更新基本信息失败", err))
		return
	}

	writeJSON(w, http.StatusOK, &CommonRes{Ok: true, Msg: "更新用户基本信息成功"})
	s.log.Info().Msg("apiWeChatUpdateProfile: update user profile success")
}
