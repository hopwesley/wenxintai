package main

import (
	"encoding/json"
	"fmt"

	"github.com/hopwesley/wenxintai/server/ai_api"
)

// AscAlignedPhyCheBio
// ============ PHY_CHE_BIO：匹配（Aligned） ============
// 物理/化学/生物：高分（Comparison/Efficacy/Achievement=5；SkillMastery=1）
// 其他学科：中性（GEO/HIS/POL 题设给 3；HIS/POL 的 Comparison 稍低 2 以拉开差距）
var AscAlignedPhyCheBio = []ai_api.ASCAnswer{
	// PHY (1–4)
	{1, ai_api.SubjectPHY, 5, false, "Comparison"},
	{2, ai_api.SubjectPHY, 5, false, "Efficacy"},
	{3, ai_api.SubjectPHY, 5, false, "AchievementExpectation"},
	{4, ai_api.SubjectPHY, 1, true, "SkillMastery"},
	// CHE (5–8)
	{5, ai_api.SubjectCHE, 5, false, "Comparison"},
	{6, ai_api.SubjectCHE, 5, false, "Efficacy"},
	{7, ai_api.SubjectCHE, 5, false, "AchievementExpectation"},
	{8, ai_api.SubjectCHE, 1, true, "SkillMastery"},
	// BIO (9–12)
	{9, ai_api.SubjectBIO, 5, false, "Comparison"},
	{10, ai_api.SubjectBIO, 5, false, "Efficacy"},
	{11, ai_api.SubjectBIO, 5, false, "AchievementExpectation"},
	{12, ai_api.SubjectBIO, 1, true, "SkillMastery"},
	// GEO (13–16)
	{13, ai_api.SubjectGEO, 3, false, "Comparison"},
	{14, ai_api.SubjectGEO, 3, false, "Efficacy"},
	{15, ai_api.SubjectGEO, 3, false, "AchievementExpectation"},
	{16, ai_api.SubjectGEO, 3, true, "SkillMastery"},
	// HIS (17–20)
	{17, ai_api.SubjectHIS, 2, false, "Comparison"},
	{18, ai_api.SubjectHIS, 3, false, "Efficacy"},
	{19, ai_api.SubjectHIS, 3, false, "AchievementExpectation"},
	{20, ai_api.SubjectHIS, 3, true, "SkillMastery"},
	// POL (21–24)
	{21, ai_api.SubjectPOL, 2, false, "Comparison"},
	{22, ai_api.SubjectPOL, 3, false, "Efficacy"},
	{23, ai_api.SubjectPOL, 3, false, "AchievementExpectation"},
	{24, ai_api.SubjectPOL, 3, true, "SkillMastery"},
}

// AscMismatchPhyCheBio
// ============ PHY_CHE_BIO：不匹配（Mismatch） ============
// 物理/化学/生物：低分（Comparison/Efficacy ~2；Achievement ~3；SkillMastery 4）
// 其他学科：中性 3，突出“不支持该理科组合”的对比效果
var AscMismatchPhyCheBio = []ai_api.ASCAnswer{
	// PHY (1–4)
	{1, ai_api.SubjectPHY, 2, false, "Comparison"},
	{2, ai_api.SubjectPHY, 2, false, "Efficacy"},
	{3, ai_api.SubjectPHY, 3, false, "AchievementExpectation"},
	{4, ai_api.SubjectPHY, 4, true, "SkillMastery"},
	// CHE (5–8)
	{5, ai_api.SubjectCHE, 2, false, "Comparison"},
	{6, ai_api.SubjectCHE, 2, false, "Efficacy"},
	{7, ai_api.SubjectCHE, 3, false, "AchievementExpectation"},
	{8, ai_api.SubjectCHE, 4, true, "SkillMastery"},
	// BIO (9–12)
	{9, ai_api.SubjectBIO, 2, false, "Comparison"},
	{10, ai_api.SubjectBIO, 3, false, "Efficacy"},
	{11, ai_api.SubjectBIO, 3, false, "AchievementExpectation"},
	{12, ai_api.SubjectBIO, 4, true, "SkillMastery"},
	// GEO (13–16)
	{13, ai_api.SubjectGEO, 3, false, "Comparison"},
	{14, ai_api.SubjectGEO, 3, false, "Efficacy"},
	{15, ai_api.SubjectGEO, 3, false, "AchievementExpectation"},
	{16, ai_api.SubjectGEO, 3, true, "SkillMastery"},
	// HIS (17–20)
	{17, ai_api.SubjectHIS, 3, false, "Comparison"},
	{18, ai_api.SubjectHIS, 3, false, "Efficacy"},
	{19, ai_api.SubjectHIS, 3, false, "AchievementExpectation"},
	{20, ai_api.SubjectHIS, 3, true, "SkillMastery"},
	// POL (21–24)
	{21, ai_api.SubjectPOL, 3, false, "Comparison"},
	{22, ai_api.SubjectPOL, 3, false, "Efficacy"},
	{23, ai_api.SubjectPOL, 3, false, "AchievementExpectation"},
	{24, ai_api.SubjectPOL, 3, true, "SkillMastery"},
}

// AscAlignedPhyCheGeo
// ============ PHY_CHE_GEO：匹配（Aligned） ============
// 物理 / 化学 / 地理 为兴趣主科：高分（5,5,5,1）
// 其他（BIO / HIS / POL）中性（3,3,3,3）
var AscAlignedPhyCheGeo = []ai_api.ASCAnswer{
	// PHY (1–4)
	{1, ai_api.SubjectPHY, 5, false, "Comparison"},
	{2, ai_api.SubjectPHY, 5, false, "Efficacy"},
	{3, ai_api.SubjectPHY, 5, false, "AchievementExpectation"},
	{4, ai_api.SubjectPHY, 1, true, "SkillMastery"},
	// CHE (5–8)
	{5, ai_api.SubjectCHE, 5, false, "Comparison"},
	{6, ai_api.SubjectCHE, 5, false, "Efficacy"},
	{7, ai_api.SubjectCHE, 5, false, "AchievementExpectation"},
	{8, ai_api.SubjectCHE, 1, true, "SkillMastery"},
	// BIO (9–12)
	{9, ai_api.SubjectBIO, 3, false, "Comparison"},
	{10, ai_api.SubjectBIO, 3, false, "Efficacy"},
	{11, ai_api.SubjectBIO, 3, false, "AchievementExpectation"},
	{12, ai_api.SubjectBIO, 3, true, "SkillMastery"},
	// GEO (13–16)
	{13, ai_api.SubjectGEO, 5, false, "Comparison"},
	{14, ai_api.SubjectGEO, 5, false, "Efficacy"},
	{15, ai_api.SubjectGEO, 5, false, "AchievementExpectation"},
	{16, ai_api.SubjectGEO, 1, true, "SkillMastery"},
	// HIS (17–20)
	{17, ai_api.SubjectHIS, 3, false, "Comparison"},
	{18, ai_api.SubjectHIS, 3, false, "Efficacy"},
	{19, ai_api.SubjectHIS, 3, false, "AchievementExpectation"},
	{20, ai_api.SubjectHIS, 3, true, "SkillMastery"},
	// POL (21–24)
	{21, ai_api.SubjectPOL, 3, false, "Comparison"},
	{22, ai_api.SubjectPOL, 3, false, "Efficacy"},
	{23, ai_api.SubjectPOL, 3, false, "AchievementExpectation"},
	{24, ai_api.SubjectPOL, 3, true, "SkillMastery"},
}

