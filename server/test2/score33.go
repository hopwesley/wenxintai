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

/**
 1. RunDemo33（3+3模式）
目标：
从6个学科（物理、化学、生物、地理、历史、政治）中选出3个学科组成一个组合，并根据多个因素计算该组合的推荐分数。

RunDemo33 (3+3模式) 数学逻辑
核心评分公式：
math
Score = W₁ × avgFit - W₂ × (rarity/10) + W₃ × globalCos + W₄ × (minA/5) - W₅ × riskPenalty
各分量详细解释：
1.1 平均兴趣匹配度 (avgFit)
math
avgFit = (Fit₁ + Fit₂ + Fit₃) / 3
其中单个学科的 Fit 计算来自 BuildScores：

math
Fit = α × (zA - zI) + β × cos(I, A) + γ × AbilityShare
zA - zI: 能力与兴趣的 z-score 差异

cos(I, A): 兴趣向量与能力向量的余弦相似度

AbilityShare: 该科能力在总能力中的占比

1.2 稀有度惩罚 (rarity)
rarity ∈ {0, 5, 12}
0: 常见组合 (PHY_CHE_BIO 等)

5: 中等稀有 (HIS_GEO_POL 等)

12: 默认稀有 (不在白名单的组合)

1.3 全局余弦相似度 (globalCos)
globalCos = cosine_similarity(I⃗, A⃗)
衡量兴趣与能力在整个学科空间的方向一致性

1.4 最低能力标准化 (minA/5)
minA/5 ∈ [0.2, 1.0]
将最低能力值从 [1,5] 缩放到 [0.2,1.0]

1.5 风险惩罚 (riskPenalty)
riskPenalty = {0.2 if minA < 3, 0 otherwise}
当任一学科能力低于阈值时的固定惩罚
默认权重配置：
ws := Weights{W1: 0.5, W2: 0.3, W3: 0.1, W4: 0.1, W5: 0.1}
*/

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
		risk := 0.0
		if minA < 2.8 {
			risk = 0.45
		}

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
		alpha, beta, gamma = 0.5, 0.2, 0.3
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
