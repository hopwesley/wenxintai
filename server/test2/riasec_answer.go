package main

import (
	"encoding/json"
	"fmt"
)

// RIASECAnswer
// ========================
// RIASEC Answer Structure
// ========================
type RIASECAnswer struct {
	ID        int    `json:"id"`
	Dimension string `json:"dimension"`
	Score     int    `json:"score"`
}

// ========================
// 固定题号与维度分布
// R:1–5, I:6–10, A:11–15, S:16–20, E:21–25, C:26–30
// ========================

// RiasecPhyCheBio
// ----------  理科核心：物理+化学+生物 ----------
var RiasecPhyCheBio = []RIASECAnswer{
	// R
	{1, "R", 5}, {2, "R", 5}, {3, "R", 4}, {4, "R", 5}, {5, "R", 4},
	// I
	{6, "I", 5}, {7, "I", 5}, {8, "I", 5}, {9, "I", 4}, {10, "I", 5},
	// A
	{11, "A", 2}, {12, "A", 3}, {13, "A", 2}, {14, "A", 2}, {15, "A", 3},
	// S
	{16, "S", 3}, {17, "S", 3}, {18, "S", 3}, {19, "S", 4}, {20, "S", 3},
	// E
	{21, "E", 2}, {22, "E", 3}, {23, "E", 2}, {24, "E", 2}, {25, "E", 3},
	// C
	{26, "C", 4}, {27, "C", 4}, {28, "C", 5}, {29, "C", 4}, {30, "C", 4},
}

// RiasecPhyCheGeo
// ---------- 理科应用型：物理+化学+地理 ----------
var RiasecPhyCheGeo = []RIASECAnswer{
	{1, "R", 5}, {2, "R", 5}, {3, "R", 5}, {4, "R", 4}, {5, "R", 5},
	{6, "I", 4}, {7, "I", 4}, {8, "I", 5}, {9, "I", 4}, {10, "I", 4},
	{11, "A", 2}, {12, "A", 3}, {13, "A", 2}, {14, "A", 2}, {15, "A", 3},
	{16, "S", 3}, {17, "S", 3}, {18, "S", 3}, {19, "S", 3}, {20, "S", 3},
	{21, "E", 3}, {22, "E", 3}, {23, "E", 2}, {24, "E", 3}, {25, "E", 3},
	{26, "C", 5}, {27, "C", 4}, {28, "C", 5}, {29, "C", 4}, {30, "C", 5},
}

// RiasecCheBioGeo
// ---------- 自然科学型：化学+生物+地理 ----------
var RiasecCheBioGeo = []RIASECAnswer{
	{1, "R", 3}, {2, "R", 4}, {3, "R", 3}, {4, "R", 3}, {5, "R", 4},
	{6, "I", 5}, {7, "I", 5}, {8, "I", 4}, {9, "I", 5}, {10, "I", 4},
	{11, "A", 2}, {12, "A", 2}, {13, "A", 3}, {14, "A", 2}, {15, "A", 3},
	{16, "S", 3}, {17, "S", 3}, {18, "S", 4}, {19, "S", 3}, {20, "S", 3},
	{21, "E", 2}, {22, "E", 2}, {23, "E", 2}, {24, "E", 3}, {25, "E", 2},
	{26, "C", 4}, {27, "C", 4}, {28, "C", 4}, {29, "C", 5}, {30, "C", 4},
}

// RiasecPhyBioGeo
// ---------- 理科探究型：物理+生物+地理 ----------
var RiasecPhyBioGeo = []RIASECAnswer{
	{1, "R", 4}, {2, "R", 5}, {3, "R", 4}, {4, "R", 4}, {5, "R", 5},
	{6, "I", 5}, {7, "I", 4}, {8, "I", 5}, {9, "I", 5}, {10, "I", 5},
	{11, "A", 2}, {12, "A", 3}, {13, "A", 2}, {14, "A", 2}, {15, "A", 3},
	{16, "S", 3}, {17, "S", 3}, {18, "S", 3}, {19, "S", 3}, {20, "S", 3},
	{21, "E", 2}, {22, "E", 2}, {23, "E", 3}, {24, "E", 2}, {25, "E", 3},
	{26, "C", 4}, {27, "C", 4}, {28, "C", 4}, {29, "C", 5}, {30, "C", 4},
}

