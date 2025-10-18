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

// OverallProfileData
// ============================================
// 全局兴趣-能力结构画像
// 来源：BuildScores 聚合层
// ============================================
type OverallProfileData struct {
	GlobalCosine     float64 `json:"global_cosine"`     // 兴趣与能力总体方向一致性 (0~1)
	AvgFitScore      float64 `json:"avg_fit_score"`     // 所有学科平均匹配度
	FitVariance      float64 `json:"fit_variance"`      // 各学科Fit的方差（匹配稳定性）
	ZGapMean         float64 `json:"z_gap_mean"`        // 兴趣-能力差异均值
	ZGapRange        float64 `json:"z_gap_range"`       // 兴趣-能力差异极差
	AbilityVariance  float64 `json:"ability_variance"`  // 能力分布方差
	InterestVariance float64 `json:"interest_variance"` // 兴趣分布方差
	BalanceIndex     float64 `json:"balance_index"`     // 能力均衡度指标（0~1）
	TotalAbility     float64 `json:"total_ability"`     // 能力总和（相对实力）
}

// SubjectProfileData
// 表示单个学科在兴趣-能力融合算法中的关键中间结果。
// 所有值均可直接由 BuildScores 推导，无需额外计算。
type SubjectProfileData struct {
	Subject       string  `json:"subject"`        // 科目标识，如 "PHY"
	InterestScore float64 `json:"interest_score"` // 兴趣综合得分 (由 RIASEC 映射得出)
	AbilityScore  float64 `json:"ability_score"`  // 能力综合得分 (来自 ASC)
	InterestPct   float64 `json:"interest_pct"`   // 兴趣百分位（相对比较）
	AbilityPct    float64 `json:"ability_pct"`    // 能力百分位
	AbilityShare  float64 `json:"ability_share"`  // 能力占总能力比
	CosineLocal   float64 `json:"cosine_local"`   // 兴趣与能力方向一致性
	ZGap          float64 `json:"z_gap"`          // 兴趣-能力差异 (zA - zI)
	Fit           float64 `json:"fit"`            // 综合匹配度 (最终得分)
	RiskFlag      bool    `json:"risk_flag"`      // 是否存在潜在风险 (低能力或兴趣过偏)
}

// DerivedIndicatorData
// ============================================
// 表示从单科与总体数据中派生出的综合性结构特征
// 不含任何文字描述，由 AI 解释
// ============================================
type DerivedIndicatorData struct {
	DominantScore          float64  `json:"dominant_score"` // [-1,1]：>0 偏STEM，<0 偏Arts，≈0混合
	FitStdDev              float64  `json:"fit_stddev"`
	HighInterestLowAbility []string `json:"high_interest_low_ability"`
	HighAbilityLowInterest []string `json:"high_ability_low_interest"`
	TopSubjects            []string `json:"top_subjects"`
	WeakSubjects           []string `json:"weak_subjects"`
	RiskScore              float64  `json:"risk_score"`
	RiskSubjects           []string `json:"risk_subjects"`
}

// Mode33Section 3+3 模式：组合推荐与匹配核心结果
type Mode33Section struct{}

// Mode312Section 3+1+2 模式：主干+辅科组合与阶段性特征
type Mode312Section struct{}
