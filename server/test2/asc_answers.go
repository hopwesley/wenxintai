package main

import (
	"encoding/json"
	"fmt"
)

type ASCAnswer struct {
	ID      int    `json:"id"`
	Subject string `json:"subject"`
	Score   int    `json:"score"`   // 1–5
	Reverse bool   `json:"reverse"` // 与题干一致；此处为“答案分”而非换算分
	Subtype string `json:"subtype"`
}

// ============ PHY_CHE_BIO：匹配（Aligned） ============
// 物理/化学/生物：高分（Comparison/Efficacy/Achievement=5；SkillMastery=1）
// 其他学科：中性（GEO/HIS/POL 题设给 3；HIS/POL 的 Comparison 稍低 2 以拉开差距）
var ASC_Aligned_PHY_CHE_BIO = []ASCAnswer{
	// PHY (1–4)
	{1, SubjectPHY, 5, false, "Comparison"},
	{2, SubjectPHY, 5, false, "Efficacy"},
	{3, SubjectPHY, 5, false, "AchievementExpectation"},
	{4, SubjectPHY, 1, true, "SkillMastery"},
	// CHE (5–8)
	{5, SubjectCHE, 5, false, "Comparison"},
	{6, SubjectCHE, 5, false, "Efficacy"},
	{7, SubjectCHE, 5, false, "AchievementExpectation"},
	{8, SubjectCHE, 1, true, "SkillMastery"},
	// BIO (9–12)
	{9, SubjectBIO, 5, false, "Comparison"},
	{10, SubjectBIO, 5, false, "Efficacy"},
	{11, SubjectBIO, 5, false, "AchievementExpectation"},
	{12, SubjectBIO, 1, true, "SkillMastery"},
	// GEO (13–16)
	{13, SubjectGEO, 3, false, "Comparison"},
	{14, SubjectGEO, 3, false, "Efficacy"},
	{15, SubjectGEO, 3, false, "AchievementExpectation"},
	{16, SubjectGEO, 3, true, "SkillMastery"},
	// HIS (17–20)
	{17, SubjectHIS, 2, false, "Comparison"},
	{18, SubjectHIS, 3, false, "Efficacy"},
	{19, SubjectHIS, 3, false, "AchievementExpectation"},
	{20, SubjectHIS, 3, true, "SkillMastery"},
	// POL (21–24)
	{21, SubjectPOL, 2, false, "Comparison"},
	{22, SubjectPOL, 3, false, "Efficacy"},
	{23, SubjectPOL, 3, false, "AchievementExpectation"},
	{24, SubjectPOL, 3, true, "SkillMastery"},
}

// ============ PHY_CHE_BIO：不匹配（Mismatch） ============
// 物理/化学/生物：低分（Comparison/Efficacy ~2；Achievement ~3；SkillMastery 4）
// 其他学科：中性 3，突出“不支持该理科组合”的对比效果
var ASC_Mismatch_PHY_CHE_BIO = []ASCAnswer{
	// PHY (1–4)
	{1, SubjectPHY, 2, false, "Comparison"},
	{2, SubjectPHY, 2, false, "Efficacy"},
	{3, SubjectPHY, 3, false, "AchievementExpectation"},
	{4, SubjectPHY, 4, true, "SkillMastery"},
	// CHE (5–8)
	{5, SubjectCHE, 2, false, "Comparison"},
	{6, SubjectCHE, 2, false, "Efficacy"},
	{7, SubjectCHE, 3, false, "AchievementExpectation"},
	{8, SubjectCHE, 4, true, "SkillMastery"},
	// BIO (9–12)
	{9, SubjectBIO, 2, false, "Comparison"},
	{10, SubjectBIO, 3, false, "Efficacy"},
	{11, SubjectBIO, 3, false, "AchievementExpectation"},
	{12, SubjectBIO, 4, true, "SkillMastery"},
	// GEO (13–16)
	{13, SubjectGEO, 3, false, "Comparison"},
	{14, SubjectGEO, 3, false, "Efficacy"},
	{15, SubjectGEO, 3, false, "AchievementExpectation"},
	{16, SubjectGEO, 3, true, "SkillMastery"},
	// HIS (17–20)
	{17, SubjectHIS, 3, false, "Comparison"},
	{18, SubjectHIS, 3, false, "Efficacy"},
	{19, SubjectHIS, 3, false, "AchievementExpectation"},
	{20, SubjectHIS, 3, true, "SkillMastery"},
	// POL (21–24)
	{21, SubjectPOL, 3, false, "Comparison"},
	{22, SubjectPOL, 3, false, "Efficacy"},
	{23, SubjectPOL, 3, false, "AchievementExpectation"},
	{24, SubjectPOL, 3, true, "SkillMastery"},
}

