package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
)

// ================== 权重配置（默认 1.0） ==================
var dimensionWeights = map[string]float64{
	// RIASEC 六维
	"R": 1.0, "I": 1.0, "A": 1.0, "S": 1.0, "E": 1.0, "C": 1.0,

	// Big Five
	"b5_O": 1.0, "b5_C": 1.0, "b5_E": 1.0, "b5_A": 1.0, "b5_N": 1.0,

	// 学科题（细分）
	"语文:A": 1.0, "语文:S": 1.0,
	"数学:I": 1.0, "数学:C": 1.0,
	"英语:A": 1.0,
	"物理:I": 1.0, "物理:R": 1.0,
	"化学:I": 1.0,
	"生物:S": 1.0,
	"政治:E": 1.0,
	"历史:A": 1.0,
	"地理:I": 1.0,

	// 生涯
	"生涯": 1.5, // 建议提高权重，比如 1.5，让生涯规划更重要

	// 效度题（不计入分数）
	"D":   0.0,
	"P_D": 0.0,

	// 家长维度（前缀 P_）
	"P_R": 1.0, "P_I": 1.0, "P_A": 1.0, "P_S": 1.0, "P_E": 1.0, "P_C": 1.0,
	"P_b5_O": 1.0, "P_b5_C": 1.0, "P_b5_E": 1.0, "P_b5_A": 1.0, "P_b5_N": 1.0,
	"P_价值观": 0.8, // 家长价值观稍弱化
}

// ================== 类型映射 ==================