// AscMismatchPhyCheGeo
// ============ PHY_CHE_GEO：不匹配（Mismatch） ============
// 物理 / 化学 / 地理 为兴趣主科但能力低（2,2,3,4）
// 其他（BIO / HIS / POL）维持中性（3,3,3,3）
var AscMismatchPhyCheGeo = []ai_api.ASCAnswer{
	// PHY (1–4)
	{1, ai_api.SubjectPHY, 2, false, "Comparison"},
	{2, ai_api.SubjectPHY, 2, false, "Efficacy"},
	{3, ai_api.SubjectPHY, 3, false, "AchievementExpectation"},
	{4, ai_api.SubjectPHY, 4, true, "SkillMastery"},
	// CHE (5–8)
	{5, ai_api.SubjectCHE, 2, false, "Comparison"},
	{6, ai_api.SubjectCHE, 2, false, "Efficacy"},
	{7, ai_api.SubjectCHE, 3, false, "AchievementExpectation"},
	{8, ai_api.SubjectCHE, 4, true, "SkillMastery"},
	// BIO (9–12)
	{9, ai_api.SubjectBIO, 3, false, "Comparison"},
	{10, ai_api.SubjectBIO, 3, false, "Efficacy"},
	{11, ai_api.SubjectBIO, 3, false, "AchievementExpectation"},
	{12, ai_api.SubjectBIO, 3, true, "SkillMastery"},
	// GEO (13–16)
	{13, ai_api.SubjectGEO, 2, false, "Comparison"},
	{14, ai_api.SubjectGEO, 2, false, "Efficacy"},
	{15, ai_api.SubjectGEO, 3, false, "AchievementExpectation"},
	{16, ai_api.SubjectGEO, 4, true, "SkillMastery"},
	// HIS (17–20)
	{17, ai_api.SubjectHIS, 3, false, "Comparison"},
	{18, ai_api.SubjectHIS, 3, false, "Efficacy"},
	{19, ai_api.SubjectHIS, 3, false, "AchievementExpectation"},
	{20, ai_api.SubjectHIS, 3, true, "SkillMastery"},
	// POL (21–24)
	{21, ai_api.SubjectPOL, 3, false, "Comparison"},
	{22, ai_api.SubjectPOL, 3, false, "Efficacy"},
	{23, ai_api.SubjectPOL, 3, false, "AchievementExpectation"},
	{24, ai_api.SubjectPOL, 3, true, "SkillMastery"},
}

// AscAlignedCheBioGeo
// ============ CHE_BIO_GEO：匹配（Aligned） ============
// 化学 / 生物 / 地理 为兴趣主科（5,5,5,1）
// 其他（PHY / HIS / POL）中性（3,3,3,3）
var AscAlignedCheBioGeo = []ai_api.ASCAnswer{
	// PHY (1–4)
	{1, ai_api.SubjectPHY, 3, false, "Comparison"},
	{2, ai_api.SubjectPHY, 3, false, "Efficacy"},
	{3, ai_api.SubjectPHY, 3, false, "AchievementExpectation"},
	{4, ai_api.SubjectPHY, 3, true, "SkillMastery"},
	// CHE (5–8)
	{5, ai_api.SubjectCHE, 5, false, "Comparison"},
	{6, ai_api.SubjectCHE, 5, false, "Efficacy"},
	{7, ai_api.SubjectCHE, 5, false, "AchievementExpectation"},
	{8, ai_api.SubjectCHE, 1, true, "SkillMastery"},
	// BIO (9–12)
	{9, ai_api.SubjectBIO, 5, false, "Comparison"},
	{10, ai_api.SubjectBIO, 5, false, "Efficacy"},
	{11, ai_api.SubjectBIO, 5, false, "AchievementExpectation"},
	{12, ai_api.SubjectBIO, 1, true, "SkillMastery"},
	// GEO (13–16)
	{13, ai_api.SubjectGEO, 5, false, "Comparison"},
	{14, ai_api.SubjectGEO, 5, false, "Efficacy"},
	{15, ai_api.SubjectGEO, 5, false, "AchievementExpectation"},
	{16, ai_api.SubjectGEO, 1, true, "SkillMastery"},
	// HIS (17–20)
	{17, ai_api.SubjectHIS, 3, false, "Comparison"},
	{18, ai_api.SubjectHIS, 3, false, "Efficacy"},
	{19, ai_api.SubjectHIS, 3, false, "AchievementExpectation"},
	{20, ai_api.SubjectHIS, 3, true, "SkillMastery"},
	// POL (21–24)
	{21, ai_api.SubjectPOL, 3, false, "Comparison"},
	{22, ai_api.SubjectPOL, 3, false, "Efficacy"},
	{23, ai_api.SubjectPOL, 3, false, "AchievementExpectation"},
	{24, ai_api.SubjectPOL, 3, true, "SkillMastery"},
}

// AscMismatchCheBioGeo
// ============ CHE_BIO_GEO：不匹配（Mismatch） ============
// 化学 / 生物 / 地理 为兴趣主科但能力低（2,2,3,4）
// 其他（PHY / HIS / POL）维持中性（3,3,3,3）
var AscMismatchCheBioGeo = []ai_api.ASCAnswer{
	// PHY (1–4)
	{1, ai_api.SubjectPHY, 3, false, "Comparison"},
	{2, ai_api.SubjectPHY, 3, false, "Efficacy"},
	{3, ai_api.SubjectPHY, 3, false, "AchievementExpectation"},
	{4, ai_api.SubjectPHY, 3, true, "SkillMastery"},
	// CHE (5–8)
	{5, ai_api.SubjectCHE, 2, false, "Comparison"},
	{6, ai_api.SubjectCHE, 2, false, "Efficacy"},
	{7, ai_api.SubjectCHE, 3, false, "AchievementExpectation"},
	{8, ai_api.SubjectCHE, 4, true, "SkillMastery"},
	// BIO (9–12)
	{9, ai_api.SubjectBIO, 2, false, "Comparison"},
	{10, ai_api.SubjectBIO, 2, false, "Efficacy"},
	{11, ai_api.SubjectBIO, 3, false, "AchievementExpectation"},
	{12, ai_api.SubjectBIO, 4, true, "SkillMastery"},
	// GEO (13–16)
	{13, ai_api.SubjectGEO, 2, false, "Comparison"},
	{14, ai_api.SubjectGEO, 2, false, "Efficacy"},
	{15, ai_api.SubjectGEO, 3, false, "AchievementExpectation"},
	{16, ai_api.SubjectGEO, 4, true, "SkillMastery"},
	// HIS (17–20)
	{17, ai_api.SubjectHIS, 3, false, "Comparison"},
	{18, ai_api.SubjectHIS, 3, false, "Efficacy"},
	{19, ai_api.SubjectHIS, 3, false, "AchievementExpectation"},
	{20, ai_api.SubjectHIS, 3, true, "SkillMastery"},
	// POL (21–24)
	{21, ai_api.SubjectPOL, 3, false, "Comparison"},
	{22, ai_api.SubjectPOL, 3, false, "Efficacy"},
	{23, ai_api.SubjectPOL, 3, false, "AchievementExpectation"},
	{24, ai_api.SubjectPOL, 3, true, "SkillMastery"},
}