// ============ PHY_CHE_GEO：匹配（Aligned） ============
// 物理 / 化学 / 地理 为兴趣主科：高分（5,5,5,1）
// 其他（BIO / HIS / POL）中性（3,3,3,3）
var ASC_Aligned_PHY_CHE_GEO = []ASCAnswer{
	// PHY (1–4)
	{1, SubjectPHY, 5, false, "Comparison"},
	{2, SubjectPHY, 5, false, "Efficacy"},
	{3, SubjectPHY, 5, false, "AchievementExpectation"},
	{4, SubjectPHY, 1, true, "SkillMastery"},
	// CHE (5–8)
	{5, SubjectCHE, 5, false, "Comparison"},
	{6, SubjectCHE, 5, false, "Efficacy"},
	{7, SubjectCHE, 5, false, "AchievementExpectation"},
	{8, SubjectCHE, 1, true, "SkillMastery"},
	// BIO (9–12)
	{9, SubjectBIO, 3, false, "Comparison"},
	{10, SubjectBIO, 3, false, "Efficacy"},
	{11, SubjectBIO, 3, false, "AchievementExpectation"},
	{12, SubjectBIO, 3, true, "SkillMastery"},
	// GEO (13–16)
	{13, SubjectGEO, 5, false, "Comparison"},
	{14, SubjectGEO, 5, false, "Efficacy"},
	{15, SubjectGEO, 5, false, "AchievementExpectation"},
	{16, SubjectGEO, 1, true, "SkillMastery"},
	// HIS (17–20)
	{17, SubjectHIS, 3, false, "Comparison"},
	{18, SubjectHIS, 3, false, "Efficacy"},
	{19, SubjectHIS, 3, false, "AchievementExpectation"},
	{20, SubjectHIS, 3, true, "SkillMastery"},
	// POL (21–24)
	{21, SubjectPOL, 3, false, "Comparison"},
	{22, SubjectPOL, 3, false, "Efficacy"},
	{23, SubjectPOL, 3, false, "AchievementExpectation"},
	{24, SubjectPOL, 3, true, "SkillMastery"},
}

// ============ PHY_CHE_GEO：不匹配（Mismatch） ============
// 物理 / 化学 / 地理 为兴趣主科但能力低（2,2,3,4）
// 其他（BIO / HIS / POL）维持中性（3,3,3,3）
var ASC_Mismatch_PHY_CHE_GEO = []ASCAnswer{
	// PHY (1–4)
	{1, SubjectPHY, 2, false, "Comparison"},
	{2, SubjectPHY, 2, false, "Efficacy"},
	{3, SubjectPHY, 3, false, "AchievementExpectation"},
	{4, SubjectPHY, 4, true, "SkillMastery"},
	// CHE (5–8)
	{5, SubjectCHE, 2, false, "Comparison"},
	{6, SubjectCHE, 2, false, "Efficacy"},
	{7, SubjectCHE, 3, false, "AchievementExpectation"},
	{8, SubjectCHE, 4, true, "SkillMastery"},
	// BIO (9–12)
	{9, SubjectBIO, 3, false, "Comparison"},
	{10, SubjectBIO, 3, false, "Efficacy"},
	{11, SubjectBIO, 3, false, "AchievementExpectation"},
	{12, SubjectBIO, 3, true, "SkillMastery"},
	// GEO (13–16)
	{13, SubjectGEO, 2, false, "Comparison"},
	{14, SubjectGEO, 2, false, "Efficacy"},
	{15, SubjectGEO, 3, false, "AchievementExpectation"},
	{16, SubjectGEO, 4, true, "SkillMastery"},
	// HIS (17–20)
	{17, SubjectHIS, 3, false, "Comparison"},
	{18, SubjectHIS, 3, false, "Efficacy"},
	{19, SubjectHIS, 3, false, "AchievementExpectation"},
	{20, SubjectHIS, 3, true, "SkillMastery"},
	// POL (21–24)
	{21, SubjectPOL, 3, false, "Comparison"},
	{22, SubjectPOL, 3, false, "Efficacy"},
	{23, SubjectPOL, 3, false, "AchievementExpectation"},
	{24, SubjectPOL, 3, true, "SkillMastery"},
}