func calculateQuota(requestID, studentID string, idx int) {

	set := AllAnswerSets[idx]
	studentAnswers, parentAnswers := set.StudentAnswers, set.ParentAnswers

	// === 从 question.json 读取题目 ===
	data, err := os.ReadFile("question.json")
	if err != nil {
		log.Fatal("读取 question.json 失败:", err)
	}

	var q Question

	if err := json.Unmarshal(data, &q); err != nil {
		log.Fatal("解析 question.json 失败:", err)
	}

	studentQuestions := q.StudentQuestions
	parentQuestions := q.ParentQuestions
	// 1. 按 type 聚合原始分数
	scoreMap := make(map[string]int)
	countMap := make(map[string]int)
	rawScoreMap := make(map[string]int)

	for _, ans := range studentAnswers {
		for _, item := range studentQuestions {
			if item.ID == ans.ID {
				score := ans.Score
				if item.Rev {
					score = 6 - score // 反向计分
				}
				rawScoreMap[item.Type] += score
				weight := dimensionWeights[item.Type]
				scoreMap[item.Type] += int(float64(score) * weight)
				countMap[item.Type]++
			}
		}
	}

	for _, ans := range parentAnswers {
		for _, item := range parentQuestions {
			if item.ID == ans.ID {
				score := ans.Score
				if item.Rev {
					score = 6 - score
				}
				key := "P_" + item.Type
				rawScoreMap[key] += score
				weight := dimensionWeights[key]
				scoreMap[key] += int(float64(score) * weight)
				countMap["P_"+item.Type]++
			}
		}
	}

	// 2. 标准化处理（真·T 分数，基于被试在各维度的均分分布）
	// 先算每个维度的均分
	dimAvg := make(map[string]float64)
	dist := make([]float64, 0, len(scoreMap))
	for k, v := range scoreMap {
		n := countMap[k]
		if n <= 0 {
			continue
		}
		avg := float64(v) / float64(n)
		dimAvg[k] = avg
		dist = append(dist, avg)
	}
	// 求总体均值和标准差
	var mean float64
	for _, x := range dist {
		mean += x
	}
	if len(dist) > 0 {
		mean /= float64(len(dist))
	}
	var varsum float64
	for _, x := range dist {
		d := x - mean
		varsum += d * d
	}
	sd := 0.0
	if len(dist) > 0 {
		sd = math.Sqrt(varsum / float64(len(dist)))
	}
	if sd == 0 {
		sd = 1 // 避免除零
	}

	// 3. interpret 表（维度解释）
	interpret := map[string]string{
		// 效度题
		"D": "效度题：用于判断答卷有效性，高分可能表示注意力不足或答题不认真",

		// RIASEC 六维
		"R": "Realistic（现实型/实践）：动手能力、喜欢操作、工程、物理实验",
		"I": "Investigative（研究型/探索）：逻辑推理、科学兴趣、好奇心",
		"A": "Artistic（艺术型/创造）：创意表达、写作、绘画、音乐等",
		"S": "Social（社会型/服务）：助人倾向、沟通合作、社会责任感",
		"E": "Enterprising（企业型/管理）：领导力、说服力、组织管理",
		"C": "Conventional（常规型/执行）：规则意识、细节准确性、秩序感",

		// 大五人格 OCEAN
		"b5_O": "开放性（Openness）：高分=富有想象力，低分=偏保守",
		"b5_C": "尽责性（Conscientiousness）：高分=自律负责，低分=随意散漫",
		"b5_E": "外向性（Extraversion）：高分=外向活跃，低分=安静内向",
		"b5_A": "宜人性（Agreeableness）：高分=友善合作，低分=竞争固执",
		"b5_N": "神经质（Neuroticism）：高分=焦虑紧张，低分=情绪稳定",

		// 学科题
		"语文:A": "语文（A=表达）：文学分析、写作、语言表达",
		"语文:S": "语文（S=合作）：讨论交流、社会文化理解",
		"数学:I": "数学（I=逻辑）：逻辑推理、数学建模、解题兴趣",
		"数学:C": "数学（C=准确）：计算精度、规则严谨性",
		"英语:A": "英语（A=表达）：语言运用、跨文化交流",
		"物理:I": "物理（I=探索）：原理推理、科学兴趣",
		"物理:R": "物理（R=操作）：实验动手、仪器操作",
		"化学:I": "化学（I=实验）：反应观察、探索兴趣",
		"生物:S": "生物（S=生命）：生命现象、研究兴趣",
		"政治:E": "政治（E=社会）：公共事务、治理参与",
		"历史:A": "历史（A=表达）：文化历史分析、叙事兴趣",
		"地理:I": "地理（I=探索）：环境空间、区域分析",

		// 生涯题
		"生涯": "生涯规划：对未来学业、职业方向的思考与信心",

		// 家长维度（前缀 P_）
		"P_D":    "家长效度题：用于判断家长答卷有效性",
		"P_R":    "家长观察：孩子的实践能力（喜欢动手、修理、体力活动）",
		"P_I":    "家长观察：孩子的研究兴趣（科学实验、逻辑问题）",
		"P_A":    "家长观察：孩子的艺术倾向（绘画、音乐、创意表达）",
		"P_S":    "家长观察：孩子的社会性（助人、合作、关心他人）",
		"P_E":    "家长观察：孩子的领导力（组织、影响力、说服力）",
		"P_C":    "家长观察：孩子的规则意识（守纪律、注重细节）",
		"P_b5_O": "家长观察：孩子的开放性（想象力、好奇心）",
		"P_b5_C": "家长观察：孩子的尽责性（自律、责任心）",
		"P_b5_E": "家长观察：孩子的外向性（活跃、喜欢社交）",
		"P_b5_A": "家长观察：孩子的宜人性（友善、合作）",
		"P_b5_N": "家长观察：孩子的情绪稳定性（是否焦虑/紧张）",
		"P_价值观":  "家长价值观：教育理念、全面发展、兴趣与就业平衡",
	}

	// 维度分组：不同组分别计算 T 分数
	groups := map[string]string{
		"R": "RIASEC", "I": "RIASEC", "A": "RIASEC", "S": "RIASEC", "E": "RIASEC", "C": "RIASEC",
		"b5_O": "Big5", "b5_C": "Big5", "b5_E": "Big5", "b5_A": "Big5", "b5_N": "Big5",
		"语文:A": "学科", "语文:S": "学科", "数学:I": "学科", "数学:C": "学科",
		"英语:A": "学科", "物理:I": "学科", "物理:R": "学科", "化学:I": "学科",
		"生物:S": "学科", "政治:E": "学科", "历史:A": "学科", "地理:I": "学科",
		"生涯":  "生涯",
		"P_R": "家长", "P_I": "家长", "P_A": "家长", "P_S": "家长", "P_E": "家长", "P_C": "家长",
		"P_b5_O": "家长", "P_b5_C": "家长", "P_b5_E": "家长", "P_b5_A": "家长", "P_b5_N": "家长",
		"P_价值观": "家长", "P_D": "家长",
	}

	stdScores := make(map[string]float64)

	// 先收集每个组的均分
	groupValues := make(map[string][]float64)
	for k, v := range scoreMap {
		n := countMap[k]
		if n <= 0 {
			continue
		}
		avg := float64(v) / float64(n)
		grp := groups[k]
		if grp == "" {
			if strings.Contains(k, ":") {
				grp = "学科"
			} else if strings.HasPrefix(k, "P_") {
				grp = "家长"
			} else {
				grp = "其他"
			}
		}
		groupValues[grp] = append(groupValues[grp], avg)
	}

	// 每组单独标准化
	for grp, vals := range groupValues {
		// 组均值、方差
		var mean float64
		for _, x := range vals {
			mean += x
		}
		mean /= float64(len(vals))

		var varsum float64
		for _, x := range vals {
			d := x - mean
			varsum += d * d
		}
		sd := math.Sqrt(varsum / float64(len(vals)))
		if sd == 0 {
			sd = 1
		}

		// 再写回各维度的 T 分数
		for k, v := range scoreMap {
			n := countMap[k]
			if n <= 0 {
				continue
			}
			avg := float64(v) / float64(n)
			if groups[k] == grp {
				t := 50 + 10*((avg-mean)/sd)
				if t < 20 {
					t = 20
				} else if t > 80 {
					t = 80
				}
				stdScores[k] = math.Round(t)
			}
		}
	}

	// 确保 interpret 覆盖所有维度
	for k := range scoreMap {
		if _, ok := interpret[k]; !ok {
			interpret[k] = "未定义的维度：" + k
		}
	}

	subjectScores := make(map[string]float64)
	subjectCounts := make(map[string]int)
	for k, v := range rawScoreMap {
		if strings.Contains(k, ":") {
			base := strings.Split(k, ":")[0]
			subjectScores[base] += float64(v)
			subjectCounts[base] += countMap[k]
		}
	}
	for base, sum := range subjectScores {
		subjectScores[base] = sum / float64(subjectCounts[base])
	}

	// 构建 interpret_table（结构化解释表）
	interpretTable := []map[string]string{}
	for k, desc := range interpret {
		grp := groups[k]
		if grp == "" {
			if strings.Contains(k, ":") {
				grp = "学科"
			} else if strings.HasPrefix(k, "P_") {
				grp = "家长"
			} else {
				grp = "其他"
			}
		}
		label, explanation := splitLabelAndExplanation(desc)

		interpretTable = append(interpretTable, map[string]string{
			"dimension":   k,
			"group":       grp,
			"label":       label,
			"explanation": explanation,
		})
	}

	// 4. 生成 quota 结果
	quota := map[string]any{
		"request_id":      requestID,
		"student_id":      studentID,
		"raw_scores":      rawScoreMap, // 原始未加权分数
		"weighted_scores": scoreMap,    // 加权后的分数（目前逻辑）
		"std_scores":      stdScores,
		"item_counts":     countMap,
		"interpret":       interpret,
		"interpret_table": interpretTable, //
		"parent_notes":    "家长维度以 P_ 前缀区分",
		"subject_counts":  subjectCounts, //
		"subject_scores":  subjectScores,
		"validity":        calValidity(q, studentAnswers, parentAnswers),
	}

	bs, _ := json.MarshalIndent(quota, "", "  ")
	_ = os.WriteFile("quota.json", bs, 0644)
	fmt.Println("概要指标已保存到 quota.json")
}

