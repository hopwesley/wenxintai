package ai_api

import "fmt"

func proBuild33Param() {

}

func proBuild312Param() {

}
func ProBuildReportParam(mode Mode, answers map[TestTyp]any) (*EngineResult, error) {
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
