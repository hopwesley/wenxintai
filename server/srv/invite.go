package srv

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/hopwesley/wenxintai/server/dbSrv"
)

type inviteVerifyRequest struct {
	InviteCode   string `json:"invite_code"`
	BusinessType string `json:"business_type"`
}

func (req *inviteVerifyRequest) parseObj(r *http.Request) *ApiErr {
	if r.Method != http.MethodPost {
		return ApiMethodInvalid
	}

	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return ApiInvalidReq("json 解析参数失败", nil)
	}

	if strings.TrimSpace(req.InviteCode) == "" {
		return ApiInvalidReq("无效的邀请码", nil)
	}

	if strings.TrimSpace(req.BusinessType) == "" {
		return ApiInvalidReq("无效的业务类型", nil)
	}

	return nil
}

type inviteVerifyResponse struct {
	OK       bool   `json:"ok"`
	Reason   string `json:"reason"`
	PublicId string `json:"public_id,omitempty"`
}

func (s *HttpSrv) handleInviteVerify(w http.ResponseWriter, r *http.Request) {

	var req inviteVerifyRequest
	if err := req.parseObj(r); err != nil {
		s.log.Err(err).Msg("decode requests error when handleInviteVerify")
		writeError(w, err)
		return
	}

	localLog := s.log.With().Str("invite_code", req.InviteCode).Str("business_type", req.BusinessType).Logger()

	ctx := r.Context()
	inv, err := dbSrv.Instance().GetInviteByCode(ctx, req.InviteCode)
	if err != nil {
		localLog.Err(err).Msg("get invite error")
		writeError(w, ApiInternalErr("查询邀请码数据库失败", err))
		return
	}

	if inv == nil {
		resp := inviteVerifyResponse{
			OK:     false,
			Reason: "无此邀请码",
		}
		localLog.Info().Msg("invite code not found")
		writeJSON(w, http.StatusOK, resp)
		return
	}

	if inv.PublicID.Valid {
		resp := inviteVerifyResponse{
			OK:       true,
			Reason:   "试卷已创建",
			PublicId: inv.PublicID.String,
		}

		localLog.Info().Msg("test record found by invite code")
		writeJSON(w, http.StatusOK, resp)
		return
	}

	now := time.Now()
	if inv.ExpiresAt.Valid && inv.ExpiresAt.Time.Before(now) {
		resp := inviteVerifyResponse{
			OK:     false,
			Reason: "邀请码过期",
		}
		localLog.Info().Msg("invite code expired")
		writeJSON(w, http.StatusOK, resp)
		return
	}

	if inv.Status != dbSrv.InviteStatusUnused {
		resp := inviteVerifyResponse{
			OK:     false,
			Reason: "当前邀请码已经被使用",
		}
		localLog.Info().Msg("invite code invalid")
		writeJSON(w, http.StatusOK, resp)
		return
	}

	publicID, dbErr := dbSrv.Instance().NewTestRecord(ctx, req.BusinessType, &req.InviteCode, nil)
	if dbErr != nil {
		localLog.Err(dbErr).Msg("failed to create test record")
		writeError(w, ApiInternalErr("创建测试问卷失败", err))
		return
	}

	resp := inviteVerifyResponse{
		OK:       true,
		Reason:   "新建试卷成功",
		PublicId: publicID,
	}

	localLog.Debug().Msg("invite code found")
	writeJSON(w, http.StatusOK, resp)
}
