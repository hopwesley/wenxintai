package srv

import (
	"context"
	"net/http"

	"github.com/hopwesley/wenxintai/server/dbSrv"
)

func (s *HttpSrv) initHobbies() error {
	hobbies, err := dbSrv.Instance().ListHobbies(context.Background())
	if err != nil {
		s.log.Err(err).Msg("init hobbies failed")
		return err
	}
	s.log.Info().Int("hobbies-in-db", len(hobbies)).Send()
	if len(hobbies) > 0 {
		s.cfg.studentHobbies = hobbies
	} else {
		s.cfg.studentHobbies = defaultHobbies
	}

	s.log.Info().Msg("init hobbies cache success")
	return err
}

func (s *HttpSrv) handleHobbies(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"hobbies": s.cfg.studentHobbies,
	})
}

type PlanInfoDTO struct {
	Key   string  `json:"key"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Desc  string  `json:"desc"`
	Tag   *string `json:"tag,omitempty"`
}

func (s *HttpSrv) handleProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	plans, err := dbSrv.Instance().ListTestPlans(ctx)
	if err != nil {
		s.log.Err(err).Msg("ListTestPlans failed")
		writeError(w, ApiInternalErr("获取产品列表失败", err))
		return
	}

	out := make([]PlanInfoDTO, 0, len(plans))
	for _, p := range plans {
		item := PlanInfoDTO{
			Key:   p.PlanKey,
			Name:  p.Name,
			Price: p.Price,
			Desc:  p.Description,
		}
		if p.Tag.Valid {
			tag := p.Tag.String
			item.Tag = &tag
		}
		out = append(out, item)
	}

	writeJSON(w, http.StatusOK, out)
}

func (s *HttpSrv) handleCurrentProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req WeChatNativeCreateReq

	if err := req.parseObj(r); err != nil {
		s.log.Err(err).Msgf("[handleCurrentProduct] invalid request")
		writeError(w, err)
		return
	}
	sLog := s.log.With().Str("public_id", req.PublicId).Str("business_type", req.BusinessType).Logger()

	plan, planErr := dbSrv.Instance().PlanByKey(ctx, req.BusinessType)
	if planErr != nil {
		sLog.Err(planErr).Msg("failed find product price info")
		writeError(w, ApiInternalErr("查询产品价格信息失败", planErr))
		return
	}

	writeJSON(w, http.StatusOK, plan)
}
