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

type ASCAnswer struct {
	ID      int    `json:"id"`
	Subject string `json:"subject"`
	Score   int    `json:"score"`   // 1–5
	Reverse bool   `json:"reverse"` // 与题干一致；此处为“答案分”而非换算分
	Subtype string `json:"subtype"`
}

type RIASECAnswer struct {
	ID        int    `json:"id"`
	Dimension string `json:"dimension"`
	Score     int    `json:"score"`
}
