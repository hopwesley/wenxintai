package dbSrv

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type UserProfile struct {
	ID          int64     `json:"id"`
	Uid         string    `json:"uid"`                  // 微信 UnionID
	NickName    string    `json:"nick_name,omitempty"`  // 微信昵称
	AvatarUrl   string    `json:"avatar_url,omitempty"` // 微信头像 URL
	Mobile      string    `json:"mobile,omitempty"`
	StudyId     string    `json:"study_id,omitempty"`    // 学号（可空）
	SchoolName  string    `json:"school_name,omitempty"` // 学校名称（可空）
	Province    string    `json:"province,omitempty"`    // 所在地区省（可空）
	City        string    `json:"city,omitempty"`        // 所在地区省（可空）
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
	LastLoginAt time.Time `json:"last_login_at,omitempty"` // 最近登录时间
}

func (pdb *psDatabase) InsertOrUpdateWeChatInfo(
	ctx context.Context,
	uid string,
	nickName string,
	avatarUrl string,
) error {
	if uid == "" {
		return errors.New("uid must be non-empty")
	}

	log := pdb.log.With().
		Str("uid", uid).
		Logger()

	log.Debug().Msg("InsertOrUpdateWeChatInfo: start")

	const q = `
		INSERT INTO app.user_profile (
			uid, nick_name, avatar_url, last_login_at
		)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (uid)
		DO UPDATE SET
			nick_name     = EXCLUDED.nick_name,
			avatar_url    = EXCLUDED.avatar_url,
			last_login_at = now(),
			updated_at    = now()
	`

	_, err := pdb.db.ExecContext(ctx, q,
		uid,
		nickName,
		avatarUrl,
	)
	if err != nil {
		log.Err(err).Msg("InsertOrUpdateWeChatInfo: exec failed")
		return err
	}

	log.Debug().Msg("InsertOrUpdateWeChatInfo: done")
	return nil
}

type UsrProfileExtra struct {
	City       *string `json:"city,omitempty"`
	Province   *string `json:"province,omitempty"`
	Mobile     *string `json:"mobile,omitempty"`
	StudyId    *string `json:"study_id,omitempty"`
	SchoolName *string `json:"school_name,omitempty"`
}

func (pdb *psDatabase) UpdateUserProfileExtra(
	ctx context.Context,
	uid string,
	extra UsrProfileExtra,
) error {
	if uid == "" {
		return errors.New("uid must be non-empty")
	}

	sLog := pdb.log.With().Str("uid", uid).Logger()
	sLog.Debug().Msg("UpdateUserProfileExtra: start")

	// 动态拼接 SET 子句
	sets := make([]string, 0, 5)
	args := make([]any, 0, 6)

	args = append(args, uid)
	idx := 2

	if extra.Mobile != nil {
		sets = append(sets, fmt.Sprintf("mobile = $%d", idx))
		args = append(args, *extra.Mobile)
		idx++
	}
	if extra.StudyId != nil {
		sets = append(sets, fmt.Sprintf("study_id = $%d", idx))
		args = append(args, *extra.StudyId)
		idx++
	}
	if extra.SchoolName != nil {
		sets = append(sets, fmt.Sprintf("school_name = $%d", idx))
		args = append(args, *extra.SchoolName)
		idx++
	}
	if extra.Province != nil {
		sets = append(sets, fmt.Sprintf("province = $%d", idx))
		args = append(args, *extra.Province)
		idx++
	}
	if extra.City != nil {
		sets = append(sets, fmt.Sprintf("city = $%d", idx))
		args = append(args, *extra.City)
		idx++
	}

	if len(sets) == 0 {
		sLog.Debug().Msg("UpdateUserProfileExtra: no fields to update")
		return nil
	}

	sets = append(sets, "updated_at = now()")

	q := fmt.Sprintf(`
        UPDATE app.user_profile
        SET %s
        WHERE uid = $1
    `, strings.Join(sets, ", "))

	res, err := pdb.db.ExecContext(ctx, q, args...)
	if err != nil {
		sLog.Err(err).Msg("UpdateUserProfileExtra: exec failed")
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		sLog.Err(err).Msg("UpdateUserProfileExtra: RowsAffected failed")
		return err
	}
	if affected == 0 {
		err := errors.New("no user_profile found for uid=" + uid)
		sLog.Warn().Err(err).Msg("UpdateUserProfileExtra: not found")
		return err
	}

	sLog.Debug().Int64("rows", affected).Msg("UpdateUserProfileExtra: done")
	return nil
}

func (pdb *psDatabase) QueryUserProfileUid(
	ctx context.Context,
	uid string,
) (*UserProfile, error) {
	if uid == "" {
		return nil, errors.New("uid must be non-empty")
	}

	log := pdb.log.With().
		Str("uid", uid).
		Logger()

	log.Debug().Msg("QueryUserProfileUid: start")

	const q = `
    SELECT
        id,
        uid,
        COALESCE(nick_name, ''),
        COALESCE(avatar_url, ''),
        COALESCE(mobile, ''),
        COALESCE(study_id, ''),
        COALESCE(school_name, ''),
        COALESCE(province, ''),
        COALESCE(city, ''),
        created_at,
        updated_at,
        last_login_at
    FROM app.user_profile
    WHERE uid = $1
    LIMIT 1
`

	row := pdb.db.QueryRowContext(ctx, q, uid)

	var (
		mobileRaw string
		u         UserProfile
	)

	if err := row.Scan(
		&u.ID,
		&u.Uid,
		&u.NickName,
		&u.AvatarUrl,
		&mobileRaw,
		&u.StudyId,
		&u.SchoolName,
		&u.Province,
		&u.City,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.LastLoginAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Warn().Msg("QueryUserProfileUid: no record")
			return nil, nil
		}
		log.Err(err).Msg("QueryUserProfileUid: query failed")
		return nil, err
	}

	u.Mobile = maskMobile(mobileRaw)

	log.Debug().Msg("QueryUserProfileUid: done")
	return &u, nil
}

type TestItem struct {
	PublicID     string `json:"public_id"`
	BusinessType string `json:"business_type"`
	Mode         string `json:"mode"`
	ReportStatus int16  `json:"report_status,omitempty"`
	CreateAt     string `json:"create_at"`
}

func (pdb *psDatabase) QueryTestInfos(ctx context.Context, wechatOpenId string) ([]*TestItem, error) {
	const q = `
SELECT
    r.public_id,
    r.business_type,
    COALESCE(r.mode, '') AS mode,
    r.created_at,
    COALESCE(rep.status, 0) AS status
FROM app.tests_record AS r
LEFT JOIN app.test_reports AS rep
    ON rep.public_id = r.public_id
WHERE r.wechat_openid = $1
ORDER BY r.created_at DESC;
`
	sLog := pdb.log.With().Str("wechat_id", wechatOpenId).Logger()

	rows, err := pdb.db.QueryContext(ctx, q, wechatOpenId)
	if err != nil {
		sLog.Err(err).Msg("query user test info failed")
		return nil, err
	}

	defer rows.Close()

	var items []*TestItem

	for rows.Next() {
		var item TestItem

		if err := rows.Scan(
			&item.PublicID,
			&item.BusinessType,
			&item.Mode,
			&item.CreateAt,
			&item.ReportStatus,
		); err != nil {
			pdb.log.Err(err).Msg("parse database item failed")
			return nil, err
		}

		items = append(items, &item)
	}

	if err := rows.Err(); err != nil {
		pdb.log.Err(err).Msg("iterate rows failed")
		return nil, err
	}

	return items, nil
}
