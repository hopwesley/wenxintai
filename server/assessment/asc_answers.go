package assessment

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

// AscAlignedPhyCheBio
// ============ PHY_CHE_BIO：匹配（Aligned） ============
// 物理/化学/生物：高分（Comparison/Efficacy/Achievement=5；SkillMastery=1）
// 其他学科：中性（GEO/HIS/POL 题设给 3；HIS/POL 的 Comparison 稍低 2 以拉开差距）
var AscAlignedPhyCheBio = []ASCAnswer{
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

// AscMismatchPhyCheBio
// ============ PHY_CHE_BIO：不匹配（Mismatch） ============
// 物理/化学/生物：低分（Comparison/Efficacy ~2；Achievement ~3；SkillMastery 4）
// 其他学科：中性 3，突出“不支持该理科组合”的对比效果
var AscMismatchPhyCheBio = []ASCAnswer{
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

// AscAlignedPhyCheGeo
// ============ PHY_CHE_GEO：匹配（Aligned） ============
// 物理 / 化学 / 地理 为兴趣主科：高分（5,5,5,1）
// 其他（BIO / HIS / POL）中性（3,3,3,3）
var AscAlignedPhyCheGeo = []ASCAnswer{
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

// AscMismatchPhyCheGeo
// ============ PHY_CHE_GEO：不匹配（Mismatch） ============
// 物理 / 化学 / 地理 为兴趣主科但能力低（2,2,3,4）
// 其他（BIO / HIS / POL）维持中性（3,3,3,3）
var AscMismatchPhyCheGeo = []ASCAnswer{
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

// AscAlignedCheBioGeo
// ============ CHE_BIO_GEO：匹配（Aligned） ============
// 化学 / 生物 / 地理 为兴趣主科（5,5,5,1）
// 其他（PHY / HIS / POL）中性（3,3,3,3）
var AscAlignedCheBioGeo = []ASCAnswer{
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

// AscMismatchCheBioGeo
// ============ CHE_BIO_GEO：不匹配（Mismatch） ============
// 化学 / 生物 / 地理 为兴趣主科但能力低（2,2,3,4）
// 其他（PHY / HIS / POL）维持中性（3,3,3,3）
var AscMismatchCheBioGeo = []ASCAnswer{
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

// AscAlignedPhyBioGeo
// ============ PHY_BIO_GEO：匹配（Aligned） ============
// 物理 / 生物 / 地理 为兴趣主科（5,5,5,1）
// 其他（CHE / HIS / POL）中性（3,3,3,3）
var AscAlignedPhyBioGeo = []ASCAnswer{
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

// AscMismatchPhyBioGeo
// ============ PHY_BIO_GEO：不匹配（Mismatch） ============
// 物理 / 生物 / 地理 为兴趣主科但能力低（2,2,3,4）
// 其他（CHE / HIS / POL）维持中性（3,3,3,3）
var AscMismatchPhyBioGeo = []ASCAnswer{
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

// AscAlignedHisGeoPol
// ============ HIS_GEO_POL：匹配（Aligned） ============
// 历史 / 地理 / 政治 为兴趣主科（5,5,5,1）
// 其他（PHY / CHE / BIO）中性（3,3,3,3）
var AscAlignedHisGeoPol = []ASCAnswer{
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

// AscMismatchHisGeoPol
// ============ HIS_GEO_POL：不匹配（Mismatch） ============
// 历史 / 地理 / 政治 为兴趣主科但能力低（2,2,3,4）
// 其他（PHY / CHE / BIO）中性（3,3,3,3）
var AscMismatchHisGeoPol = []ASCAnswer{
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

// AscAlignedHisGeoBio
// ============ HIS_GEO_BIO：匹配（Aligned） ============
// 历史 / 地理 / 生物 为兴趣主科（5,5,5,1）
// 其他（PHY / CHE / POL）中性（3,3,3,3）
var AscAlignedHisGeoBio = []ASCAnswer{
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

// AscMismatchHisGeoBio
// ============ HIS_GEO_BIO：不匹配（Mismatch） ============
// 历史 / 地理 / 生物 为兴趣主科但能力低（2,2,3,4）
// 其他（PHY / CHE / POL）中性（3,3,3,3）
var AscMismatchHisGeoBio = []ASCAnswer{
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

// AscAlignedHisPolBio
// ============ HIS_POL_BIO：匹配（Aligned） ============
// 历史 / 政治 / 生物 为兴趣主科（5,5,5,1）
// 其他（PHY / CHE / GEO）中性（3,3,3,3）
var AscAlignedHisPolBio = []ASCAnswer{
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

// AscMismatchHisPolBio
// ============ HIS_POL_BIO：不匹配（Mismatch） ============
// 历史 / 政治 / 生物 为兴趣主科但能力低（2,2,3,4）
// 其他（PHY / CHE / GEO）中性（3,3,3,3）
var AscMismatchHisPolBio = []ASCAnswer{
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

// AscAlignedPhyChePol
// ============ PHY_CHE_POL：匹配（Aligned） ============
// 物理 / 化学 / 政治 为兴趣主科（5,5,5,1；POL的Comparison为2）
// 其他（BIO / GEO / HIS）中性（3,3,3,3）
var AscAlignedPhyChePol = []ASCAnswer{
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
	{21, SubjectPOL, 2, false, "Comparison"}, // 参考理科组合，POL的Comparison稍低
	{22, SubjectPOL, 5, false, "Efficacy"},
	{23, SubjectPOL, 5, false, "AchievementExpectation"},
	{24, SubjectPOL, 1, true, "SkillMastery"},
}

// AscMismatchPhyChePol
// ============ PHY_CHE_POL：不匹配（Mismatch） ============
// 物理 / 化学 / 政治 为兴趣主科但能力低（2,2,3,4）
// 其他（BIO / GEO / HIS）中性（3,3,3,3）
var AscMismatchPhyChePol = []ASCAnswer{
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
	{21, SubjectPOL, 2, false, "Comparison"},
	{22, SubjectPOL, 2, false, "Efficacy"},
	{23, SubjectPOL, 3, false, "AchievementExpectation"},
	{24, SubjectPOL, 4, true, "SkillMastery"},
}

// AscAlignedPhyBioPol
// ============ PHY_BIO_POL：匹配（Aligned） ============
// 物理 / 生物 / 政治 为兴趣主科（5,5,5,1；POL的Comparison为2）
// 其他（CHE / GEO / HIS）中性（3,3,3,3）
var AscAlignedPhyBioPol = []ASCAnswer{
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
	{21, SubjectPOL, 2, false, "Comparison"},
	{22, SubjectPOL, 5, false, "Efficacy"},
	{23, SubjectPOL, 5, false, "AchievementExpectation"},
	{24, SubjectPOL, 1, true, "SkillMastery"},
}

// AscMismatchPhyBioPol
// ============ PHY_BIO_POL：不匹配（Mismatch） ============
// 物理 / 生物 / 政治 为兴趣主科但能力低（2,2,3,4）
// 其他（CHE / GEO / HIS）中性（3,3,3,3）
var AscMismatchPhyBioPol = []ASCAnswer{
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
	{21, SubjectPOL, 2, false, "Comparison"},
	{22, SubjectPOL, 2, false, "Efficacy"},
	{23, SubjectPOL, 3, false, "AchievementExpectation"},
	{24, SubjectPOL, 4, true, "SkillMastery"},
}

// AscAlignedPhyGeoPol
// ============ PHY_GEO_POL：匹配（Aligned） ============
// 物理 / 地理 / 政治 为兴趣主科（5,5,5,1；POL的Comparison为2）
// 其他（CHE / BIO / HIS）中性（3,3,3,3）
var AscAlignedPhyGeoPol = []ASCAnswer{
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
	{21, SubjectPOL, 2, false, "Comparison"},
	{22, SubjectPOL, 5, false, "Efficacy"},
	{23, SubjectPOL, 5, false, "AchievementExpectation"},
	{24, SubjectPOL, 1, true, "SkillMastery"},
}

// AscMismatchPhyGeoPol
// ============ PHY_GEO_POL：不匹配（Mismatch） ============
// 物理 / 地理 / 政治 为兴趣主科但能力低（2,2,3,4）
// 其他（CHE / BIO / HIS）中性（3,3,3,3）
var AscMismatchPhyGeoPol = []ASCAnswer{
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
	{21, SubjectPOL, 2, false, "Comparison"},
	{22, SubjectPOL, 2, false, "Efficacy"},
	{23, SubjectPOL, 3, false, "AchievementExpectation"},
	{24, SubjectPOL, 4, true, "SkillMastery"},
}

// AscAlignedHisCheBio
// ============ HIS_CHE_BIO：匹配（Aligned） ============
// 历史 / 化学 / 生物 为兴趣主科（5,5,5,1；HIS的Comparison为2）
// 其他（PHY / GEO / POL）中性（3,3,3,3）
var AscAlignedHisCheBio = []ASCAnswer{
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
	{13, SubjectGEO, 3, false, "Comparison"},
	{14, SubjectGEO, 3, false, "Efficacy"},
	{15, SubjectGEO, 3, false, "AchievementExpectation"},
	{16, SubjectGEO, 3, true, "SkillMastery"},
	// HIS (17–20)
	{17, SubjectHIS, 2, false, "Comparison"}, // 文科Comparison稍低
	{18, SubjectHIS, 5, false, "Efficacy"},
	{19, SubjectHIS, 5, false, "AchievementExpectation"},
	{20, SubjectHIS, 1, true, "SkillMastery"},
	// POL (21–24)
	{21, SubjectPOL, 3, false, "Comparison"},
	{22, SubjectPOL, 3, false, "Efficacy"},
	{23, SubjectPOL, 3, false, "AchievementExpectation"},
	{24, SubjectPOL, 3, true, "SkillMastery"},
}

// AscMismatchHisCheBio
// ============ HIS_CHE_BIO：不匹配（Mismatch） ============
// 历史 / 化学 / 生物 为兴趣主科但能力低（2,2,3,4）
// 其他（PHY / GEO / POL）中性（3,3,3,3）
var AscMismatchHisCheBio = []ASCAnswer{
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
	{21, SubjectPOL, 3, false, "Comparison"},
	{22, SubjectPOL, 3, false, "Efficacy"},
	{23, SubjectPOL, 3, false, "AchievementExpectation"},
	{24, SubjectPOL, 3, true, "SkillMastery"},
}

// AscAlignedHisChePol
// ============ HIS_CHE_POL：匹配（Aligned） ============
// 历史 / 化学 / 政治 为兴趣主科（5,5,5,1；HIS/POL的Comparison为2）
// 其他（PHY / BIO / GEO）中性（3,3,3,3）
var AscAlignedHisChePol = []ASCAnswer{
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
	{9, SubjectBIO, 3, false, "Comparison"},
	{10, SubjectBIO, 3, false, "Efficacy"},
	{11, SubjectBIO, 3, false, "AchievementExpectation"},
	{12, SubjectBIO, 3, true, "SkillMastery"},
	// GEO (13–16)
	{13, SubjectGEO, 3, false, "Comparison"},
	{14, SubjectGEO, 3, false, "Efficacy"},
	{15, SubjectGEO, 3, false, "AchievementExpectation"},
	{16, SubjectGEO, 3, true, "SkillMastery"},
	// HIS (17–20)
	{17, SubjectHIS, 2, false, "Comparison"},
	{18, SubjectHIS, 5, false, "Efficacy"},
	{19, SubjectHIS, 5, false, "AchievementExpectation"},
	{20, SubjectHIS, 1, true, "SkillMastery"},
	// POL (21–24)
	{21, SubjectPOL, 2, false, "Comparison"},
	{22, SubjectPOL, 5, false, "Efficacy"},
	{23, SubjectPOL, 5, false, "AchievementExpectation"},
	{24, SubjectPOL, 1, true, "SkillMastery"},
}

// AscMismatchHisChePol
// ============ HIS_CHE_POL：不匹配（Mismatch） ============
// 历史 / 化学 / 政治 为兴趣主科但能力低（2,2,3,4）
// 其他（PHY / BIO / GEO）中性（3,3,3,3）
var AscMismatchHisChePol = []ASCAnswer{
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
	{9, SubjectBIO, 3, false, "Comparison"},
	{10, SubjectBIO, 3, false, "Efficacy"},
	{11, SubjectBIO, 3, false, "AchievementExpectation"},
	{12, SubjectBIO, 3, true, "SkillMastery"},
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

// AllASCCombos
// =======================
// 统一映射表 AllASCCombos
// =======================
var AllASCCombos = map[string]map[string][]ASCAnswer{
	ComboPHY_CHE_BIO: {
		"aligned":  AscAlignedPhyCheBio,
		"mismatch": AscMismatchPhyCheBio,
	},
	ComboPHY_CHE_GEO: {
		"aligned":  AscAlignedPhyCheGeo,
		"mismatch": AscMismatchPhyCheGeo,
	},
	ComboCHE_BIO_GEO: {
		"aligned":  AscAlignedCheBioGeo,
		"mismatch": AscMismatchCheBioGeo,
	},
	ComboPHY_BIO_GEO: {
		"aligned":  AscAlignedPhyBioGeo,
		"mismatch": AscMismatchPhyBioGeo,
	},
	ComboHIS_GEO_POL: {
		"aligned":  AscAlignedHisGeoPol,
		"mismatch": AscMismatchHisGeoPol,
	},
	ComboHIS_GEO_BIO: {
		"aligned":  AscAlignedHisGeoBio,
		"mismatch": AscMismatchHisGeoBio,
	},
	ComboHIS_POL_BIO: {
		"aligned":  AscAlignedHisPolBio,
		"mismatch": AscMismatchHisPolBio,
	},
	ComboPHY_CHE_POL: {
		"aligned":  AscAlignedPhyChePol,
		"mismatch": AscMismatchPhyChePol,
	},
	ComboPHY_BIO_POL: {
		"aligned":  AscAlignedPhyBioPol,
		"mismatch": AscMismatchPhyBioPol,
	},
	ComboPHY_GEO_POL: {
		"aligned":  AscAlignedPhyGeoPol,
		"mismatch": AscMismatchPhyGeoPol,
	},
	ComboHIS_CHE_BIO: {
		"aligned":  AscAlignedHisCheBio,
		"mismatch": AscMismatchHisCheBio,
	},
	ComboHIS_CHE_POL: {
		"aligned":  AscAlignedHisChePol,
		"mismatch": AscMismatchHisChePol,
	},
}

func TestASCAnswer() {
	// 示例：获取 HIS_GEO_POL 的匹配型 ASC 答案
	combo := ComboHIS_GEO_POL
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
