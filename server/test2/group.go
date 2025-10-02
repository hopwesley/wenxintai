package main

import "fmt"

// 科目组合结构体
type SubjectCombination struct {
	Name     string // 组合名称
	Subjects string // 具体科目
	Desc     string // 组合描述
}

// 3+3 模式组合
type ThreePlusThreeModel struct {
	Combinations []SubjectCombination
}

// 3+1+2 模式下的物理类组合
type PhysicsCombinations struct {
	Combinations []SubjectCombination
}

// 3+1+2 模式下的历史类组合
type HistoryCombinations struct {
	Combinations []SubjectCombination
}

// 3+1+2 模式
type ThreePlusOnePlusTwoModel struct {
	Physics PhysicsCombinations
	History HistoryCombinations
}

// 高考选科模型
type GaokaoSelectionModel struct {
	ThreePlusThree      ThreePlusThreeModel
	ThreePlusOnePlusTwo ThreePlusOnePlusTwoModel
}

// 常量定义
var (
	// 3+3 模式所有组合
	ThreePlusThree = ThreePlusThreeModel{
		Combinations: []SubjectCombination{
			{Name: "物化生", Subjects: "物理+化学+生物", Desc: "纯理科"},
			{Name: "物化地", Subjects: "物理+化学+地理", Desc: "理科基础"},
			{Name: "物化政", Subjects: "物理+化学+政治", Desc: "工科潜力"},
			{Name: "物生地", Subjects: "物理+生物+地理", Desc: "交叉组合"},
			{Name: "物生政", Subjects: "物理+生物+政治", Desc: "交叉组合"},
			{Name: "物地政", Subjects: "物理+地理+政治", Desc: "交叉组合"},
			{Name: "史地政", Subjects: "历史+政治+地理", Desc: "纯文科"},
			{Name: "史政生", Subjects: "历史+政治+生物", Desc: "文科基础"},
			{Name: "史政化", Subjects: "历史+政治+化学", Desc: "文科潜力"},
			{Name: "史地生", Subjects: "历史+地理+生物", Desc: "交叉组合"},
			{Name: "史地化", Subjects: "历史+地理+化学", Desc: "交叉组合"},
			{Name: "史生化", Subjects: "历史+生物+化学", Desc: "交叉组合"},
			{Name: "化生政", Subjects: "化学+生物+政治", Desc: "交叉组合"},
			{Name: "化生地", Subjects: "化学+生物+地理", Desc: "交叉组合"},
			{Name: "化政地", Subjects: "化学+政治+地理", Desc: "交叉组合"},
			{Name: "生政地", Subjects: "生物+政治+地理", Desc: "交叉组合"},
			{Name: "物化史", Subjects: "物理+化学+历史", Desc: "文理均衡"},
			{Name: "物生史", Subjects: "物理+生物+历史", Desc: "文理均衡"},
			{Name: "物地史", Subjects: "物理+地理+历史", Desc: "文理均衡"},
			{Name: "物政史", Subjects: "物理+政治+历史", Desc: "文理均衡"},
		},
	}

	// 3+1+2 模式 - 物理类组合
	PhysicsCombos = PhysicsCombinations{
		Combinations: []SubjectCombination{
			{Name: "物化生", Subjects: "物理+化学+生物", Desc: "纯理科组合，难度最大，是顶尖医学院校和理工科专业的标配"},
			{Name: "物化地", Subjects: "物理+化学+地理", Desc: "专业覆盖率极高，仅次于物化生"},
			{Name: "物化政", Subjects: "物理+化学+政治", Desc: "专业覆盖率极高，政治对考研、考公有帮助"},
			{Name: "物生地", Subjects: "物理+生物+地理", Desc: "规避了难度较高的化学，学习压力相对较小"},
			{Name: "物生政", Subjects: "物理+生物+政治", Desc: "规避了难度较高的化学，学习压力相对较小"},
			{Name: "物地政", Subjects: "物理+地理+政治", Desc: "规避了难度较高的化学，学习压力相对较小"},
		},
	}

	// 3+1+2 模式 - 历史类组合
	HistoryCombos = HistoryCombinations{
		Combinations: []SubjectCombination{
			{Name: "史化政", Subjects: "历史+化学+政治", Desc: "带有化学的组合，可以覆盖一部分对化学有要求的专业"},
			{Name: "史化地", Subjects: "历史+化学+地理", Desc: "带有化学的组合，可以覆盖一部分对化学有要求的专业"},
			{Name: "史生政", Subjects: "历史+生物+政治", Desc: "在传统文科基础上增加了理科思维"},
			{Name: "史生地", Subjects: "历史+生物+地理", Desc: "在传统文科基础上增加了理科思维"},
			{Name: "史地政", Subjects: "历史+地理+政治", Desc: "纯文科组合，是传统文科生的选择"},
			{Name: "史生化", Subjects: "历史+生物+化学", Desc: "带有化学的组合，可以覆盖一部分对化学有要求的专业"},
		},
	}

	// 3+1+2 模式
	ThreePlusOnePlusTwo = ThreePlusOnePlusTwoModel{
		Physics: PhysicsCombos,
		History: HistoryCombos,
	}

	// 完整的高考选科模型
	GaokaoModel = GaokaoSelectionModel{
		ThreePlusThree:      ThreePlusThree,
		ThreePlusOnePlusTwo: ThreePlusOnePlusTwo,
	}
)

