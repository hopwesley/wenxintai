package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// 仅为命名提取 student_id
type payloadMeta struct {
	Meta struct {
		StudentID string `json:"student_id"`
	} `json:"meta"`
}

// —— I/O 辅助 ——
func readText(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func writeText(path, s string) error {
	return os.WriteFile(path, []byte(s), 0644)
}

func buildSystemPrompt() string {
	return `
你是融合心理学权威理论与AI智能算法的“新高考科学选科决策支持平台”。
你的职责：把提供的测评JSON数据转化为科学、可执行的《选科战略分析报告》。

必须遵守的边界与底线：
- 仅基于提供的 JSON 数据生成内容，禁止编造外部事实、学校或专业名单；
- 推荐组合必须符合当期新高考政策，并与 meta.mode 一致；
- 推荐的每一门学科必须存在于 JSON 的 subjects 数据中（名称需与 JSON 保持一致，不做同义转换）；
- deltas 仅代表家长与学生的“认知差异”，不得写成“能力高低”的结论；
- 引用任何维度均值时结合 count 给出稳定性提示（count=1/2 → 数据有限/谨慎解读；count≥3 → 相对较可靠但避免绝对化）；
- 禁止使用排名化表述（如“第几名/最高/最低/Top/Bottom”），也不要对连续分数做硬分档；仅用连续比较的中性表述；
- 语气专业、中立、可执行；输出中文 Markdown；若发现不满足上述约束，应在生成前自我纠偏。`
}

func buildUserPrompt(payloadJSON string) string {
	template := `
请仅依据下方提供的 JSON 数据，生成《选科战略分析报告》。不要输出推理过程，不得编造外部事实、学校或专业名单。

【任务与语境】
- 你的职责是为高中阶段（初二至高一）学生与家长提供**系统化选科决策支撑**。
- 生成的选科组合必须**符合当期新高考选科政策**，并与 meta.mode 一致；必须通过合法性自检，确保所有推荐学科均存在于 payload.subjects。

【数据与解释纪律】
- 只使用提供的 JSON 数据；关键处需**点名引用**：学科均值、RIASEC/OCEAN 维度均值、mode_support 指标。
- 引用维度均值时结合 count 给出稳定性提示：count=1 → “数据有限，仅供参考”；count=2 → “数据有限，不宜下绝对结论”；count≥3 → “相对较可靠但避免绝对化”。同一段落最多出现一次稳定性提示。
- deltas 仅表示家长与学生的**认知差异**，不得写成能力高低。
- 禁止使用排名化表述（如“第几名/最高/最低/Top/Bottom”），不做硬性分档；使用“更高/偏向/相对较强”等中性表述，避免“显著”。

【写作结构（5 段）】
1) 抬头信息
   - 学生编号（meta.student_id）
   - 分析日期（优先 meta.generated_at；缺失则用当前日期）
   - 核心发现：写 2–3 个关键洞察（如某些学科均值突出、RIASEC/OCEAN 特征明显）
   - 必须点名一次 b5_C 与一次 b5_N，并说明作用（如 b5_C=尽责性→坚持学习，b5_N=低→情绪更稳定→更能应对考试压力）
   - 本节末尾单独一行：效度：有效/存疑/无效（依据 validity_section.checks 的事实）

2) 最优选科组合及适配度评分
   - 输出 2 套组合（合法性自检，meta.mode 一致，学科存在于 payload.subjects）
   - 每套组合标注方向（偏文/偏理/偏工/偏艺）
   - 给出适配度评分（0–100 分，四舍五入）
   - 说明评分依据：结合 subjects 与 riasec 均值、mode_support 指标；不得引入外部标准
   - 每套组合附“类别级别”的专业方向预览和未来职业方向预测，细分到具体路径（如工学→新能源/机电/环境工程）

3) 推荐理由（逐套组合）
   - 开头写宏观引导：概述推荐方向的价值，并点名一次 mode_support 的总体或分向数值对比（如 3.67 vs 2.75）
   - 对每套组合，用自然语言解释：结合学科优势、兴趣人格特征、职业契合路径
   - 必须再次点名一次 b5_C 与一次 b5_N，并说明其对该组合的支持作用
   - 在理由中点出关键评分支撑点（如“物理均值 5.0 + b5_C=4.5 → 学科基础+学习坚持性”）
   - 若两套组合评分差 ≤3 分或关键维度差 <0.2，需说明“差异有限”，并提出决策锚点（志趣/地域/培养体系）

4) 发展风险与机会成本（逐套组合）
   - 点出关键挑战（如某些薄弱学科或人格风险）
   - 指出选择该组合可能放弃的方向（类别级别示例即可）
   - 如存在 deltas 且 |Δ| ≥ 1.0，需点名并说明（如“A 维差异≈1.2”），提出沟通/体验类缓解措施，并强调这是认知差异而非能力差距

5) 总结建议（聚焦参考，而非任务清单）
   - 归纳 2 条参考性建议：
     1. 哪类方向最适合优先考虑（如偏理+工学相关）
     2. 哪些方向可作为备选，以及学生/家长在选科过程中需保持关注的重点（如兴趣与学科匹配、情绪稳定性对压力应对的作用）
   - 建议应简洁、面向决策参考，而不是布置学习改进任务

【数据（JSON）】
${PAYLOAD}

【输出与自检】
- 输出 700–1200 字中文 Markdown 正文
- 文末加：“**数据来源**：request_id（meta.request_id） / 生成时间（meta.generated_at）”
- 自检清单：
  1) 所有组合与 meta.mode 一致，且学科均在 payload.subjects；
  2) “核心发现”与“推荐理由”各必须点名一次 b5_C 与一次 b5_N，且 b5_N=低→更稳定语义不得反转；
  3) 推荐理由需点出关键支撑点（至少包含学科数据+人格特质各一个）；
  4) 若存在 deltas 且 |Δ| ≥ 1.0，必须提及并声明为认知差异；
  5) count≤2 必须提示“数据有限/谨慎解读”，每段最多出现一次；
  6) 禁用分档/排名/显著字样，统一中性表述；
  7) 总结建议仅限于选科参考方向，不得扩展为学习任务清单。
`
	return strings.ReplaceAll(template, "${PAYLOAD}", payloadJSON)
}

// —— Step5 主流程 ——
// 读取 report_payload.json → 构建 System/User 提示词 → 调用 callDeepSeek（流式）→ 落盘 report.md
func step5(apiKey string) error {
	// 1) 读取 payload
	payloadPath := "report_payload.json"
	payloadJSON, err := readText(payloadPath)
	if err != nil {
		return fmt.Errorf("读取 %s 失败: %w", payloadPath, err)
	}

	// 2) 提取 StudentID 仅用于命名
	var pm payloadMeta
	_ = json.Unmarshal([]byte(payloadJSON), &pm)
	studentID := strings.TrimSpace(pm.Meta.StudentID)
	if studentID == "" {
		studentID = "unknown"
	}

	// 3) 构建提示词（分工清晰）
	systemPrompt := buildSystemPrompt()
	userPrompt := buildUserPrompt(payloadJSON)

	// 4) 选择模型（默认 chat；可通过环境变量切换 reasoner）
	model := "deepseek-chat"
	// model := "deepseek-reasoner"

	req := Request{
		Model:       model,
		Temperature: 0.2, // 建议保持 0.2：稳定 + 少量多样性；如输出机械可试 0.3，如偶有跑偏可降至 0.1
		MaxTokens:   8000,
		Stream:      true,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	fmt.Println("=== Step5 开始生成报告 ===")
	content := callDeepSeek(apiKey, req)

	// 5) 落盘
	ts := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("report_%s_%s.md", studentID, ts)
	if err := writeText(filename, content); err != nil {
		return fmt.Errorf("写入报告失败: %w", err)
	}
	fmt.Println("报告已保存到", filename)
	return nil
}

func TestStep5(apiKey string) {
	if err := step5(apiKey); err != nil {
		fmt.Println("Step5 执行失败:", err)
		return
	}
	fmt.Println("Step5 测试执行完成 ✅")
}
