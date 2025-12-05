package srv

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/hopwesley/wenxintai/server/dbSrv"
	"github.com/rs/zerolog"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
)

// ========================= 回调入口 =========================

func (s *HttpSrv) apiWeChatPayCallBack(w http.ResponseWriter, r *http.Request) {

	if len(s.cfg.PaymentForward) > 0 {
		s.forwardCallback(w, r, s.cfg.PaymentForward)
		return
	}
	s.processWeChatPayment(w, r)
}

type PaymentStatus int16

const (
	PaymentInit PaymentStatus = iota
	PaymentSuccess
	PaymentFailed
)

func (s *HttpSrv) processWeChatPayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sLog := s.log.With().Str("handler", "processWeChatPayment").Logger()

	if s.wxNotifyHandler == nil {
		sLog.Error().Msg("wxNotifyHandler not initialized")
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	tx := new(payments.Transaction)
	notifyReq, err := s.wxNotifyHandler.ParseNotifyRequest(ctx, r, tx)
	if err != nil {
		sLog.Err(err).Msg("parse notify failed")
		http.Error(w, "invalid notify", http.StatusBadRequest)
		return
	}

	outTradeNo := safeStr(tx.OutTradeNo)
	tradeState := safeStr(tx.TradeState)
	transID := safeStr(tx.TransactionId)
	openID := safeStr(tx.Payer.Openid)
	paidAt, _ := time.Parse(time.RFC3339, *tx.SuccessTime)
	rawBody, _ := json.Marshal(tx)

	sLog.Info().
		Str("event_type", notifyReq.EventType).
		Str("summary", notifyReq.Summary).
		Str("out_trade_no", outTradeNo).
		Str("trade_state", tradeState).
		Str("transaction_id", safeStr(tx.TransactionId)).
		Msg("wechat payment notify")

	var tState = PaymentInit
	switch tradeState {
	case "SUCCESS":
		tState = PaymentSuccess
		break
	case "REFUND", "CLOSED", "PAYERROR":
		tState = PaymentFailed
		break
	default:
		tState = PaymentInit
	}

	if err := dbSrv.Instance().UpdateWeChatOrderStatus(
		ctx,
		outTradeNo,
		int16(tState),
		transID,
		openID,
		paidAt,
		rawBody,
	); err != nil {
		sLog.Err(err).Str("out_trade_no", outTradeNo).Msg("updateWeChatOrderStatus failed")
		writeJSON(w, http.StatusOK, map[string]string{
			"code":    "FAILED",
			"message": "失败",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"code":    "SUCCESS",
		"message": "成功",
	})
}

// ========================= 创建 Native 订单 =========================

type WeChatNativeCreateReq struct {
	PublicId string `json:"public_id"`
}

func (req *WeChatNativeCreateReq) parseObj(r *http.Request) *ApiErr {
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return ApiInvalidReq("invalid request body", err)
	}

	if !IsValidPublicID(req.PublicId) {
		return ApiInvalidReq("无效的问卷编号", nil)
	}
	return nil
}