// 工具函数：打印所有组合
func PrintAllCombinations() {
	fmt.Println("=== 高考选科组合大全 ===")

	fmt.Println("\n--- 3+3 模式 (20种组合) ---")
	for i, combo := range GaokaoModel.ThreePlusThree.Combinations {
		fmt.Printf("%2d. %s: %s (%s)\n", i+1, combo.Name, combo.Subjects, combo.Desc)
	}

	fmt.Println("\n--- 3+1+2 模式 - 物理类 (6种组合) ---")
	for i, combo := range GaokaoModel.ThreePlusOnePlusTwo.Physics.Combinations {
		fmt.Printf("%2d. %s: %s\n   %s\n", i+1, combo.Name, combo.Subjects, combo.Desc)
	}

	fmt.Println("\n--- 3+1+2 模式 - 历史类 (6种组合) ---")
	for i, combo := range GaokaoModel.ThreePlusOnePlusTwo.History.Combinations {
		fmt.Printf("%2d. %s: %s\n   %s\n", i+1, combo.Name, combo.Subjects, combo.Desc)
	}
}

// 根据名称查找组合
func FindCombinationByName(name string) *SubjectCombination {
	// 在 3+3 模式中查找
	for _, combo := range GaokaoModel.ThreePlusThree.Combinations {
		if combo.Name == name {
			return &combo
		}
	}

	// 在 3+1+2 物理类中查找
	for _, combo := range GaokaoModel.ThreePlusOnePlusTwo.Physics.Combinations {
		if combo.Name == name {
			return &combo
		}
	}

	// 在 3+1+2 历史类中查找
	for _, combo := range GaokaoModel.ThreePlusOnePlusTwo.History.Combinations {
		if combo.Name == name {
			return &combo
		}
	}

	return nil
}

// 获取特定模式的组合数量
func GetCombinationCount() map[string]int {
	return map[string]int{
		"3+3模式总组合数":   len(GaokaoModel.ThreePlusThree.Combinations),
		"3+1+2物理类组合数": len(GaokaoModel.ThreePlusOnePlusTwo.Physics.Combinations),
		"3+1+2历史类组合数": len(GaokaoModel.ThreePlusOnePlusTwo.History.Combinations),
		"3+1+2总组合数": len(GaokaoModel.ThreePlusOnePlusTwo.Physics.Combinations) +
			len(GaokaoModel.ThreePlusOnePlusTwo.History.Combinations),
	}
}

func testCombination() {
	// 示例使用
	PrintAllCombinations()

	fmt.Println("\n=== 统计信息 ===")
	counts := GetCombinationCount()
	for key, value := range counts {
		fmt.Printf("%s: %d\n", key, value)
	}

	fmt.Println("\n=== 组合查询示例 ===")
	if combo := FindCombinationByName("物化生"); combo != nil {
		fmt.Printf("找到组合: %s - %s (%s)\n", combo.Name, combo.Subjects, combo.Desc)
	}

	if combo := FindCombinationByName("史地政"); combo != nil {
		fmt.Printf("找到组合: %s - %s (%s)\n", combo.Name, combo.Subjects, combo.Desc)
	}
}
