package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"time"
)

// Weights 组合打分
type Weights struct{ W1, W2, W3, W4, W5 float64 }

// ScoreCombos33
// ------------------------
// 组合打分逻辑
// ------------------------
func ScoreCombos33(scores []SubjectScores, w Weights) *Mode33Section {
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
		score := w.W1*avgFit -
			w.W2*rarity/10.0 +
			w.W3*comboCos +
			w.W4*minA/5.0 -
			w.W5*risk

		combos = append(combos, Combo33CoreData{
			Subjects:    [3]string{s1, s2, s3},
			AvgFit:      round3(avgFit),
			MinAbility:  round3(minA),
			Rarity:      round3(rarity), // 虽是离散值，保持统一风格
			RiskPenalty: round3(risk),
			ComboCosine: round3(comboCos),
			Score:       round3(score),
		})
	}

	sort.Slice(combos, func(i, j int) bool {
		if combos[i].Score == combos[j].Score {
			return combos[i].MinAbility > combos[j].MinAbility
		}
		return combos[i].Score > combos[j].Score
	})

	// 仅保留前5名
	if len(combos) > 5 {
		combos = combos[:5]
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
	case ComboPHY_CHE_BIO, ComboHIS_GEO_POL:
		return 0
	case ComboPHY_CHE_GEO, ComboPHY_BIO_GEO, ComboCHE_BIO_GEO, ComboHIS_GEO_BIO:
		return 5
	case ComboPHY_GEO_CHE, ComboHIS_POL_BIO:
		return 12
	default:
		return 5 // 默认中等稀有
	}
}

// RunDemo33
// ---------------------------------------
// 演示入口
// ---------------------------------------

func RunDemo33(riasecAnswers []RIASECAnswer, ascAnswers []ASCAnswer, alpha, beta, gamma float64) *ParamForAIPrompt {
	if alpha == 0 && beta == 0 && gamma == 0 {
		alpha, beta, gamma = 0.4, 0.4, 0.2
	}

	var paramPrompt ParamForAIPrompt
	scores, result := BuildScores(riasecAnswers, ascAnswers, Wfinal, DimCalib, alpha, beta, gamma)

	paramPrompt.Common = result.Common

	// 组合余弦相似度 + 风险约束的科学权重方案
	ws := Weights{
		W1: 0.45, // avgFit: 主导
		W2: 0.20, // rarity: 竞争性平衡
		W3: 0.25, // comboCos: 组合方向一致性
		W4: 0.15, // minA: 能力底线保障
		W5: 0.20, // riskPenalty: 触发式惩罚（0.04 扣分）
	}

	paramPrompt.Mode33 = ScoreCombos33(scores, ws)

	content, _ := json.MarshalIndent(&paramPrompt, "", "  ")
	ts := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("report_param_%s_%s.json", "3+3", ts) // 增加了模块名
	_ = os.WriteFile(filename, content, 0644)

	fmt.Printf("Radar Visualization:\n%+v\n", result.Radar)

	return &paramPrompt
}
