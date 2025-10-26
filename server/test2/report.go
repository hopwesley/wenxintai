package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// ========================
// Prompt Builder for AI Reports
// ========================

// BuildPromptForMode33 生成 3+3 模式的 System/User 提示词
func BuildPromptForMode33(param ParamForAIPrompt) (systemPrompt string, userPrompt string, err error) {
	data, err := json.MarshalIndent(param, "", "  ")
	if err != nil {
		return "", "", err
	}

	commonPart := systemPromptCommon()
	mode33Part := systemPromptMode33()
	systemPrompt = fmt.Sprintf("%s\n%s\n【输入数据】\n%s\n", commonPart, mode33Part, string(data))

	userPrompt = `
请基于上述数据，生成一份《选科战略分析报告（3+3 模式）》。
要求：
1. 报告使用 Markdown 格式；
2. 报告结构包括：
   - 一、核心发现（兴趣与能力总体特征）
   - 二、组合模式分析（3+3 模式）
   - 三、最优组合与推荐理由
   - 四、风险提示与改进方向
   - 五、总结展望
3. 报告用语应通俗易懂，面向高中学生与家长；
4. 不得编造学校、专业或外部事实。
`
	return systemPrompt, userPrompt, nil
}

// BuildPromptForMode312 生成 3+1+2 模式的 System/User 提示词
func BuildPromptForMode312(param ParamForAIPrompt) (systemPrompt string, userPrompt string, err error) {
	data, err := json.MarshalIndent(param, "", "  ")
	if err != nil {
		return "", "", err
	}

	commonPart := systemPromptCommon()
	mode312Part := systemPromptMode312()
	systemPrompt = fmt.Sprintf("%s\n%s\n【输入数据】\n%s\n", commonPart, mode312Part, string(data))

	userPrompt = `
请基于上述数据，生成一份《选科战略分析报告（3+1+2 模式）》。
要求：
1. 报告使用 Markdown 格式；
2. 报告结构包括：
   - 一、核心发现（兴趣与能力总体特征）
   - 二、主干方向分析（Anchor 对比）
   - 三、最佳辅科组合与扩展潜能
   - 四、风险提示与改进方向
   - 五、总结展望
3. 报告用语应通俗易懂，面向高中学生与家长；
4. 不得编造学校、专业或外部事实。
`
	return systemPrompt, userPrompt, nil
}

// ========================
// 通用提示词部分
// ========================
func systemPromptCommon() string {
	return `
【身份与任务】
你是融合心理学与AI算法的《新高考科学选科决策支持平台》。
目标：基于兴趣与能力匹配理论（RIASEC + OCEAN）与算法输出的关键参数，
生成一份科学、可解释、面向学生与家长的《选科战略分析报告》。

【算法背景】
报告依托以下心理与数据基础：
- RIASEC（霍兰德职业兴趣类型）
- OCEAN（大五人格模型）
- 兴趣–能力匹配模型（Interest–Ability Fit）
- 全局一致性指标（Global Cosine）
所有分析均基于学生自身测评数据，不得编造外部事实。

【算法解释原则】
1. 若推荐结果与预期组合不同，必须指出算法判断的关键原因（如 Fit 偏低、Cosine 过低、稀有性惩罚、跨簇惩罚等）。
2. 若 global_cosine < 0.3，说明兴趣与能力方向冲突，应在报告中解释。
3. 若 quality_score < 0.7，应提示“数据可信度较低”。
4. 对每个推荐组合，需解释其高分来源（结构一致性、匹配平衡、能力支撑）。

【字段说明】
| 字段 | 含义 |
|------|------|
| global_cosine | 兴趣与能力总体方向一致性（−1～1，越高越协调） |
| quality_score | 测评数据可信度（0–1） |
| interest_z / ability_z / zgap | 兴趣与能力的标准化Z值与差异 |
| ability_share | 各科在总体能力中的比重 |
| fit | 单科兴趣–能力匹配度 |
| avg_fit / min_fit | 三科组合的平均与最低匹配度 |
| combo_cosine | 三科在兴趣–能力空间中的方向一致性 |
| rarity | 组合稀有性（0=常见，5=谨慎，8–12=稀有） |
| risk_penalty | 组合风险惩罚，用于反映不稳定或方向冲突 |
| coverage | 招生专业覆盖率 |
| mix_penalty | 跨簇惩罚（理文错配时增加惩罚） |
| min_ability | 三科中最低能力水平，用于识别潜在短板 |
| ability_norm | 主干科能力标准化值 |
| auxAbility | 辅科平均能力标准化值 |
| s1 / s23 / s_final | 各阶段综合得分（3+1+2 模式） |
`
}

