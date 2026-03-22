-- 创建报告反馈表
CREATE TABLE IF NOT EXISTS app.report_feedbacks (
    id SERIAL PRIMARY KEY,
    public_id VARCHAR(64) NOT NULL,
    uid VARCHAR(128) NOT NULL,
    invite_code VARCHAR(64) NOT NULL,
    rating_score INTEGER NOT NULL CHECK (rating_score >= 0 AND rating_score <= 10),
    feedback_content TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_report_feedbacks_public_id
        FOREIGN KEY (public_id)
        REFERENCES app.tests_record(public_id)
        ON DELETE CASCADE,

    CONSTRAINT fk_report_feedbacks_uid
        FOREIGN KEY (uid)
        REFERENCES app.user_profile(uid)
        ON DELETE CASCADE
);

-- 创建索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_report_feedbacks_public_id ON app.report_feedbacks(public_id);
CREATE INDEX IF NOT EXISTS idx_report_feedbacks_uid ON app.report_feedbacks(uid);
CREATE INDEX IF NOT EXISTS idx_report_feedbacks_invite_code ON app.report_feedbacks(invite_code);
CREATE INDEX IF NOT EXISTS idx_report_feedbacks_created_at ON app.report_feedbacks(created_at DESC);
