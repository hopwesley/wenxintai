package dbSrv

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type UserProfile struct {
	ID          int64     `json:"id"`
	Uid         string    `json:"uid"`                   // 微信 UnionID
	NickName    string    `json:"nick_name,omitempty"`   // 微信昵称
	AvatarUrl   string    `json:"avatar_url,omitempty"`  // 微信头像 URL
	Mobile      string    `json:"mobile,omitempty"`      // 手机号（可空）
	StudyId     string    `json:"study_id,omitempty"`    // 学号（可空）
	SchoolName  string    `json:"school_name,omitempty"` // 学校名称（可空）
	Province    string    `json:"province,omitempty"`    // 所在地区省（可空）
	City        string    `json:"city,omitempty"`        // 所在地区省（可空）
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	LastLoginAt time.Time `json:"last_login_at"` // 最近登录时间
}

func (pdb *psDatabase) InsertOrUpdateUserProfileBasic(
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

	log.Debug().Msg("InsertOrUpdateUserProfileBasic: start")

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
		log.Err(err).Msg("InsertOrUpdateUserProfileBasic: exec failed")
		return err
	}

	log.Debug().Msg("InsertOrUpdateUserProfileBasic: done")
	return nil
}

func (pdb *psDatabase) UpdateUserProfileExtra(
	ctx context.Context,
	uid string,
	mobile string,
	studyId string,
	schoolName string,
	province string,
	city string,
) error {
	if uid == "" {
		return errors.New("uid must be non-empty")
	}

	log := pdb.log.With().
		Str("uid", uid).
		Logger()

	log.Debug().Msg("UpdateUserProfileExtra: start")

	const q = `
		UPDATE app.user_profile
		SET
			mobile      = $2,
			study_id    = $3,
			school_name = $4,
			province    = $5,
			city    = $6,
			updated_at  = now()
		WHERE uid = $1
	`

	res, err := pdb.db.ExecContext(ctx, q,
		uid,
		mobile,
		studyId,
		schoolName,
		province,
		city,
	)
	if err != nil {
		log.Err(err).Msg("UpdateUserProfileExtra: exec failed")
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		log.Err(err).Msg("UpdateUserProfileExtra: RowsAffected failed")
		return err
	}
	if affected == 0 {
		err := errors.New("no user_profile found for uid=" + uid)
		log.Warn().Err(err).Msg("UpdateUserProfileExtra: not found")
		return err
	}

	log.Debug().Int64("rows", affected).Msg("UpdateUserProfileExtra: done")
	return nil
}

func (pdb *psDatabase) FindUserProfileByUid(
	ctx context.Context,
	uid string,
) (*UserProfile, error) {
	if uid == "" {
		return nil, errors.New("uid must be non-empty")
	}

	log := pdb.log.With().
		Str("uid", uid).
		Logger()

	log.Debug().Msg("FindUserProfileByUid: start")

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

	var u UserProfile
	if err := row.Scan(
		&u.ID,
		&u.Uid,
		&u.NickName,
		&u.AvatarUrl,
		&u.Mobile,
		&u.StudyId,
		&u.SchoolName,
		&u.Province,
		&u.City,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.LastLoginAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Warn().Msg("FindUserProfileByUid: no record")
			return nil, nil
		}
		log.Err(err).Msg("FindUserProfileByUid: query failed")
		return nil, err
	}

	log.Debug().Msg("FindUserProfileByUid: done")
	return &u, nil
}