// AscAlignedPhyBioGeo
// ============ PHY_BIO_GEO：匹配（Aligned） ============
// 物理 / 生物 / 地理 为兴趣主科（5,5,5,1）
// 其他（CHE / HIS / POL）中性（3,3,3,3）
var AscAlignedPhyBioGeo = []ai_api.ASCAnswer{
	// PHY (1–4)
	{1, ai_api.SubjectPHY, 5, false, "Comparison"},
	{2, ai_api.SubjectPHY, 5, false, "Efficacy"},
	{3, ai_api.SubjectPHY, 5, false, "AchievementExpectation"},
	{4, ai_api.SubjectPHY, 1, true, "SkillMastery"},
	// CHE (5–8)
	{5, ai_api.SubjectCHE, 3, false, "Comparison"},
	{6, ai_api.SubjectCHE, 3, false, "Efficacy"},
	{7, ai_api.SubjectCHE, 3, false, "AchievementExpectation"},
	{8, ai_api.SubjectCHE, 3, true, "SkillMastery"},
	// BIO (9–12)
	{9, ai_api.SubjectBIO, 5, false, "Comparison"},
	{10, ai_api.SubjectBIO, 5, false, "Efficacy"},
	{11, ai_api.SubjectBIO, 5, false, "AchievementExpectation"},
	{12, ai_api.SubjectBIO, 1, true, "SkillMastery"},
	// GEO (13–16)
	{13, ai_api.SubjectGEO, 5, false, "Comparison"},
	{14, ai_api.SubjectGEO, 5, false, "Efficacy"},
	{15, ai_api.SubjectGEO, 5, false, "AchievementExpectation"},
	{16, ai_api.SubjectGEO, 1, true, "SkillMastery"},
	// HIS (17–20)
	{17, ai_api.SubjectHIS, 3, false, "Comparison"},
	{18, ai_api.SubjectHIS, 3, false, "Efficacy"},
	{19, ai_api.SubjectHIS, 3, false, "AchievementExpectation"},
	{20, ai_api.SubjectHIS, 3, true, "SkillMastery"},
	// POL (21–24)
	{21, ai_api.SubjectPOL, 3, false, "Comparison"},
	{22, ai_api.SubjectPOL, 3, false, "Efficacy"},
	{23, ai_api.SubjectPOL, 3, false, "AchievementExpectation"},
	{24, ai_api.SubjectPOL, 3, true, "SkillMastery"},
}

// AscMismatchPhyBioGeo
// ============ PHY_BIO_GEO：不匹配（Mismatch） ============
// 物理 / 生物 / 地理 为兴趣主科但能力低（2,2,3,4）
// 其他（CHE / HIS / POL）维持中性（3,3,3,3）
var AscMismatchPhyBioGeo = []ai_api.ASCAnswer{
	// PHY (1–4)
	{1, ai_api.SubjectPHY, 2, false, "Comparison"},
	{2, ai_api.SubjectPHY, 2, false, "Efficacy"},
	{3, ai_api.SubjectPHY, 3, false, "AchievementExpectation"},
	{4, ai_api.SubjectPHY, 4, true, "SkillMastery"},
	// CHE (5–8)
	{5, ai_api.SubjectCHE, 3, false, "Comparison"},
	{6, ai_api.SubjectCHE, 3, false, "Efficacy"},
	{7, ai_api.SubjectCHE, 3, false, "AchievementExpectation"},
	{8, ai_api.SubjectCHE, 3, true, "SkillMastery"},
	// BIO (9–12)
	{9, ai_api.SubjectBIO, 2, false, "Comparison"},
	{10, ai_api.SubjectBIO, 2, false, "Efficacy"},
	{11, ai_api.SubjectBIO, 3, false, "AchievementExpectation"},
	{12, ai_api.SubjectBIO, 4, true, "SkillMastery"},
	// GEO (13–16)
	{13, ai_api.SubjectGEO, 2, false, "Comparison"},
	{14, ai_api.SubjectGEO, 2, false, "Efficacy"},
	{15, ai_api.SubjectGEO, 3, false, "AchievementExpectation"},
	{16, ai_api.SubjectGEO, 4, true, "SkillMastery"},
	// HIS (17–20)
	{17, ai_api.SubjectHIS, 3, false, "Comparison"},
	{18, ai_api.SubjectHIS, 3, false, "Efficacy"},
	{19, ai_api.SubjectHIS, 3, false, "AchievementExpectation"},
	{20, ai_api.SubjectHIS, 3, true, "SkillMastery"},
	// POL (21–24)
	{21, ai_api.SubjectPOL, 3, false, "Comparison"},
	{22, ai_api.SubjectPOL, 3, false, "Efficacy"},
	{23, ai_api.SubjectPOL, 3, false, "AchievementExpectation"},
	{24, ai_api.SubjectPOL, 3, true, "SkillMastery"},
}

// AscAlignedHisGeoPol
// ============ HIS_GEO_POL：匹配（Aligned） ============
// 历史 / 地理 / 政治 为兴趣主科（5,5,5,1）
// 其他（PHY / CHE / BIO）中性（3,3,3,3）
var AscAlignedHisGeoPol = []ai_api.ASCAnswer{
	// PHY (1–4)
	{1, ai_api.SubjectPHY, 3, false, "Comparison"},
	{2, ai_api.SubjectPHY, 3, false, "Efficacy"},
	{3, ai_api.SubjectPHY, 3, false, "AchievementExpectation"},
	{4, ai_api.SubjectPHY, 3, true, "SkillMastery"},
	// CHE (5–8)
	{5, ai_api.SubjectCHE, 3, false, "Comparison"},
	{6, ai_api.SubjectCHE, 3, false, "Efficacy"},
	{7, ai_api.SubjectCHE, 3, false, "AchievementExpectation"},
	{8, ai_api.SubjectCHE, 3, true, "SkillMastery"},
	// BIO (9–12)
	{9, ai_api.SubjectBIO, 3, false, "Comparison"},
	{10, ai_api.SubjectBIO, 3, false, "Efficacy"},
	{11, ai_api.SubjectBIO, 3, false, "AchievementExpectation"},
	{12, ai_api.SubjectBIO, 3, true, "SkillMastery"},
	// GEO (13–16)
	{13, ai_api.SubjectGEO, 5, false, "Comparison"},
	{14, ai_api.SubjectGEO, 5, false, "Efficacy"},
	{15, ai_api.SubjectGEO, 5, false, "AchievementExpectation"},
	{16, ai_api.SubjectGEO, 1, true, "SkillMastery"},
	// HIS (17–20)
	{17, ai_api.SubjectHIS, 5, false, "Comparison"},
	{18, ai_api.SubjectHIS, 5, false, "Efficacy"},
	{19, ai_api.SubjectHIS, 5, false, "AchievementExpectation"},
	{20, ai_api.SubjectHIS, 1, true, "SkillMastery"},
	// POL (21–24)
	{21, ai_api.SubjectPOL, 5, false, "Comparison"},
	{22, ai_api.SubjectPOL, 5, false, "Efficacy"},
	{23, ai_api.SubjectPOL, 5, false, "AchievementExpectation"},
	{24, ai_api.SubjectPOL, 1, true, "SkillMastery"},
}

