package main

import (
	"encoding/json"
	"fmt"
	"math"
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

var AllowedCombos = map[string]bool{
	ComboPHY_CHE_BIO: true,
	ComboPHY_CHE_GEO: true,
	ComboPHY_BIO_GEO: true,
	ComboCHE_BIO_GEO: true,
	ComboHIS_GEO_POL: true,
	ComboHIS_GEO_BIO: true,
	ComboPHY_GEO_CHE: true,
	ComboHIS_POL_BIO: true,
}

var ComboRarity = map[string]float64{
	ComboPHY_CHE_BIO: 0,
	ComboPHY_CHE_GEO: 0,
	ComboCHE_BIO_GEO: 0,
	ComboPHY_BIO_GEO: 0,
	ComboHIS_GEO_POL: 5,
	ComboHIS_GEO_BIO: 5,
	ComboHIS_POL_BIO: 5,
}

// RarityPenalty 稀有度惩罚
func RarityPenalty(subs [3]string) float64 {
	key := strings.Join(subs[:], "_")
	if v, ok := ComboRarity[key]; ok {
		return v
	}
	return 12
}

// Weights 组合打分
type Weights struct{ W1, W2, W3, W4, W5 float64 }

// ScoreCombos33
// ------------------------
// 修正版：组合打分逻辑
// ------------------------
func ScoreCombos33(scores []SubjectScores, globalCos float64, w Weights) []Combo {
	m := map[string]SubjectScores{}
	for _, s := range scores {
		m[s.Subject] = s
	}

	// 生成所有 6选3 组合
	combos := generateAllCombos(Subjects)
	var out []Combo

	for _, subs := range combos {
		key := strings.Join([]string{subs[0], subs[1], subs[2]}, "_")

		if !AllowedCombos[key] {
			continue // 跳过不在允许列表内的组合
		}

		// ------------------------
		// 计算加权得分
		// ------------------------
		var fitSum, minA, minI float64 = 0, 999, 999
		for _, s := range subs {
			fitSum += m[s].Fit
			if m[s].A < minA {
				minA = m[s].A
			}
			if m[s].I < minI {
				minI = m[s].I
			}
		}

		avgFit := fitSum / 3.0
		rarity := RarityPenalty(subs)
		// 风险惩罚：若最低能力低于 3，减 0.2 分
		riskPenalty := 0.0
		if minA < 3 {
			riskPenalty = 0.2
		}

		// 计算组合最终分
		score := w.W1*avgFit -
			w.W2*rarity/10.0 +
			w.W3*globalCos +
			w.W4*minA/5.0 -
			w.W5*riskPenalty

		out = append(out, Combo{
			Subs:  [3]string{subs[0], subs[1], subs[2]},
			Score: math.Round(score*100) / 100,
			Reason: fmt.Sprintf("平均Fit=%.2f, 最低能力=%.1f, 稀有度=%.1f",
				avgFit, minA, rarity),
		})
	}

	// ------------------------
	// 排序与输出
	// ------------------------
	sort.Slice(out, func(i, j int) bool { return out[i].Score > out[j].Score })

	// 若没有推荐结果，输出调试信息
	if len(out) == 0 {
		fmt.Println("[Warn] No valid combination found. Check key format or rarity values.")
		return out
	}

	// 仅保留前 5 个
	if len(out) > 5 {
		out = out[:5]
	}
	return out
}

func generateAllCombos(ss []string) [][3]string {
	var res [][3]string
	n := len(ss)
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			for k := j + 1; k < n; k++ {
				res = append(res, [3]string{ss[i], ss[j], ss[k]})
			}
		}
	}
	return res
}

// RunDemo33
// ---------------------------------------
// 演示入口
// ---------------------------------------

func RunDemo33(riasecAnswers []RIASECAnswer, ascAnswers []ASCAnswer, alpha, beta, gamma float64) ComboExplainLog {
	if alpha == 0 && beta == 0 && gamma == 0 {
		alpha, beta, gamma = 0.4, 0.4, 0.2
	}

	scores, globalCos, log := BuildScores(riasecAnswers, ascAnswers, Wfinal, DimCalib, alpha, beta, gamma)
	b, _ := json.MarshalIndent(log, "", "  ")
	fmt.Println(string(b))

	ws := Weights{W1: 0.5, W2: 0.3, W3: 0.1, W4: 0.1, W5: 0.1}
	combRank := ScoreCombos33(scores, globalCos, ws)

	if len(combRank) > 5 {
		combRank = combRank[:5]
	}

	// 生成日志结构
	log2 := ComboExplainLog{
		Mode:         "3+3",
		GlobalCosine: globalCos,
		Version:      "v1.0.0",
		Timestamp:    time.Now().Format(time.RFC3339),
		Summary:      buildSummary(scores, combRank, globalCos),
		TopCombos:    buildExplainCombos(scores, combRank),
	}

	return log2
}
