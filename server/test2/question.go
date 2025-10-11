package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

var SuperQuestions = []map[string]interface{}{
	{
		"id":        "Super-1",
		"module":    "Super",
		"dimension": "Values",
		"type":      "sort", // 固定：排序题
		"text":      "请从以下5个核心价值中选出最重要的3个并排序",
		"options":   []string{"成就感", "经济回报", "社会利他", "独立性", "创造性"},
		"reverse":   false,
	},
	{
		"id":        "Super-2",
		"module":    "Super",
		"dimension": "Values",
		"type":      "single", // 固定：单选题
		"text":      "请从上题选出的3个价值观中，选择1个‘无论如何都不能放弃’的核心锚点",
		"options":   []string{"成就感", "经济回报", "社会利他", "独立性", "创造性"},
		"reverse":   false,
	},
}

// --- 通用 System Prompt 头部 ---
var systemPromptHeader = `
【身份与任务】
你是融合心理学权威理论与AI算法的《新高考科学选科决策支持平台》系统。你的目标：基于霍兰德职业兴趣理论（RIASEC）和大五人格模型（OCEAN），
为中国高中学生及其家长设计一份科学有效、文化适配的综合选科测评问卷。题目需参考标准量表（如BFI-20和RIASEC SDS）进行本土化改编，
最终支持《选科战略分析报告》，为科目组合（偏文、偏理、偏工、偏艺）提供科学推荐参考。
`

var systemPromptASC = systemPromptHeader + `
【学科自我概念量表（24题）】
- 参考Marsh的SDQ-III结构，旨在测量学生对核心高考选考科目的能力信心和表现认知。
- 覆盖6个高考科目（物理、化学、生物、地理、历史、政治），每个学科4题，共24题。
- 每学科固定生成 3 道正向题 + 1 道反向题，且必须标明 reverse:true，聚焦能力短板认知。
- **题目分配**：每个学科的4题应涵盖：
  1. 比较性能力自信（与同伴比较）
  2. 学习任务效能感  
  3. 学业成就预期
  4. 特定技能掌握信心（反向题）

【行为要求】 
- 题干基于中国初三/高一课程标准（如物理力学基础、化学简单反应），聚焦课堂或实验场景（如解答练习、完成实验）。
- 题干示例：解答课堂练习、完成实验步骤、复习考试内容、参与小组讨论。
- **Efficacy 维度**：用“解答、记忆、完成、分析”等动词，禁止“建模、论证、批判”等高阶词汇。

【反向题规则（SkillMastery）】
- 必须围绕该学科的**常见学习难点**（如复杂公式、图表解读、历史细节、政治概念）。
- 反向题应体现**低水平倾向**，采用“偏好/习惯对比”而非“困难/出错”的否定。
- 推荐模式：
  - 正向：我通常能够熟练掌握……
  - 反向：我通常不太倾向于在……中表现出优势 / 我相对较少在……任务中展现熟练度
- **禁止**：“感到困难”“经常出错”等偶发或情绪化表述。

- subject 字段请统一使用以下编码：
  PHY=物理, CHE=化学, BIO=生物, GEO=地理, HIS=历史, POL=政治

【学科特色要求】
- 物理：侧重公式应用、实验操作、现象解释
- 化学：侧重反应记忆、实验安全、物质性质  
- 生物：侧重结构识别、分类记忆、生命过程
- 地理：侧重地图阅读、空间思维、自然人文
- 历史：侧重时间脉络、事件关联、材料分析
- 政治：侧重概念理解、时事讨论、案例分析

【JSON 格式示例】
{
  "id": "int",              // 题号
  "subject": "PHY",         // 学科编码（必填）
  "subject_label": "物理",  // 中文学科名（可选）
  "text": "string",         // 中文题干，基于具体学科任务或情境
  "reverse": false          // 是否为反向题: true | false
  "subtype": "Comparison" | "Efficacy" | "AchievementExpectation" | "SkillMastery" // 对应 4 个分配维度
}
`
var systemPromptOCEAN = systemPromptHeader + `
【OCEAN 大五人格（20题）】
- 基于 BFI-20 框架：5 个维度（O/C/E/A/N）各 4 题，共 20 题。
- **严格约束**：每个维度仅允许 1 道反向题（reverse:true），其余 3 道必须为正向题。
- 聚焦学习风格、学习态度、同伴交往与心理适应。题目需以高中学习和校园生活为主要语境（如课堂、作业、同伴关系、考试压力），可适度包含一般性人格行为以保持心理学有效性。
- **多样性要求**：同一维度下的 4 道题目必须从不同角度反映该特质，涵盖多种典型校园情境（如课堂学习、作业管理、小组合作、考试应对、社团活动、同伴关系等），确保场景分布均衡。

【反向题规则】
- 反向题是**同一维度的低水平表征**，不得跨维度。
- 维度对照示例：
  - O（开放性）：正向=喜欢尝试新方法 → 反向=偏好固定、传统方式
  - C（尽责性）：正向=有计划、守时 → 反向=容易拖延、缺乏条理
  - E（外向性）：正向=喜欢与人互动 → 反向=更倾向独处、较少参与
  - A（宜人性）：正向=乐于合作、体贴 → 反向=更倾向坚持自我、不易妥协
  - N（神经质）：正向=焦虑/易紧张 → 反向=冷静/稳定
- 表达：使用“较少/不太倾向/通常不”等自然表述；**禁止**“讨厌/不感兴趣”等强否定。

【JSON 格式示例】
{
  "id": "int",              // 题号
  "dimension": "O" | "C" | "E" | "A" | "N" 
  "text": "string",            // 中文题干
  "reverse": false             // 是否为反向题: true | false
}
`

