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

func (s *HttpSrv) processWeChatPayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := s.log.With().Str("handler", "processWeChatPayment").Logger()

	if s.wxNotifyHandler == nil {
		log.Error().Msg("wxNotifyHandler not initialized")
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	var txn payments.Transaction
	notifyReq, err := s.wxNotifyHandler.ParseNotifyRequest(ctx, r, &txn)
	if err != nil {
		log.Error().Err(err).Msg("ParseNotifyRequest failed")
		http.Error(w, "invalid notify", http.StatusBadRequest)
		return
	}

	outTradeNo := ""
	if txn.OutTradeNo != nil {
		outTradeNo = *txn.OutTradeNo
	}
	tradeState := ""
	if txn.TradeState != nil {
		tradeState = *txn.TradeState
	}
	transactionID := ""
	if txn.TransactionId != nil {
		transactionID = *txn.TransactionId
	}

	log.Info().
		Str("event_type", notifyReq.EventType).
		Str("summary", notifyReq.Summary).
		Str("out_trade_no", outTradeNo).
		Str("trade_state", tradeState).
		Str("transaction_id", transactionID).
		Msg("wechat payment notify")

	switch tradeState {
	case "SUCCESS", "REFUND", "CLOSED", "PAYERROR":
		if err := s.updateWeChatOrderStatusFromTransaction(ctx, &txn); err != nil {
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
	PlanKey      string `json:"plan_key"`
}

type WeChatNativeCreateRes struct {
	Ok          bool   `json:"ok"`
	OrderID     string `json:"order_id"`
	CodeURL     string `json:"code_url"`
	Amount      int64  `json:"amount"`
	Description string `json:"description"`
	ErrMessage  string `json:"err_message,omitempty"`
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

	amount, desc, err := s.lookupWeChatPlan(req.BusinessType, req.PlanKey)
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

	outTradeNo := s.generateOutTradeNo()

	if s.wxNativeService == nil {
		log.Error().Msg("wxNativeService not initialized")
		writeJSON(w, http.StatusInternalServerError, &WeChatNativeCreateRes{
			Ok:         false,
			ErrMessage: "pay service not ready",
		})
		return
	}

	resp, _, err := s.wxNativeService.Prepay(ctx, native.PrepayRequest{
		Appid:       core.String(s.payment.AppID),
		Mchid:       core.String(s.payment.MchID),
		Description: core.String(desc),
		OutTradeNo:  core.String(outTradeNo),
		NotifyUrl:   core.String(s.payment.NotifyURL),
		Amount: &native.Amount{
			Total:    core.Int64(amount),
			Currency: core.String("CNY"),
		},
	})
	if err != nil {
		log.Error().Err(err).Str("out_trade_no", outTradeNo).Msg("wechat native prepay failed")
		writeJSON(w, http.StatusBadGateway, &WeChatNativeCreateRes{
			Ok:         false,
			ErrMessage: "create wechat order failed",
		})
		return
	}
	if resp.CodeUrl == nil || *resp.CodeUrl == "" {
		writeJSON(w, http.StatusInternalServerError, &WeChatNativeCreateRes{
			Ok:         false,
			ErrMessage: "empty code_url",
		})
		return
	}

	// 落库
	var wxUnionID string
	if c, err := r.Cookie("wx_user"); err == nil && c.Value != "" {
		wxUnionID = c.Value
	}
	if err := s.saveWeChatOrder(ctx, outTradeNo, req.BusinessType, req.PlanKey, amount, desc, wxUnionID); err != nil {
		log.Error().Err(err).Str("out_trade_no", outTradeNo).Msg("save order failed")
	}

	writeJSON(w, http.StatusOK, &WeChatNativeCreateRes{
		Ok:          true,
		OrderID:     outTradeNo,
		CodeURL:     *resp.CodeUrl,
		Amount:      amount,
		Description: desc,
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

// ========================= 本地配置 / DB =========================

func (s *HttpSrv) lookupWeChatPlan(businessType, planKey string) (int64, string, error) {
	switch businessType {
	case "riasec":
		switch planKey {
		case "basic":
			return 9900, "智择未来 · 基础版测评", nil
		case "pro":
			return 19900, "智择未来 · 专业版测评", nil
		}
	}
	return 0, "", fmt.Errorf("unknown plan: %s/%s", businessType, planKey)
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
