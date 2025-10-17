package main

// ParamForAIPrompt
// ===========================================
// 用于向 AI 提供报告生成的关键参数（非原始算法数据）
// 由三层组成：
//  1. Common — 通用部分（兴趣-能力总体特征）
//  2. Mode33 — 3+3 模式的选科结果与指标
//  3. Mode312 — 3+1+2 模式的选科结果与指标
//
// ===========================================
type ParamForAIPrompt struct {
	Common  CommonSection  `json:"common"`     // 通用部分：兴趣-能力整体特征
	Mode33  Mode33Section  `json:"mode_3_3"`   // 3+3 模式部分
	Mode312 Mode312Section `json:"mode_3_1_2"` // 3+1+2 模式部分
}

// CommonSection
// ===========================================
// 表示学生在兴趣-能力层面的总体特征。
// 来源：BuildScores()
// 不包含任何语言描述，仅存储可解释性数据。
// ===========================================
type CommonSection struct {
	OverallProfile    OverallProfileData   `json:"overall_profile"`    // 全局一致性与方向
	SubjectProfiles   []SubjectProfileData `json:"subject_profiles"`   // 各学科兴趣/能力指标
	DerivedIndicators DerivedIndicatorData `json:"derived_indicators"` // 平衡性、一致性等衍生指标
}

type OverallProfileData struct {
	GlobalCosine     float64 `json:"global_cosine"`     // 兴趣与能力方向一致性 (0~1)
	AvgFitScore      float64 `json:"avg_fit_score"`     // 所有学科平均匹配度
	FitVariance      float64 `json:"fit_variance"`      // 各学科Fit方差
	AbilityVariance  float64 `json:"ability_variance"`  // 能力分布方差
	InterestVariance float64 `json:"interest_variance"` // 兴趣分布方差
	ZGapMean         float64 `json:"z_gap_mean"`        // 兴趣-能力差异均值
	ZGapRange        float64 `json:"z_gap_range"`       // 兴趣-能力差异极差
}

type SubjectProfileData struct {
	Subject     string  `json:"subject"`      // 科目名
	Interest    float64 `json:"interest"`     // 兴趣原始得分或标准分
	Ability     float64 `json:"ability"`      // 能力原始得分或标准分
	InterestPct float64 `json:"interest_pct"` // 兴趣百分位
	AbilityPct  float64 `json:"ability_pct"`  // 能力百分位
	Fit         float64 `json:"fit"`          // 综合匹配度
	ZGap        float64 `json:"z_gap"`        // 兴趣与能力差
}

type DerivedIndicatorData struct {
	TopInterestSubjects []string `json:"top_interest_subjects"` // 兴趣最高科目
	TopAbilitySubjects  []string `json:"top_ability_subjects"`  // 能力最强科目
	WeakSubjects        []string `json:"weak_subjects"`         // 弱项科目
	BalanceIndex        float64  `json:"balance_index"`         // 能力分布均衡度指标
	StabilityIndex      float64  `json:"stability_index"`       // 匹配度稳定性指标
}

// Mode33Section 3+3 模式：组合推荐与匹配核心结果
type Mode33Section struct{}

// Mode312Section 3+1+2 模式：主干+辅科组合与阶段性特征
type Mode312Section struct{}
