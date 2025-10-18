package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
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
目标：基于兴趣与能力匹配理论（RIASEC + OCEAN）和算法输出的关键参数，
生成一份科学、客观、可解释的《选科战略分析报告》。

【算法背景】
报告依托以下心理与数据基础：
- RIASEC（霍兰德职业兴趣类型）
- OCEAN（大五人格模型）
- 兴趣-能力匹配模型（Interest–Ability Fit）
- 学科适配性计算与全局一致性评估（Global Cosine）
所有结论均基于学生自身测评数据，不包含外部统计或虚构信息。

【字段解释简表】
- global_cosine：兴趣与能力总体方向一致性（0–1，越高越匹配）
- avg_fit_score：平均学科匹配度（越高越平衡）
- fit_variance：学科间匹配度差异（越低越均衡）
- ability_variance：能力离散度（越低越稳定）
- z_gap_mean / z_gap_range：兴趣与能力标准差距（正值代表兴趣超前能力）
- balance_index：综合平衡指数（衡量兴趣与能力协同程度）
- top_subjects / weak_subjects：最强与最弱学科
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
基于每科兴趣-能力匹配度（Fit）、能力标准化分（AbilityNorm）、
稀有性惩罚（Rarity）与风险惩罚（RiskPenalty），
计算每个三科组合的综合得分（Score），并选出前 5 名推荐组合。

【报告重点】
1. 综合能力与兴趣特征分析
2. 组合匹配度比较与最优解
3. 稀有性、风险影响解释
4. 发展方向与学业建议

【输出风格】
- 语言需逻辑清晰、条理分明；
- 使用表格或分点方式说明组合优劣；
- 避免出现公式或算法符号。
`
}

// ========================
// 3+1+2 模式专属提示词部分
// ========================
func systemPromptMode312() string {
	return `
【模式类型】
3+1+2 模式：以“主干科目（Anchor）”为核心，结合两门辅科构建个性化选科组合。

【算法逻辑】
算法分为三个阶段：
1. 阶段一：计算每个主干科的基础适配度（Fit, AbilityNorm）
2. 阶段二：在每个主干科下枚举所有辅科组合（Aux1, Aux2），计算扩展适配度（S23）
3. 阶段三：结合阶段一与阶段二得出最终综合得分（SFinal）

【报告重点】
1. 主干方向分析（Anchor 对比）
2. 辅科组合对比与拓展潜能解释
3. 综合得分比较与建议
4. 学业风险与平衡性提示

【输出风格】
- 注重“方向性”与“个性化潜能”；
- 强调主干与辅科之间的匹配逻辑；
- 使用分点说明或表格呈现结果；
- 避免学术化术语。
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
		"model":           "deepseek-chat",
		"temperature":     0.4, // 报告生成偏稳
		"max_tokens":      8000,
		"stream":          true, // 报告一次性拿完整内容
		"response_format": map[string]string{"type": "json_object"},
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

	// ② 反序列化为结构体
	var param ParamForAIPrompt
	if err := json.Unmarshal(data, &param); err != nil {
		return fmt.Errorf("解析参数JSON失败: %w", err)
	}

	ts := time.Now().Format("20060102_150405")
	outputFile := fmt.Sprintf("report_%s_%s.md", Mode33, ts)
	return callAPIAndSaveReport(Mode33, param, apiKey, outputFile)
}