// ============ CHE_BIO_GEO：匹配（Aligned） ============
// 化学 / 生物 / 地理 为兴趣主科（5,5,5,1）
// 其他（PHY / HIS / POL）中性（3,3,3,3）
var ASC_Aligned_CHE_BIO_GEO = []ASCAnswer{
	// PHY (1–4)
	{1, SubjectPHY, 3, false, "Comparison"},
	{2, SubjectPHY, 3, false, "Efficacy"},
	{3, SubjectPHY, 3, false, "AchievementExpectation"},
	{4, SubjectPHY, 3, true, "SkillMastery"},
	// CHE (5–8)
	{5, SubjectCHE, 5, false, "Comparison"},
	{6, SubjectCHE, 5, false, "Efficacy"},
	{7, SubjectCHE, 5, false, "AchievementExpectation"},
	{8, SubjectCHE, 1, true, "SkillMastery"},
	// BIO (9–12)
	{9, SubjectBIO, 5, false, "Comparison"},
	{10, SubjectBIO, 5, false, "Efficacy"},
	{11, SubjectBIO, 5, false, "AchievementExpectation"},
	{12, SubjectBIO, 1, true, "SkillMastery"},
	// GEO (13–16)
	{13, SubjectGEO, 5, false, "Comparison"},
	{14, SubjectGEO, 5, false, "Efficacy"},
	{15, SubjectGEO, 5, false, "AchievementExpectation"},
	{16, SubjectGEO, 1, true, "SkillMastery"},
	// HIS (17–20)
	{17, SubjectHIS, 3, false, "Comparison"},
	{18, SubjectHIS, 3, false, "Efficacy"},
	{19, SubjectHIS, 3, false, "AchievementExpectation"},
	{20, SubjectHIS, 3, true, "SkillMastery"},
	// POL (21–24)
	{21, SubjectPOL, 3, false, "Comparison"},
	{22, SubjectPOL, 3, false, "Efficacy"},
	{23, SubjectPOL, 3, false, "AchievementExpectation"},
	{24, SubjectPOL, 3, true, "SkillMastery"},
}

// ============ CHE_BIO_GEO：不匹配（Mismatch） ============
// 化学 / 生物 / 地理 为兴趣主科但能力低（2,2,3,4）
// 其他（PHY / HIS / POL）维持中性（3,3,3,3）
var ASC_Mismatch_CHE_BIO_GEO = []ASCAnswer{
	// PHY (1–4)
	{1, SubjectPHY, 3, false, "Comparison"},
	{2, SubjectPHY, 3, false, "Efficacy"},
	{3, SubjectPHY, 3, false, "AchievementExpectation"},
	{4, SubjectPHY, 3, true, "SkillMastery"},
	// CHE (5–8)
	{5, SubjectCHE, 2, false, "Comparison"},
	{6, SubjectCHE, 2, false, "Efficacy"},
	{7, SubjectCHE, 3, false, "AchievementExpectation"},
	{8, SubjectCHE, 4, true, "SkillMastery"},
	// BIO (9–12)
	{9, SubjectBIO, 2, false, "Comparison"},
	{10, SubjectBIO, 2, false, "Efficacy"},
	{11, SubjectBIO, 3, false, "AchievementExpectation"},
	{12, SubjectBIO, 4, true, "SkillMastery"},
	// GEO (13–16)
	{13, SubjectGEO, 2, false, "Comparison"},
	{14, SubjectGEO, 2, false, "Efficacy"},
	{15, SubjectGEO, 3, false, "AchievementExpectation"},
	{16, SubjectGEO, 4, true, "SkillMastery"},
	// HIS (17–20)
	{17, SubjectHIS, 3, false, "Comparison"},
	{18, SubjectHIS, 3, false, "Efficacy"},
	{19, SubjectHIS, 3, false, "AchievementExpectation"},
	{20, SubjectHIS, 3, true, "SkillMastery"},
	// POL (21–24)
	{21, SubjectPOL, 3, false, "Comparison"},
	{22, SubjectPOL, 3, false, "Efficacy"},
	{23, SubjectPOL, 3, false, "AchievementExpectation"},
	{24, SubjectPOL, 3, true, "SkillMastery"},
}

