package srv

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/hopwesley/wenxintai/server/dbSrv"
)

type inviteVerifyRequest struct {
	InviteCode string `json:"invite_code"`
}

type inviteVerifyResponse struct {
	OK       bool   `json:"ok"`
	Reason   string `json:"reason"`
	PublicId string `json:"public_id,omitempty"`
}

func (s *HttpSrv) handleInviteVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.log.Error().Interface("method", r.Method).Interface("request", r).Msg("invalid method")
		writeError(w, ApiMethodInvalid)
		return
	}

	var req inviteVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.log.Err(err).Msg("decode json error when handleInviteVerify")
		writeError(w, ApiInvalidReq("无效的邀请码", nil))
		return
	}

	code := strings.TrimSpace(req.InviteCode)
	if code == "" {
		s.log.Error().Msg("code is empty when handleInviteVerify")
		writeError(w, ApiInvalidReq("无效的邀请码", nil))
		return
	}

	ctx := r.Context()
	// ---------- 第一步：先查 tests_record ----------
	rec, err := dbSrv.Instance().FindTestRecordByUid(ctx, code, "")
	if err != nil {
		s.log.Err(err).Str("invite_code", code).Msg("find test record error")
		writeError(w, NewApiError(http.StatusInternalServerError, "db_error_tests_record", "查询问卷数据库失败", err))
		return
	}

	if rec != nil {
		resp := inviteVerifyResponse{
			OK:       true,
			Reason:   "ok_existing_record",
			PublicId: rec.PublicId,
		}
		s.log.Info().Str("invite_code", code).Msg("test record found by invite code")
		writeJSON(w, http.StatusOK, resp)
		return
	}

	// ---------- 第二步：再查 invites 表 ----------
	inv, err := dbSrv.Instance().GetInviteByCode(ctx, code)
	if err != nil {
		s.log.Err(err).Str("invite_code", code).Msg("get invite error")
		writeError(w, NewApiError(http.StatusInternalServerError, "db_error_invites", "查询邀请码数据库失败", err))
		return
	}

	if inv == nil {
		resp := inviteVerifyResponse{
			OK:     false,
			Reason: "无此邀请码",
		}
		s.log.Info().Str("invite_code", code).Msg("invite code not found")
		writeJSON(w, http.StatusOK, resp)
		return
	}

	now := time.Now()
	if inv.ExpiresAt.Valid && inv.ExpiresAt.Time.Before(now) {
		resp := inviteVerifyResponse{
			OK:     false,
			Reason: "邀请码过期",
		}
		s.log.Info().Str("invite_code", code).Msg("invite code expired")
		writeJSON(w, http.StatusOK, resp)
		return
	}

	// 检查 status 是否可用（这里只把 status=0 当作可用）
	if inv.Status != dbSrv.InviteStatusUnused {
		resp := inviteVerifyResponse{
			OK:     false,
			Reason: "当前邀请码已经被使用",
		}
		s.log.Info().Str("invite_code", code).Msg("invite code invalid")
		writeJSON(w, http.StatusOK, resp)
		return
	}

	resp := inviteVerifyResponse{
		OK:     true,
		Reason: "ok_no_record",
	}

	s.log.Debug().Str("invite_code", code).Msg("invite code found")
	writeJSON(w, http.StatusOK, resp)
}
