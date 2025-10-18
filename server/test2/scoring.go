package main

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// =======================================
// score_fusion.go
// 方案 A（维度加权平均）+ 方案 C（余弦匹配）的融合计分实现
// =======================================
// - 兴趣六维 → 维度加权 → W_final 投影到 6 科（Interest 1–5）
// - 能力：ASC 每科 4 题（含反向题换算）平均（Ability 1–5）
// - 标准化：z-score（在 6 科内部） & 0–100（用于雷达图）
// - 余弦相似度：衡量兴趣与能力在“学科空间”的方向一致性
// - Fit = α·(zA - zI) + β·cos(I, A) + γ·AbilityShare
//   （默认 α=0.4, β=0.4, γ=0.2；可按省/校样本校准）

// SubjectScores
// ---------------------------------------
// 固定变量与类型定义
// ---------------------------------------
type SubjectScores struct {
	Subject string  `json:"subject"`
	I       float64 `json:"interest"`     // 1–5
	A       float64 `json:"ability"`      // 1–5
	IPct    float64 `json:"interest_pct"` // 0–100
	APct    float64 `json:"ability_pct"`  // 0–100
	ZGap    float64 `json:"zgap"`         // z(A) - z(I)
	Fit     float64 `json:"fit"`          // 融合评分（越高越推荐）
}

type RadarPayload struct {
	Subjects []string  `json:"subjects"`
	Interest []float64 `json:"interest_pct"`
	Ability  []float64 `json:"ability_pct"`
}

// DimCalib
// ---------------------------------------
// 参数配置
// ---------------------------------------
var DimCalib = map[string]float64{
	"R": 0.82, "I": 0.87, "A": 0.78, "S": 0.80, "E": 0.75, "C": 0.72,
}

// Wfinal 兴趣→学科权重矩阵（最终版）
var Wfinal = map[string]map[string]float64{
	"PHY": {"R": 0.30, "I": 0.35, "C": 0.15, "E": 0.10, "S": 0.05, "A": 0.05},
	"CHE": {"R": 0.25, "I": 0.35, "C": 0.20, "E": 0.10, "S": 0.05, "A": 0.05},
	"BIO": {"R": 0.20, "I": 0.35, "S": 0.15, "C": 0.15, "A": 0.10, "E": 0.05},
	"GEO": {"R": 0.25, "I": 0.25, "C": 0.15, "S": 0.15, "E": 0.10, "A": 0.10},
	"HIS": {"A": 0.30, "S": 0.25, "E": 0.15, "I": 0.15, "C": 0.10, "R": 0.05},
	"POL": {"E": 0.30, "S": 0.25, "A": 0.15, "I": 0.15, "C": 0.10, "R": 0.05},
}

// ---------------------------------------
// 数据结构与基础运算
// ---------------------------------------

type riasecMean struct{ R, I, A, S, E, C float64 }

func safeDiv(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	return a / b
}

func meanRIASEC(ans []RIASECAnswer) riasecMean {
	var sum riasecMean
	cnt := map[string]int{}
	for _, a := range ans {
		switch strings.ToUpper(a.Dimension) {
		case "R":
			sum.R += float64(a.Score)
			cnt["R"]++
		case "I":
			sum.I += float64(a.Score)
			cnt["I"]++
		case "A":
			sum.A += float64(a.Score)
			cnt["A"]++
		case "S":
			sum.S += float64(a.Score)
			cnt["S"]++
		case "E":
			sum.E += float64(a.Score)
			cnt["E"]++
		case "C":
			sum.C += float64(a.Score)
			cnt["C"]++
		}
	}
	return riasecMean{
		R: safeDiv(sum.R, float64(cnt["R"])),
		I: safeDiv(sum.I, float64(cnt["I"])),
		A: safeDiv(sum.A, float64(cnt["A"])),
		S: safeDiv(sum.S, float64(cnt["S"])),
		E: safeDiv(sum.E, float64(cnt["E"])),
		C: safeDiv(sum.C, float64(cnt["C"])),
	}
}

func getDim(v riasecMean, d string) float64 {
	switch d {
	case "R":
		return v.R
	case "I":
		return v.I
	case "A":
		return v.A
	case "S":
		return v.S
	case "E":
		return v.E
	case "C":
		return v.C
	}
	return 0
}

// ASC 能力平均（含反向题）
func subjectAbility(asc []ASCAnswer) map[string]float64 {
	sum := map[string]float64{}
	cnt := map[string]float64{}
	for _, s := range Subjects {
		sum[s], cnt[s] = 0, 0
	}
	for _, a := range asc {
		sc := float64(a.Score)
		if a.Reverse {
			sc = 6 - sc
		}
		sub := strings.ToUpper(a.Subject)
		sum[sub] += sc
		cnt[sub]++
	}
	out := map[string]float64{}
	for _, s := range Subjects {
		out[s] = safeDiv(sum[s], cnt[s])
	}
	return out
}

