package main

// =======================================
// 组合常量定义（3+3 + 3+1+2 模式通用修正版）
// =======================================
const (
	// —— 原 3+3 模式 ——

	ComboPHY_CHE_BIO = "PHY_CHE_BIO"
	ComboPHY_CHE_GEO = "PHY_CHE_GEO"
	ComboPHY_BIO_GEO = "PHY_BIO_GEO"
	ComboCHE_BIO_GEO = "CHE_BIO_GEO"
	ComboHIS_GEO_POL = "HIS_GEO_POL"
	ComboHIS_GEO_BIO = "HIS_GEO_BIO"
	ComboPHY_GEO_CHE = "PHY_GEO_CHE"
	ComboHIS_POL_BIO = "HIS_POL_BIO"

	// —— 新增：3+1+2 物理组（Anchor=PHY）——

	ComboPHY_CHE_POL = "PHY_CHE_POL"
	ComboPHY_BIO_POL = "PHY_BIO_POL"
	ComboPHY_GEO_POL = "PHY_GEO_POL"

	// —— 新增：3+1+2 历史组（Anchor=HIS）——

	ComboHIS_CHE_BIO = "HIS_CHE_BIO"
	ComboHIS_CHE_POL = "HIS_CHE_POL"
	ComboHIS_BIO_GEO = "HIS_BIO_GEO"
	ComboHIS_POL_GEO = "HIS_POL_GEO"
)

const (
	SubjectPHY = "PHY" // 物理
	SubjectCHE = "CHE" // 化学
	SubjectBIO = "BIO" // 生物
	SubjectGEO = "GEO" // 地理
	SubjectHIS = "HIS" // 历史
	SubjectPOL = "POL" // 政治
)

// Subjects
// 方便遍历的固定顺序（六边形绘制顺序）
var Subjects = []string{
	SubjectPHY,
	SubjectCHE,
	SubjectBIO,
	SubjectGEO,
	SubjectHIS,
	SubjectPOL,
}

// AllCombos33 用于 3+3 模式遍历
var AllCombos33 = []string{
	ComboPHY_CHE_BIO,
	ComboPHY_CHE_GEO,
	ComboPHY_BIO_GEO,
	ComboCHE_BIO_GEO,
	ComboHIS_GEO_POL,
	ComboHIS_GEO_BIO,
	ComboPHY_GEO_CHE,
	ComboHIS_POL_BIO,
}

// Aux pools for Mode 3+1+2

var AuxPoolPHY = []string{ // 物理主干下的辅科池
	SubjectCHE, SubjectBIO, SubjectGEO, SubjectPOL,
}

var AuxPoolHIS = []string{ // 历史主干下的辅科池
	SubjectGEO, SubjectPOL, SubjectCHE, SubjectBIO,
}