var systemPromptRIASEC = systemPromptHeader + `
【RIASEC 基础题（30题）】
- 6 个维度（R/I/A/S/E/C）各 5 题（固定生成 4 道正向题 + 1 道反向题），且必须标明 reverse:true，共 30 题。
- 聚焦基础兴趣倾向，题目需涵盖多样场景（如学习活动、社团活动、职业兴趣探索），确保维度准确性。
- **活动导向限制**：题干必须侧重于具体的活动、情境或行为偏好，**严禁直接使用"数学"、"物理"、"历史"等高考科目名称或专业学术术语**。
- **多样性要求**：同一维度下的5道题目必须覆盖该兴趣维度的不同表现形式，确保场景分布均衡。

【反向题规则】
- 每个维度固定 1 道反向题；必须保持**活动导向**，与正向题在行为偏好上对立，且**不跨维度**。
- 示例对照：
  - R（现实型）：正向=喜欢动手实践 → 反向=更少参与操作、避免动手活动
  - I（研究型）：正向=喜欢探索分析 → 反向=缺少探究兴趣、不主动研究
  - A（艺术型）：正向=喜欢表达创意 → 反向=较少创造性表达、偏好常规活动
  - S（社会型）：正向=喜欢帮助他人 → 反向=更少参与助人或社交活动
  - E（企业型）：正向=喜欢组织领导 → 反向=较少承担领导角色、偏向跟随
  - C（常规型）：正向=喜欢结构和秩序 → 反向=对规则/条理依赖较低，偏好灵活
- 表达：使用“较少/不太倾向/通常不”等自然表述；**禁止**“讨厌/不感兴趣”等强否定。

【JSON 格式示例】
{
  "id": "int",              // 题号
  "dimension": "R" | "I" | "A" | "S" | "E" | "C", 
  "text": "string",            // 中文题干
  "reverse": false             // 是否为反向题: true | false
}
`
var systemPromptFooter = `
【场景覆盖约束】  
- 每个维度/学科下的题目必须覆盖至少 3 种不同校园场景，不得集中在单一语境。  
- 校园场景示例（至少覆盖其三）：课堂学习、作业/复习、考试准备、实验探究、小组合作、社团活动、体育/艺术活动、同伴交往、家庭沟通、志愿服务。  
- 如出现兴趣因子，应优先用于**扩展新的场景类型**，而非简单修饰已有场景。  
- 同一模块内不得出现高度重复的题干语境或措辞（如连续 3 道“按时完成作业”）。

【题干要求】
- 题干必须使用简洁自然的中文，适合高中生阅读；保持正式清晰的问卷语气，不得使用英文、拼音或外来词，不得过度学术化或网络化；禁止双重否定、引导性或价值判断。
- 所有题目使用1-5 Likert量表
- 所有题干必须严格基于中国高中生的校园生活与学习场景（如课堂学习、考试准备、完成作业、实验探究、学科竞赛、班级/社团活动、同伴关系、家庭沟通、志愿服务）。
- 严禁出现与成人工作、社会职场、财务或职业行为相关的情境。

【量表锚点要求】
- OCEAN / RIASEC：1=完全不符合，2=不太符合，3=一般，4=比较符合，5=完全符合
- ASC：1=非常不擅长，2=不太擅长，3=一般，4=比较擅长，5=非常擅长
- 必须严格按照对应模块的锚点表述，不得混用。

【输出格式要求】
- 仅以 JSON 对象数组输出，无任何解释。
- 所有字段必须完整填写；不得生成未定义字段。
- id 从 1 开始，连续编号。
`

// --- Prompt 组装（根据模块拼接） ---
func composeSystemPrompt(module string) (string, error) {
	var modulePrompt string
	switch module {
	case "OCEAN":
		modulePrompt = systemPromptOCEAN
	case "RIASEC":
		modulePrompt = systemPromptRIASEC
	case "ASC":
		modulePrompt = systemPromptASC
	default:
		return "", fmt.Errorf("未知模块: %s", module)
	}
	return strings.TrimSpace(modulePrompt + "\n" + systemPromptFooter), nil
}

