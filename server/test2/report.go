package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// ======================================================
// systemPromptUnified
// —— 仅定义角色、任务与阶段逻辑（移除输出结构与纪律）
// ======================================================
func systemPromptUnified() string {
	return `
【身份与任务】
你是融合心理学与 AI 算法的《新高考科学选科决策支持平台》。
你的目标：基于输入的兴趣–能力–组合匹配数据，生成科学、可解释、可执行的《选科战略报告》（JSON 格式）。

【阶段逻辑】
1️⃣ 阶段一（Common 基础分析）：解读兴趣与能力总体结构与关键指标。
2️⃣ 阶段二（Mode 模式分析）：分析 3+3 或 3+1+2 模式的组合得分、稳定性与风险。
3️⃣ 阶段三（Final 汇总整合）：综合前两阶段，输出最终《选科战略报告》（JSON）。`
}

// ======================================================
// 保留 Common 层，不修改内容
// ======================================================
func systemPromptCommon() string {
	return `
【算法背景 - 通用层】
- RIASEC（霍兰德职业兴趣类型）
- OCEAN（大五人格）
- 兴趣–能力匹配（Interest–Ability Fit）
- 全局一致性（Global Cosine）

【关键判断】
- global_cosine < 0.3 → 兴趣与能力方向可能冲突；
- quality_score < 0.7 → 数据可信度偏低；
- fit 越高越协调；ability_share 用于识别优势学科。`
}

// ======================================================
// systemPromptMode33
// —— 含算法背景 + 推荐理由生成策略
// ======================================================
func systemPromptMode33() string {
	return `
【算法背景 - 3+3 模式】
- 自 6 科中选择 3 科等权组合。

【核心指标】
- avg_fit：三科平均匹配度
- combo_cosine：三科方向一致性
- rarity：组合稀有性
- risk_penalty：风险惩罚
- score：综合推荐得分

【报告要点】
- 解释高分组合来源；未入选组合的抑制因子；稀有组合的风险提示。

【推荐理由生成策略】
在 3+3 模式下，推荐理由必须依据真实数据字段生成：
1) 若 avg_fit、combo_cosine 均高且 rarity 低 → 表示组合平衡、方向一致；
2) 若 risk_penalty 高或 combo_cosine 偏低 → 表示潜在冲突或稳定性不足；
3) 若 rarity 较高或 min_ability 偏低 → 表示需谨慎选择；
4) 模型需自动从字段名与数值差异中提炼差异化理由，不可套模板。`
}

// ======================================================
// systemPromptMode312
// —— 含算法背景 + 推荐理由生成策略
// ======================================================
func systemPromptMode312() string {
	return `
【算法背景 - 3+1+2 模式】
- 主干科（Anchor）+ 两门辅科。

【核心指标】
- s1：主干阶段得分
- s23：辅科阶段得分
- s_final：综合得分
- mix_penalty：跨簇惩罚
- coverage：覆盖率

【报告要点】
- 比较不同 Anchor 潜力；解释高分辅科组合来源；指出稳定性/扩展性不足的改进方向。

【推荐理由生成策略】
在 3+1+2 模式下，推荐理由需基于各阶段得分与风险字段生成：
1) 若 s1 与 s23 均高且 mix_penalty 较低 → 表示能力结构平衡；
2) 若 s23 较高但 mix_penalty 偏高 → 表示高潜力但跨簇风险；
3) 若 coverage 广且 s_final 稳定 → 说明适配面广、发展潜力大；
4) 若 anchor 优势明显但辅科差距大 → 提醒专业选择风险；
5) 允许根据字段名自动推断正负作用，禁止使用固定语句模板。`
}

// ======================================================
// 新增 systemPromptFinal —— 输出结构与纪律独立化
// ======================================================
func systemPromptFinal() string {
	return `
【输出结构】
仅输出一个 JSON 对象：
{
  "summary_student": "...",
  "summary_parent": "...",
  "recommendation": [
    {"combo": "组合名称", "reason": "推荐理由（基于字段差异）", "score": 0.00}
  ],
  "risk_analysis": "...",
  "conclusion": "..."
}

【输出纪律】
- 仅基于输入数据；不编造外部事实；
- 不展示数学公式；
- 语言积极、鼓励、易读；
- 内部思考不得输出；
- 唯一允许输出的为 JSON 对象本身，前后不得包含其他文字。`
}

