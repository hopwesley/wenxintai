package ai_api

import "fmt"

func basicBuild33Param() {
}

func basicBuild312Param() {
}

// 兴趣→学科权重矩阵（最终版）
var iToAWeight = map[string]map[string]float64{
	"PHY": {"R": 0.30, "I": 0.35, "C": 0.15, "E": 0.10, "S": 0.05, "A": 0.05},
	"CHE": {"R": 0.25, "I": 0.35, "C": 0.20, "E": 0.10, "S": 0.05, "A": 0.05},
	"BIO": {"R": 0.20, "I": 0.35, "S": 0.15, "C": 0.15, "A": 0.10, "E": 0.05},
	"GEO": {"R": 0.25, "I": 0.25, "C": 0.15, "S": 0.15, "E": 0.10, "A": 0.10},
	"HIS": {"A": 0.30, "S": 0.25, "E": 0.15, "I": 0.15, "C": 0.10, "R": 0.05},
	"POL": {"E": 0.30, "S": 0.25, "A": 0.15, "I": 0.15, "C": 0.10, "R": 0.05},
}

var dimWeight = map[string]float64{
	"R": 0.82, "I": 0.87, "A": 0.78, "S": 0.80, "E": 0.75, "C": 0.72,
}

var subWeight = SubjectWeight{alpha: 0.4, beta: 0.4, gamma: 0.2}

func BasicBuildReportParam(mode Mode, answers map[TestTyp]any) (*EngineResult, error) {

	riasecAnswers, ok := answers[TypRIASEC].([]RIASECAnswer)
	if !ok {
		return nil, fmt.Errorf("invalid RIASEC answer data")
	}

	ascAnswers, ok := answers[TypASC].([]ASCAnswer)
	if !ok {
		return nil, fmt.Errorf("invalid ASC answer data")
	}

	var result = &EngineResult{}

	scores, scoreForUsr := BuildScores(riasecAnswers,
		ascAnswers, iToAWeight, dimWeight, subWeight)
	result.CommonScore = scoreForUsr

	switch mode {
	case Mode33:
		result.Recommend33 = ScoreCombos33(scores)
	case Mode312:
		result.Recommend312 = ScoreCombos312(scores)
	default:
		return nil, fmt.Errorf("invalid mode:%s", mode)
	}

	return result, nil
}