// ============ PHY_BIO_GEO：匹配（Aligned） ============
// 物理 / 生物 / 地理 为兴趣主科（5,5,5,1）
// 其他（CHE / HIS / POL）中性（3,3,3,3）
var ASC_Aligned_PHY_BIO_GEO = []ASCAnswer{
	// PHY (1–4)
	{1, SubjectPHY, 5, false, "Comparison"},
	{2, SubjectPHY, 5, false, "Efficacy"},
	{3, SubjectPHY, 5, false, "AchievementExpectation"},
	{4, SubjectPHY, 1, true, "SkillMastery"},
	// CHE (5–8)
	{5, SubjectCHE, 3, false, "Comparison"},
	{6, SubjectCHE, 3, false, "Efficacy"},
	{7, SubjectCHE, 3, false, "AchievementExpectation"},
	{8, SubjectCHE, 3, true, "SkillMastery"},
	// BIO (9–12)
	{9, SubjectBIO, 5, false, "Comparison"},
	{10, SubjectBIO, 5, false, "Efficacy"},
	{11, SubjectBIO, 5, false, "AchievementExpectation"},
	{12, SubjectBIO, 1, true, "SkillMastery"},
	// GEO (13–16)
	{13, SubjectGEO, 5, false, "Comparison"},
	{14, SubjectGEO, 5, false, "Efficacy"},
	{15, SubjectGEO, 5, false, "AchievementExpectation"},
	{16, SubjectGEO, 1, true, "SkillMastery"},
	// HIS (17–20)
	{17, SubjectHIS, 3, false, "Comparison"},
	{18, SubjectHIS, 3, false, "Efficacy"},
	{19, SubjectHIS, 3, false, "AchievementExpectation"},
	{20, SubjectHIS, 3, true, "SkillMastery"},
	// POL (21–24)
	{21, SubjectPOL, 3, false, "Comparison"},
	{22, SubjectPOL, 3, false, "Efficacy"},
	{23, SubjectPOL, 3, false, "AchievementExpectation"},
	{24, SubjectPOL, 3, true, "SkillMastery"},
}

// ============ PHY_BIO_GEO：不匹配（Mismatch） ============
// 物理 / 生物 / 地理 为兴趣主科但能力低（2,2,3,4）
// 其他（CHE / HIS / POL）维持中性（3,3,3,3）
var ASC_Mismatch_PHY_BIO_GEO = []ASCAnswer{
	// PHY (1–4)
	{1, SubjectPHY, 2, false, "Comparison"},
	{2, SubjectPHY, 2, false, "Efficacy"},
	{3, SubjectPHY, 3, false, "AchievementExpectation"},
	{4, SubjectPHY, 4, true, "SkillMastery"},
	// CHE (5–8)
	{5, SubjectCHE, 3, false, "Comparison"},
	{6, SubjectCHE, 3, false, "Efficacy"},
	{7, SubjectCHE, 3, false, "AchievementExpectation"},
	{8, SubjectCHE, 3, true, "SkillMastery"},
	// BIO (9–12)
	{9, SubjectBIO, 2, false, "Comparison"},
	{10, SubjectBIO, 2, false, "Efficacy"},
	{11, SubjectBIO, 3, false, "AchievementExpectation"},
	{12, SubjectBIO, 4, true, "SkillMastery"},
	// GEO (13–16)
	{13, SubjectGEO, 2, false, "Comparison"},
	{14, SubjectGEO, 2, false, "Efficacy"},
	{15, SubjectGEO, 3, false, "AchievementExpectation"},
	{16, SubjectGEO, 4, true, "SkillMastery"},
	// HIS (17–20)
	{17, SubjectHIS, 3, false, "Comparison"},
	{18, SubjectHIS, 3, false, "Efficacy"},
	{19, SubjectHIS, 3, false, "AchievementExpectation"},
	{20, SubjectHIS, 3, true, "SkillMastery"},
	// POL (21–24)
	{21, SubjectPOL, 3, false, "Comparison"},
	{22, SubjectPOL, 3, false, "Efficacy"},
	{23, SubjectPOL, 3, false, "AchievementExpectation"},
	{24, SubjectPOL, 3, true, "SkillMastery"},
}

