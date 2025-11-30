package ai_api

import (
	"math"
	"strings"
)

type SubjectScores struct {
	Subject string  `json:"subject"`
	I       float64 `json:"interest"`   // 原始兴趣值1-5
	A       float64 `json:"ability"`    // 原始能力值1-5
	IZ      float64 `json:"interest_z"` // 兴趣Z-Score
	AZ      float64 `json:"ability_z"`  // 能力Z-Score
	Fit     float64 `json:"fit"`        // 匹配度
}

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

func toPct(x float64) float64 {
	if x < 1 {
		x = 1
	}
	if x > 5 {
		x = 5
	}
	return (x - 1.0) / 4.0 * 100.0
}

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

func BuildScores(
	riasecAnswers []RIASECAnswer,
	ascAnswers []ASCAnswer,
	W map[string]map[string]float64,
	f map[string]float64,
	subWeight SubjectWeight,
) ([]SubjectScores, *FullScoreResult) {

	var common CommonSection

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
	cos := cosineSim(IZ, AZ)
	common.GlobalCosine = round3(cos)
	common.GlobalCosineScore = NormalizeMetric("common.global_cosine", common.GlobalCosine)

	// ---- 6. 能力占比 ----
	sumA := 0.0
	for _, s := range Subjects {
		sumA += A[s]
	}
	shareA := make(map[string]float64)
	for _, s := range Subjects {
		shareA[s] = safeDiv(A[s], sumA)
	}

	qualityScore := assessInterestQuality(riasecAnswers)

	// 动态调整权重。假设 alpha, beta, gamma 是 BuildScores 的输入参数
	newSW := subWeight.adjustWeights(qualityScore)
	common.QualityScore = round3(qualityScore)
	common.QualityScoreScore = NormalizeMetric("common.quality_score", qualityScore)
	
	// ---- 7. 每科 Fit ----
	out := make([]SubjectScores, 0, len(Subjects))
	for _, s := range Subjects {

		diff := math.Abs(AZ[s] - IZ[s])

		p := 1.1
		if diff >= 1.0 {
			p = 1.2
		}
		zgap := -math.Pow(diff, p)

		gate := 1.0 / (1.0 + math.Exp(-(AZ[s]+1.0)/0.45))
		//share := safeDiv(A[s], sumA)
		share := shareA[s]

		alphaAdj := newSW.alpha
		if AZ[s] > IZ[s] {
			alphaAdj *= 0.5
		}
		fit := gate * (alphaAdj*zgap + newSW.beta*cos + newSW.gamma*share)

		out = append(out, SubjectScores{
			Subject: s,
			I:       I[s],
			A:       A[s],
			AZ:      AZ[s],
			IZ:      IZ[s],
			Fit:     fit,
		})
	}

	// ---- 8. 构建 CommonSection.Subjects ----
	var subjectProfiles []SubjectProfileData
	for _, s := range Subjects {

		rawFit := findFit(out, s)
		fitScore := NormalizeMetric("subjects.fit", rawFit)

		subjectProfiles = append(subjectProfiles, SubjectProfileData{
			Subject:      s,
			InterestZ:    round3(IZ[s]),
			AbilityZ:     round3(AZ[s]),
			AbilityShare: round3(shareA[s]),
			ZGap:         round3(ZGap[s]),
			Fit:          round3(rawFit),

			FitScore: fitScore,
		})
	}
	common.Subjects = subjectProfiles

	var radar RadarData
	for _, s := range Subjects {
		radar.Subjects = append(radar.Subjects, s)
		radar.InterestPct = append(radar.InterestPct, round3(toPct(I[s])))
		radar.AbilityPct = append(radar.AbilityPct, round3(toPct(A[s])))
	}

	result := FullScoreResult{
		Common: &common,
		Radar:  &radar,
	}

	return out, &result
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

func round3(x float64) float64 { return math.Round(x*1000) / 1000 }

func calcComboCos(aux []SubjectScores) float64 {
	a := make(map[string]float64)
	b := make(map[string]float64)
	for _, sc := range aux {
		a[sc.Subject] = sc.AZ
		b[sc.Subject] = sc.IZ
	}

	return cosineSim(a, b)
}

func calculateRiskPenalty(minA, avgFit float64) float64 {
	const (
		MinAThreshold = 3.0
		MaxPenalty    = 0.2
	)

	base := math.Max(0, (MinAThreshold-minA)/(MinAThreshold-1.0))
	risk := MaxPenalty * base

	if avgFit < 0 {
		risk *= 1 + 0.3*math.Abs(avgFit)
	} else {
		risk *= 1 - 0.3*avgFit
	}

	risk = math.Min(math.Max(risk, 0), 0.25)

	return risk
}

func assessInterestQuality(riasecAnswers []RIASECAnswer) float64 {

	consistency := checkRIASECConsistency(riasecAnswers)

	pattern := checkResponsePattern(riasecAnswers)

	quality := 0.7*consistency + 0.3*pattern

	return math.Max(quality, 0.4)

}

func checkRIASECConsistency(answers []RIASECAnswer) float64 {

	scores := extractRIASECScores(answers)
	mean, std := calcStats(scores)

	stdPenalty := math.Max(0, 1-std/0.5)
	meanPenalty := math.Abs(mean-3.0) / 2.0

	return math.Max(0, 1.0-0.5*stdPenalty-0.5*meanPenalty)
}

func checkResponsePattern(answers []RIASECAnswer) float64 {

	counts := make(map[int]int)
	for _, a := range answers {
		counts[a.Score]++
	}

	n := float64(len(answers))
	if n == 0 {
		return 0.5
	}

	extremeRatio := float64(counts[1]+counts[5]) / n
	if extremeRatio <= 0.2 {

		return 1.0

	}

	return math.Max(0, 1.0-(extremeRatio-0.2)/0.8)
}

func calcStats(values []float64) (mean, std float64) {
	n := float64(len(values))
	if n == 0 {
		return 0, 0
	}

	var sum float64
	for _, v := range values {
		sum += v
	}
	mean = sum / n

	if n <= 1 {
		return mean, 0.0
	}

	var variance float64
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}

	std = math.Sqrt(variance / (n - 1))
	return mean, std
}

func extractRIASECScores(answers []RIASECAnswer) []float64 {
	if len(answers) == 0 {
		return []float64{3, 3, 3, 3, 3, 3} // 返回中性分，防止除零
	}

	dimSum := make(map[string]float64)
	dimCount := make(map[string]float64)

	for _, a := range answers {
		dim := strings.ToUpper(strings.TrimSpace(a.Dimension))
		dimSum[dim] += float64(a.Score)
		dimCount[dim]++
	}

	dims := []string{"R", "I", "A", "S", "E", "C"}
	result := make([]float64, 0, 6)

	for _, d := range dims {
		if dimCount[d] > 0 {
			result = append(result, dimSum[d]/dimCount[d])
		} else {
			result = append(result, 3.0)
		}
	}

	return result
}
