package main

const (
	ComboPHY_CHE_BIO = "PHY_CHE_BIO"
	ComboPHY_CHE_GEO = "PHY_CHE_GEO"
	ComboPHY_BIO_GEO = "PHY_BIO_GEO"
	ComboCHE_BIO_GEO = "CHE_BIO_GEO"
	ComboHIS_GEO_POL = "HIS_GEO_POL"
	ComboHIS_GEO_BIO = "HIS_GEO_BIO"
	ComboPHY_GEO_CHE = "PHY_GEO_CHE"
	ComboHIS_POL_BIO = "HIS_POL_BIO"
)

const (
	SubjectPHY = "PHY" // 物理
	SubjectCHE = "CHE" // 化学
	SubjectBIO = "BIO" // 生物
	SubjectGEO = "GEO" // 地理
	SubjectHIS = "HIS" // 历史
	SubjectPOL = "POL" // 政治
)

// 方便遍历的固定顺序（六边形绘制顺序）
var Subjects = []string{
	SubjectPHY,
	SubjectCHE,
	SubjectBIO,
	SubjectGEO,
	SubjectHIS,
	SubjectPOL,
}
