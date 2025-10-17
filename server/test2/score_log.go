package main

import "sort"

// =======================
// 结构定义部分
// =======================

// LogParams 参数
type LogParams struct {
	Alpha        float64 `json:"alpha"`
	Beta         float64 `json:"beta"`
	Gamma        float64 `json:"gamma"`
	GlobalCosine float64 `json:"global_cosine"`
}

// LogRIASECMean RIASEC 兴趣均值
type LogRIASECMean struct {
	Values map[string]float64 `json:"values"`
}

// LogInterestProjection 兴趣投影（RIASEC → 学科）
type LogInterestProjection struct {
	Values  map[string]float64 `json:"values"`
	Ranking []string           `json:"ranking,omitempty"`
}

// LogAbility 学科能力
type LogAbility struct {
	Values  map[string]float64 `json:"values"`
	Percent map[string]float64 `json:"percent,omitempty"`
}

// LogStandardization 标准化
type LogStandardization struct {
	ZInterest map[string]float64 `json:"z_interest"`
	ZAbility  map[string]float64 `json:"z_ability"`
	ZGap      map[string]float64 `json:"z_gap"`
}

// LogConsistency 一致性
type LogConsistency struct {
	AbilityShare map[string]float64 `json:"ability_share"`
	TotalAbility float64            `json:"total_ability"`
}

// LogFitDetail Fit 细节
type LogFitDetail struct {
	Fit         float64 `json:"fit"`
	AlphaTerm   float64 `json:"alpha_term"`
	BetaTerm    float64 `json:"beta_term"`
	GammaTerm   float64 `json:"gamma_term"`
	AbilityPct  float64 `json:"ability_pct"`
	InterestPct float64 `json:"interest_pct"`
}

// BuildScoresLog 主体结构
type BuildScoresLog struct {
	Params             LogParams               `json:"params"`
	RIASECMean         LogRIASECMean           `json:"riasec_mean"`
	InterestProjection LogInterestProjection   `json:"interest_projection"`
	Ability            LogAbility              `json:"ability"`
	Standardization    LogStandardization      `json:"standardization"`
	Consistency        LogConsistency          `json:"consistency"`
	FitDetail          map[string]LogFitDetail `json:"fit_detail"`
}

func toPercentMap(m map[string]float64) map[string]float64 {
	out := make(map[string]float64)
	for k, v := range m {
		out[k] = toPct(v)
	}
	return out
}

func sortByValueDesc(m map[string]float64) []string {
	type kv struct {
		Key   string
		Value float64
	}
	arr := make([]kv, 0, len(m))
	for k, v := range m {
		arr = append(arr, kv{k, v})
	}
	sort.Slice(arr, func(i, j int) bool { return arr[i].Value > arr[j].Value })
	out := make([]string, len(arr))
	for i, kv := range arr {
		out[i] = kv.Key
	}
	return out
}

type ComboExplainLog struct {
	Mode         string             `json:"mode"`          // "3+3" or "3+1+2"
	GlobalCosine float64            `json:"global_cosine"` // 兴趣能力方向一致性
	Summary      ComboSummary       `json:"summary"`       // 全局推荐概况
	TopCombos    []ComboExplainItem `json:"top_combos"`    // 推荐组合
	// --- 以下为3+1+2专用字段，可为空 ---
	FixedSubject   string             `json:"fixed_subject,omitempty"`   // 主科
	GroupsOverview []GroupExplainItem `json:"groups_overview,omitempty"` // 组别选择结果
	// --- 通用元信息 ---
	Version   string `json:"version"`
	Timestamp string `json:"timestamp"`
}

type ComboSummary struct {
	BestCategory      string  `json:"best_category"`      // 推荐方向：理科 / 文科 / 综合 / 工科
	BestFit           float64 `json:"best_fit"`           // 最优组合平均匹配度（非原始分数）
	Stability         float64 `json:"stability"`          // 匹配度稳定性（= 1 - std(Fit)）
	Balance           float64 `json:"balance"`            // 能力分布均衡度（0~1）
	InterestAlignment float64 `json:"interest_alignment"` // 兴趣方向一致性（由余弦或聚类计算）
	RiskLevel         string  `json:"risk_level"`         // 高/中/低
	Commentary        string  `json:"commentary"`         // AI 报告引言，例如“整体理科倾向明显，匹配度稳定”
}

type ComboExplainItem struct {
	Rank         int        `json:"rank"`          // 排名（1,2,3,...）
	Subjects     []string   `json:"subjects"`      // 组合学科 ["PHY","CHE","BIO"]
	Category     string     `json:"category"`      // 组合方向：理科/文科/综合
	AvgFit       float64    `json:"avg_fit"`       // 平均匹配度（兴趣-能力融合）
	Stability    float64    `json:"stability"`     // 稳定性（Fit标准差的反比）
	Balance      float64    `json:"balance"`       // 能力均衡度
	InterestBias string     `json:"interest_bias"` // 兴趣偏向（如“偏探究/动手”）
	Strengths    []string   `json:"strengths"`     // 优势点（如“兴趣高、能力强、风险低”）
	Weaknesses   []string   `json:"weaknesses"`    // 劣势点（如“地理匹配度偏低”）
	FitProfile   []FitPoint `json:"fit_profile"`   // 每个科目的 Fit 概况
	RiskLevel    string     `json:"risk_level"`    // 低/中/高
	SummaryText  string     `json:"summary_text"`  // AI 报告生成模板文本（推荐理由）
}

type FitPoint struct {
	Subject        string  `json:"subject"`        // "PHY"
	Fit            float64 `json:"fit"`            // 综合匹配度
	AbilityPct     float64 `json:"ability_pct"`    // 能力百分位
	InterestPct    float64 `json:"interest_pct"`   // 兴趣百分位
	ZGap           float64 `json:"z_gap"`          // 兴趣-能力差距
	Interpretation string  `json:"interpretation"` // “兴趣高且能力强”“兴趣偏低但能力突出”
}

type GroupExplainItem struct {
	GroupName     string   `json:"group_name"`         // “理科组”/“文科组”
	CandidateSubs []string `json:"candidate_subjects"` // ["BIO","GEO"]
	BestChoice    string   `json:"best_choice"`        // "BIO"
	Rationale     string   `json:"rationale"`          // “BIO 兴趣与能力匹配度高 15%”
}