func splitLabelAndExplanation(desc string) (label, explanation string) {
	parts := strings.SplitN(desc, "：", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return desc, desc
}

// 计算 int 切片的均值
func meanInt(xs []int) float64 {
	if len(xs) == 0 {
		return 0
	}
	sum := 0
	for _, v := range xs {
		sum += v
	}
	return float64(sum) / float64(len(xs))
}

// 保留两位小数，写 JSON 更友好
func round2(f float64) float64 {
	return math.Round(f*100) / 100
}

// calValidity 计算问卷的效度情况
// 规则：
//  1. 学生D or 家长D“全相同（全1/全5等）” → 记风险
//  2. 学生D均值 <=2.0 或 >=4.5 → 判为无效（给出原因）
//  3. 家长D均值 <=2.0 或 >=4.5 → 作为辅助风险（合并进原因）
//     注：效度题 D 在问卷定义中已是 rev=true，高分意味着潜在不一致/不认真。
//     这里我们使用答题“原始分”，不做反向处理。
func calValidity(qs Question, studentAnswers, parentAnswers []Answer) map[string]any {
	// 收集学生/家长效度题ID
	stuDIDs := []int{}
	for _, it := range qs.StudentQuestions {
		if it.Type == "D" {
			stuDIDs = append(stuDIDs, it.ID)
		}
	}
	parDIDs := []int{}
	for _, it := range qs.ParentQuestions {
		if it.Type == "D" {
			parDIDs = append(parDIDs, it.ID)
		}
	}

	// 抽取学生/家长效度题得分（用原始分，不反向）
	stuDScores := []int{}
	for _, ans := range studentAnswers {
		for _, id := range stuDIDs {
			if ans.ID == id {
				stuDScores = append(stuDScores, ans.Score)
				break
			}
		}
	}
	parDScores := []int{}
	for _, ans := range parentAnswers {
		for _, id := range parDIDs {
			if ans.ID == id {
				parDScores = append(parDScores, ans.Score)
				break
			}
		}
	}

	flags := []string{}
	// 极端一致性：全相同
	if len(stuDScores) > 1 {
		allSame := true
		first := stuDScores[0]
		for _, s := range stuDScores[1:] {
			if s != first {
				allSame = false
				break
			}
		}
		if allSame {
			flags = append(flags, "学生效度题全部相同分（可能模式化作答）")
		}
	}
	if len(parDScores) > 1 {
		allSame := true
		first := parDScores[0]
		for _, s := range parDScores[1:] {
			if s != first {
				allSame = false
				break
			}
		}
		if allSame {
			flags = append(flags, "家长效度题全部相同分（可能模式化作答）")
		}
	}

	// 阈值判定：学生为主、家长为辅
	valid := true
	reason := "回答有效"

	if len(stuDScores) > 0 {
		stuAvg := meanInt(stuDScores)
		if stuAvg <= 2.0 {
			valid = false
			reason = "学生效度题平均分过低（≤2.0），可能答题不认真"
		} else if stuAvg >= 4.5 {
			valid = false
			reason = "学生效度题平均分过高（≥4.5），可能存在随意作答"
		}
	}

	// 家长阈值提示（不直接改变有效性结论，但写入原因）
	if len(parDScores) > 0 {
		parAvg := meanInt(parDScores)
		if parAvg <= 2.0 {
			flags = append(flags, "家长效度题平均分较低（≤2.0），家长问卷可信度待核查")
		} else if parAvg >= 4.5 {
			flags = append(flags, "家长效度题平均分较高（≥4.5），家长问卷可信度待核查")
		}
	}

	if len(flags) > 0 {
		if valid {
			// 有风险但未到“无效”的阈值
			reason = "存在效度风险：" + strings.Join(flags, "；")
		} else {
			// 已判无效，再追加风险说明
			reason = reason + "；" + strings.Join(flags, "；")
		}
	}

	return map[string]any{
		"is_valid":     valid,
		"reason":       reason,
		"student_D":    stuDScores,
		"parent_D":     parDScores,
		"student_Davg": round2(meanInt(stuDScores)),
		"parent_Davg":  round2(meanInt(parDScores)),
	}
}