// ============ HIS_GEO_POL：匹配（Aligned） ============
// 历史 / 地理 / 政治 为兴趣主科（5,5,5,1）
// 其他（PHY / CHE / BIO）中性（3,3,3,3）
var ASC_Aligned_HIS_GEO_POL = []ASCAnswer{
	// PHY (1–4)
	{1, SubjectPHY, 3, false, "Comparison"},
	{2, SubjectPHY, 3, false, "Efficacy"},
	{3, SubjectPHY, 3, false, "AchievementExpectation"},
	{4, SubjectPHY, 3, true, "SkillMastery"},
	// CHE (5–8)
	{5, SubjectCHE, 3, false, "Comparison"},
	{6, SubjectCHE, 3, false, "Efficacy"},
	{7, SubjectCHE, 3, false, "AchievementExpectation"},
	{8, SubjectCHE, 3, true, "SkillMastery"},
	// BIO (9–12)
	{9, SubjectBIO, 3, false, "Comparison"},
	{10, SubjectBIO, 3, false, "Efficacy"},
	{11, SubjectBIO, 3, false, "AchievementExpectation"},
	{12, SubjectBIO, 3, true, "SkillMastery"},
	// GEO (13–16)
	{13, SubjectGEO, 5, false, "Comparison"},
	{14, SubjectGEO, 5, false, "Efficacy"},
	{15, SubjectGEO, 5, false, "AchievementExpectation"},
	{16, SubjectGEO, 1, true, "SkillMastery"},
	// HIS (17–20)
	{17, SubjectHIS, 5, false, "Comparison"},
	{18, SubjectHIS, 5, false, "Efficacy"},
	{19, SubjectHIS, 5, false, "AchievementExpectation"},
	{20, SubjectHIS, 1, true, "SkillMastery"},
	// POL (21–24)
	{21, SubjectPOL, 5, false, "Comparison"},
	{22, SubjectPOL, 5, false, "Efficacy"},
	{23, SubjectPOL, 5, false, "AchievementExpectation"},
	{24, SubjectPOL, 1, true, "SkillMastery"},
}

// ============ HIS_GEO_POL：不匹配（Mismatch） ============
// 历史 / 地理 / 政治 为兴趣主科但能力低（2,2,3,4）
// 其他（PHY / CHE / BIO）中性（3,3,3,3）
var ASC_Mismatch_HIS_GEO_POL = []ASCAnswer{
	// PHY (1–4)
	{1, SubjectPHY, 3, false, "Comparison"},
	{2, SubjectPHY, 3, false, "Efficacy"},
	{3, SubjectPHY, 3, false, "AchievementExpectation"},
	{4, SubjectPHY, 3, true, "SkillMastery"},
	// CHE (5–8)
	{5, SubjectCHE, 3, false, "Comparison"},
	{6, SubjectCHE, 3, false, "Efficacy"},
	{7, SubjectCHE, 3, false, "AchievementExpectation"},
	{8, SubjectCHE, 3, true, "SkillMastery"},
	// BIO (9–12)
	{9, SubjectBIO, 3, false, "Comparison"},
	{10, SubjectBIO, 3, false, "Efficacy"},
	{11, SubjectBIO, 3, false, "AchievementExpectation"},
	{12, SubjectBIO, 3, true, "SkillMastery"},
	// GEO (13–16)
	{13, SubjectGEO, 2, false, "Comparison"},
	{14, SubjectGEO, 2, false, "Efficacy"},
	{15, SubjectGEO, 3, false, "AchievementExpectation"},
	{16, SubjectGEO, 4, true, "SkillMastery"},
	// HIS (17–20)
	{17, SubjectHIS, 2, false, "Comparison"},
	{18, SubjectHIS, 2, false, "Efficacy"},
	{19, SubjectHIS, 3, false, "AchievementExpectation"},
	{20, SubjectHIS, 4, true, "SkillMastery"},
	// POL (21–24)
	{21, SubjectPOL, 2, false, "Comparison"},
	{22, SubjectPOL, 2, false, "Efficacy"},
	{23, SubjectPOL, 3, false, "AchievementExpectation"},
	{24, SubjectPOL, 4, true, "SkillMastery"},
}

// ============ HIS_GEO_BIO：匹配（Aligned） ============
// 历史 / 地理 / 生物 为兴趣主科（5,5,5,1）
// 其他（PHY / CHE / POL）中性（3,3,3,3）
var ASC_Aligned_HIS_GEO_BIO = []ASCAnswer{
	// PHY (1–4)
	{1, SubjectPHY, 3, false, "Comparison"},
	{2, SubjectPHY, 3, false, "Efficacy"},
	{3, SubjectPHY, 3, false, "AchievementExpectation"},
	{4, SubjectPHY, 3, true, "SkillMastery"},
	// CHE (5–8)
	{5, SubjectCHE, 3, false, "Comparison"},
	{6, SubjectCHE, 3, false, "Efficacy"},
	{7, SubjectCHE, 3, false, "AchievementExpectation"},
	{8, SubjectCHE, 3, true, "SkillMastery"},
	// BIO (9–12)
	{9, SubjectBIO, 5, false, "Comparison"},
	{10, SubjectBIO, 5, false, "Efficacy"},
	{11, SubjectBIO, 5, false, "AchievementExpectation"},
	{12, SubjectBIO, 1, true, "SkillMastery"},
	// GEO (13–16)
	{13, SubjectGEO, 5, false, "Comparison"},
	{14, SubjectGEO, 5, false, "Efficacy"},
	{15, SubjectGEO, 5, false, "AchievementExpectation"},
	{16, SubjectGEO, 1, true, "SkillMastery"},
	// HIS (17–20)
	{17, SubjectHIS, 5, false, "Comparison"},
	{18, SubjectHIS, 5, false, "Efficacy"},
	{19, SubjectHIS, 5, false, "AchievementExpectation"},
	{20, SubjectHIS, 1, true, "SkillMastery"},
	// POL (21–24)
	{21, SubjectPOL, 3, false, "Comparison"},
	{22, SubjectPOL, 3, false, "Efficacy"},
	{23, SubjectPOL, 3, false, "AchievementExpectation"},
	{24, SubjectPOL, 3, true, "SkillMastery"},
}