// AscMismatchHisGeoPol
// ============ HIS_GEO_POL：不匹配（Mismatch） ============
// 历史 / 地理 / 政治 为兴趣主科但能力低（2,2,3,4）
// 其他（PHY / CHE / BIO）中性（3,3,3,3）
var AscMismatchHisGeoPol = []ai_api.ASCAnswer{
	// PHY (1–4)
	{1, ai_api.SubjectPHY, 3, false, "Comparison"},
	{2, ai_api.SubjectPHY, 3, false, "Efficacy"},
	{3, ai_api.SubjectPHY, 3, false, "AchievementExpectation"},
	{4, ai_api.SubjectPHY, 3, true, "SkillMastery"},
	// CHE (5–8)
	{5, ai_api.SubjectCHE, 3, false, "Comparison"},
	{6, ai_api.SubjectCHE, 3, false, "Efficacy"},
	{7, ai_api.SubjectCHE, 3, false, "AchievementExpectation"},
	{8, ai_api.SubjectCHE, 3, true, "SkillMastery"},
	// BIO (9–12)
	{9, ai_api.SubjectBIO, 3, false, "Comparison"},
	{10, ai_api.SubjectBIO, 3, false, "Efficacy"},
	{11, ai_api.SubjectBIO, 3, false, "AchievementExpectation"},
	{12, ai_api.SubjectBIO, 3, true, "SkillMastery"},
	// GEO (13–16)
	{13, ai_api.SubjectGEO, 2, false, "Comparison"},
	{14, ai_api.SubjectGEO, 2, false, "Efficacy"},
	{15, ai_api.SubjectGEO, 3, false, "AchievementExpectation"},
	{16, ai_api.SubjectGEO, 4, true, "SkillMastery"},
	// HIS (17–20)
	{17, ai_api.SubjectHIS, 2, false, "Comparison"},
	{18, ai_api.SubjectHIS, 2, false, "Efficacy"},
	{19, ai_api.SubjectHIS, 3, false, "AchievementExpectation"},
	{20, ai_api.SubjectHIS, 4, true, "SkillMastery"},
	// POL (21–24)
	{21, ai_api.SubjectPOL, 2, false, "Comparison"},
	{22, ai_api.SubjectPOL, 2, false, "Efficacy"},
	{23, ai_api.SubjectPOL, 3, false, "AchievementExpectation"},
	{24, ai_api.SubjectPOL, 4, true, "SkillMastery"},
}

// AscAlignedHisGeoBio
// ============ HIS_GEO_BIO：匹配（Aligned） ============
// 历史 / 地理 / 生物 为兴趣主科（5,5,5,1）
// 其他（PHY / CHE / POL）中性（3,3,3,3）
var AscAlignedHisGeoBio = []ai_api.ASCAnswer{
	// PHY (1–4)
	{1, ai_api.SubjectPHY, 3, false, "Comparison"},
	{2, ai_api.SubjectPHY, 3, false, "Efficacy"},
	{3, ai_api.SubjectPHY, 3, false, "AchievementExpectation"},
	{4, ai_api.SubjectPHY, 3, true, "SkillMastery"},
	// CHE (5–8)
	{5, ai_api.SubjectCHE, 3, false, "Comparison"},
	{6, ai_api.SubjectCHE, 3, false, "Efficacy"},
	{7, ai_api.SubjectCHE, 3, false, "AchievementExpectation"},
	{8, ai_api.SubjectCHE, 3, true, "SkillMastery"},
	// BIO (9–12)
	{9, ai_api.SubjectBIO, 5, false, "Comparison"},
	{10, ai_api.SubjectBIO, 5, false, "Efficacy"},
	{11, ai_api.SubjectBIO, 5, false, "AchievementExpectation"},
	{12, ai_api.SubjectBIO, 1, true, "SkillMastery"},
	// GEO (13–16)
	{13, ai_api.SubjectGEO, 5, false, "Comparison"},
	{14, ai_api.SubjectGEO, 5, false, "Efficacy"},
	{15, ai_api.SubjectGEO, 5, false, "AchievementExpectation"},
	{16, ai_api.SubjectGEO, 1, true, "SkillMastery"},
	// HIS (17–20)
	{17, ai_api.SubjectHIS, 5, false, "Comparison"},
	{18, ai_api.SubjectHIS, 5, false, "Efficacy"},
	{19, ai_api.SubjectHIS, 5, false, "AchievementExpectation"},
	{20, ai_api.SubjectHIS, 1, true, "SkillMastery"},
	// POL (21–24)
	{21, ai_api.SubjectPOL, 3, false, "Comparison"},
	{22, ai_api.SubjectPOL, 3, false, "Efficacy"},
	{23, ai_api.SubjectPOL, 3, false, "AchievementExpectation"},
	{24, ai_api.SubjectPOL, 3, true, "SkillMastery"},
}

// AscMismatchHisGeoBio
// ============ HIS_GEO_BIO：不匹配（Mismatch） ============
// 历史 / 地理 / 生物 为兴趣主科但能力低（2,2,3,4）
// 其他（PHY / CHE / POL）中性（3,3,3,3）
var AscMismatchHisGeoBio = []ai_api.ASCAnswer{
	// PHY (1–4)
	{1, ai_api.SubjectPHY, 3, false, "Comparison"},
	{2, ai_api.SubjectPHY, 3, false, "Efficacy"},
	{3, ai_api.SubjectPHY, 3, false, "AchievementExpectation"},
	{4, ai_api.SubjectPHY, 3, true, "SkillMastery"},
	// CHE (5–8)
	{5, ai_api.SubjectCHE, 3, false, "Comparison"},
	{6, ai_api.SubjectCHE, 3, false, "Efficacy"},
	{7, ai_api.SubjectCHE, 3, false, "AchievementExpectation"},
	{8, ai_api.SubjectCHE, 3, true, "SkillMastery"},
	// BIO (9–12)
	{9, ai_api.SubjectBIO, 2, false, "Comparison"},
	{10, ai_api.SubjectBIO, 2, false, "Efficacy"},
	{11, ai_api.SubjectBIO, 3, false, "AchievementExpectation"},
	{12, ai_api.SubjectBIO, 4, true, "SkillMastery"},
	// GEO (13–16)
	{13, ai_api.SubjectGEO, 2, false, "Comparison"},
	{14, ai_api.SubjectGEO, 2, false, "Efficacy"},
	{15, ai_api.SubjectGEO, 3, false, "AchievementExpectation"},
	{16, ai_api.SubjectGEO, 4, true, "SkillMastery"},
	// HIS (17–20)
	{17, ai_api.SubjectHIS, 2, false, "Comparison"},
	{18, ai_api.SubjectHIS, 2, false, "Efficacy"},
	{19, ai_api.SubjectHIS, 3, false, "AchievementExpectation"},
	{20, ai_api.SubjectHIS, 4, true, "SkillMastery"},
	// POL (21–24)
	{21, ai_api.SubjectPOL, 3, false, "Comparison"},
	{22, ai_api.SubjectPOL, 3, false, "Efficacy"},
	{23, ai_api.SubjectPOL, 3, false, "AchievementExpectation"},
	{24, ai_api.SubjectPOL, 3, true, "SkillMastery"},
}

