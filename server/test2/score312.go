package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
)

var SubjectCluster = map[string]string{
	"PHY": "STEM", "CHE": "STEM", "BIO": "STEM",
	"HIS": "HUM", "GEO": "HUM", "POL": "HUM",
}

// Coverage312 全国本科招生专业覆盖率 每个选科组合可以报考的全国高校专业比例
var Coverage312 = map[string]float64{
	// ===== 物理组 (Anchor = PHY) =====
	ComboPHY_CHE_POL: 0.99, // 物化政 — 覆盖率最高，接近全开放
	ComboPHY_CHE_BIO: 0.96, // 物化生 — 理工+医学主干
	ComboPHY_CHE_GEO: 0.95, // 物化地 — 地质/材料方向
	ComboPHY_BIO_GEO: 0.88, // 物生地 — 无化学组合中最优
	ComboPHY_BIO_POL: 0.85, // 物生政 — 无化学+跨社科，下降明显
	ComboPHY_GEO_POL: 0.83, // 物地政 — 理工边缘

	// ===== 历史组 (Anchor = HIS) =====
	ComboHIS_GEO_POL: 0.50, // 史地政 — 纯文科主流
	ComboHIS_GEO_BIO: 0.48, // 史地生 — 文理交叉
	ComboHIS_POL_BIO: 0.46, // 史政生 — 覆盖有限
	ComboHIS_CHE_POL: 0.44, // 史化政 — 文理夹层
	ComboHIS_CHE_BIO: 0.46, // 史化生 — 化学略加分
}

// AnchorBaseCoverage Anchor 基线覆盖率（全国平均）
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

var (
	wAvgFit     = 0.25 // 三科平均Fit
	wMinFit     = 0.15 // 三科最小Fit
	wAuxAbility = 0.15 // 辅科能力：0.6*均值 + 0.4*最小（标准化）
	wCoverage   = 0.25 // 覆盖率（查表）
	wCosPos     = 0.10 // 三科兴趣-能力方向一致性（余弦正值）
	wMixPenalty = 0.10 // 结构惩罚（跨簇×低能力×低覆盖^1.2），最终要“减去”
)

// 阶段三权重（综合）
var (
	lambda1, lambda2 = 0.6, 0.4
)

// ScoreCombos312
// =============================
// 核心算法逻辑
// =============================
func ScoreCombos312(scores []SubjectScores) *Mode312Section {
	m := map[string]SubjectScores{}
	for _, s := range scores {
		m[s.Subject] = s
	}

	// 生成两个固定方向的 Anchor
	anchorPHY := buildAnchor312(SubjectPHY, m)
	anchorHIS := buildAnchor312(SubjectHIS, m)

	return &Mode312Section{
		AnchorPHY: anchorPHY,
		AnchorHIS: anchorHIS,
	}
}

// 计算辅科竞争能力（标准化后）
func calculateAuxAbility(s2, s3 string, m map[string]SubjectScores) float64 {
	// 将 A (1~5) 归一化到 [0,1]
	avgAbility := ((m[s2].A / 5.0) + (m[s3].A / 5.0)) / 2.0
	minAbility := math.Min(m[s2].A/5.0, m[s3].A/5.0)

	// 按照 0.6:0.4 加权组合，代表整体实力与短板平衡
	return 0.6*avgAbility + 0.4*minAbility
}

// 计算结构惩罚 MixPenalty（带软跨簇调节）
func calculateMixPenalty(anchor, s2, s3 string, m map[string]SubjectScores, coverage float64) float64 {
	mixRatio := 0.0
	anchorCluster := SubjectCluster[anchor]

	// —— 跨簇检测 ——
	if SubjectCluster[s2] != anchorCluster {
		mixRatio += 0.5
	}
	if SubjectCluster[s3] != anchorCluster {
		mixRatio += 0.5
	}

	minAbilityNorm := math.Min(m[anchor].A/5.0, math.Min(m[s2].A/5.0, m[s3].A/5.0))
	abilityRisk := 1 - minAbilityNorm
	coverageRisk := math.Pow(1-coverage, 1.2)

	// ✅ 新增：软跨簇调节（理↔文组合减半惩罚）
	softCross := false
	if (anchorCluster == "STEM" && (SubjectCluster[s2] == "HUM" || SubjectCluster[s3] == "HUM")) ||
		(anchorCluster == "HUM" && (SubjectCluster[s2] == "STEM" || SubjectCluster[s3] == "STEM")) {
		softCross = true
	}

	basePenalty := mixRatio * abilityRisk * coverageRisk

	if softCross {
		basePenalty *= 0.5 // 理↔文 组合惩罚减半
	}

	return basePenalty // 范围仍在 [0,1]
}

