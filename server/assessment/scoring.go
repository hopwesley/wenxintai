package assessment

import (
	"math"
	"strings"
)

// =======================================
// score_fusion.go
// 方案 IZ（维度加权平均）+ 方案 C（余弦匹配）的融合计分实现
// =======================================
// - 兴趣六维 → 维度加权 → W_final 投影到 6 科（Interest 1–5）
// - 能力：ASC 每科 4 题（含反向题换算）平均（Ability 1–5）
// - 标准化：z-score（在 6 科内部） & 0–100（用于雷达图）
// - 余弦相似度：衡量兴趣与能力在“学科空间”的方向一致性
// - Fit = α·(zA - zI) + β·cos(AZ, IZ) + γ·AbilityShare
//   （默认 α=0.4, β=0.4, γ=0.2；可按省/校样本校准）

// SubjectScores
// ---------------------------------------
// 固定变量与类型定义
// ---------------------------------------
type SubjectScores struct {
	Subject string  `json:"subject"`
	I       float64 `json:"interest"`   // 原始兴趣值1-5
	A       float64 `json:"ability"`    // 原始能力值1-5
	IZ      float64 `json:"interest_z"` // 兴趣Z-Score
	AZ      float64 `json:"ability_z"`  // 能力Z-Score
	Fit     float64 `json:"fit"`        // 匹配度
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

func abilityGate(az, center, steep, floor float64) float64 {
	// floor 可传 0 表示不用软下限；建议 0.2~0.25
	x := (az - center) / steep
	g := 1.0 / (1.0 + math.Exp(-x))
	if floor > 0 && g < floor {
		return floor
	}
	return g
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

	// ---- 6. 能力占比 ----
	sumA := 0.0
	for _, s := range Subjects {
		sumA += A[s]
	}
	shareA := make(map[string]float64)
	for _, s := range Subjects {
		shareA[s] = safeDiv(A[s], sumA)
	}

	// 注意：assessInterestQuality 在这个简化版中不再需要 subjectScores
	qualityScore := assessInterestQuality(riasecAnswers)

	// 动态调整权重。假设 alpha, beta, gamma 是 BuildScores 的输入参数
	alpha, beta, gamma = adjustWeights(alpha, beta, gamma, qualityScore)
	common.QualityScore = round3(qualityScore)
	// ---- 7. 每科 Fit ----
	out := make([]SubjectScores, 0, len(Subjects))
	for _, s := range Subjects {

		diff := math.Abs(AZ[s] - IZ[s])

		// 分段幂次：中度差距从重惩罚，重度差距仍维持强惩罚
		p := 1.1
		if diff >= 1.0 {
			p = 1.2
		}
		zgap := -math.Pow(diff, p)

		// 软门控：低能力科目下调 fit，但不硬过滤
		gate := 1.0 / (1.0 + math.Exp(-(AZ[s]+1.0)/0.45))
		share := safeDiv(A[s], sumA)

		alphaAdj := alpha
		if AZ[s] > IZ[s] {
			alphaAdj *= 0.5
		}
		fit := gate * (alphaAdj*zgap + beta*cos + gamma*share)

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
		subjectProfiles = append(subjectProfiles, SubjectProfileData{
			Subject:      s,
			InterestZ:    round3(IZ[s]),
			AbilityZ:     round3(AZ[s]),
			AbilityShare: round3(shareA[s]),
			ZGap:         round3(ZGap[s]),
			Fit:          round3(findFit(out, s)),
		})
	}
	common.Subjects = subjectProfiles

	// ---- 9. 构建 RadarData ----
	var radar RadarData
	for _, s := range Subjects {
		radar.Subjects = append(radar.Subjects, s)
		radar.InterestPct = append(radar.InterestPct, round3(toPct(I[s])))
		radar.AbilityPct = append(radar.AbilityPct, round3(toPct(A[s])))
	}

	// ---- 10. 构建 FullScoreResult ----
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

	// 降低放大系数：负Fit时最多放大30%（原50%）
	if avgFit < 0 {
		risk *= 1 + 0.3*math.Abs(avgFit) // 温和放大
	} else {
		risk *= 1 - 0.3*avgFit // 保留正向减轻
	}

	// 严格上限 + 最小值
	risk = math.Min(math.Max(risk, 0), 0.25) // 原0.3 → 0.25

	return risk
}

// 兴趣质量评估 - 最终推荐版

func assessInterestQuality(riasecAnswers []RIASECAnswer) float64 {

	consistency := checkRIASECConsistency(riasecAnswers) // 内部一致性

	pattern := checkResponsePattern(riasecAnswers) // 答题模式

	quality := 0.7*consistency + 0.3*pattern // 加权平均

	return math.Max(quality, 0.4) // 保证最低质量

}

// 检查内部一致性：分化程度和均值合理性

func checkRIASECConsistency(answers []RIASECAnswer) float64 {

	scores := extractRIASECScores(answers) // 获取六维平均分

	mean, std := calcStats(scores) // 计算均值和标准差

	// 标准差惩罚：std<0.5时渐进惩罚

	stdPenalty := math.Max(0, 1-std/0.5)

	// 均值惩罚：偏离3.0时渐进惩罚

	meanPenalty := math.Abs(mean-3.0) / 2.0

	return math.Max(0, 1.0-0.5*stdPenalty-0.5*meanPenalty)

}

// 检查答题模式：极端回答比例

func checkResponsePattern(answers []RIASECAnswer) float64 {

	counts := make(map[int]int)

	for _, a := range answers {
		counts[a.Score]++
	}

	n := float64(len(answers))

	if n == 0 {
		return 0.5
	} // 防除零

	// 计算极端答案比例（1分或5分）

	extremeRatio := float64(counts[1]+counts[5]) / n

	// 极端比例>0.2时开始线性惩罚

	if extremeRatio <= 0.2 {

		return 1.0

	}

	// 线性惩罚：从0.2到1.0，质量从1.0降到0.0

	return math.Max(0, 1.0-(extremeRatio-0.2)/0.8)

}

// 权重动态调整 - 归一化线性吸收

func adjustWeights(alpha, beta, gamma, quality float64) (float64, float64, float64) {

	// 兴趣权重衰减

	a := alpha * quality

	b := beta * quality

	// 能力权重吸收（吸收50%的释放权重）

	g := gamma + (alpha+beta)*(1-quality)*0.5

	// 归一化

	sum := a + b + g

	return a / sum, b / sum, g / sum

}

// 计算一组float64数据的均值和样本标准差（带Bessel校正）
// 计算均值与样本标准差（带Bessel校正）
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

// 提取六个RIASEC维度的平均得分
// 假设 RIASECAnswer 结构体至少包含字段：Dimension string, Score int
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

	// 维度顺序固定：R, I, A, S, E, C
	dims := []string{"R", "I", "A", "S", "E", "C"}
	result := make([]float64, 0, 6)

	for _, d := range dims {
		if dimCount[d] > 0 {
			result = append(result, dimSum[d]/dimCount[d])
		} else {
			result = append(result, 3.0) // 缺失维度用中性分补齐
		}
	}
	return result
}