// AscAlignedHisPolBio
// ============ HIS_POL_BIO：匹配（Aligned） ============
// 历史 / 政治 / 生物 为兴趣主科（5,5,5,1）
// 其他（PHY / CHE / GEO）中性（3,3,3,3）
var AscAlignedHisPolBio = []ai_api.ASCAnswer{
	// PHY (1–4)
	{1, ai_api.SubjectPHY, 3, false, "Comparison"},
	{2, ai_api.SubjectPHY, 3, false, "Efficacy"},
	{3, ai_api.SubjectPHY, 3, false, "AchievementExpectation"},
	{4, ai_api.SubjectPHY, 3, true, "SkillMastery"},
	// CHE (5–8)
	{5, ai_api.SubjectCHE, 3, false, "Comparison"},
	{6, ai_api.SubjectCHE, 3, false, "Efficacy"},
	{7, ai_api.SubjectCHE, 3, false, "AchievementExpectation"},
	{8, ai_api.SubjectCHE, 3, true, "SkillMastery"},
	// BIO (9–12)
	{9, ai_api.SubjectBIO, 5, false, "Comparison"},
	{10, ai_api.SubjectBIO, 5, false, "Efficacy"},
	{11, ai_api.SubjectBIO, 5, false, "AchievementExpectation"},
	{12, ai_api.SubjectBIO, 1, true, "SkillMastery"},
	// GEO (13–16)
	{13, ai_api.SubjectGEO, 3, false, "Comparison"},
	{14, ai_api.SubjectGEO, 3, false, "Efficacy"},
	{15, ai_api.SubjectGEO, 3, false, "AchievementExpectation"},
	{16, ai_api.SubjectGEO, 3, true, "SkillMastery"},
	// HIS (17–20)
	{17, ai_api.SubjectHIS, 5, false, "Comparison"},
	{18, ai_api.SubjectHIS, 5, false, "Efficacy"},
	{19, ai_api.SubjectHIS, 5, false, "AchievementExpectation"},
	{20, ai_api.SubjectHIS, 1, true, "SkillMastery"},
	// POL (21–24)
	{21, ai_api.SubjectPOL, 5, false, "Comparison"},
	{22, ai_api.SubjectPOL, 5, false, "Efficacy"},
	{23, ai_api.SubjectPOL, 5, false, "AchievementExpectation"},
	{24, ai_api.SubjectPOL, 1, true, "SkillMastery"},
}

// AscMismatchHisPolBio
// ============ HIS_POL_BIO：不匹配（Mismatch） ============
// 历史 / 政治 / 生物 为兴趣主科但能力低（2,2,3,4）
// 其他（PHY / CHE / GEO）中性（3,3,3,3）
var AscMismatchHisPolBio = []ai_api.ASCAnswer{
	// PHY (1–4)
	{1, ai_api.SubjectPHY, 3, false, "Comparison"},
	{2, ai_api.SubjectPHY, 3, false, "Efficacy"},
	{3, ai_api.SubjectPHY, 3, false, "AchievementExpectation"},
	{4, ai_api.SubjectPHY, 3, true, "SkillMastery"},
	// CHE (5–8)
	{5, ai_api.SubjectCHE, 3, false, "Comparison"},
	{6, ai_api.SubjectCHE, 3, false, "Efficacy"},
	{7, ai_api.SubjectCHE, 3, false, "AchievementExpectation"},
	{8, ai_api.SubjectCHE, 3, true, "SkillMastery"},
	// BIO (9–12)
	{9, ai_api.SubjectBIO, 2, false, "Comparison"},
	{10, ai_api.SubjectBIO, 2, false, "Efficacy"},
	{11, ai_api.SubjectBIO, 3, false, "AchievementExpectation"},
	{12, ai_api.SubjectBIO, 4, true, "SkillMastery"},
	// GEO (13–16)
	{13, ai_api.SubjectGEO, 3, false, "Comparison"},
	{14, ai_api.SubjectGEO, 3, false, "Efficacy"},
	{15, ai_api.SubjectGEO, 3, false, "AchievementExpectation"},
	{16, ai_api.SubjectGEO, 3, true, "SkillMastery"},
	// HIS (17–20)
	{17, ai_api.SubjectHIS, 2, false, "Comparison"},
	{18, ai_api.SubjectHIS, 2, false, "Efficacy"},
	{19, ai_api.SubjectHIS, 3, false, "AchievementExpectation"},
	{20, ai_api.SubjectHIS, 4, true, "SkillMastery"},
	// POL (21–24)
	{21, ai_api.SubjectPOL, 2, false, "Comparison"},
	{22, ai_api.SubjectPOL, 2, false, "Efficacy"},
	{23, ai_api.SubjectPOL, 3, false, "AchievementExpectation"},
	{24, ai_api.SubjectPOL, 4, true, "SkillMastery"},
}

// AscAlignedPhyChePol
// ============ PHY_CHE_POL：匹配（Aligned） ============
// 物理 / 化学 / 政治 为兴趣主科（5,5,5,1；POL的Comparison为2）
// 其他（BIO / GEO / HIS）中性（3,3,3,3）
var AscAlignedPhyChePol = []ai_api.ASCAnswer{
	// PHY (1–4)
	{1, ai_api.SubjectPHY, 5, false, "Comparison"},
	{2, ai_api.SubjectPHY, 5, false, "Efficacy"},
	{3, ai_api.SubjectPHY, 5, false, "AchievementExpectation"},
	{4, ai_api.SubjectPHY, 1, true, "SkillMastery"},
	// CHE (5–8)
	{5, ai_api.SubjectCHE, 5, false, "Comparison"},
	{6, ai_api.SubjectCHE, 5, false, "Efficacy"},
	{7, ai_api.SubjectCHE, 5, false, "AchievementExpectation"},
	{8, ai_api.SubjectCHE, 1, true, "SkillMastery"},
	// BIO (9–12)
	{9, ai_api.SubjectBIO, 3, false, "Comparison"},
	{10, ai_api.SubjectBIO, 3, false, "Efficacy"},
	{11, ai_api.SubjectBIO, 3, false, "AchievementExpectation"},
	{12, ai_api.SubjectBIO, 3, true, "SkillMastery"},
	// GEO (13–16)
	{13, ai_api.SubjectGEO, 3, false, "Comparison"},
	{14, ai_api.SubjectGEO, 3, false, "Efficacy"},
	{15, ai_api.SubjectGEO, 3, false, "AchievementExpectation"},
	{16, ai_api.SubjectGEO, 3, true, "SkillMastery"},
	// HIS (17–20)
	{17, ai_api.SubjectHIS, 3, false, "Comparison"},
	{18, ai_api.SubjectHIS, 3, false, "Efficacy"},
	{19, ai_api.SubjectHIS, 3, false, "AchievementExpectation"},
	{20, ai_api.SubjectHIS, 3, true, "SkillMastery"},
	// POL (21–24)
	{21, ai_api.SubjectPOL, 2, false, "Comparison"}, // 参考理科组合，POL的Comparison稍低
	{22, ai_api.SubjectPOL, 5, false, "Efficacy"},
	{23, ai_api.SubjectPOL, 5, false, "AchievementExpectation"},
	{24, ai_api.SubjectPOL, 1, true, "SkillMastery"},
}

// AscMismatchPhyChePol
// ============ PHY_CHE_POL：不匹配（Mismatch） ============
// 物理 / 化学 / 政治 为兴趣主科但能力低（2,2,3,4）
// 其他（BIO / GEO / HIS）中性（3,3,3,3）
var AscMismatchPhyChePol = []ai_api.ASCAnswer{
	// PHY (1–4)
	{1, ai_api.SubjectPHY, 2, false, "Comparison"},
	{2, ai_api.SubjectPHY, 2, false, "Efficacy"},
	{3, ai_api.SubjectPHY, 3, false, "AchievementExpectation"},
	{4, ai_api.SubjectPHY, 4, true, "SkillMastery"},
	// CHE (5–8)
	{5, ai_api.SubjectCHE, 2, false, "Comparison"},
	{6, ai_api.SubjectCHE, 2, false, "Efficacy"},
	{7, ai_api.SubjectCHE, 3, false, "AchievementExpectation"},
	{8, ai_api.SubjectCHE, 4, true, "SkillMastery"},
	// BIO (9–12)
	{9, ai_api.SubjectBIO, 3, false, "Comparison"},
	{10, ai_api.SubjectBIO, 3, false, "Efficacy"},
	{11, ai_api.SubjectBIO, 3, false, "AchievementExpectation"},
	{12, ai_api.SubjectBIO, 3, true, "SkillMastery"},
	// GEO (13–16)
	{13, ai_api.SubjectGEO, 3, false, "Comparison"},
	{14, ai_api.SubjectGEO, 3, false, "Efficacy"},
	{15, ai_api.SubjectGEO, 3, false, "AchievementExpectation"},
	{16, ai_api.SubjectGEO, 3, true, "SkillMastery"},
	// HIS (17–20)
	{17, ai_api.SubjectHIS, 3, false, "Comparison"},
	{18, ai_api.SubjectHIS, 3, false, "Efficacy"},
	{19, ai_api.SubjectHIS, 3, false, "AchievementExpectation"},
	{20, ai_api.SubjectHIS, 3, true, "SkillMastery"},
	// POL (21–24)
	{21, ai_api.SubjectPOL, 2, false, "Comparison"},
	{22, ai_api.SubjectPOL, 2, false, "Efficacy"},
	{23, ai_api.SubjectPOL, 3, false, "AchievementExpectation"},
	{24, ai_api.SubjectPOL, 4, true, "SkillMastery"},
}

