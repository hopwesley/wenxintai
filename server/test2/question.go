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

// --- 拆分后的 OCEAN 专属 Prompt ---
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

【行为特异性与高阶思维要求】
- **核心原则**：题干必须是具体的、可观察的**高阶学习行为**或**复杂操作**，严禁使用“容易”、“简单”、“解决问题”等笼统描述。

- **Efficacy (效能感) 维度强化**：
  - 必须使用能体现**学科核心思维**或**复杂应用**的动词（如：**推导、建模、评估、设计、论证、批判、整合、辨析**）。
  - 题干应聚焦于该学科**最具代表性的综合题型**或**实验分析任务**。

- **SkillMastery (反向题) 维度强化**：
  - 必须聚焦于该学科**特有的学习难点、抽象概念**或**精细的技能操作**（如：回避需要**精确计算的步骤**、回避**多因素的逻辑推理**、回避**复杂的图表解读**）。
  - 必须采用稳定的倾向或对比结构来表达相反偏好。

- subject 字段请统一使用以下编码：
  PHY=物理, CHE=化学, BIO=生物, GEO=地理, HIS=历史, POL=政治

- JSON 格式示例：
{
  "id": "int",              // 题号
  "subject": "PHY",         // 学科编码（必填）
  "subject_label": "物理",  // 中文学科名（可选）
  "text": "string",         // 中文题干，基于具体学科任务或情境
  "reverse": false          // 是否为反向题: true | false
  "subtype": "Comparison" | "Efficacy" | "AchievementExpectation" | "SkillMastery" // 对应 4 个分配维度
}
`

// --- 拆分后的 OCEAN 专属 Prompt ---
var systemPromptOCEAN = systemPromptHeader + `
【OCEAN 大五人格（20题）】
- 基于 BFI-20 框架：5 个维度（O/C/E/A/N）各 4 题（固定生成 3 道正向题 + 1 道反向题），且必须标明 reverse:true，共 20 题。
- 聚焦学习风格、学习态度、同伴交往与心理适应，题目需以高中学习和校园生活为主要语境（如课堂、作业、同伴关系、考试压力），可适度包含一般性人格行为以保持心理学有效性。
- **多样性要求**：同一维度下的4道题目必须从不同角度反映该特质，涵盖多种典型校园情境（如课堂学习、作业管理、小组合作、考试应对、社团活动、同伴关系等），确保场景分布均衡。

【JSON 格式示例】
{
  "id": "int",              // 题号
  "dimension": "O" | "C" | "E" | "A" | "N" 
  "text": "string",            // 中文题干
  "reverse": false             // 是否为反向题: true | false
}
`

// --- 拆分后的 RIASEC 专属 Prompt ---
var systemPromptRIASEC = systemPromptHeader + `
【RIASEC 基础题（30题）】
- 6 个维度（R/I/A/S/E/C）各 5 题 (固定生成 4 道正向题 + 1 道反向题)，且必须标明 reverse:true，共 30 题。
- 聚焦基础兴趣倾向，题目需涵盖多样场景（如学习活动、社团活动、职业兴趣探索），确保维度准确性。
- **活动导向限制**：题干必须侧重于具体的活动、情境或行为偏好，**严禁直接使用"数学"、"物理"、"历史"等高考科目名称或专业学术术语**。
- **多样性要求**：同一维度下的5道题目必须覆盖该兴趣维度的不同表现形式，确保场景分布均衡。

【JSON 格式示例】
{
  "id": "int",              // 题号
  "dimension": "R" | "I" | "A" | "S" | "E" | "C", 
  "text": "string",            // 中文题干
  "reverse": false             // 是否为反向题: true | false
}
`

// --- 通用 Prompt 尾部：格式和题干要求 ---
var systemPromptFooter = `
【反向题优化要求】
- **核心原则**：语义清晰，与正向题在行为倾向或态度方向上形成明确对立
- **禁止项**：严格禁止使用"避免"、"讨厌"、"不感兴趣"等否定词
- **推荐表达**：使用"倾向于"、"更偏好"、"通常选择"等对比结构来表达相反偏好
- **稳定性要求**：应描述为一种稳定的倾向或习惯，而非偶发情况 例如：‘我通常喜欢和同伴讨论题目’（稳定） vs ‘有时因为心情不好不想讨论’（偶发，不可用）

【题干要求】
- 题干必须使用简洁自然的中文，适合高中生阅读；保持正式清晰的问卷语气，不得使用英文、拼音或外来词，不得过度学术化或网络化；禁止双重否定、引导性或价值判断。
- 所有题目使用1-5 Likert量表
- 所有题干必须严格基于中国高中生的校园生活与学习场景（如课堂学习、考试准备、完成作业、实验探究、学科竞赛、班级/社团活动、同伴关系、家庭沟通、志愿服务）。
- 严禁出现与成人工作、社会职场、财务或职业行为相关的情境。
- 题干只能围绕高中学习与校园活动语境展开。

【量表锚点要求】
- OCEAN 模块和 RIASEC 模块题目使用符合度锚点：
  1=完全不符合，2=不太符合，3=一般，4=比较符合，5=完全符合
- ASC 模块题目使用擅长程度锚点：
  1=非常不擅长，2=不太擅长，3=一般，4=比较擅长，5=非常擅长
- 必须严格按照对应模块的锚点表述，不得混用。


【输出格式要求】
- 仅以 JSON 对象数组输出，无任何解释。
- 所有字段必须完整填写，不得留空；不得生成未定义的额外字段。
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
			"如果提供了兴趣因子，请仅将其作为题干的背景情境修饰元素，不得改变题目的心理学维度或测量方向。  \n"+
			"兴趣因子题目出现的频率不得超过 2 道题，其余题目保持通用场景。  \n"+
			"如果未提供兴趣因子，则全部使用通用场景，不得随意编造。\n\n",
		requestID, gender, grade, mode, hobby,
	)

	// --- 维持原有 DeepSeek 请求体字段 ---
	reqBody := map[string]interface{}{
		"model":           "deepseek-chat",
		"temperature":     0.7,
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
