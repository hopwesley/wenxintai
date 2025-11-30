package ai_api

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
)

// ScoreCombos33
// ------------------------
// 组合打分逻辑
// ------------------------

var factorWeight33 = Weights33{
	W1: 0.45,
	W2: 0.10,
	W3: 0.25,
	W4: 0.20,
	W5: 0.25,
}

func ScoreCombos33(scores []SubjectScores) *Mode33Section {
	m := map[string]SubjectScores{}
	for _, s := range scores {
		m[s.Subject] = s
	}

	var combos []Combo33CoreData

	for _, comboKey := range AllCombos33 {
		subjs := strings.Split(comboKey, "_")
		if len(subjs) != 3 {
			continue
		}
		s1, s2, s3 := subjs[0], subjs[1], subjs[2]

		// 获取三科的 SubjectScores
		sc1, sc2, sc3 := m[s1], m[s2], m[s3]

		// 平均匹配度
		avgFit := (sc1.Fit + sc2.Fit + sc3.Fit) / 3.0

		// 计算最低能力
		minA := math.Min(m[s1].A, math.Min(m[s2].A, m[s3].A))

		// 稀有性值
		rarity := RarityValue(comboKey)

		// 风险惩罚
		risk := calculateRiskPenalty(minA, avgFit)

		comboCos := calcComboCos([]SubjectScores{sc1, sc2, sc3})

		// 计算组合最终分
		score := factorWeight33.W1*avgFit -
			factorWeight33.W2*rarity/10.0 +
			factorWeight33.W3*comboCos +
			factorWeight33.W4*minA/5.0 -
			factorWeight33.W5*risk

		recommendScore := NormalizeMetric("combo33.score", score)

		combos = append(combos, Combo33CoreData{
			Subjects:       [3]string{s1, s2, s3},
			AvgFit:         round3(avgFit),
			MinAbility:     round3(minA),
			Rarity:         round3(rarity),
			RiskPenalty:    round3(risk),
			ComboCosine:    round3(comboCos),
			Score:          round3(score),
			RecommendScore: recommendScore,
		})
	}

	sort.Slice(combos, func(i, j int) bool {
		if combos[i].Score == combos[j].Score {
			return combos[i].MinAbility > combos[j].MinAbility
		}
		return combos[i].Score > combos[j].Score
	})

	if len(combos) > 3 {
		combos = combos[:3]
	}

	return &Mode33Section{
		TopCombinations: combos,
	}
}

// RarityValue
// ===========================================
// 返回组合的稀有性数值：0=常见，5=中等，12=稀有
// ===========================================
func RarityValue(combo string) float64 {
	switch combo {
	// === 强烈推荐组合（0分）===
	case ComboPHY_CHE_BIO, ComboPHY_CHE_POL, ComboPHY_CHE_GEO, ComboHIS_GEO_POL:
		return 0

	// === 谨慎考虑组合（5分）===
	case ComboPHY_BIO_GEO, ComboPHY_BIO_POL, ComboCHE_BIO_GEO, ComboHIS_GEO_BIO, ComboPHY_GEO_POL:
		return 5

	// === 避免组合（8分）===
	case ComboHIS_POL_BIO, ComboHIS_CHE_BIO:
		return 8

	// === 其他所有组合（12分）===
	default:
		return 12
	}
}

// BuildFullParam 结合兴趣与能力答案，生成报告输入所需的完整参数。
func BuildFullParam(riasecAnswers []RIASECAnswer, ascAnswers []ASCAnswer, alpha, beta, gamma float64) (*ParamForAIPrompt, *FullScoreResult, []SubjectScores) {
	if alpha == 0 && beta == 0 && gamma == 0 {
		alpha, beta, gamma = 0.4, 0.4, 0.2
	}

	scores, result := BuildScores(riasecAnswers, ascAnswers, iToAWeight, dimWeight, subWeight)

	param := &ParamForAIPrompt{
		Common:  result.Common,
		Mode33:  ScoreCombos33(scores),
		Mode312: ScoreCombos312(scores),
	}

	return param, result, scores
}

// RunDemo33
// ---------------------------------------
// 演示入口
// ---------------------------------------

func RunDemo33(riasecAnswers []RIASECAnswer, ascAnswers []ASCAnswer, alpha, beta, gamma float64, idx, yesno, combo string) *ParamForAIPrompt {
	param, _, _ := BuildFullParam(riasecAnswers, ascAnswers, alpha, beta, gamma)

	payload, _ := json.MarshalIndent(&param, "", "  ")
	filename := fmt.Sprintf("param_%s_%s_%s_%s.json", idx, combo, "3+3", yesno)
	_ = os.WriteFile(filename, payload, 0o644)

	return param
}
