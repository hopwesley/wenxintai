package srv

import (
	"net/http"
)

func (s *HttpSrv) handleInviteVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, ApiMethodInvalid)
		return
	}

	writeJSON(w, http.StatusOK, nil)
}