// ============ HIS_GEO_BIO：不匹配（Mismatch） ============
// 历史 / 地理 / 生物 为兴趣主科但能力低（2,2,3,4）
// 其他（PHY / CHE / POL）中性（3,3,3,3）
var ASC_Mismatch_HIS_GEO_BIO = []ASCAnswer{
	// PHY (1–4)
	{1, SubjectPHY, 3, false, "Comparison"},
	{2, SubjectPHY, 3, false, "Efficacy"},
	{3, SubjectPHY, 3, false, "AchievementExpectation"},
	{4, SubjectPHY, 3, true, "SkillMastery"},
	// CHE (5–8)
	{5, SubjectCHE, 3, false, "Comparison"},
	{6, SubjectCHE, 3, false, "Efficacy"},
	{7, SubjectCHE, 3, false, "AchievementExpectation"},
	{8, SubjectCHE, 3, true, "SkillMastery"},
	// BIO (9–12)
	{9, SubjectBIO, 2, false, "Comparison"},
	{10, SubjectBIO, 2, false, "Efficacy"},
	{11, SubjectBIO, 3, false, "AchievementExpectation"},
	{12, SubjectBIO, 4, true, "SkillMastery"},
	// GEO (13–16)
	{13, SubjectGEO, 2, false, "Comparison"},
	{14, SubjectGEO, 2, false, "Efficacy"},
	{15, SubjectGEO, 3, false, "AchievementExpectation"},
	{16, SubjectGEO, 4, true, "SkillMastery"},
	// HIS (17–20)
	{17, SubjectHIS, 2, false, "Comparison"},
	{18, SubjectHIS, 2, false, "Efficacy"},
	{19, SubjectHIS, 3, false, "AchievementExpectation"},
	{20, SubjectHIS, 4, true, "SkillMastery"},
	// POL (21–24)
	{21, SubjectPOL, 3, false, "Comparison"},
	{22, SubjectPOL, 3, false, "Efficacy"},
	{23, SubjectPOL, 3, false, "AchievementExpectation"},
	{24, SubjectPOL, 3, true, "SkillMastery"},
}

// ============ PHY_GEO_CHE：匹配（Aligned） ============
// 物理 / 地理 / 化学 为兴趣主科（5,5,5,1）
// 其他（BIO / HIS / POL）中性（3,3,3,3）
var ASC_Aligned_PHY_GEO_CHE = []ASCAnswer{
	// PHY (1–4)
	{1, SubjectPHY, 5, false, "Comparison"},
	{2, SubjectPHY, 5, false, "Efficacy"},
	{3, SubjectPHY, 5, false, "AchievementExpectation"},
	{4, SubjectPHY, 1, true, "SkillMastery"},
	// CHE (5–8)
	{5, SubjectCHE, 5, false, "Comparison"},
	{6, SubjectCHE, 5, false, "Efficacy"},
	{7, SubjectCHE, 5, false, "AchievementExpectation"},
	{8, SubjectCHE, 1, true, "SkillMastery"},
	// BIO (9–12)
	{9, SubjectBIO, 3, false, "Comparison"},
	{10, SubjectBIO, 3, false, "Efficacy"},
	{11, SubjectBIO, 3, false, "AchievementExpectation"},
	{12, SubjectBIO, 3, true, "SkillMastery"},
	// GEO (13–16)
	{13, SubjectGEO, 5, false, "Comparison"},
	{14, SubjectGEO, 5, false, "Efficacy"},
	{15, SubjectGEO, 5, false, "AchievementExpectation"},
	{16, SubjectGEO, 1, true, "SkillMastery"},
	// HIS (17–20)
	{17, SubjectHIS, 3, false, "Comparison"},
	{18, SubjectHIS, 3, false, "Efficacy"},
	{19, SubjectHIS, 3, false, "AchievementExpectation"},
	{20, SubjectHIS, 3, true, "SkillMastery"},
	// POL (21–24)
	{21, SubjectPOL, 3, false, "Comparison"},
	{22, SubjectPOL, 3, false, "Efficacy"},
	{23, SubjectPOL, 3, false, "AchievementExpectation"},
	{24, SubjectPOL, 3, true, "SkillMastery"},
}

