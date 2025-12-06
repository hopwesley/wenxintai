package dbSrv

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type WeChatOrder struct {
	ID            int64
	OrderID       string
	PublicID      string
	PlanKey       string
	AmountTotal   int64
	Currency      string
	Description   string
	PayerOpenId   sql.NullString
	TransactionID sql.NullString
	CodeUrl       sql.NullString
	TradeState    int16
	NotifyRaw     []byte // 存 jsonb，可以 Marshal/Unmarshal
	PaidAt        sql.NullTime
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (pdb *psDatabase) InsertWeChatOrder(ctx context.Context, po *WeChatOrder) error {
	log := pdb.log.With().
		Str("order_id", po.OrderID).
		Str("public_id", po.PublicID).
		Str("plan_key", po.PlanKey).
		Logger()

	const q = `
INSERT INTO app.pay_orders (
    order_id,
    public_id,
    plan_key,
    amount_total,
    currency,
    description,
    code_url
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)`

	_, err := pdb.db.ExecContext(
		ctx,
		q,
		po.OrderID,
		po.PublicID,
		po.PlanKey,
		po.AmountTotal,
		po.Currency,
		po.Description,
		po.CodeUrl, // 新增：把 CodeUrl 写入
	)
	if err != nil {
		log.Error().Err(err).Msg("InsertWeChatOrder failed")
		return err
	}

	log.Info().Msg("InsertWeChatOrder ok")
	return nil
}

func (pdb *psDatabase) QueryWeChatOrderByOrderID(ctx context.Context, orderID string) (*WeChatOrder, error) {
	sLog := pdb.log.With().
		Str("order_id", orderID).
		Logger()

	const q = `
SELECT
    id,
    order_id,
    public_id,
    plan_key,
    amount_total,
    currency,
    description,
    code_url,
    wx_payer_openid,
    wx_transaction_id,
    trade_state,
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
		&po.PublicID,
		&po.PlanKey,
		&po.AmountTotal,
		&po.Currency,
		&po.Description,
		&po.CodeUrl, // 新增：code_url 列
		&po.PayerOpenId,
		&po.TransactionID,
		&po.TradeState,
		&notifyRaw,
		&po.PaidAt,
		&po.CreatedAt,
		&po.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		sLog.Debug().Msg("QueryWeChatOrderByOrderID no rows")
		return nil, nil
	}
	if err != nil {
		sLog.Error().Err(err).Msg("QueryWeChatOrderByOrderID failed")
		return nil, err
	}

	if notifyRaw.Valid {
		po.NotifyRaw = []byte(notifyRaw.String)
	}

	return &po, nil
}
func (pdb *psDatabase) QueryUnfinishedOrder(
	ctx context.Context,
	publicId string,
	timeout time.Time, // timeout 表示“最早可接受的 created_at 时间”，例如 now-90min
) (*WeChatOrder, error) {
	sLog := pdb.log.With().
		Str("public_id", publicId).
		Logger()

	const q = `
SELECT
    id,
    order_id,
    public_id,
    plan_key,
    amount_total,
    currency,
    description,
    code_url,
    wx_payer_openid,
    wx_transaction_id,
    trade_state,
    wx_notify_raw,
    paid_at,
    created_at,
    updated_at
FROM app.pay_orders
WHERE public_id = $1
  AND trade_state = 0        -- 未支付
  AND paid_at IS NULL        -- 没有支付时间
  AND code_url IS NOT NULL   -- 支付二维码存在
  AND created_at >= $2       -- 在超时时间窗口内
ORDER BY created_at DESC
LIMIT 1
`
	row := pdb.db.QueryRowContext(ctx, q, publicId, timeout)

	var po WeChatOrder
	var notifyRaw sql.NullString

	err := row.Scan(
		&po.ID,
		&po.OrderID,
		&po.PublicID,
		&po.PlanKey,
		&po.AmountTotal,
		&po.Currency,
		&po.Description,
		&po.CodeUrl,
		&po.PayerOpenId,
		&po.TransactionID,
		&po.TradeState,
		&notifyRaw,
		&po.PaidAt,
		&po.CreatedAt,
		&po.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		sLog.Debug().Msg("QueryUnfinishedOrder no rows")
		return nil, nil
	}
	if err != nil {
		sLog.Error().Err(err).Msg("QueryUnfinishedOrder failed")
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
	tradeState int16,
	transactionID string,
	payerOpenID string,
	paidAt time.Time,
	notifyRaw []byte,
) error {
	sLog := pdb.log.With().
		Str("order_id", orderID).
		Int16("trade_state", tradeState).
		Logger()

	return pdb.WithTx(ctx, func(tx *sql.Tx) error {
		// --- 第一步：更新 pay_orders 并返回 public_id ---
		const qUpdateOrder = `
UPDATE app.pay_orders
SET
    trade_state       = COALESCE($2, trade_state),
    wx_transaction_id = COALESCE($3, wx_transaction_id),
    wx_payer_openid   = COALESCE($4, wx_payer_openid),
    paid_at           = COALESCE($5, paid_at),
    wx_notify_raw     = COALESCE($6, wx_notify_raw),
    updated_at        = NOW()
WHERE order_id = $1
RETURNING public_id
`

		var publicID string
		if err := tx.QueryRowContext(
			ctx,
			qUpdateOrder,
			orderID,
			tradeState,
			transactionID,
			payerOpenID,
			paidAt,
			notifyRaw,
		).Scan(&publicID); err != nil {
			sLog.Error().Err(err).Msg("update pay_orders failed or order not found")
			return err
		}

		// --- 第二步：更新 tests_record ---
		const qUpdateTestRecord = `
UPDATE app.tests_record
SET
    pay_order_id = $2,
    paid_time    = NOW()
WHERE public_id = $1
`

		res, err := tx.ExecContext(ctx, qUpdateTestRecord, publicID, orderID)
		if err != nil {
			sLog.Error().Err(err).Str("public_id", publicID).Msg("update tests_record failed")
			return err
		}

		rows, err := res.RowsAffected()
		if err != nil {
			sLog.Error().Err(err).Str("public_id", publicID).Msg("RowsAffected failed")
			return err
		}

		if rows == 0 {
			err = fmt.Errorf("no tests_record updated for public_id=%s", publicID)
			sLog.Error().Err(err).Str("public_id", publicID).Msg("update tests_record affected 0 rows")
			return err
		}

		sLog.Info().Str("public_id", publicID).Msg("UpdateWeChatOrderStatus ok")
		return nil
	})
}
