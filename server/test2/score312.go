package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"time"
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

// ScoreCombos312
// =============================
// 核心算法逻辑
// =============================
func ScoreCombos312(scores []SubjectScores, globalCos float64) *Mode312Section {
	m := map[string]SubjectScores{}
	for _, s := range scores {
		m[s.Subject] = s
	}

	// 生成两个固定方向的 Anchor
	anchorPHY := buildAnchor312(SubjectPHY, m, globalCos)
	anchorHIS := buildAnchor312(SubjectHIS, m, globalCos)

	return &Mode312Section{
		AnchorPHY: anchorPHY,
		AnchorHIS: anchorHIS,
	}
}

// buildAnchor312
// --------------------------------------
// 按主干方向（PHY/HIS）计算阶段一与阶段二结果
// --------------------------------------
func buildAnchor312(anchor string, m map[string]SubjectScores, globalCos float64) AnchorCoreData {
	// 阶段一计算
	fit := m[anchor].Fit
	abNorm := m[anchor].A / 5.0
	baseCov := AnchorBaseCoverage[anchor]

	termFit := anchW_Fit * fit
	termAbility := anchW_Ability * abNorm
	termCoverage := anchW_Cover * baseCov

	S1 := termFit + termAbility + termCoverage

	// 构建辅科候选池
	var auxPool []string
	if anchor == SubjectPHY {
		auxPool = AuxPoolPHY
	} else {
		auxPool = AuxPoolHIS
	}

	// 阶段二：计算辅科组合
	var combos []ComboCoreData
	var maxSFinal float64

	for i := 0; i < len(auxPool); i++ {
		for j := i + 1; j < len(auxPool); j++ {
			s2, s3 := auxPool[i], auxPool[j]
			key := strings.Join([]string{anchor, s2, s3}, "_")

			cov, ok := Coverage312[key]
			if !ok {
				continue
			}

			avgFit := (m[s2].Fit + m[s3].Fit) / 2.0
			minFit := math.Min(m[s2].Fit, m[s3].Fit)
			expansion := math.Max(0, cov-baseCov)

			// 阶段二计算
			termAvgFit := auxW_AvgFit * avgFit
			termMinFit := auxW_MinFit * minFit
			termExpansion := auxW_Expansion * expansion
			termGlobalCos := auxW_GlobalCos * globalCos

			S23 := termAvgFit + termMinFit + termExpansion + termGlobalCos

			// 阶段三计算
			SFinal := lambda1*S1 + lambda2*S23
			if SFinal > maxSFinal {
				maxSFinal = SFinal
			}

			combos = append(combos, ComboCoreData{
				Aux1:          s2,
				Aux2:          s3,
				AvgFit:        round3(avgFit),
				MinFit:        round3(minFit),
				Expansion:     round3(expansion),
				TermAvgFit:    round3(termAvgFit),
				TermMinFit:    round3(termMinFit),
				TermExpansion: round3(termExpansion),
				TermGlobalCos: round3(termGlobalCos),
				S23:           round3(S23),
			})
		}
	}

	return AnchorCoreData{
		Subject:      anchor,
		Fit:          round3(fit),
		AbilityNorm:  round3(abNorm),
		TermFit:      round3(termFit),
		TermAbility:  round3(termAbility),
		TermCoverage: round3(termCoverage),
		S1:           round3(S1),
		SFinal:       round3(maxSFinal),
		Combos:       combos,
	}
}

// =============================
// RunDemo312
// =============================

func RunDemo312(riasecAnswers []RIASECAnswer, ascAnswers []ASCAnswer, alpha, beta, gamma float64) *ParamForAIPrompt {
	if alpha == 0 && beta == 0 && gamma == 0 {
		alpha, beta, gamma = 0.4, 0.4, 0.2
	}

	scores, globalCos, commonParam := BuildScores(riasecAnswers, ascAnswers, Wfinal, DimCalib, alpha, beta, gamma)

	var paramForPrompt ParamForAIPrompt
	paramForPrompt.Common = commonParam
	paramForPrompt.Mode312 = ScoreCombos312(scores, globalCos)

	content, _ := json.MarshalIndent(&paramForPrompt, "", "  ")
	ts := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("report_param_%s_%s.json", "3+1+2", ts) // 增加了模块名
	_ = os.WriteFile(filename, content, 0644)

	return &paramForPrompt
}
