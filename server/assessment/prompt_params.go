package assessment

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
	Common  *CommonSection  `json:"common"`               // 通用部分：兴趣-能力整体特征
	Mode33  *Mode33Section  `json:"mode_3_3,omitempty"`   // 3+3 模式部分
	Mode312 *Mode312Section `json:"mode_3_1_2,omitempty"` // 3+1+2 模式部分
}

// CommonSection
// ===========================================
// 表示学生在兴趣-能力层面的总体特征。
// 来源：BuildScores()
// 不包含任何语言描述，仅存储可解释性数据。
// ===========================================
type CommonSection struct {
	GlobalCosine float64              `json:"global_cosine"` // 全局方向一致性
	Subjects     []SubjectProfileData `json:"subjects"`      // 六科详细信息
	QualityScore float64              `json:"quality_score"`
}

// SubjectProfileData
// 表示单个学科在兴趣-能力融合算法中的关键中间结果。
// 所有值均可直接由 BuildScores 推导，无需额外计算。
type SubjectProfileData struct {
	Subject      string  `json:"subject"`       // 学科名，例如 PHY
	InterestZ    float64 `json:"interest_z"`    // 标准化兴趣（z(Ij)）
	AbilityZ     float64 `json:"ability_z"`     // 标准化能力（z(Aj)）
	ZGap         float64 `json:"zgap"`          // 差值 z(Aj)-z(Ij)
	AbilityShare float64 `json:"ability_share"` // 能力占比 Aj / ΣAk
	Fit          float64 `json:"fit"`           // 该学科最终匹配度 Fitj
}

// Mode312Section 3+1+2 模式：主干+辅科组合与阶段性特征
// ===========================================
// 3+1+2 模式的核心过程参数（仅记录必要信息）
// 用于解释算法过程与结果，非展示层。
// ===========================================
type Mode312Section struct {
	AnchorPHY AnchorCoreData `json:"anchor_phy"` // 理科主干（物理组）
	AnchorHIS AnchorCoreData `json:"anchor_his"` // 文科主干（历史组）
}

// AnchorCoreData
// 表示单个主干学科（Anchor）的完整阶段信息
type AnchorCoreData struct {
	Subject      string          `json:"subject"`       // 主干学科 ("PHY" / "HIS")
	Fit          float64         `json:"fit"`           // 匹配度（兴趣-能力契合）
	AbilityNorm  float64         `json:"ability_norm"`  // 归一化能力（0~1）
	TermFit      float64         `json:"term_fit"`      // 契合度项贡献
	TermAbility  float64         `json:"term_ability"`  // 能力项贡献
	TermCoverage float64         `json:"term_coverage"` // 覆盖项贡献
	S1           float64         `json:"s1"`            // 阶段一综合得分（主干稳定性）
	Combos       []ComboCoreData `json:"combos"`        // 阶段二组合结果
	SFinal       float64         `json:"s_final"`       // 阶段三综合分（用于排序）
}

// ComboCoreData
// 属于特定 Anchor 的辅科组合信息
type ComboCoreData struct {
	Aux1        string  `json:"aux1"`
	Aux2        string  `json:"aux2"`
	AvgFit      float64 `json:"avg_fit"`
	MinFit      float64 `json:"min_fit"`
	ComboCos    float64 `json:"combo_cos"`
	AuxAbility  float64 `json:"auxAbility"`
	Coverage    float64 `json:"coverage"`
	MixPenalty  float64 `json:"mix_penalty"`
	S23         float64 `json:"s23"`
	SFinalCombo float64 `json:"s_final_combo"`
}

// Mode33Section 3+3 模式：组合推荐与匹配核心结果
type Mode33Section struct {
	TopCombinations []Combo33CoreData `json:"top_combinations"` // 前5推荐组合
}

// Combo33CoreData
// ===========================================
// 表示单个 3科组合的可解释数据
// ===========================================
type Combo33CoreData struct {
	Subjects    [3]string `json:"subjects"`     // 三科组合
	AvgFit      float64   `json:"avg_fit"`      // 平均匹配度（原始值）
	MinAbility  float64   `json:"min_ability"`  // 最低能力（原始值）
	Rarity      float64   `json:"rarity"`       // 稀有性（原始值 0/5/12）
	RiskPenalty float64   `json:"risk_penalty"` // 风险惩罚（原始值 0 或 0.2）
	Score       float64   `json:"score"`        // 综合推荐得分（最终输出）
	ComboCosine float64   `json:"combo_cosine"`
}

type RadarData struct {
	Subjects    []string  `json:"subjects"`     // ["PHY","CHE","BIO","GEO","HIS","POL"]
	InterestPct []float64 `json:"interest_pct"` // [61, 60, 58, 55, 40, 39]
	AbilityPct  []float64 `json:"ability_pct"`  // [100, 100, 100, 50, 44, 44]
}

type FullScoreResult struct {
	Common *CommonSection `json:"common"` // 算法核心因子（Fit 计算）
	Radar  *RadarData     `json:"radar"`  // 展示数据（兴趣/能力雷达）
}