// AscAlignedPhyBioPol
// ============ PHY_BIO_POL：匹配（Aligned） ============
// 物理 / 生物 / 政治 为兴趣主科（5,5,5,1；POL的Comparison为2）
// 其他（CHE / GEO / HIS）中性（3,3,3,3）
var AscAlignedPhyBioPol = []ai_api.ASCAnswer{
	// PHY (1–4)
	{1, ai_api.SubjectPHY, 5, false, "Comparison"},
	{2, ai_api.SubjectPHY, 5, false, "Efficacy"},
	{3, ai_api.SubjectPHY, 5, false, "AchievementExpectation"},
	{4, ai_api.SubjectPHY, 1, true, "SkillMastery"},
	// CHE (5–8)
	{5, ai_api.SubjectCHE, 3, false, "Comparison"},
	{6, ai_api.SubjectCHE, 3, false, "Efficacy"},
	{7, ai_api.SubjectCHE, 3, false, "AchievementExpectation"},
	{8, ai_api.SubjectCHE, 3, true, "SkillMastery"},
	// BIO (9–12)
	{9, ai_api.SubjectBIO, 5, false, "Comparison"},
	{10, ai_api.SubjectBIO, 5, false, "Efficacy"},
	{11, ai_api.SubjectBIO, 5, false, "AchievementExpectation"},
	{12, ai_api.SubjectBIO, 1, true, "SkillMastery"},
	// GEO (13–16)
	{13, ai_api.SubjectGEO, 3, false, "Comparison"},
	{14, ai_api.SubjectGEO, 3, false, "Efficacy"},
	{15, ai_api.SubjectGEO, 3, false, "AchievementExpectation"},
	{16, ai_api.SubjectGEO, 3, true, "SkillMastery"},
	// HIS (17–20)
	{17, ai_api.SubjectHIS, 3, false, "Comparison"},
	{18, ai_api.SubjectHIS, 3, false, "Efficacy"},
	{19, ai_api.SubjectHIS, 3, false, "AchievementExpectation"},
	{20, ai_api.SubjectHIS, 3, true, "SkillMastery"},
	// POL (21–24)
	{21, ai_api.SubjectPOL, 2, false, "Comparison"},
	{22, ai_api.SubjectPOL, 5, false, "Efficacy"},
	{23, ai_api.SubjectPOL, 5, false, "AchievementExpectation"},
	{24, ai_api.SubjectPOL, 1, true, "SkillMastery"},
}

// AscMismatchPhyBioPol
// ============ PHY_BIO_POL：不匹配（Mismatch） ============
// 物理 / 生物 / 政治 为兴趣主科但能力低（2,2,3,4）
// 其他（CHE / GEO / HIS）中性（3,3,3,3）
var AscMismatchPhyBioPol = []ai_api.ASCAnswer{
	// PHY (1–4)
	{1, ai_api.SubjectPHY, 2, false, "Comparison"},
	{2, ai_api.SubjectPHY, 2, false, "Efficacy"},
	{3, ai_api.SubjectPHY, 3, false, "AchievementExpectation"},
	{4, ai_api.SubjectPHY, 4, true, "SkillMastery"},
	// CHE (5–8)
	{5, ai_api.SubjectCHE, 3, false, "Comparison"},
	{6, ai_api.SubjectCHE, 3, false, "Efficacy"},
	{7, ai_api.SubjectCHE, 3, false, "AchievementExpectation"},
	{8, ai_api.SubjectCHE, 3, true, "SkillMastery"},
	// BIO (9–12)
	{9, ai_api.SubjectBIO, 2, false, "Comparison"},
	{10, ai_api.SubjectBIO, 2, false, "Efficacy"},
	{11, ai_api.SubjectBIO, 3, false, "AchievementExpectation"},
	{12, ai_api.SubjectBIO, 4, true, "SkillMastery"},
	// GEO (13–16)
	{13, ai_api.SubjectGEO, 3, false, "Comparison"},
	{14, ai_api.SubjectGEO, 3, false, "Efficacy"},
	{15, ai_api.SubjectGEO, 3, false, "AchievementExpectation"},
	{16, ai_api.SubjectGEO, 3, true, "SkillMastery"},
	// HIS (17–20)
	{17, ai_api.SubjectHIS, 3, false, "Comparison"},
	{18, ai_api.SubjectHIS, 3, false, "Efficacy"},
	{19, ai_api.SubjectHIS, 3, false, "AchievementExpectation"},
	{20, ai_api.SubjectHIS, 3, true, "SkillMastery"},
	// POL (21–24)
	{21, ai_api.SubjectPOL, 2, false, "Comparison"},
	{22, ai_api.SubjectPOL, 2, false, "Efficacy"},
	{23, ai_api.SubjectPOL, 3, false, "AchievementExpectation"},
	{24, ai_api.SubjectPOL, 4, true, "SkillMastery"},
}

// AscAlignedPhyGeoPol
// ============ PHY_GEO_POL：匹配（Aligned） ============
// 物理 / 地理 / 政治 为兴趣主科（5,5,5,1；POL的Comparison为2）
// 其他（CHE / BIO / HIS）中性（3,3,3,3）
var AscAlignedPhyGeoPol = []ai_api.ASCAnswer{
	// PHY (1–4)
	{1, ai_api.SubjectPHY, 5, false, "Comparison"},
	{2, ai_api.SubjectPHY, 5, false, "Efficacy"},
	{3, ai_api.SubjectPHY, 5, false, "AchievementExpectation"},
	{4, ai_api.SubjectPHY, 1, true, "SkillMastery"},
	// CHE (5–8)
	{5, ai_api.SubjectCHE, 3, false, "Comparison"},
	{6, ai_api.SubjectCHE, 3, false, "Efficacy"},
	{7, ai_api.SubjectCHE, 3, false, "AchievementExpectation"},
	{8, ai_api.SubjectCHE, 3, true, "SkillMastery"},
	// BIO (9–12)
	{9, ai_api.SubjectBIO, 3, false, "Comparison"},
	{10, ai_api.SubjectBIO, 3, false, "Efficacy"},
	{11, ai_api.SubjectBIO, 3, false, "AchievementExpectation"},
	{12, ai_api.SubjectBIO, 3, true, "SkillMastery"},
	// GEO (13–16)
	{13, ai_api.SubjectGEO, 5, false, "Comparison"},
	{14, ai_api.SubjectGEO, 5, false, "Efficacy"},
	{15, ai_api.SubjectGEO, 5, false, "AchievementExpectation"},
	{16, ai_api.SubjectGEO, 1, true, "SkillMastery"},
	// HIS (17–20)
	{17, ai_api.SubjectHIS, 3, false, "Comparison"},
	{18, ai_api.SubjectHIS, 3, false, "Efficacy"},
	{19, ai_api.SubjectHIS, 3, false, "AchievementExpectation"},
	{20, ai_api.SubjectHIS, 3, true, "SkillMastery"},
	// POL (21–24)
	{21, ai_api.SubjectPOL, 2, false, "Comparison"},
	{22, ai_api.SubjectPOL, 5, false, "Efficacy"},
	{23, ai_api.SubjectPOL, 5, false, "AchievementExpectation"},
	{24, ai_api.SubjectPOL, 1, true, "SkillMastery"},
}

