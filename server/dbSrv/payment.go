package dbSrv

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type WeChatOrder struct {
	ID            int64
	OrderID       string
	BusinessType  string
	PlanKey       string
	AmountTotal   int64
	Currency      string
	Description   string
	WeChatUnionID sql.NullString
	WeChatOpenID  sql.NullString
	TransactionID sql.NullString
	TradeState    string
	NotifyRaw     []byte // 存 jsonb，可以 Marshal/Unmarshal
	PaidAt        sql.NullTime
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (pdb *psDatabase) InsertWeChatOrder(ctx context.Context, po *WeChatOrder) error {
	log := pdb.log.With().
		Str("order_id", po.OrderID).
		Str("business_type", po.BusinessType).
		Str("plan_key", po.PlanKey).
		Logger()

	const q = `
INSERT INTO app.pay_orders (
    order_id,
    business_type,
    plan_key,
    amount_total,
    currency,
    description,
    wx_unionid,
    wx_trade_state
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, 'NOTPAY'
)`

	_, err := pdb.db.ExecContext(
		ctx,
		q,
		po.OrderID,
		po.BusinessType,
		po.PlanKey,
		po.AmountTotal,
		po.Currency,
		po.Description,
		nullStringValue(po.WeChatUnionID), // 可空
	)
	if err != nil {
		log.Error().Err(err).Msg("InsertWeChatOrder failed")
		return err
	}

	log.Info().Msg("InsertWeChatOrder ok")
	return nil
}

func nullStringValue(ns sql.NullString) any {
	if ns.Valid {
		return ns.String
	}
	return nil
}

func (pdb *psDatabase) FindWeChatOrderByID(ctx context.Context, orderID string) (*WeChatOrder, error) {
	log := pdb.log.With().
		Str("order_id", orderID).
		Logger()

	const q = `
SELECT
    id,
    order_id,
    business_type,
    plan_key,
    amount_total,
    currency,
    description,
    wx_unionid,
    wx_payer_openid,
    wx_transaction_id,
    wx_trade_state,
    wx_notify_raw,
    paid_at,
    created_at,
    updated_at
FROM app.pay_orders
WHERE order_id = $1
LIMIT 1
`
	row := pdb.db.QueryRowContext(ctx, q, orderID)

	var po WeChatOrder
	var notifyRaw sql.NullString

	err := row.Scan(
		&po.ID,
		&po.OrderID,
		&po.BusinessType,
		&po.PlanKey,
		&po.AmountTotal,
		&po.Currency,
		&po.Description,
		&po.WeChatUnionID,
		&po.WeChatOpenID,
		&po.TransactionID,
		&po.TradeState,
		&notifyRaw,
		&po.PaidAt,
		&po.CreatedAt,
		&po.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		log.Debug().Msg("FindWeChatOrderByID no rows")
		return nil, nil
	}
	if err != nil {
		log.Error().Err(err).Msg("FindWeChatOrderByID failed")
		return nil, err
	}

	if notifyRaw.Valid {
		po.NotifyRaw = []byte(notifyRaw.String)
	}

	return &po, nil
}

func (pdb *psDatabase) UpdateWeChatOrderStatus(
	ctx context.Context,
	orderID string,
	tradeState string,
	transactionID *string,
	payerOpenID *string,
	paidAt *time.Time,
	notifyRaw []byte,
) error {
	log := pdb.log.With().
		Str("order_id", orderID).
		Str("trade_state", tradeState).
		Logger()

	const q = `
UPDATE app.pay_orders
SET
    wx_trade_state   = COALESCE($2, wx_trade_state),
    wx_transaction_id = COALESCE($3, wx_transaction_id),
    wx_payer_openid  = COALESCE($4, wx_payer_openid),
    paid_at          = COALESCE($5, paid_at),
    wx_notify_raw    = COALESCE($6, wx_notify_raw),
    updated_at       = NOW()
WHERE order_id = $1
`

	var paidAtVal any
	if paidAt != nil {
		paidAtVal = *paidAt
	} else {
		paidAtVal = nil
	}

	var transIDVal any
	if transactionID != nil && *transactionID != "" {
		transIDVal = *transactionID
	} else {
		transIDVal = nil
	}

	var openIDVal any
	if payerOpenID != nil && *payerOpenID != "" {
		openIDVal = *payerOpenID
	} else {
		openIDVal = nil
	}

	var notifyVal any
	if len(notifyRaw) > 0 {
		notifyVal = string(notifyRaw)
	} else {
		notifyVal = nil
	}

	_, err := pdb.db.ExecContext(
		ctx,
		q,
		orderID,
		tradeState,
		transIDVal,
		openIDVal,
		paidAtVal,
		notifyVal,
	)
	if err != nil {
		log.Error().Err(err).Msg("UpdateWeChatOrderStatus failed")
		return err
	}

	log.Info().Msg("UpdateWeChatOrderStatus ok")
	return nil
}