// ========================
// 3+3 模式专属提示词部分
// ========================
func systemPromptMode33() string {
	return `
【模式类型】
3+3 模式：学生从 6 门科目中选择 3 门作为选考科目，三科等权。

【算法逻辑】
算法综合考虑以下权重：
- Fit：兴趣与能力的匹配度；
- ComboCosine：三科方向一致性；
- Rarity：组合稀有性；
- RiskPenalty：组合风险惩罚；
- MinAbility：最低能力标准化值。
最终得分用于确定最优推荐顺序。

【报告重点】
1. 分析兴趣与能力的总体协调性（Global Cosine）；
2. 说明最优组合的算法原因与心理特征；
3. 若目标组合未入选，指出关键抑制因子（如负 Fit、低 Cosine、Rarity 过高）；
4. 对稀有组合给出教育建议与风险提示；
5. 以辅导老师口吻解释选择逻辑与改进方向。

【输出要求】
- 使用自然语言分段说明；
- 说明算法与心理学逻辑；
- 禁止使用数学符号；
- 语言简洁、逻辑透明；
- 以“为什么推荐”与“为什么未推荐”为主线。`
}

// ========================
// 3+1+2 模式专属提示词部分
// ========================
func systemPromptMode312() string {
	return `
【模式类型】
3+1+2 模式：以主干科（Anchor）为核心，结合两门辅科形成个性化方案。

【算法逻辑】
1. 阶段一（S1）：计算 Anchor 科目的匹配与能力贡献；
2. 阶段二（S23）：评估辅科组合的平均匹配度、能力水平、一致性；
3. 阶段三（SFinal）：综合 Anchor 稳定性与辅科扩展性形成最终得分。

【报告重点】
1. 对比不同主干方向（理/文）Anchor 的表现；
2. 分析每组辅科组合的高分来源；
3. 若目标组合未入选，指出算法层面的核心原因；
4. 结合心理特征说明发展方向；
5. 提出兴趣–能力均衡的改进建议。

【字段解析补充】
- s1 / s23 / s_final：主干、辅科、综合阶段得分；
- term_fit / term_ability / term_coverage：S1 分项贡献；
- auxAbility：辅科平均能力；
- mix_penalty：跨簇惩罚；
- coverage：招生专业覆盖率；
- rarity：组合稀有度。

【输出风格】
- 以学生与家长易懂的语言；
- 强调 Anchor → 辅科的逻辑层次；
- 不使用符号或复杂公式；
- 给出“算法结果+心理解释+教育建议”三层结论。
`
}

func callAPIAndSaveReport(modeStr Mode, param ParamForAIPrompt, apiKey, outPath string) error {
	var (
		systemPrompt string
		userPrompt   string
		err          error
	)

	switch modeStr {
	case Mode33:
		systemPrompt, userPrompt, err = BuildPromptForMode33(param)
	case Mode312:
		systemPrompt, userPrompt, err = BuildPromptForMode312(param)
	default:
		return fmt.Errorf("未知报告模式: %s", modeStr)
	}
	if err != nil {
		return err
	}

	// —— DeepSeek 请求体，与 question.go 的 callAPIAndSave 同风格 ——
	reqBody := map[string]interface{}{
		"model":       "deepseek-chat",
		"temperature": 0.4, // 报告生成偏稳
		"max_tokens":  8000,
		"stream":      true, // 报告一次性拿完整内容
		//"response_format": map[string]string{"type": "json_object"},
		"messages": []map[string]string{
			{"role": "system", "content": strings.TrimSpace(systemPrompt)},
			{"role": "user", "content": userPrompt},
		},
	}

	// —— 直接复用你已有的 callDeepSeek(apiKey, reqBody) ——
	content := callDeepSeek(apiKey, reqBody)
	raw := strings.TrimSpace(content)
	if raw == "" {
		// 给点上下文信息，方便定位
		b, _ := json.Marshal(reqBody)
		return fmt.Errorf("报告生成返回空内容, mode=%s, reqBody=%s", modeStr, string(b))
	}

	// —— 落地到本地文件（Markdown） ——
	if err := os.WriteFile(outPath, []byte(raw), 0644); err != nil {
		return fmt.Errorf("写入报告失败: %w", err)
	}
	return nil
}

func TestReport(apiKey, filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("读取参数文件失败: %w", err)
	}

	var param ParamForAIPrompt
	if err := json.Unmarshal(data, &param); err != nil {
		return fmt.Errorf("解析参数JSON失败: %w", err)
	}

	outputFile := fmt.Sprintf("report_%s.md", filepath)
	return callAPIAndSaveReport(Mode33, param, apiKey, outputFile)
}