// AscMismatchPhyGeoPol
// ============ PHY_GEO_POL：不匹配（Mismatch） ============
// 物理 / 地理 / 政治 为兴趣主科但能力低（2,2,3,4）
// 其他（CHE / BIO / HIS）中性（3,3,3,3）
var AscMismatchPhyGeoPol = []ai_api.ASCAnswer{
	// PHY (1–4)
	{1, ai_api.SubjectPHY, 2, false, "Comparison"},
	{2, ai_api.SubjectPHY, 2, false, "Efficacy"},
	{3, ai_api.SubjectPHY, 3, false, "AchievementExpectation"},
	{4, ai_api.SubjectPHY, 4, true, "SkillMastery"},
	// CHE (5–8)
	{5, ai_api.SubjectCHE, 3, false, "Comparison"},
	{6, ai_api.SubjectCHE, 3, false, "Efficacy"},
	{7, ai_api.SubjectCHE, 3, false, "AchievementExpectation"},
	{8, ai_api.SubjectCHE, 3, true, "SkillMastery"},
	// BIO (9–12)
	{9, ai_api.SubjectBIO, 3, false, "Comparison"},
	{10, ai_api.SubjectBIO, 3, false, "Efficacy"},
	{11, ai_api.SubjectBIO, 3, false, "AchievementExpectation"},
	{12, ai_api.SubjectBIO, 3, true, "SkillMastery"},
	// GEO (13–16)
	{13, ai_api.SubjectGEO, 2, false, "Comparison"},
	{14, ai_api.SubjectGEO, 2, false, "Efficacy"},
	{15, ai_api.SubjectGEO, 3, false, "AchievementExpectation"},
	{16, ai_api.SubjectGEO, 4, true, "SkillMastery"},
	// HIS (17–20)
	{17, ai_api.SubjectHIS, 3, false, "Comparison"},
	{18, ai_api.SubjectHIS, 3, false, "Efficacy"},
	{19, ai_api.SubjectHIS, 3, false, "AchievementExpectation"},
	{20, ai_api.SubjectHIS, 3, true, "SkillMastery"},
	// POL (21–24)
	{21, ai_api.SubjectPOL, 2, false, "Comparison"},
	{22, ai_api.SubjectPOL, 2, false, "Efficacy"},
	{23, ai_api.SubjectPOL, 3, false, "AchievementExpectation"},
	{24, ai_api.SubjectPOL, 4, true, "SkillMastery"},
}

// AscAlignedHisCheBio
// ============ HIS_CHE_BIO：匹配（Aligned） ============
// 历史 / 化学 / 生物 为兴趣主科（5,5,5,1；HIS的Comparison为2）
// 其他（PHY / GEO / POL）中性（3,3,3,3）
var AscAlignedHisCheBio = []ai_api.ASCAnswer{
	// PHY (1–4)
	{1, ai_api.SubjectPHY, 3, false, "Comparison"},
	{2, ai_api.SubjectPHY, 3, false, "Efficacy"},
	{3, ai_api.SubjectPHY, 3, false, "AchievementExpectation"},
	{4, ai_api.SubjectPHY, 3, true, "SkillMastery"},
	// CHE (5–8)
	{5, ai_api.SubjectCHE, 5, false, "Comparison"},
	{6, ai_api.SubjectCHE, 5, false, "Efficacy"},
	{7, ai_api.SubjectCHE, 5, false, "AchievementExpectation"},
	{8, ai_api.SubjectCHE, 1, true, "SkillMastery"},
	// BIO (9–12)
	{9, ai_api.SubjectBIO, 5, false, "Comparison"},
	{10, ai_api.SubjectBIO, 5, false, "Efficacy"},
	{11, ai_api.SubjectBIO, 5, false, "AchievementExpectation"},
	{12, ai_api.SubjectBIO, 1, true, "SkillMastery"},
	// GEO (13–16)
	{13, ai_api.SubjectGEO, 3, false, "Comparison"},
	{14, ai_api.SubjectGEO, 3, false, "Efficacy"},
	{15, ai_api.SubjectGEO, 3, false, "AchievementExpectation"},
	{16, ai_api.SubjectGEO, 3, true, "SkillMastery"},
	// HIS (17–20)
	{17, ai_api.SubjectHIS, 2, false, "Comparison"}, // 文科Comparison稍低
	{18, ai_api.SubjectHIS, 5, false, "Efficacy"},
	{19, ai_api.SubjectHIS, 5, false, "AchievementExpectation"},
	{20, ai_api.SubjectHIS, 1, true, "SkillMastery"},
	// POL (21–24)
	{21, ai_api.SubjectPOL, 3, false, "Comparison"},
	{22, ai_api.SubjectPOL, 3, false, "Efficacy"},
	{23, ai_api.SubjectPOL, 3, false, "AchievementExpectation"},
	{24, ai_api.SubjectPOL, 3, true, "SkillMastery"},
}

// AscMismatchHisCheBio
// ============ HIS_CHE_BIO：不匹配（Mismatch） ============
// 历史 / 化学 / 生物 为兴趣主科但能力低（2,2,3,4）
// 其他（PHY / GEO / POL）中性（3,3,3,3）
var AscMismatchHisCheBio = []ai_api.ASCAnswer{
	// PHY (1–4)
	{1, ai_api.SubjectPHY, 3, false, "Comparison"},
	{2, ai_api.SubjectPHY, 3, false, "Efficacy"},
	{3, ai_api.SubjectPHY, 3, false, "AchievementExpectation"},
	{4, ai_api.SubjectPHY, 3, true, "SkillMastery"},
	// CHE (5–8)
	{5, ai_api.SubjectCHE, 2, false, "Comparison"},
	{6, ai_api.SubjectCHE, 2, false, "Efficacy"},
	{7, ai_api.SubjectCHE, 3, false, "AchievementExpectation"},
	{8, ai_api.SubjectCHE, 4, true, "SkillMastery"},
	// BIO (9–12)
	{9, ai_api.SubjectBIO, 2, false, "Comparison"},
	{10, ai_api.SubjectBIO, 2, false, "Efficacy"},
	{11, ai_api.SubjectBIO, 3, false, "AchievementExpectation"},
	{12, ai_api.SubjectBIO, 4, true, "SkillMastery"},
	// GEO (13–16)
	{13, ai_api.SubjectGEO, 3, false, "Comparison"},
	{14, ai_api.SubjectGEO, 3, false, "Efficacy"},
	{15, ai_api.SubjectGEO, 3, false, "AchievementExpectation"},
	{16, ai_api.SubjectGEO, 3, true, "SkillMastery"},
	// HIS (17–20)
	{17, ai_api.SubjectHIS, 2, false, "Comparison"},
	{18, ai_api.SubjectHIS, 2, false, "Efficacy"},
	{19, ai_api.SubjectHIS, 3, false, "AchievementExpectation"},
	{20, ai_api.SubjectHIS, 4, true, "SkillMastery"},
	// POL (21–24)
	{21, ai_api.SubjectPOL, 3, false, "Comparison"},
	{22, ai_api.SubjectPOL, 3, false, "Efficacy"},
	{23, ai_api.SubjectPOL, 3, false, "AchievementExpectation"},
	{24, ai_api.SubjectPOL, 3, true, "SkillMastery"},
}

