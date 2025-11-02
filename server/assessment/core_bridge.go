package assessment

import (
	"encoding/json"

	"github.com/hopwesley/wenxintai/server/core"
)

type (
	Mode               = core.Mode
	QuestionsResult    = core.QuestionsResult
	ParamForAIPrompt   = core.ParamForAIPrompt
	CommonSection      = core.CommonSection
	SubjectProfileData = core.SubjectProfileData
	Mode312Section     = core.Mode312Section
	Mode33Section      = core.Mode33Section
	AnchorCoreData     = core.AnchorCoreData
	ComboCoreData      = core.ComboCoreData
	Combo33CoreData    = core.Combo33CoreData
	RadarData          = core.RadarData
	FullScoreResult    = core.FullScoreResult
	SubjectScores      = core.SubjectScores
	RIASECAnswer       = core.RIASECAnswer
	ASCAnswer          = core.ASCAnswer
	Input              = core.Input
	Output             = core.Output
)

const (
	Mode33  = core.Mode33
	Mode312 = core.Mode312

	ComboPHY_CHE_BIO = core.ComboPHY_CHE_BIO
	ComboPHY_CHE_GEO = core.ComboPHY_CHE_GEO
	ComboPHY_CHE_POL = core.ComboPHY_CHE_POL
	ComboPHY_BIO_GEO = core.ComboPHY_BIO_GEO
	ComboPHY_BIO_POL = core.ComboPHY_BIO_POL
	ComboPHY_GEO_POL = core.ComboPHY_GEO_POL
	ComboHIS_GEO_POL = core.ComboHIS_GEO_POL
	ComboHIS_GEO_BIO = core.ComboHIS_GEO_BIO
	ComboHIS_POL_BIO = core.ComboHIS_POL_BIO
	ComboHIS_CHE_BIO = core.ComboHIS_CHE_BIO
	ComboHIS_CHE_POL = core.ComboHIS_CHE_POL
	ComboCHE_BIO_GEO = core.ComboCHE_BIO_GEO

	SubjectPHY = core.SubjectPHY
	SubjectCHE = core.SubjectCHE
	SubjectBIO = core.SubjectBIO
	SubjectGEO = core.SubjectGEO
	SubjectHIS = core.SubjectHIS
	SubjectPOL = core.SubjectPOL
)

var (
	Subjects        = core.Subjects
	AllCombos33     = core.AllCombos33
	AuxPoolPHY      = core.AuxPoolPHY
	AuxPoolHIS      = core.AuxPoolHIS
	StudentHobbies  = core.StudentHobbies
	AllRIASECCombos = core.AllRIASECCombos
	AllASCCombos    = core.AllASCCombos
)

func ParseMode(s string) (Mode, bool) {
	return core.ParseMode(s)
}

func GenerateQuestions(mode Mode, apiKey, gender, grade, hobby string) (*QuestionsResult, error) {
	core.DeepSeekCaller = DeepSeekCaller
	return core.GenerateQuestions(mode, apiKey, gender, grade, hobby)
}

func GenerateUnifiedReport(apiKey string, param ParamForAIPrompt, mode Mode) (json.RawMessage, error) {
	core.DeepSeekCaller = DeepSeekCaller
	return core.GenerateUnifiedReport(apiKey, param, mode)
}

func TestUnifiedReport(apiKey string, param ParamForAIPrompt, mode Mode) (json.RawMessage, error) {
	core.DeepSeekCaller = DeepSeekCaller
	return core.TestUnifiedReport(apiKey, param, mode)
}

func BuildFullParam(riasecAnswers []RIASECAnswer, ascAnswers []ASCAnswer, alpha, beta, gamma float64) (*ParamForAIPrompt, *FullScoreResult, []SubjectScores) {
	return core.BuildFullParam(riasecAnswers, ascAnswers, alpha, beta, gamma)
}

func RunDemo33(riasecAnswers []RIASECAnswer, ascAnswers []ASCAnswer, alpha, beta, gamma float64, idx, yesno, combo string) *ParamForAIPrompt {
	return core.RunDemo33(riasecAnswers, ascAnswers, alpha, beta, gamma, idx, yesno, combo)
}

func RunDemo312(riasecAnswers []RIASECAnswer, ascAnswers []ASCAnswer, alpha, beta, gamma float64, idx, yesno, combo string) *ParamForAIPrompt {
	return core.RunDemo312(riasecAnswers, ascAnswers, alpha, beta, gamma, idx, yesno, combo)
}

func Run(in Input, mode Mode) (Output, error) {
	return core.Run(in, mode)
}
