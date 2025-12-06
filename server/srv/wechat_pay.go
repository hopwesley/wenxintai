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

type PaymentStatus int16

const (
	PaymentInit PaymentStatus = iota
	PaymentSuccess
	PaymentFailed
	PayOrderTimeout = time.Minute * 90
)

type WeChatNativeCreateRes struct {
	Ok          bool    `json:"ok"`
	OrderID     string  `json:"order_id"`
	CodeURL     string  `json:"code_url"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	ErrMessage  string  `json:"err_message,omitempty"`
}

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

// ========================= 回调入口 =========================

func (s *HttpSrv) apiWeChatPayCallBack(w http.ResponseWriter, r *http.Request) {

	if len(s.cfg.PaymentForward) > 0 {
		s.forwardCallback(w, r, s.cfg.PaymentForward)
		return
	}
	s.processWeChatPayment(w, r)
}

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

	testRecord, dbErr := dbSrv.Instance().QueryTestRecord(ctx, req.PublicId, uid)
	if dbErr != nil || testRecord == nil {
		sLog.Err(dbErr).Msg("failed find test testRecord")
		writeError(w, ApiInvalidNoTestRecord(dbErr))
		return
	}

	if testRecord.PayOrderId.Valid || testRecord.PaidTime.Valid {
		sLog.Error().Msg("the test testRecord already paid")
		writeError(w, ApiInternalErr("重复的支付", nil))
		return
	}

	cutoff := time.Now().Add(-90 * time.Minute)
	order, orderErr := dbSrv.Instance().QueryUnfinishedOrder(ctx, req.PublicId, cutoff)
	if orderErr != nil {
		sLog.Err(orderErr).Msg("the test testRecord already paid")
		writeError(w, ApiInternalErr("确认订单状态时数据库错误", orderErr))
		return
	}

	plan, planErr := dbSrv.Instance().PlanByKey(ctx, testRecord.BusinessType)
	if planErr != nil {
		sLog.Err(planErr).Msg("failed find product price info")
		writeError(w, ApiInternalErr("查询产品价格信息失败", planErr))
		return
	}

	amount := int64(plan.Price * 100)

	if order != nil && amount == order.AmountTotal {
		sLog.Info().Str("order_id", order.OrderID).Msg("order found")

		writeJSON(w, http.StatusOK, &WeChatNativeCreateRes{
			Ok:          true,
			OrderID:     order.OrderID,
			CodeURL:     order.CodeUrl.String,
			Amount:      plan.Price,
			Description: plan.Description,
		})

		return
	}

	outTradeNo := s.generateOutTradeNo()

	if s.wxNativeService == nil {
		sLog.Error().Msg("wxNativeService not initialized")
		writeError(w, ApiInternalErr("支付系统初始化异常", nil))
		return
	}

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
		CodeUrl:     toNullString(resp.CodeUrl),
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

	order, err := dbSrv.Instance().QueryWeChatOrderByOrderID(ctx, orderID)
	if err != nil || order == nil {
		sLog.Err(err).Msg("getWeChatOrder failed")
		writeError(w, ApiInternalErr("无效的订单编号"+orderID, err))
		return
	}

	if order.TradeState == int16(PaymentInit) &&
		time.Since(order.UpdatedAt) > time.Duration(s.cfg.WxPaymentTimeout)*time.Second &&
		s.wxNativeService != nil {
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
		sLog.Info().Str("trade_status", *resp.TradeState).Msg("query wechat payment status success")
		switch *resp.TradeState {
		case "SUCCESS":
			tState = PaymentSuccess
			break
		case "REFUND", "CLOSED", "PAYERROR":
			tState = PaymentFailed
			break
		default:
			tState = PaymentInit
			return
		}
	}

	transID := safeStr(resp.TransactionId)
	openID := ""
	if resp.Payer != nil {
		openID = safeStr(resp.Payer.Openid)
	}

	var paidAt time.Time
	if resp.SuccessTime != nil {
		if t, pErr := time.Parse(time.RFC3339, *resp.SuccessTime); pErr == nil {
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