// ======================================================
// userPromptUnified —— 精简，仅保留任务说明与数据展示
// ======================================================
func userPromptUnified(param ParamForAIPrompt, mode Mode) string {
	dataCommon, _ := json.Marshal(param.Common)
	var dataMode []byte
	if mode == Mode33 {
		dataMode, _ = json.Marshal(param.Mode33)
	} else {
		dataMode, _ = json.Marshal(param.Mode312)
	}
	fdCommon := fieldDefinitionCommon()
	fdMode := ""
	if mode == Mode33 {
		fdMode = fieldDefinition33()
	} else {
		fdMode = fieldDefinition312()
	}

	return fmt.Sprintf(`
=========================
阶段一：基础数据（Common）
=========================
【输入数据】
%s

【字段定义】
%s

分析任务：
- 识别兴趣与能力的总体趋势；
- 发现潜在冲突或协调的学科特征；
- 概括能力优势结构。

=========================
阶段二：模式分析（%s 模式）
=========================
【输入数据】
%s

【字段定义】
%s

分析任务：
- 识别高分组合与风险因子；
- 对比各组合在匹配度、一致性、惩罚值等方面的差异；
- 关联基础层特征进行综合分析。

=========================
阶段三：综合整合（Final）
=========================
请基于以上理解，输出完整《选科战略报告》。
仅输出 JSON 对象，结构在系统提示中已定义。`,
		string(dataCommon),
		fdCommon,
		string(mode),
		string(dataMode),
		fdMode)
}

// ======================================================
// 其余函数保持不变
// ======================================================
func fieldDefinitionCommon() string {
	return `
| 字段 | 含义 |
|------|------|
| global_cosine | 兴趣与能力总体方向一致性（−1~1） |
| quality_score | 测评数据可信度（0–1） |
| fit | 单科兴趣–能力匹配度 |
| zgap | 能力与兴趣的差距（z(A) - z(I)） |
| ability_share | 各学科在总体能力中的占比 |
`
}

func fieldDefinition33() string {
	return `
| 字段 | 含义 |
|------|------|
| avg_fit | 三科平均匹配度 |
| combo_cosine | 三科方向一致性 |
| rarity | 稀有性（0=常见，5=谨慎，8–12=稀有） |
| min_ability | 三科中最低能力值 |
| risk_penalty | 风险惩罚 |
| score | 综合推荐得分 |
`
}

func fieldDefinition312() string {
	return `
| 字段 | 含义 |
|------|------|
| s1 | 主干阶段得分 |
| s23 | 辅科阶段得分 |
| s_final | 综合阶段得分 |
| mix_penalty | 理↔文跨簇惩罚 |
| coverage | 招生专业覆盖率 |
| s_final_combo | 综合推荐得分 |
`
}

// ======================================================
// callUnifiedReport —— 调试打印 + 新拼接顺序
// ======================================================
func callUnifiedReport(apiKey string, param ParamForAIPrompt, mode Mode, outPath string) error {
	systemPrompt := systemPromptUnified() + "\n" + systemPromptCommon()
	if mode == Mode33 {
		systemPrompt += "\n" + systemPromptMode33()
	} else {
		systemPrompt += "\n" + systemPromptMode312()
	}
	systemPrompt += "\n" + systemPromptFinal()

	userPrompt := userPromptUnified(param, mode)

	fmt.Println("========== SYSTEM PROMPT ==========")
	fmt.Println(systemPrompt)
	fmt.Println("========== USER PROMPT ==========")
	fmt.Println(userPrompt)

	reqBody := map[string]interface{}{
		"model":       "deepseek-chat",
		"temperature": 0.4,
		"max_tokens":  8000,
		"stream":      true,
		"response_format": map[string]string{
			"type": "json_object",
		},
		"messages": []map[string]string{
			{"role": "system", "content": strings.TrimSpace(systemPrompt)},
			{"role": "user", "content": strings.TrimSpace(userPrompt)},
		},
	}

	content := callDeepSeek(apiKey, reqBody)
	raw := strings.TrimSpace(content)
	if raw == "" {
		return fmt.Errorf("AI 返回空内容（mode=%s）", mode)
	}

	if err := validateJSON(raw); err != nil {
		return fmt.Errorf("AI 返回非合法 JSON：%w", err)
	}

	if err := os.WriteFile(outPath, []byte(raw), 0644); err != nil {
		return fmt.Errorf("写入报告失败: %w", err)
	}

	fmt.Printf("选科战略报告已生成：%s（mode=%s）\n", outPath, mode)
	return nil
}

func validateJSON(s string) error {
	var t map[string]interface{}
	return json.Unmarshal([]byte(s), &t)
}

func TestUnifiedReport(apiKey, filepath string, mode Mode) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("读取参数文件失败: %w", err)
	}
	var param ParamForAIPrompt
	if err := json.Unmarshal(data, &param); err != nil {
		return fmt.Errorf("解析参数JSON失败: %w", err)
	}
	outputFile := fmt.Sprintf("report_unified_v5_%s.json", mode)
	return callUnifiedReport(apiKey, param, mode, outputFile)
}