// RiasecHisGeoPol
// ---------- 文科核心：历史+地理+政治 ----------
var RiasecHisGeoPol = []RIASECAnswer{
	{1, "R", 2}, {2, "R", 2}, {3, "R", 3}, {4, "R", 2}, {5, "R", 3},
	{6, "I", 3}, {7, "I", 4}, {8, "I", 3}, {9, "I", 3}, {10, "I", 3},
	{11, "A", 5}, {12, "A", 4}, {13, "A", 5}, {14, "A", 5}, {15, "A", 4},
	{16, "S", 4}, {17, "S", 4}, {18, "S", 5}, {19, "S", 4}, {20, "S", 4},
	{21, "E", 4}, {22, "E", 3}, {23, "E", 4}, {24, "E", 4}, {25, "E", 4},
	{26, "C", 3}, {27, "C", 3}, {28, "C", 3}, {29, "C", 4}, {30, "C", 3},
}

// RiasecHisGeoBio
// ----------  文理交叉：历史+地理+生物 ----------
var RiasecHisGeoBio = []RIASECAnswer{
	{1, "R", 3}, {2, "R", 3}, {3, "R", 4}, {4, "R", 3}, {5, "R", 3},
	{6, "I", 4}, {7, "I", 4}, {8, "I", 4}, {9, "I", 5}, {10, "I", 4},
	{11, "A", 4}, {12, "A", 3}, {13, "A", 4}, {14, "A", 4}, {15, "A", 3},
	{16, "S", 4}, {17, "S", 3}, {18, "S", 4}, {19, "S", 4}, {20, "S", 3},
	{21, "E", 3}, {22, "E", 3}, {23, "E", 3}, {24, "E", 4}, {25, "E", 3},
	{26, "C", 3}, {27, "C", 4}, {28, "C", 3}, {29, "C", 3}, {30, "C", 4},
}

// RiasecHisPolBio
// ----------  教育社会：历史+政治+生物 ----------
var RiasecHisPolBio = []RIASECAnswer{
	{1, "R", 2}, {2, "R", 3}, {3, "R", 2}, {4, "R", 3}, {5, "R", 2},
	{6, "I", 3}, {7, "I", 4}, {8, "I", 3}, {9, "I", 4}, {10, "I", 3},
	{11, "A", 4}, {12, "A", 4}, {13, "A", 3}, {14, "A", 4}, {15, "A", 3},
	{16, "S", 5}, {17, "S", 5}, {18, "S", 4}, {19, "S", 5}, {20, "S", 4},
	{21, "E", 4}, {22, "E", 4}, {23, "E", 3}, {24, "E", 4}, {25, "E", 3},
	{26, "C", 3}, {27, "C", 3}, {28, "C", 4}, {29, "C", 3}, {30, "C", 3},
}

var RiasecPhyChePol = []RIASECAnswer{
	// R: High due to physics and chemistry (practical, hands-on)
	{1, "R", 5}, {2, "R", 4}, {3, "R", 5}, {4, "R", 4}, {5, "R", 5},
	// I: High due to scientific inquiry
	{6, "I", 5}, {7, "I", 4}, {8, "I", 5}, {9, "I", 5}, {10, "I", 4},
	// A: Low, as creativity is not a focus
	{11, "A", 2}, {12, "A", 3}, {13, "A", 2}, {14, "A", 2}, {15, "A", 3},
	// S: Moderate, influenced by politics (social issues)
	{16, "S", 4}, {17, "S", 3}, {18, "S", 4}, {19, "S", 3}, {20, "S", 4},
	// E: Moderate, influenced by politics (leadership, decision-making)
	{21, "E", 3}, {22, "E", 4}, {23, "E", 3}, {24, "E", 3}, {25, "E", 4},
	// C: Moderate to high, as science and politics involve structured work
	{26, "C", 4}, {27, "C", 4}, {28, "C", 5}, {29, "C", 4}, {30, "C", 4},
}

var RiasecPhyBioPol = []RIASECAnswer{
	// R: High due to physics and biology (hands-on experiments)
	{1, "R", 4}, {2, "R", 5}, {3, "R", 4}, {4, "R", 5}, {5, "R", 4},
	// I: High due to scientific inquiry in physics and biology
	{6, "I", 5}, {7, "I", 5}, {8, "I", 4}, {9, "I", 5}, {10, "I", 4},
	// A: Low, minimal artistic focus
	{11, "A", 2}, {12, "A", 2}, {13, "A", 3}, {14, "A", 2}, {15, "A", 3},
	// S: Moderate to high, influenced by biology (ecology) and politics (social issues)
	{16, "S", 4}, {17, "S", 4}, {18, "S", 3}, {19, "S", 4}, {20, "S", 4},
	// E: Moderate, influenced by politics
	{21, "E", 3}, {22, "E", 4}, {23, "E", 3}, {24, "E", 3}, {25, "E", 4},
	// C: Moderate, structured tasks in science and politics
	{26, "C", 4}, {27, "C", 3}, {28, "C", 4}, {29, "C", 4}, {30, "C", 3},
}

