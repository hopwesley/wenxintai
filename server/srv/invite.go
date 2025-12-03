package srv

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/hopwesley/wenxintai/server/dbSrv"
)

type payByInviteReq struct {
	InviteCode string `json:"invite_code"`
	PublicID   string `json:"public_id"`
}

func (req *payByInviteReq) parseObj(r *http.Request) *ApiErr {
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return ApiInvalidReq("json 解析参数失败", nil)
	}
	if strings.TrimSpace(req.InviteCode) == "" {
		return ApiInvalidReq("无效的邀请码", nil)
	}
	if !IsValidPublicID(req.PublicID) {
		return ApiInvalidReq("无效的问卷编号", nil)
	}

	return nil
}

func (s *HttpSrv) apiPayByInvite(w http.ResponseWriter, r *http.Request) {
	var req payByInviteReq
	if err := req.parseObj(r); err != nil {
		s.log.Err(err).Msg("decode requests error when payByInvite")
		writeError(w, err)
		return
	}

	sLog := s.log.With().Str("invite_code", req.InviteCode).Str("public_id", req.PublicID).Logger()

	ctx := r.Context()
	inv, err := dbSrv.Instance().GetInviteByCode(ctx, req.InviteCode)
	if err != nil {
		sLog.Err(err).Msg("get invite error")
		writeError(w, ApiInternalErr("查询邀请码数据库失败", err))
		return
	}

	if inv == nil {
		sLog.Info().Msg("invite code not found")
		writeError(w, ApiInternalErr("无此邀请码", nil))
		return
	}

	now := time.Now()
	if inv.ExpiresAt.Valid && inv.ExpiresAt.Time.Before(now) {
		sLog.Info().Str("expired", inv.ExpiresAt.Time.String()).Msg("invite code expired")
		writeError(w, ApiInternalErr("邀请码过期", nil))
		return
	}

	if inv.Status != dbSrv.InviteStatusUnused {
		sLog.Info().Int16("status", inv.Status).Msg("invite code invalid")
		writeError(w, ApiInternalErr("当前邀请码已经被使用", nil))
		return
	}

	if dbErr := dbSrv.Instance().PayByInviteCode(ctx, req.PublicID, req.InviteCode); dbErr != nil {
		sLog.Err(dbErr).Msg("pay error")
		writeError(w, ApiInternalErr("更新支付状态失败", nil))
		return
	}

	sLog.Info().Msg("create test record success")
	writeJSON(w, http.StatusOK, CommonRes{Ok: true, Msg: "邀请码支付成功"})
}
