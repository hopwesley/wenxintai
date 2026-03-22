package dbSrv

import (
	"context"
	"database/sql"
)

type ReportFeedback struct {
	ID              int64
	PublicID        string
	Uid             string
	InviteCode      string
	RatingScore     int
	FeedbackContent string
	CreatedAt       string
}

// InsertReportFeedback 插入报告反馈
func (pdb *psDatabase) InsertReportFeedback(ctx context.Context, feedback *ReportFeedback) error {
	sLog := pdb.log.With().
		Str("public_id", feedback.PublicID).
		Str("uid", feedback.Uid).
		Str("invite_code", feedback.InviteCode).
		Int("rating_score", feedback.RatingScore).
		Logger()

	sLog.Debug().Msg("InsertReportFeedback")

	const q = `
		INSERT INTO app.report_feedbacks
		    (public_id, uid, invite_code, rating_score, feedback_content, created_at)
		VALUES
		    ($1, $2, $3, $4, $5, NOW())
		RETURNING id, created_at
	`

	err := pdb.db.QueryRowContext(
		ctx,
		q,
		feedback.PublicID,
		feedback.Uid,
		feedback.InviteCode,
		feedback.RatingScore,
		feedback.FeedbackContent,
	).Scan(&feedback.ID, &feedback.CreatedAt)

	if err != nil {
		sLog.Err(err).Msg("insert report feedback failed")
		return err
	}

	sLog.Info().Int64("id", feedback.ID).Msg("insert report feedback success")
	return nil
}

// QueryTestRecordPaymentInfo 查询测试记录的支付信息
func (pdb *psDatabase) QueryTestRecordPaymentInfo(ctx context.Context, publicId string) (payOrderId string, paidTime sql.NullTime, err error) {
	sLog := pdb.log.With().Str("public_id", publicId).Logger()
	sLog.Debug().Msg("QueryTestRecordPaymentInfo")

	const q = `
		SELECT pay_order_id, paid_time
		FROM app.tests_record
		WHERE public_id = $1
	`

	err = pdb.db.QueryRowContext(ctx, q, publicId).Scan(&payOrderId, &paidTime)
	if err != nil {
		if err == sql.ErrNoRows {
			sLog.Debug().Msg("test record not found")
			return "", sql.NullTime{}, nil
		}
		sLog.Err(err).Msg("query test record payment info failed")
		return "", sql.NullTime{}, err
	}

	sLog.Debug().Str("pay_order_id", payOrderId).Msg("query payment info success")
	return payOrderId, paidTime, nil
}