type WeChatNativeCreateRes struct {
	Ok          bool    `json:"ok"`
	OrderID     string  `json:"order_id"`
	CodeURL     string  `json:"code_url"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	ErrMessage  string  `json:"err_message,omitempty"`
}

func (s *HttpSrv) apiWeChatCreateNativeOrder(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	var req WeChatNativeCreateReq

	if err := req.parseObj(r); err != nil {
		s.log.Err(err).Msgf("[order creation]invalid request")
		writeError(w, err)
		return
	}

	sLog := s.log.With().Str("public_id", req.PublicId).Logger()
	uid := userIDFromContext(ctx)

	record, dbErr := dbSrv.Instance().QueryRecordByPid(ctx, req.PublicId)
	if dbErr != nil || record == nil {
		sLog.Err(dbErr).Msg("failed find test record")
		writeError(w, ApiInvalidNoTestRecord(dbErr))
		return
	}
	if record.WeChatID.String != uid {
		sLog.Err(dbErr).Msg("no right to operate test record")
		writeError(w, NewApiError(http.StatusForbidden, ErrorCodeForbidden, "无权操作", nil))
		return
	}

	if record.PayOrderId.Valid || record.PaidTime.Valid {
		sLog.Err(dbErr).Msg("the test record already paid")
		writeError(w, ApiInternalErr("重复的支付", nil))
		return
	}

	plan, planErr := dbSrv.Instance().PlanByKey(ctx, record.BusinessType)
	if planErr != nil {
		sLog.Err(planErr).Msg("failed find product price info")
		writeError(w, ApiInternalErr("查询产品价格信息失败", planErr))
		return
	}

	outTradeNo := s.generateOutTradeNo()

	if s.wxNativeService == nil {
		sLog.Error().Msg("wxNativeService not initialized")
		writeError(w, ApiInternalErr("支付系统初始化异常", nil))
		return
	}

	amount := int64(plan.Price * 100)
	resp, _, payErr := s.wxNativeService.Prepay(ctx, native.PrepayRequest{
		Appid:       core.String(s.payment.AppID),
		Mchid:       core.String(s.payment.MchID),
		Description: core.String(plan.Description),
		OutTradeNo:  core.String(outTradeNo),
		NotifyUrl:   core.String(s.payment.NotifyURL),
		Amount: &native.Amount{
			Total:    core.Int64(amount),
			Currency: core.String("CNY"),
		},
	})

	if payErr != nil {
		sLog.Err(payErr).Str("out_trade_no", outTradeNo).Msg("wechat native prepay failed")
		writeError(w, ApiInternalErr("创建支付订单失败", payErr))
		return
	}

	if resp.CodeUrl == nil || *resp.CodeUrl == "" {
		sLog.Error().Str("out_trade_no", outTradeNo).Msg("wechat native prepay failed")
		writeError(w, ApiInternalErr("生成支付二维码失败", nil))
		return
	}

	po := &dbSrv.WeChatOrder{
		OrderID:     outTradeNo,
		PublicID:    req.PublicId,
		PlanKey:     plan.PlanKey,
		AmountTotal: amount,
		Currency:    "CNY",
		Description: plan.Description,
	}

	if err := dbSrv.Instance().InsertWeChatOrder(ctx, po); err != nil {
		sLog.Err(err).Str("out_trade_no", outTradeNo).Msg("save order failed")
		writeError(w, ApiInternalErr("保存原始订单失败", nil))
		return
	}

	writeJSON(w, http.StatusOK, &WeChatNativeCreateRes{
		Ok:          true,
		OrderID:     outTradeNo,
		CodeURL:     *resp.CodeUrl,
		Amount:      plan.Price,
		Description: plan.Description,
	})
}

// ========================= 查询订单状态 =========================

type WeChatOrderStatusRes struct {
	Ok         bool   `json:"ok"`
	OrderID    string `json:"order_id"`
	Status     int16  `json:"status"`
	ErrMessage string `json:"err_message,omitempty"`
}

func (s *HttpSrv) generateOutTradeNo() string {
	return fmt.Sprintf("WXT_%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}

func (s *HttpSrv) apiWeChatOrderStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orderID := strings.TrimSpace(r.URL.Query().Get("order_id"))
	if orderID == "" {
		writeError(w, ApiInvalidReq("查询状态时需要订单编号", nil))
		return
	}

	sLog := s.log.With().Str("order_id", orderID).Logger()

	order, err := dbSrv.Instance().QueryWeChatOrderByID(ctx, orderID)
	if err != nil || order == nil {
		sLog.Err(err).Msg("getWeChatOrder failed")
		writeError(w, ApiInternalErr("无效的订单编号"+orderID, err))
		return
	}

	if order.TradeState == int16(PaymentInit) && time.Since(order.UpdatedAt) > 50*time.Second && s.wxNativeService != nil {
		s.queryStatusFromWeChatSrv(ctx, order, sLog)
	}

	writeJSON(w, http.StatusOK, &WeChatOrderStatusRes{
		Ok:      true,
		OrderID: orderID,
		Status:  order.TradeState,
	})

	sLog.Debug().
		Int16("status", order.TradeState).
		Str("order_id", orderID).
		Msg("query payment order status success")
}

func (s *HttpSrv) queryStatusFromWeChatSrv(ctx context.Context, order *dbSrv.WeChatOrder, sLog zerolog.Logger) {

	sLog.Info().Msg("too long time has no payment result, query from wechat server")
	resp, _, qErr := s.wxNativeService.QueryOrderByOutTradeNo(ctx, native.QueryOrderByOutTradeNoRequest{
		OutTradeNo: core.String(order.OrderID),
		Mchid:      core.String(s.payment.MchID),
	})

	if qErr != nil || resp == nil {
		sLog.Err(qErr).Msg("QueryOrderByOutTradeNo to wechat failed")
		return
	}

	var tState = PaymentInit
	if resp.TradeState != nil {
		switch *resp.TradeState {
		case "SUCCESS":
			tState = PaymentSuccess
		case "REFUND", "CLOSED", "PAYERROR":
			tState = PaymentFailed
		default:
			tState = PaymentInit
		}
	}

	transID := safeStr(resp.TransactionId)
	openID := ""
	if resp.Payer != nil {
		openID = safeStr(resp.Payer.Openid)
	}

	var paidAt time.Time
	if resp.SuccessTime != nil {
		if t, perr := time.Parse(time.RFC3339, *resp.SuccessTime); perr == nil {
			paidAt = t
		}
	}

	rawBody, _ := json.Marshal(resp)

	if err := dbSrv.Instance().UpdateWeChatOrderStatus(
		ctx,
		order.OrderID,
		int16(tState),
		transID,
		openID,
		paidAt,
		rawBody,
	); err != nil {
		sLog.Err(err).Msg("UpdateWeChatOrderStatus from query failed")
	} else {
		order.TradeState = int16(tState)
		sLog.Info().
			Int16("status", order.TradeState).
			Msg("order status refreshed from wechat server")
	}
}
