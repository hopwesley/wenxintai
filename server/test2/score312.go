package main

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
)

/*
RunDemo312（3+1+2模式）
目标：
先选择1个主干学科（物理或历史），然后从剩余的4个学科中选择2个作为辅科，形成3学科组合。

步骤：
阶段一：主干学科（Anchor）评分（S1）
对于每个主干学科（物理或历史），计算：

S1 = anchW_Fit * Fit(anchor) + anchW_Ability * (Ability(anchor)/5) + anchW_Cover * AnchorBaseCoverage(anchor)
其中：

Fit(anchor)：该主干学科的Fit值。

Ability(anchor)：该主干学科的能力值。

AnchorBaseCoverage(anchor)：该主干学科的基线覆盖率（来自AnchorBaseCoverage表）。

阶段二：辅科组合评分（S23）
对于每个主干学科，从剩余的4个学科中任选2个作为辅科，计算：

平均兴趣匹配度（avgFit）：两个辅科的Fit值的平均值。

最小兴趣匹配度（minFit）：两个辅科的Fit值的最小值。

扩展覆盖率（expansion）：组合的覆盖率（来自Coverage312表）减去主干学科的基线覆盖率，且最小为0。

全局余弦相似度（globalCos）：同3+3模式。

则：

S23 = auxW_AvgFit * avgFit + auxW_MinFit * minFit + auxW_Expansion * expansion + auxW_GlobalCos * globalCos

阶段三：综合评分（Sfinal）
Sfinal = lambda1 * S1 + lambda2 * S23
权重说明（默认权重）：
阶段一权重：

anchW_Fit（主干学科兴趣匹配度）：0.5

anchW_Ability（主干学科能力）：0.3

anchW_Cover（主干学科基线覆盖率）：0.2

阶段二权重：

auxW_AvgFit（辅科平均兴趣匹配度）：0.4

auxW_MinFit（辅科最小兴趣匹配度）：0.3

auxW_Expansion（扩展覆盖率）：0.3

auxW_GlobalCos（全局余弦相似度）：0.1

阶段三权重：

lambda1（主干学科得分权重）：0.6

lambda2（辅科组合得分权重）：0.4

注意：
在阶段二中，扩展覆盖率（expansion）的计算使用了math.Max(0, cov - baseCov)，确保不会因为覆盖率低于基线而出现负值。

*/

// Coverage312 全国平均覆盖率表（键名需与 Combo 常量一致）
var Coverage312 = map[string]float64{
	ComboPHY_CHE_BIO: 0.95,
	ComboPHY_CHE_GEO: 0.92,
	ComboPHY_CHE_POL: 0.78,
	ComboPHY_BIO_GEO: 0.88,
	ComboPHY_BIO_POL: 0.72,
	ComboPHY_GEO_POL: 0.70,

	ComboHIS_POL_GEO: 0.50,
	ComboHIS_POL_BIO: 0.47,
	ComboHIS_GEO_POL: 0.48,
	ComboHIS_CHE_BIO: 0.42,
	ComboHIS_CHE_POL: 0.47,
	ComboHIS_BIO_GEO: 0.45,
}

// Anchor 基线覆盖率（全国平均）
var AnchorBaseCoverage = map[string]float64{
	SubjectPHY: 0.90, // 理科方向覆盖较高
	SubjectHIS: 0.50, // 文科方向中等
}

// =============================
// 权重配置（按最终确认版）
// =============================

// 阶段一权重（Anchor 主科）
var (
	anchW_Fit, anchW_Ability, anchW_Cover = 0.5, 0.3, 0.2
)

// 阶段二权重（辅科组合）
var (
	auxW_AvgFit, auxW_MinFit, auxW_Expansion, auxW_GlobalCos = 0.4, 0.3, 0.3, 0.1
)

// 阶段三权重（综合）
var (
	lambda1, lambda2 = 0.6, 0.4
)

// =============================
// 核心算法逻辑
// =============================

func ScoreCombos312(scores []SubjectScores, globalCos float64) []Combo {
	m := map[string]SubjectScores{}
	for _, s := range scores {
		m[s.Subject] = s
	}

	anchors := []string{SubjectPHY, SubjectHIS}
	var out []Combo

	for _, anchor := range anchors {
		// 阶段一：Anchor 主干得分
		fit := m[anchor].Fit
		ab := m[anchor].A
		baseCov := AnchorBaseCoverage[anchor]
		S1 := anchW_Fit*fit + anchW_Ability*(ab/5.0) + anchW_Cover*baseCov

		// 辅科候选池
		var auxPool []string
		if anchor == SubjectPHY {
			auxPool = []string{SubjectCHE, SubjectBIO, SubjectGEO, SubjectPOL}
		} else {
			auxPool = []string{SubjectGEO, SubjectPOL, SubjectCHE, SubjectBIO}
		}

		// 生成4选2组合
		for i := 0; i < len(auxPool); i++ {
			for j := i + 1; j < len(auxPool); j++ {
				s2, s3 := auxPool[i], auxPool[j]
				key := strings.Join([]string{anchor, s2, s3}, "_")

				// 检查覆盖率合法性
				cov, ok := Coverage312[key]
				if !ok {
					continue
				}

				// 阶段二：辅科组合优化
				avgFit := (m[s2].Fit + m[s3].Fit) / 2.0
				minFit := math.Min(m[s2].Fit, m[s3].Fit)
				// 商业/展示版（更平滑）
				expansion := math.Max(0, cov-baseCov)

				S23 := auxW_AvgFit*avgFit +
					auxW_MinFit*minFit +
					auxW_Expansion*expansion +
					auxW_GlobalCos*globalCos

				// 阶段三：综合评分
				final := lambda1*S1 + lambda2*S23

				out = append(out, Combo{
					Subs:  [3]string{anchor, s2, s3},
					Score: math.Round(final*100) / 100,
					Reason: fmt.Sprintf("S1=%.2f(主干:%s, Fit=%.2f, Ab=%.1f, Cov=%.2f); "+
						"S23=%.2f(辅科:%s+%s, AvgFit=%.2f, MinFit=%.2f, Expansion=%.2f, GlobalCos=%.2f)",
						S1, anchor, fit, ab, baseCov,
						S23, s2, s3, avgFit, minFit, expansion, globalCos),
				})
			}
		}
	}

	// 排序
	sort.Slice(out, func(i, j int) bool { return out[i].Score > out[j].Score })

	// 仅保留前5
	if len(out) > 5 {
		out = out[:5]
	}
	return out
}

// =============================
// RunDemo312
// =============================

func RunDemo312(riasecAnswers []RIASECAnswer, ascAnswers []ASCAnswer, alpha, beta, gamma float64) {
	if alpha == 0 && beta == 0 && gamma == 0 {
		alpha, beta, gamma = 0.4, 0.4, 0.2
	}

	// 阶段0：构建测评数据
	scores, globalCos, log := BuildScores(riasecAnswers, ascAnswers, Wfinal, DimCalib, alpha, beta, gamma)

	b, _ := json.MarshalIndent(log, "", "  ")
	fmt.Println(string(b))

	// 阶段1–3：组合评分
	combRank := ScoreCombos312(scores, globalCos)

	// 取Top3
	limit := 3
	if len(combRank) < limit {
		limit = len(combRank)
	}
	rec := combRank[:limit]

	// 雷达图载荷
	radar := Radar(scores)

	// 输出
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
