package srv

import (
	"encoding/json"
	"net/http"
)

type CommonRes struct {
	Ok  bool   `json:"ok"`
	Msg string `json:"msg,omitempty"`
}

type Mode string

const (
	Mode33  Mode = "Mode33"
	Mode312 Mode = "Mode312"
)

func (m Mode) IsValid() bool {
	switch m {
	case Mode33, Mode312:
		return true
	default:
		return false
	}
}

type Grade string

const (
	GradeChuEr  Grade = "初二"
	GradeChuSan Grade = "初三"
	GradeGaoYi  Grade = "高一"
)

func (g Grade) IsValid() bool {
	switch g {
	case GradeChuEr, GradeChuSan, GradeGaoYi:
		return true
	default:
		return false
	}
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}

func writeError(w http.ResponseWriter, err *ApiErr) {
	writeJSON(w, err.status, err)
}
