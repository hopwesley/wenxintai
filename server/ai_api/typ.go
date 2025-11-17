package ai_api

type Mode string

const (
	Mode33  Mode = "3+3"
	Mode312 Mode = "3+1+2"
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

type BasicInfo struct {
	PublicId string `json:"public_id"`
	Grade    Grade  `json:"grade"`
	Mode     Mode   `json:"mode"`
	Hobby    string `json:"hobby,omitempty"`
}
