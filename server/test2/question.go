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

var systemPromptCommon = `
【身份与任务】
你是融合心理学权威理论与AI算法的《新高考科学选科决策支持平台》系统。你的目标：基于霍兰德职业兴趣理论（RIASEC）和大五人格模型（OCEAN），
为中国高中学生及其家长设计一份科学有效、文化适配的综合选科测评问卷。题目需参考标准量表（如BFI-20和RIASEC SDS）进行本土化改编，
确保题目内容贴合中国高考选科场景（如学科偏好、职业倾向），最终支持《选科战略分析报告》，为科目组合（偏文、偏理、偏工、偏艺）提供科学推荐参考。

【OCEAN 大五人格（20题）】
- 基于 BFI-20 框架：5 个维度（O/C/E/A/N）各 4 题，共 20 题；每个维度包含 1 道反向题。
- 聚焦学习风格、学习态度、同伴交往与心理适应，题目需以高中学习和校园生活为主要语境（如课堂、作业、同伴关系、考试压力），可适度包含一般性人格行为以保持心理学有效性。。
- **多样性要求**：同一维度下的4道题目必须从不同角度反映该特质，涵盖多种典型校园情境（如课堂学习、作业管理、小组合作、考试应对、社团活动、同伴关系等），确保场景分布均衡。

【RIASEC 兴趣（30题）】
- 6 个维度（R/I/A/S/E/C）各 5 题，共 30 题；每个维度包含 1 道反向题。
- 聚焦基础兴趣倾向，题目需涵盖多样场景（如学习活动、社团活动、职业兴趣探索），确保维度准确性。
- **活动导向限制**：题干必须侧重于具体的活动、情境或行为偏好，**严禁直接使用"数学"、"物理"、"历史"等高考科目名称或专业学术术语**。
- **多样性要求**：同一维度下的5道题目必须覆盖该兴趣维度的不同表现形式，确保场景分布均衡。

【题干要求】
- 题干必须使用简洁自然的中文，适合高中生阅读；保持正式清晰的问卷语气，不得使用英文、拼音或外来词，不得过度学术化或网络化；禁止双重否定、引导性或价值判断。
- **反向题优化要求**：
  - 语义清晰，与正向题在行为倾向或态度方向上形成明确对立
  - **严格禁止使用"避免"、"讨厌"、"不感兴趣"等否定词**
  - **优先使用"倾向于"、"更偏好"、"通常选择"等对比结构来表达相反偏好**
  - 反向题应描述为一种稳定的倾向或习惯，而非偶发情况
- 所有题目使用1-5 Likert量表，锚点为：1=完全不符合，2=不太符合，3=一般，4=比较符合，5=完全符合
- 每个维度必须包含且仅包含 1 道 reverse=true 的题目。
- **反向题句式多样化要求**：反向题不得重复使用相同的句式（如“倾向于…而不是…”），需在全卷中体现多种表述方式（如“更常…而较少…”、“通常选择…而很少…”、“更偏向于…而不太…”），保证自然多样。
- **场景覆盖要求**：同一维度的题目必须覆盖至少 3 种不同的场景（如课堂学习、作业管理、同伴关系、社团活动、考试压力、家庭支持），不得集中在单一场景。

【输出格式要求】
- 仅以 JSON 对象输出，无任何解释。
- 所有字段必须完整填写，不得留空；不得生成未定义的额外字段。
- id 全局从 1 开始，连续编号。
{
  "id": "int",              // 唯一题号
  "module": "OCEAN" | "RIASEC",// 题目所属模块
  "dimension": "O" | "C" | "E" | "A" | "N"   // 如果 module=OCEAN
              | "R" | "I" | "A" | "S" | "E" | "C", // 如果 module=RIASEC
  "text": "string",            // 中文题干
  "reverse": false             // 是否为反向题: true | false
}
`

// ------------------------- 模式：3+3 配额/补充 -------------------------
var systemPrompt33 = `
`

// ------------------------- 模式：3+1+2 配额/补充 -------------------------
var systemPrompt312 = `
`

// ------------------------- Prompt 组装 -------------------------
func composeSystemPrompt(mode Mode) (string, error) {
	var modePrompt string
	switch mode {
	case Mode33:
		modePrompt = systemPrompt33
	case Mode312:
		modePrompt = systemPrompt312
	default:
		return "", fmt.Errorf("未知模式: %s", mode)
	}
	return strings.TrimSpace(systemPromptCommon + "\n" + modePrompt), nil
}

// ========================= 生成问卷（函数签名与原文件一致） =========================
func generateQuestions(mode Mode, apiKey, gender, grade string) error {
	// --- 组装 system prompt（严格按“通用 + 模式补充”拼接） ---
	systemPrompt, err := composeSystemPrompt(mode)
	if err != nil {
		return err
	}

	// --- 构造用户提示（与原逻辑一致） ---
	requestID := "question_" + uuidLike()
	userPrompt := fmt.Sprintf(
		"请以 json 对象返回，仅输出合法 json：\n"+
			"request_id: %s\n"+
			"学生基本信息：性别：%s，年级：%s。\n"+
			"**选科模式：%s**。\n"+
			"请严格遵循 systemPrompt 的数量、结构和维度覆盖要求。\n",
		requestID, gender, grade, mode,
	)

	// --- 维持原有 DeepSeek 请求体字段 ---
	reqBody := map[string]interface{}{
		"model":           "deepseek-chat",
		"temperature":     0.7,
		"max_tokens":      8000,
		"stream":          true,
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
		return fmt.Errorf("模型返回空内容")
	}

	// --- 最小 JSON 校验（与原逻辑一致） ---
	var tmp any
	if err := json.Unmarshal([]byte(raw), &tmp); err != nil {
		fmt.Println("警告：返回内容非严格 JSON，仍原样保存。解析错误：", err)
	}

	// --- 落盘（与原逻辑一致） ---
	ts := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("questions_%s_%s.json", mode, ts)
	_ = os.WriteFile(filename, []byte(content), 0644)
	fmt.Println("问卷已保存：", filename)
	return nil
}