// ============ PHY_GEO_CHE：不匹配（Mismatch） ============
// 物理 / 地理 / 化学 为兴趣主科但能力低（2,2,3,4）
// 其他（BIO / HIS / POL）维持中性（3,3,3,3）
var ASC_Mismatch_PHY_GEO_CHE = []ASCAnswer{
	// PHY (1–4)
	{1, SubjectPHY, 2, false, "Comparison"},
	{2, SubjectPHY, 2, false, "Efficacy"},
	{3, SubjectPHY, 3, false, "AchievementExpectation"},
	{4, SubjectPHY, 4, true, "SkillMastery"},
	// CHE (5–8)
	{5, SubjectCHE, 2, false, "Comparison"},
	{6, SubjectCHE, 2, false, "Efficacy"},
	{7, SubjectCHE, 3, false, "AchievementExpectation"},
	{8, SubjectCHE, 4, true, "SkillMastery"},
	// BIO (9–12)
	{9, SubjectBIO, 3, false, "Comparison"},
	{10, SubjectBIO, 3, false, "Efficacy"},
	{11, SubjectBIO, 3, false, "AchievementExpectation"},
	{12, SubjectBIO, 3, true, "SkillMastery"},
	// GEO (13–16)
	{13, SubjectGEO, 2, false, "Comparison"},
	{14, SubjectGEO, 2, false, "Efficacy"},
	{15, SubjectGEO, 3, false, "AchievementExpectation"},
	{16, SubjectGEO, 4, true, "SkillMastery"},
	// HIS (17–20)
	{17, SubjectHIS, 3, false, "Comparison"},
	{18, SubjectHIS, 3, false, "Efficacy"},
	{19, SubjectHIS, 3, false, "AchievementExpectation"},
	{20, SubjectHIS, 3, true, "SkillMastery"},
	// POL (21–24)
	{21, SubjectPOL, 3, false, "Comparison"},
	{22, SubjectPOL, 3, false, "Efficacy"},
	{23, SubjectPOL, 3, false, "AchievementExpectation"},
	{24, SubjectPOL, 3, true, "SkillMastery"},
}

// ============ HIS_POL_BIO：匹配（Aligned） ============
// 历史 / 政治 / 生物 为兴趣主科（5,5,5,1）
// 其他（PHY / CHE / GEO）中性（3,3,3,3）
var ASC_Aligned_HIS_POL_BIO = []ASCAnswer{
	// PHY (1–4)
	{1, SubjectPHY, 3, false, "Comparison"},
	{2, SubjectPHY, 3, false, "Efficacy"},
	{3, SubjectPHY, 3, false, "AchievementExpectation"},
	{4, SubjectPHY, 3, true, "SkillMastery"},
	// CHE (5–8)
	{5, SubjectCHE, 3, false, "Comparison"},
	{6, SubjectCHE, 3, false, "Efficacy"},
	{7, SubjectCHE, 3, false, "AchievementExpectation"},
	{8, SubjectCHE, 3, true, "SkillMastery"},
	// BIO (9–12)
	{9, SubjectBIO, 5, false, "Comparison"},
	{10, SubjectBIO, 5, false, "Efficacy"},
	{11, SubjectBIO, 5, false, "AchievementExpectation"},
	{12, SubjectBIO, 1, true, "SkillMastery"},
	// GEO (13–16)
	{13, SubjectGEO, 3, false, "Comparison"},
	{14, SubjectGEO, 3, false, "Efficacy"},
	{15, SubjectGEO, 3, false, "AchievementExpectation"},
	{16, SubjectGEO, 3, true, "SkillMastery"},
	// HIS (17–20)
	{17, SubjectHIS, 5, false, "Comparison"},
	{18, SubjectHIS, 5, false, "Efficacy"},
	{19, SubjectHIS, 5, false, "AchievementExpectation"},
	{20, SubjectHIS, 1, true, "SkillMastery"},
	// POL (21–24)
	{21, SubjectPOL, 5, false, "Comparison"},
	{22, SubjectPOL, 5, false, "Efficacy"},
	{23, SubjectPOL, 5, false, "AchievementExpectation"},
	{24, SubjectPOL, 1, true, "SkillMastery"},
}