// 兴趣投影：RIASEC 六维 → 学科六维
func projectInterest(ria riasecMean, W map[string]map[string]float64, f map[string]float64) map[string]float64 {
	res := map[string]float64{}
	for subj, dimW := range W {
		var total, wsum float64
		for d, w := range dimW {
			total += w * getDim(ria, d) * f[d]
			wsum += w
		}
		if wsum == 0 {
			res[subj] = 0
		} else {
			res[subj] = total / wsum
		}
	}
	return res
}

// 百分制
func toPct(x float64) float64 {
	if x < 1 {
		x = 1
	}
	if x > 5 {
		x = 5
	}
	return (x - 1.0) / 4.0 * 100.0
}

// 六科内 z-score
func z6(m map[string]float64) map[string]float64 {
	mean, sd := 0.0, 0.0
	for _, s := range Subjects {
		mean += m[s]
	}
	mean /= 6.0
	for _, s := range Subjects {
		sd += (m[s] - mean) * (m[s] - mean)
	}
	sd = math.Sqrt(sd / 6.0)
	if sd == 0 {
		sd = 1
	}
	out := map[string]float64{}
	for _, s := range Subjects {
		out[s] = (m[s] - mean) / sd
	}
	return out
}

// 余弦相似度
func cosineSim(a, b map[string]float64) float64 {
	var dot, na2, nb2 float64
	for _, s := range Subjects {
		dot += a[s] * b[s]
		na2 += a[s] * a[s]
		nb2 += b[s] * b[s]
	}
	na := math.Sqrt(na2)
	nb := math.Sqrt(nb2)
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (na * nb)
}