// buildAnchor312
// --------------------------------------
// 按主干方向（PHY/HIS）计算阶段一与阶段二结果
// --------------------------------------
func buildAnchor312(anchor string, m map[string]SubjectScores) AnchorCoreData {
	// 阶段一计算
	fit := m[anchor].Fit
	abNorm := m[anchor].A / 5.0
	baseCov := AnchorBaseCoverage[anchor]

	termFit := anchW_Fit * fit
	termAbility := anchW_Ability * abNorm
	termCoverage := anchW_Cover * baseCov

	S1 := termFit + termAbility + termCoverage

	// 阶段二：计算辅科组合（改进版：直接遍历 Coverage312，过滤匹配 anchor 的键）
	var combos []ComboCoreData
	var maxSFinal = math.Inf(-1)

	for key, cov := range Coverage312 {
		parts := strings.Split(key, "_")
		if len(parts) != 3 || parts[0] != anchor {
			continue
		}
		s2, s3 := parts[1], parts[2]

		avgFit := (m[s2].Fit + m[s3].Fit) / 2.0
		minFit := math.Min(m[s2].Fit, m[s3].Fit)
		comboCosPos := calcComboCos([]SubjectScores{m[anchor], m[s2], m[s3]})
		auxAbility := calculateAuxAbility(s2, s3, m)
		mixPenalty := calculateMixPenalty(anchor, s2, s3, m, cov)

		S23 := wAvgFit*avgFit +
			wMinFit*minFit +
			wAuxAbility*auxAbility +
			wCoverage*cov +
			wCosPos*comboCosPos -
			wMixPenalty*mixPenalty

		// 阶段三计算
		SFinal := lambda1*S1 + lambda2*S23
		if SFinal > maxSFinal {
			maxSFinal = SFinal
		}

		combos = append(combos, ComboCoreData{
			Aux1:        s2,
			Aux2:        s3,
			AvgFit:      round3(avgFit),
			MinFit:      round3(minFit),
			AuxAbility:  round3(auxAbility),
			Coverage:    round3(cov),
			ComboCos:    round3(comboCosPos),
			MixPenalty:  round3(mixPenalty),
			S23:         round3(S23),
			SFinalCombo: round3(SFinal),
		})
	}

	sort.Slice(combos, func(i, j int) bool {
		return combos[i].SFinalCombo > combos[j].SFinalCombo
	})

	if len(combos) > 3 {
		combos = combos[:3]
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

func RunDemo312(riasecAnswers []RIASECAnswer, ascAnswers []ASCAnswer, alpha, beta, gamma float64, idx, yesno, combo string) *ParamForAIPrompt {
	if alpha == 0 && beta == 0 && gamma == 0 {
		alpha, beta, gamma = 0.4, 0.4, 0.2
	}

	scores, result := BuildScores(riasecAnswers, ascAnswers, Wfinal, DimCalib, alpha, beta, gamma)

	var paramForPrompt ParamForAIPrompt
	paramForPrompt.Common = result.Common
	paramForPrompt.Mode312 = ScoreCombos312(scores)

	content, _ := json.MarshalIndent(&paramForPrompt, "", "  ")
	filename := fmt.Sprintf("%s_rp_%s_%s_%s.json", idx, combo, "3+1+2", yesno)
	_ = os.WriteFile(filename, content, 0644)

	fmt.Printf("Radar Visualization:\n%+v\n", result.Radar)

	return &paramForPrompt
}