// ============ HIS_POL_BIO：不匹配（Mismatch） ============
// 历史 / 政治 / 生物 为兴趣主科但能力低（2,2,3,4）
// 其他（PHY / CHE / GEO）中性（3,3,3,3）
var ASC_Mismatch_HIS_POL_BIO = []ASCAnswer{
	// PHY (1–4)
	{1, SubjectPHY, 3, false, "Comparison"},
	{2, SubjectPHY, 3, false, "Efficacy"},
	{3, SubjectPHY, 3, false, "AchievementExpectation"},
	{4, SubjectPHY, 3, true, "SkillMastery"},
	// CHE (5–8)
	{5, SubjectCHE, 3, false, "Comparison"},
	{6, SubjectCHE, 3, false, "Efficacy"},
	{7, SubjectCHE, 3, false, "AchievementExpectation"},
	{8, SubjectCHE, 3, true, "SkillMastery"},
	// BIO (9–12)
	{9, SubjectBIO, 2, false, "Comparison"},
	{10, SubjectBIO, 2, false, "Efficacy"},
	{11, SubjectBIO, 3, false, "AchievementExpectation"},
	{12, SubjectBIO, 4, true, "SkillMastery"},
	// GEO (13–16)
	{13, SubjectGEO, 3, false, "Comparison"},
	{14, SubjectGEO, 3, false, "Efficacy"},
	{15, SubjectGEO, 3, false, "AchievementExpectation"},
	{16, SubjectGEO, 3, true, "SkillMastery"},
	// HIS (17–20)
	{17, SubjectHIS, 2, false, "Comparison"},
	{18, SubjectHIS, 2, false, "Efficacy"},
	{19, SubjectHIS, 3, false, "AchievementExpectation"},
	{20, SubjectHIS, 4, true, "SkillMastery"},
	// POL (21–24)
	{21, SubjectPOL, 2, false, "Comparison"},
	{22, SubjectPOL, 2, false, "Efficacy"},
	{23, SubjectPOL, 3, false, "AchievementExpectation"},
	{24, SubjectPOL, 4, true, "SkillMastery"},
}

// =======================
// 统一映射表 AllASCCombos
// =======================

var AllASCCombos = map[string]map[string][]ASCAnswer{
	ComboPHY_CHE_BIO: {
		"aligned":  ASC_Aligned_PHY_CHE_BIO,
		"mismatch": ASC_Mismatch_PHY_CHE_BIO,
	},
	ComboPHY_CHE_GEO: {
		"aligned":  ASC_Aligned_PHY_CHE_GEO,
		"mismatch": ASC_Mismatch_PHY_CHE_GEO,
	},
	ComboCHE_BIO_GEO: {
		"aligned":  ASC_Aligned_CHE_BIO_GEO,
		"mismatch": ASC_Mismatch_CHE_BIO_GEO,
	},
	ComboPHY_BIO_GEO: {
		"aligned":  ASC_Aligned_PHY_BIO_GEO,
		"mismatch": ASC_Mismatch_PHY_BIO_GEO,
	},
	ComboHIS_GEO_POL: {
		"aligned":  ASC_Aligned_HIS_GEO_POL,
		"mismatch": ASC_Mismatch_HIS_GEO_POL,
	},
	ComboHIS_GEO_BIO: {
		"aligned":  ASC_Aligned_HIS_GEO_BIO,
		"mismatch": ASC_Mismatch_HIS_GEO_BIO,
	},
	ComboPHY_GEO_CHE: {
		"aligned":  ASC_Aligned_PHY_GEO_CHE,
		"mismatch": ASC_Mismatch_PHY_GEO_CHE,
	},
	ComboHIS_POL_BIO: {
		"aligned":  ASC_Aligned_HIS_POL_BIO,
		"mismatch": ASC_Mismatch_HIS_POL_BIO,
	},
}

func TestASCAnswer() {
	// 示例：获取 HIS_GEO_POL 的匹配型 ASC 答案
	combo := "HIS_GEO_POL"
	category := "aligned" // 可改为 "mismatch"

	if comboData, ok := AllASCCombos[combo]; ok {
		if answers, ok := comboData[category]; ok {
			data, _ := json.MarshalIndent(answers, "", "  ")
			fmt.Printf("组合 %s (%s)：\n%s\n", combo, category, string(data))
		} else {
			fmt.Printf("未找到组合 %s 的类别 %s。\n", combo, category)
		}
	} else {
		fmt.Printf("未找到组合 %s。\n", combo)
	}
}