var RiasecPhyGeoPol = []RIASECAnswer{
	// R: Moderate to high, physics and geography (fieldwork, practical)
	{1, "R", 4}, {2, "R", 4}, {3, "R", 5}, {4, "R", 4}, {5, "R", 4},
	// I: High, physics and geography (scientific analysis)
	{6, "I", 4}, {7, "I", 5}, {8, "I", 4}, {9, "I", 4}, {10, "I", 5},
	// A: Low, minimal artistic focus
	{11, "A", 2}, {12, "A", 3}, {13, "A", 2}, {14, "A", 2}, {15, "A", 3},
	// S: Moderate to high, geography (human geography) and politics (social issues)
	{16, "S", 4}, {17, "S", 4}, {18, "S", 5}, {19, "S", 4}, {20, "S", 4},
	// E: Moderate to high, politics (leadership, policy-making)
	{21, "E", 4}, {22, "E", 3}, {23, "E", 4}, {24, "E", 4}, {25, "E", 3},
	// C: Moderate, structured tasks in geography and politics
	{26, "C", 3}, {27, "C", 4}, {28, "C", 3}, {29, "C", 4}, {30, "C", 3},
}

var RiasecHisCheBio = []RIASECAnswer{
	// R: Moderate, chemistry and biology (lab work)
	{1, "R", 3}, {2, "R", 4}, {3, "R", 3}, {4, "R", 4}, {5, "R", 3},
	// I: High, chemistry and biology (scientific inquiry)
	{6, "I", 5}, {7, "I", 4}, {8, "I", 5}, {9, "I", 4}, {10, "I", 5},
	// A: Moderate, history (narrative, creativity)
	{11, "A", 4}, {12, "A", 3}, {13, "A", 4}, {14, "A", 3}, {15, "A", 4},
	// S: Moderate, history and biology (social and ecological concerns)
	{16, "S", 4}, {17, "S", 3}, {18, "S", 4}, {19, "S", 3}, {20, "S", 4},
	// E: Low to moderate, minimal leadership focus
	{21, "E", 2}, {22, "E", 3}, {23, "E", 2}, {24, "E", 3}, {25, "E", 2},
	// C: Moderate, structured tasks in science and history
	{26, "C", 4}, {27, "C", 3}, {28, "C", 4}, {29, "C", 4}, {30, "C", 3},
}

var RiasecHisChePol = []RIASECAnswer{
	// R: Moderate, chemistry (practical work)
	{1, "R", 3}, {2, "R", 3}, {3, "R", 4}, {4, "R", 3}, {5, "R", 3},
	// I: High, chemistry (scientific inquiry)
	{6, "I", 4}, {7, "I", 5}, {8, "I", 4}, {9, "I", 5}, {10, "I", 4},
	// A: High, history (narrative, creativity)
	{11, "A", 4}, {12, "A", 5}, {13, "A", 4}, {14, "A", 4}, {15, "A", 5},
	// S: High, history and politics (social issues, service)
	{16, "S", 4}, {17, "S", 5}, {18, "S", 4}, {19, "S", 5}, {20, "S", 4},
	// E: Moderate to high, politics (leadership, decision-making)
	{21, "E", 4}, {22, "E", 3}, {23, "E", 4}, {24, "E", 4}, {25, "E", 3},
	// C: Moderate, structured tasks in chemistry and politics
	{26, "C", 3}, {27, "C", 4}, {28, "C", 3}, {29, "C", 4}, {30, "C", 3},
}

// AllRIASECCombos
// ========================
// 索引表
// ========================
var AllRIASECCombos = map[string][]RIASECAnswer{
	ComboPHY_CHE_BIO: RiasecPhyCheBio,
	ComboPHY_CHE_GEO: RiasecPhyCheGeo,
	ComboCHE_BIO_GEO: RiasecCheBioGeo,
	ComboPHY_BIO_GEO: RiasecPhyBioGeo,
	ComboHIS_GEO_POL: RiasecHisGeoPol,
	ComboHIS_GEO_BIO: RiasecHisGeoBio,
	ComboHIS_POL_BIO: RiasecHisPolBio,
	ComboPHY_CHE_POL: RiasecPhyChePol,
	ComboPHY_BIO_POL: RiasecPhyBioPol,
	ComboPHY_GEO_POL: RiasecPhyGeoPol,
	ComboHIS_CHE_BIO: RiasecHisCheBio,
	ComboHIS_CHE_POL: RiasecHisChePol,
}

func TestRIASECAnswer() {
	combo := ComboPHY_CHE_BIO
	answers := AllRIASECCombos[combo]
	data, _ := json.MarshalIndent(answers, "", "  ")
	fmt.Println(string(data))
}