// BuildScores
// ===========================================
// 计算兴趣、能力及综合匹配指标，并生成可解释性结构。
// 返回：
//   - SubjectScores: 每科分数
//   - globalCosine: 全局一致性
//   - CommonSection: 用于 AI 报告解释的核心参数
//
// ===========================================
func BuildScores(
	riasecAnswers []RIASECAnswer,
	ascAnswers []ASCAnswer,
	W map[string]map[string]float64,
	f map[string]float64,
	alpha, beta, gamma float64,
) ([]SubjectScores, float64, *CommonSection) {

	var log CommonSection

	// ---- 1. RIASEC 兴趣均值 ----
	ria := meanRIASEC(riasecAnswers)

	// ---- 2. 兴趣投影 ----
	I := projectInterest(ria, W, f)

	// ---- 3. 能力 ----
	A := subjectAbility(ascAnswers)

	// ---- 4. 标准化 ----
	IZ := z6(I)
	AZ := z6(A)
	ZGap := make(map[string]float64)
	for _, s := range Subjects {
		ZGap[s] = AZ[s] - IZ[s]
	}

	// ---- 5. 一致性 ----
	cos := cosineSim(I, A)

	// ---- 6. 能力占比 ----
	sumA := 0.0
	for _, s := range Subjects {
		sumA += A[s]
	}
	shareA := make(map[string]float64)
	for _, s := range Subjects {
		shareA[s] = safeDiv(A[s], sumA)
	}

	// ---- 7. 每科 Fit ----
	out := make([]SubjectScores, 0, len(Subjects))
	for _, s := range Subjects {
		ipct := toPct(I[s])
		apct := toPct(A[s])
		zgap := AZ[s] - IZ[s]
		share := safeDiv(A[s], sumA)
		fit := alpha*zgap + beta*cos + gamma*share

		out = append(out, SubjectScores{
			Subject: s,
			I:       I[s],
			A:       A[s],
			IPct:    math.Round(ipct),
			APct:    math.Round(apct),
			ZGap:    zgap,
			Fit:     fit,
		})
	}

	// ---- 8. 构建 SubjectProfiles ----
	var subjectProfiles []SubjectProfileData
	for _, s := range Subjects {
		subjectProfiles = append(subjectProfiles, SubjectProfileData{
			Subject:       s,
			InterestScore: round3(I[s]),
			AbilityScore:  round3(A[s]),
			InterestPct:   round3(toPct(I[s])),
			AbilityPct:    round3(toPct(A[s])),
			AbilityShare:  round3(shareA[s]),
			CosineLocal:   round3(cos),
			ZGap:          round3(ZGap[s]),
			Fit:           round3(findFit(out, s)),
		})
	}

	// ---- 9. 计算 OverallProfileData ----
	fits := extractFit(out)
	overall := OverallProfileData{
		GlobalCosine:     round3(cos),
		AvgFitScore:      round3(mean(fits)),
		FitVariance:      round3(variance(fits)),
		ZGapMean:         round3(mean(mapValues(ZGap))),
		ZGapRange:        round3(rangeOf(mapValues(ZGap))),
		AbilityVariance:  round3(variance(mapValues(A))),
		InterestVariance: round3(variance(mapValues(I))),
		BalanceIndex:     round3(1.0 / (1.0 + math.Sqrt(variance(mapValues(A))))),
		TotalAbility:     round3(sumA),
	}

	// ---- 10. 计算 DerivedIndicatorData ----

	// ① 学科方向判定
	stemShare := shareA[SubjectPHY] + shareA[SubjectCHE] + shareA[SubjectBIO]
	artsShare := shareA[SubjectHIS] + shareA[SubjectPOL] + shareA[SubjectGEO]

	den := stemShare + artsShare
	dominantScore := 0.0
	if den > 0 {
		dominantScore = (stemShare - artsShare) / den
	}
	dominantScore = round3(dominantScore)

	// ② 稳定性
	fitStd := math.Sqrt(variance(fits))

	// ③ 错配统计
	var hiLow, loHi, riskSubs []string
	for _, s := range Subjects {
		ip := toPct(I[s])
		ap := toPct(A[s])
		if ip > 70 && ap < 40 {
			hiLow = append(hiLow, s)
		}
		if ap > 70 && ip < 40 {
			loHi = append(loHi, s)
		}
		if A[s] < 3.0 || math.Abs(ZGap[s]) > 1.0 {
			riskSubs = append(riskSubs, s)
		}
	}

	// ④ 风险数值化计算
	n := float64(len(Subjects))
	lowAbilityRatio := float64(len(riskSubs)) / n
	zgapMeanAbs := meanAbs(mapValues(ZGap))
	mismatchRatio := float64(len(hiLow)+len(loHi)) / n
	riskScore := 0.4*lowAbilityRatio + 0.3*zgapMeanAbs/2.0 + 0.2*mismatchRatio + 0.1*fitStd
	if riskScore > 1.0 {
		riskScore = 1.0
	}

	// ⑤ 优劣势学科
	sort.Slice(out, func(i, j int) bool { return out[i].Fit > out[j].Fit })
	var topSubs []string
	var weakSubs []string
	for i := 0; i < minInt(3, len(out)); i++ {
		topSubs = append(topSubs, out[i].Subject)
	}
	for i := len(out) - 1; i >= 0 && len(weakSubs) < 3; i-- {
		weakSubs = append(weakSubs, out[i].Subject)
	}

	// ---- 11. 汇总 DerivedIndicators ----
	log.DerivedIndicators = DerivedIndicatorData{
		DominantScore:          dominantScore,
		FitStdDev:              round3(fitStd),
		HighInterestLowAbility: hiLow,
		HighAbilityLowInterest: loHi,
		TopSubjects:            topSubs,
		WeakSubjects:           weakSubs,
		RiskScore:              round3(riskScore),
		RiskSubjects:           riskSubs,
	}

	// ---- 12. 汇总 CommonSection ----
	log.OverallProfile = overall
	log.SubjectProfiles = subjectProfiles

	fmt.Printf("Radar Visualization:\n%v\n", Radar(out))
	return out, cos, &log
}

// ========== 工具函数 ==========

func findFit(arr []SubjectScores, subj string) float64 {
	for _, s := range arr {
		if s.Subject == subj {
			return s.Fit
		}
	}
	return 0
}

func extractFit(arr []SubjectScores) []float64 {
	out := make([]float64, len(arr))
	for i, s := range arr {
		out[i] = s.Fit
	}
	return out
}

func mapValues(m map[string]float64) []float64 {
	out := make([]float64, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}

func variance(xs []float64) float64 {
	if len(xs) < 2 {
		return 0
	}
	m := mean(xs)
	sum := 0.0
	for _, v := range xs {
		d := v - m
		sum += d * d
	}
	return sum / float64(len(xs))
}

func mean(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range xs {
		sum += v
	}
	return sum / float64(len(xs))
}

func meanAbs(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range xs {
		sum += math.Abs(v)
	}
	return sum / float64(len(xs))
}

func rangeOf(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	minV, maxV := xs[0], xs[0]
	for _, v := range xs {
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}
	return maxV - minV
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func round3(x float64) float64 { return math.Round(x*1000) / 1000 }

// Radar 雷达图载荷
func Radar(scores []SubjectScores) RadarPayload {
	ip, ap := make([]float64, 6), make([]float64, 6)
	idx := map[string]int{
		SubjectPHY: 0,
		SubjectCHE: 1,
		SubjectBIO: 2,
		SubjectGEO: 3,
		SubjectHIS: 4,
		SubjectPOL: 5,
	}
	for _, s := range scores {
		k := idx[s.Subject]
		ip[k] = s.IPct
		ap[k] = s.APct
	}
	return RadarPayload{Subjects: Subjects, Interest: ip, Ability: ap}
}
