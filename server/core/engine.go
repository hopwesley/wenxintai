package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type Mode int

const (
	Mode33 Mode = iota
	Mode312
)

var modeStrings = map[Mode]string{
	Mode33:  "3+3",
	Mode312: "3+1+2",
}

var stringToMode = map[string]Mode{
	"3+3":   Mode33,
	"3+1+2": Mode312,
}

func (m Mode) String() string {
	if s, ok := modeStrings[m]; ok {
		return s
	}
	return fmt.Sprintf("unknown(%d)", int(m))
}

func (m Mode) MarshalJSON() ([]byte, error) {
	if s, ok := modeStrings[m]; ok {
		return json.Marshal(s)
	}
	return nil, fmt.Errorf("core: unknown mode %d", m)
}

func (m *Mode) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, ok := stringToMode[s]
	if !ok {
		return fmt.Errorf("core: invalid mode %q", s)
	}
	*m = parsed
	return nil
}

func ParseMode(s string) (Mode, bool) {
	m, ok := stringToMode[s]
	return m, ok
}

type Input struct {
	RIASECAnswers []RIASECAnswer
	ASCAnswers    []ASCAnswer
	Alpha         float64
	Beta          float64
	Gamma         float64
}

type Output struct {
	Scores []SubjectScores
	Result *FullScoreResult
	Param  *ParamForAIPrompt
}

func Run(in Input, mode Mode) (Output, error) {
	if _, ok := modeStrings[mode]; !ok {
		return Output{}, fmt.Errorf("core: unsupported mode %d", int(mode))
	}

	alpha, beta, gamma := in.Alpha, in.Beta, in.Gamma
	if alpha == 0 && beta == 0 && gamma == 0 {
		alpha, beta, gamma = 0.4, 0.4, 0.2
	}

	scores, result := BuildScores(in.RIASECAnswers, in.ASCAnswers, Wfinal, DimCalib, alpha, beta, gamma)

	if result == nil {
		return Output{}, errors.New("core: empty score result")
	}

	out := Output{
		Scores: scores,
		Result: result,
		Param: &ParamForAIPrompt{
			Common: result.Common,
		},
	}

	switch mode {
	case Mode33:
		ws := Weights{W1: 0.45, W2: 0.10, W3: 0.25, W4: 0.20, W5: 0.25}
		out.Param.Mode33 = ScoreCombos33(scores, ws)
	case Mode312:
		out.Param.Mode312 = ScoreCombos312(scores)
	}

	return out, nil
}

var DeepSeekCaller = func(apiKey string, reqBody interface{}) (string, error) {
	return "", errors.New("core: DeepSeekCaller not configured")
}

func uuidLike() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}
