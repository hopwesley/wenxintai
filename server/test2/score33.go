package main

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
)

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

// ---------------------------------------
// 演示入口
// ---------------------------------------
func RunDemo33(riasecAnswers []RIASECAnswer, ascAnswers []ASCAnswer, alpha, beta, gamma float64) {
	if alpha == 0 && beta == 0 && gamma == 0 {
		alpha, beta, gamma = 0.4, 0.4, 0.2
	}

	scores, globalCos := BuildScores(riasecAnswers, ascAnswers, Wfinal, DimCalib, alpha, beta, gamma)
	ws := Weights{W1: 0.5, W2: 0.3, W3: 0.1, W4: 0.1, W5: 0.1}
	combRank := ScoreCombos33(scores, globalCos, ws)

	limit := 3
	if len(combRank) < limit {
		limit = len(combRank)
	}
	rec := combRank[:limit]

	radar := Radar(scores)

	fmt.Printf("Global Cosine (Interest vs Ability): %.3f\n", globalCos)

	jsScores, _ := json.MarshalIndent(scores, "", "  ")
	jsRec, _ := json.MarshalIndent(rec, "", "  ")
	jsRadar, _ := json.MarshalIndent(radar, "", "  ")

	fmt.Println("\n[Scores]")
	fmt.Println(string(jsScores))

	fmt.Println("\n[Recommendation]")
	fmt.Println(string(jsRec))

	fmt.Println("\n[Radar Payload]")
	fmt.Println(string(jsRadar))
}