// AscAlignedHisChePol
// ============ HIS_CHE_POL：匹配（Aligned） ============
// 历史 / 化学 / 政治 为兴趣主科（5,5,5,1；HIS/POL的Comparison为2）
// 其他（PHY / BIO / GEO）中性（3,3,3,3）
var AscAlignedHisChePol = []ai_api.ASCAnswer{
	// PHY (1–4)
	{1, ai_api.SubjectPHY, 3, false, "Comparison"},
	{2, ai_api.SubjectPHY, 3, false, "Efficacy"},
	{3, ai_api.SubjectPHY, 3, false, "AchievementExpectation"},
	{4, ai_api.SubjectPHY, 3, true, "SkillMastery"},
	// CHE (5–8)
	{5, ai_api.SubjectCHE, 5, false, "Comparison"},
	{6, ai_api.SubjectCHE, 5, false, "Efficacy"},
	{7, ai_api.SubjectCHE, 5, false, "AchievementExpectation"},
	{8, ai_api.SubjectCHE, 1, true, "SkillMastery"},
	// BIO (9–12)
	{9, ai_api.SubjectBIO, 3, false, "Comparison"},
	{10, ai_api.SubjectBIO, 3, false, "Efficacy"},
	{11, ai_api.SubjectBIO, 3, false, "AchievementExpectation"},
	{12, ai_api.SubjectBIO, 3, true, "SkillMastery"},
	// GEO (13–16)
	{13, ai_api.SubjectGEO, 3, false, "Comparison"},
	{14, ai_api.SubjectGEO, 3, false, "Efficacy"},
	{15, ai_api.SubjectGEO, 3, false, "AchievementExpectation"},
	{16, ai_api.SubjectGEO, 3, true, "SkillMastery"},
	// HIS (17–20)
	{17, ai_api.SubjectHIS, 2, false, "Comparison"},
	{18, ai_api.SubjectHIS, 5, false, "Efficacy"},
	{19, ai_api.SubjectHIS, 5, false, "AchievementExpectation"},
	{20, ai_api.SubjectHIS, 1, true, "SkillMastery"},
	// POL (21–24)
	{21, ai_api.SubjectPOL, 2, false, "Comparison"},
	{22, ai_api.SubjectPOL, 5, false, "Efficacy"},
	{23, ai_api.SubjectPOL, 5, false, "AchievementExpectation"},
	{24, ai_api.SubjectPOL, 1, true, "SkillMastery"},
}

// AscMismatchHisChePol
// ============ HIS_CHE_POL：不匹配（Mismatch） ============
// 历史 / 化学 / 政治 为兴趣主科但能力低（2,2,3,4）
// 其他（PHY / BIO / GEO）中性（3,3,3,3）
var AscMismatchHisChePol = []ai_api.ASCAnswer{
	// PHY (1–4)
	{1, ai_api.SubjectPHY, 3, false, "Comparison"},
	{2, ai_api.SubjectPHY, 3, false, "Efficacy"},
	{3, ai_api.SubjectPHY, 3, false, "AchievementExpectation"},
	{4, ai_api.SubjectPHY, 3, true, "SkillMastery"},
	// CHE (5–8)
	{5, ai_api.SubjectCHE, 2, false, "Comparison"},
	{6, ai_api.SubjectCHE, 2, false, "Efficacy"},
	{7, ai_api.SubjectCHE, 3, false, "AchievementExpectation"},
	{8, ai_api.SubjectCHE, 4, true, "SkillMastery"},
	// BIO (9–12)
	{9, ai_api.SubjectBIO, 3, false, "Comparison"},
	{10, ai_api.SubjectBIO, 3, false, "Efficacy"},
	{11, ai_api.SubjectBIO, 3, false, "AchievementExpectation"},
	{12, ai_api.SubjectBIO, 3, true, "SkillMastery"},
	// GEO (13–16)
	{13, ai_api.SubjectGEO, 3, false, "Comparison"},
	{14, ai_api.SubjectGEO, 3, false, "Efficacy"},
	{15, ai_api.SubjectGEO, 3, false, "AchievementExpectation"},
	{16, ai_api.SubjectGEO, 3, true, "SkillMastery"},
	// HIS (17–20)
	{17, ai_api.SubjectHIS, 2, false, "Comparison"},
	{18, ai_api.SubjectHIS, 2, false, "Efficacy"},
	{19, ai_api.SubjectHIS, 3, false, "AchievementExpectation"},
	{20, ai_api.SubjectHIS, 4, true, "SkillMastery"},
	// POL (21–24)
	{21, ai_api.SubjectPOL, 2, false, "Comparison"},
	{22, ai_api.SubjectPOL, 2, false, "Efficacy"},
	{23, ai_api.SubjectPOL, 3, false, "AchievementExpectation"},
	{24, ai_api.SubjectPOL, 4, true, "SkillMastery"},
}

// AllASCCombos
// =======================
// 统一映射表 AllASCCombos
// =======================
var AllASCCombos = map[string]map[string][]ai_api.ASCAnswer{
	ai_api.ComboPHY_CHE_BIO: {
		"aligned":  AscAlignedPhyCheBio,
		"mismatch": AscMismatchPhyCheBio,
	},
	ai_api.ComboPHY_CHE_GEO: {
		"aligned":  AscAlignedPhyCheGeo,
		"mismatch": AscMismatchPhyCheGeo,
	},
	ai_api.ComboCHE_BIO_GEO: {
		"aligned":  AscAlignedCheBioGeo,
		"mismatch": AscMismatchCheBioGeo,
	},
	ai_api.ComboPHY_BIO_GEO: {
		"aligned":  AscAlignedPhyBioGeo,
		"mismatch": AscMismatchPhyBioGeo,
	},
	ai_api.ComboHIS_GEO_POL: {
		"aligned":  AscAlignedHisGeoPol,
		"mismatch": AscMismatchHisGeoPol,
	},
	ai_api.ComboHIS_GEO_BIO: {
		"aligned":  AscAlignedHisGeoBio,
		"mismatch": AscMismatchHisGeoBio,
	},
	ai_api.ComboHIS_POL_BIO: {
		"aligned":  AscAlignedHisPolBio,
		"mismatch": AscMismatchHisPolBio,
	},
	ai_api.ComboPHY_CHE_POL: {
		"aligned":  AscAlignedPhyChePol,
		"mismatch": AscMismatchPhyChePol,
	},
	ai_api.ComboPHY_BIO_POL: {
		"aligned":  AscAlignedPhyBioPol,
		"mismatch": AscMismatchPhyBioPol,
	},
	ai_api.ComboPHY_GEO_POL: {
		"aligned":  AscAlignedPhyGeoPol,
		"mismatch": AscMismatchPhyGeoPol,
	},
	ai_api.ComboHIS_CHE_BIO: {
		"aligned":  AscAlignedHisCheBio,
		"mismatch": AscMismatchHisCheBio,
	},
	ai_api.ComboHIS_CHE_POL: {
		"aligned":  AscAlignedHisChePol,
		"mismatch": AscMismatchHisChePol,
	},
}

func TestASCAnswer() {
	// 示例：获取 HIS_GEO_POL 的匹配型 ASC 答案
	combo := ai_api.ComboHIS_GEO_POL
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
