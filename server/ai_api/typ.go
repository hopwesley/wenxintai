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

type TestTyp string

const (
	TypUnknown TestTyp = "Unknown"
	TypRIASEC  TestTyp = "RIASEC"
	TypOCEAN   TestTyp = "OCEAN"
	TypASC     TestTyp = "ASC"
)

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

type OCEANCAnswer struct {
	ID        int    `json:"id"`
	Dimension string `json:"dimension"`
	Score     int    `json:"score"`
	Reverse   bool   `json:"reverse"`
}

// SubjectWeight 科目基础计算
type SubjectWeight struct{ alpha, beta, gamma float64 }

func (sw *SubjectWeight) adjustWeights(quality float64) *SubjectWeight {
	a := sw.alpha * quality
	b := sw.beta * quality
	g := sw.gamma + (sw.alpha+sw.beta)*(1-quality)*0.5
	sum := a + b + g
	return &SubjectWeight{
		a / sum, b / sum, g / sum,
	}
}

// Weights33 组合打分
type Weights33 struct{ W1, W2, W3, W4, W5 float64 }

type EngineResult struct {
	CommonScore  *FullScoreResult `json:"common_score"`
	Recommend33  *Mode33Section   `json:"recommend_33"`
	Recommend312 *Mode312Section  `json:"recommend_312"`
}
