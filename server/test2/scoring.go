package main

import (
	"encoding/json"
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

// ---------------------------------------
// 参数配置
// ---------------------------------------

var DimCalib = map[string]float64{
	"R": 0.82, "I": 0.87, "A": 0.78, "S": 0.80, "E": 0.75, "C": 0.72,
}

// 兴趣→学科权重矩阵（最终版）
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

func round1(x float64) float64 { return math.Round(x*10) / 10 }
func round2(x float64) float64 { return math.Round(x*100) / 100 }

// ---------------------------------------
// 核心融合评分流程
// ---------------------------------------

func BuildScores(
	riasecAnswers []RIASECAnswer,
	ascAnswers []ASCAnswer,
	W map[string]map[string]float64,
	f map[string]float64,
	alpha, beta, gamma float64,
) ([]SubjectScores, float64) {

	ria := meanRIASEC(riasecAnswers)
	I := projectInterest(ria, W, f)
	A := subjectAbility(ascAnswers)

	IZ := z6(I)
	AZ := z6(A)

	cos := cosineSim(I, A)
	sumA := 0.0
	for _, s := range Subjects {
		sumA += A[s]
	}

	out := make([]SubjectScores, 0, 6)
	for _, s := range Subjects {
		apct, ipct := toPct(A[s]), toPct(I[s])
		zgap := AZ[s] - IZ[s]
		shareA := 0.0
		if sumA > 0 {
			shareA = A[s] / sumA
		}
		fit := alpha*zgap + beta*cos + gamma*shareA

		out = append(out, SubjectScores{
			Subject: s,
			I:       round1(I[s]),
			A:       round1(A[s]),
			IPct:    math.Round(ipct),
			APct:    math.Round(apct),
			ZGap:    round2(zgap),
			Fit:     round2(fit),
		})
	}

	return out, cos
}

// 雷达图载荷
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

// ---------------------------------------
// 组合推荐模块
// ---------------------------------------

type Combo struct {
	Subs   [3]string
	Score  float64
	Reason string
}

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

// 稀有度惩罚
func RarityPenalty(subs [3]string) float64 {
	key := strings.Join(subs[:], "_")
	if v, ok := ComboRarity[key]; ok {
		return v
	}
	return 12
}

// 组合打分
type Weights struct{ W1, W2, W3, W4, W5 float64 }

func ScoreCombos(scores []SubjectScores, globalCos float64, w Weights) []Combo {
	m := map[string]SubjectScores{}
	for _, s := range scores {
		m[s.Subject] = s
	}

	combos := generateAllCombos(Subjects)
	out := []Combo{}

	for _, subs := range combos {
		key := strings.Join([]string{subs[0], subs[1], subs[2]}, "+")
		if !AllowedCombos[key] {
			continue
		}

		a := m[subs[0]]
		b := m[subs[1]]
		c := m[subs[2]]

		fitMean := (a.Fit + b.Fit + c.Fit) / 3.0
		minAbilityPct := min3(a.APct, b.APct, c.APct)

		// 风险惩罚：高兴趣低能力（差值>15）
		riskPenalty := 0.0
		for _, x := range []SubjectScores{a, b, c} {
			if x.APct < 50 && x.IPct > 75 && (x.IPct-x.APct) > 15 {
				riskPenalty += 1.0
			}
		}

		rarity := RarityPenalty(subs)
		score := w.W1*fitMean + w.W2*minAbilityPct + w.W3*globalCos - w.W4*rarity - w.W5*riskPenalty

		reason := fmt.Sprintf(
			"组合[%s,%s,%s]：平均Fit=%.1f；最低能力=%.0f；方向一致性=%.2f；稀有度惩罚=%.0f；风险项=%.0f。",
			subs[0], subs[1], subs[2], fitMean, minAbilityPct, globalCos, rarity, riskPenalty,
		)
		out = append(out, Combo{Subs: subs, Score: round2(score), Reason: reason})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Score > out[j].Score })
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

func min3(a, b, c float64) float64 {
	if a > b {
		a = b
	}
	if a > c {
		a = c
	}
	return a
}

// ---------------------------------------
// 演示入口
// ---------------------------------------

func RunDemo(riasecAnswers []RIASECAnswer, ascAnswers []ASCAnswer, alpha, beta, gamma float64) {
	if alpha == 0 && beta == 0 && gamma == 0 {
		alpha, beta, gamma = 0.4, 0.4, 0.2
	}

	scores, globalCos := BuildScores(riasecAnswers, ascAnswers, Wfinal, DimCalib, alpha, beta, gamma)
	ws := Weights{W1: 0.5, W2: 0.3, W3: 0.1, W4: 0.1, W5: 0.1}
	combRank := ScoreCombos(scores, globalCos, ws)

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
