package srv

import (
	"context"
	"net/http"

	"github.com/hopwesley/wenxintai/server/assessment"
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

func (s *HttpSrv) handleHobbies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, ApiMethodInvalid)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"hobbies": assessment.StudentHobbies,
	})
}
