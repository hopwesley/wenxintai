package srv

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/hopwesley/wenxintai/server/dbSrv"
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

func safeStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func (s *HttpSrv) processWeChatPayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := s.log.With().Str("handler", "processWeChatPayment").Logger()

	if s.wxNotifyHandler == nil {
		log.Error().Msg("wxNotifyHandler not initialized")
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	tx := new(payments.Transaction)
	notifyReq, err := s.wxNotifyHandler.ParseNotifyRequest(ctx, r, tx)
	if err != nil {
		log.Error().Err(err).Msg("parse notify failed")
		http.Error(w, "invalid notify", http.StatusBadRequest)
		return
	}

	outTradeNo := safeStr(tx.OutTradeNo)
	tradeState := safeStr(tx.TradeState)

	log.Info().
		Str("event_type", notifyReq.EventType).
		Str("summary", notifyReq.Summary).
		Str("out_trade_no", outTradeNo).
		Str("trade_state", tradeState).
		Str("transaction_id", safeStr(tx.TransactionId)).
		Msg("wechat payment notify")

	switch tradeState {
	case "SUCCESS", "REFUND", "CLOSED", "PAYERROR":
		if err := s.updateWeChatOrderStatusFromTransaction(ctx, tx); err != nil {
			log.Error().Err(err).Str("out_trade_no", outTradeNo).Msg("updateWeChatOrderStatus failed")
		}
	default:
		// NOTPAY 等状态，看你业务要不要处理
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"code":    "SUCCESS",
		"message": "成功",
	})
}

func (s *HttpSrv) updateWeChatOrderStatusFromTransaction(
	ctx context.Context,
	t *payments.Transaction,
) error {
	if t == nil || t.OutTradeNo == nil {
		return fmt.Errorf("transaction or out_trade_no nil")
	}

	orderID := *t.OutTradeNo

	tradeState := ""
	if t.TradeState != nil {
		tradeState = *t.TradeState
	}

	var transID *string
	if t.TransactionId != nil && *t.TransactionId != "" {
		transID = t.TransactionId
	}

	var openID *string
	if t.Payer != nil && t.Payer.Openid != nil && *t.Payer.Openid != "" {
		openID = t.Payer.Openid
	}

	var paidAt *time.Time
	if tradeState == "SUCCESS" && t.SuccessTime != nil && *t.SuccessTime != "" {
		if tt, err := time.Parse(time.RFC3339, *t.SuccessTime); err == nil {
			paidAt = &tt
		}
	}

	rawBody, _ := json.Marshal(t)

	return dbSrv.Instance().UpdateWeChatOrderStatus(
		ctx,
		orderID,
		tradeState,
		transID,
		openID,
		paidAt,
		rawBody,
	)
}

// ========================= 创建 Native 订单 =========================

type WeChatNativeCreateReq struct {
	BusinessType string `json:"business_type"`
	PublicId     string `json:"public_id"`
}

func (req *WeChatNativeCreateReq) parseObj(r *http.Request) *ApiErr {
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return ApiInvalidReq("invalid request body", err)
	}

	if !isValidBusinessType(req.BusinessType) {
		return ApiInvalidReq("无效的测试类型", nil)
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

	sLog := s.log.With().Str("public_id", req.PublicId).Str("business_type", req.BusinessType).Logger()

	record, dbErr := dbSrv.Instance().QueryUnfinishedTest(ctx, req.PublicId)
	if dbErr != nil || record == nil {
		sLog.Err(dbErr).Msg("failed find test record")
		writeError(w, ApiInvalidNoTestRecord(dbErr))
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

	// 落库
	var wxUnionID = userIDFromContext(ctx)

	if err := s.saveWeChatOrder(ctx, outTradeNo, req.BusinessType, record.BusinessType, amount, plan.Description, wxUnionID); err != nil {
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
	Paid       bool   `json:"paid"`
	Status     string `json:"status"`
	ErrMessage string `json:"err_message,omitempty"`
}

func (s *HttpSrv) apiWeChatOrderStatus(w http.ResponseWriter, r *http.Request) {
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
		log.Error().Err(err).Str("order_id", orderID).Msg("getWeChatOrder failed")
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

	// 如果不是终态，可以调用微信侧查询最新状态（可选）
	if order.TradeState != "SUCCESS" && order.TradeState != "REFUND" && order.TradeState != "CLOSED" {
		if s.wxNativeService != nil {
			tx, _, err := s.wxNativeService.QueryOrderByOutTradeNo(ctx, native.QueryOrderByOutTradeNoRequest{
				OutTradeNo: core.String(orderID),
				Mchid:      core.String(s.payment.MchID),
			})
			if err == nil && tx != nil && tx.TradeState != nil {
				order.TradeState = *tx.TradeState
				_ = s.updateWeChatOrderStatusFromTransaction(ctx, (*payments.Transaction)(tx))
			}
		}
	}

	writeJSON(w, http.StatusOK, &WeChatOrderStatusRes{
		Ok:      true,
		OrderID: orderID,
		Paid:    order.TradeState == "SUCCESS",
		Status:  order.TradeState,
	})
}

func (s *HttpSrv) generateOutTradeNo() string {
	return fmt.Sprintf("WXT%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}

func (s *HttpSrv) saveWeChatOrder(
	ctx context.Context,
	orderID, businessType, planKey string,
	amount int64,
	desc string,
	wxUnionID string,
) error {
	po := &dbSrv.WeChatOrder{
		OrderID:      orderID,
		BusinessType: businessType,
		PlanKey:      planKey,
		AmountTotal:  amount,
		Currency:     "CNY",
		Description:  desc,
		TradeState:   "NOTPAY",
	}
	if wxUnionID != "" {
		po.WeChatUnionID = sql.NullString{String: wxUnionID, Valid: true}
	}
	return dbSrv.Instance().InsertWeChatOrder(ctx, po)
}

func (s *HttpSrv) getWeChatOrder(ctx context.Context, orderID string) (*dbSrv.WeChatOrder, error) {
	return dbSrv.Instance().FindWeChatOrderByID(ctx, orderID)
}