// ========================= 生成问卷（修改后，双重调用） =========================
// 新增一个内部函数用于处理单次 API 调用和文件保存
func callAPIAndSave(module string, mode Mode, apiKey, gender, grade, hobby string) error {
	// --- 组装 system prompt（严格按“通用 + 模块补充 + 尾部要求”拼接） ---
	systemPrompt, err := composeSystemPrompt(module)
	if err != nil {
		return err
	}

	// --- 构造用户提示（与原逻辑一致） ---
	requestID := "question_" + uuidLike()
	userPrompt := fmt.Sprintf(
		"请以 json 对象数组返回，仅输出合法 json：\n"+
			"request_id: %s\n"+
			"学生基本信息：性别：%s，年级：%s。\n"+
			"选科模式：%s。\n"+
			"**学生兴趣因子：%s**。\n"+
			"请严格遵循 systemPrompt 的数量、结构和维度覆盖要求，"+
			"【兴趣因子使用约束】\n"+
			"- 兴趣因子仅作为“场景种子”，用于派生校园语境，不得直接作为心理学变量。\n"+
			"- 允许：将兴趣因子转化为典型校园场景，再结合学习/社交/任务行为形成题干。\n"+
			"  例如：篮球 → 体育课/课余锻炼 → 团队合作、专注力\n"+
			"        音乐 → 社团/艺术节表演 → 坚持练习、表现焦虑\n"+
			"- 禁止：直接硬写兴趣因子或将其作为考察对象。\n"+
			"  不合理示例：'我在篮球比赛失利后容易焦虑'（错误：把篮球当成心理变量）\n"+
			"  合理示例：'我喜欢在体育课上积极参与团队合作活动'（场景为篮球引申的体育课，主体为团队合作行为）\n"+
			"- 出现频率：兴趣因子衍生的场景题目最多4题，且不得在同一维度/学科中连续出现。如果兴趣与当前模块不匹配，则兴趣因子可以不生成题目\n"+
			"- 如果未提供兴趣因子，则全部使用通用校园场景。\n",
		requestID, gender, grade, mode, hobby,
	)

	//temperature := getTemperature(module)
	// --- 维持原有 DeepSeek 请求体字段 ---
	reqBody := map[string]interface{}{
		"model":           "deepseek-chat",
		"temperature":     0.75,
		"max_tokens":      8000,
		"stream":          true, // 双重调用最好不要用 stream，防止中断。
		"response_format": map[string]string{"type": "json_object"},
		"messages": []map[string]string{
			{"role": "system", "content": strings.TrimSpace(systemPrompt)},
			{"role": "user", "content": userPrompt},
		},
	}

	// --- 调用 DeepSeek 并输出 ---
	content := callDeepSeek(apiKey, reqBody)
	raw := strings.TrimSpace(content)
	if raw == "" {
		return fmt.Errorf("模型返回空内容 for %s", module)
	}

	// --- 最小 JSON 校验（与原逻辑一致） ---
	var tmp any
	if err := json.Unmarshal([]byte(raw), &tmp); err != nil {
		fmt.Printf("警告：[%s] 返回内容非严格 JSON，仍原样保存。解析错误：%v\n", module, err)
	}

	// --- 落盘，文件名设计为区分模块 ---
	ts := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("questions_%s_%s_%s.json", mode, module, ts) // 增加了模块名
	err = os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("保存文件失败 [%s]: %w", module, err)
	}

	fmt.Printf("问卷已保存：[%s] -> %s\n", module, filename)
	return nil
}

func getTemperature(module string) float64 {
	switch module {
	case "OCEAN":
		return 0.8 // 稍高，鼓励创意和多样性
	case "RIASEC":
		return 0.7 // 中等，平衡稳定与多样性
	case "ASC":
		return 0.6 // 稍低，确保结构准确性和一致性
	default:
		return 0.75 // 默认值
	}
}

// 主生成函数，拆分为两次调用
func generateQuestions(mode Mode, apiKey, gender, grade, hobby string) error {
	fmt.Println("--- 开始 题目 ---", hobby)
	fmt.Println("--- 开始生成 OCEAN 题目 (20题) ---")
	// 第一次调用：生成 OCEAN 题目
	errOCEAN := callAPIAndSave("OCEAN", mode, apiKey, gender, grade, hobby)
	if errOCEAN != nil {
		fmt.Printf("生成 OCEAN 题目失败: %v\n", errOCEAN)
	}

	fmt.Println("\n--- 开始生成 RIASEC 题目 (30题) ---")
	// 第二次调用：生成 RIASEC 题目
	errRIASEC := callAPIAndSave("RIASEC", mode, apiKey, gender, grade, hobby)
	if errRIASEC != nil {
		fmt.Printf("生成 RIASEC 题目失败: %v\n", errRIASEC)
	}

	fmt.Println("\n--- 开始生成 ASC 题目 (24题) ---")
	// 第二次调用：生成 RIASEC 题目
	errASC := callAPIAndSave("ASC", mode, apiKey, gender, grade, hobby)
	if errASC != nil {
		fmt.Printf("生成 ASC 题目失败: %v\n", errASC)
	}

	// 综合返回错误
	if errOCEAN != nil || errRIASEC != nil || errASC != nil {
		return fmt.Errorf("部分或全部题目生成失败: OCEAN 错误: %v, RIASEC 错误: %v, ASC 错误: %v  ", errOCEAN, errRIASEC, errASC)
	}

	fmt.Println("\n--- 三个模块题目均已成功生成并保存 ---")
	return nil
}
