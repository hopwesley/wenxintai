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

// —— System 提示词（角色 + 边界 + 底线约束） ——
// 不包含写作结构与细节，避免与 User 重复
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
- 禁用固定阈值分档与 Top/Bottom 排名措辞，使用连续比较与中性表述；
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
- 引用维度均值时结合 count 给出稳定性提示：count=1 → “数据有限，仅供参考”；count=2 → “数据有限，不宜下绝对结论”；count≥3 → “相对较可靠但避免绝对化”。同一段落最多出现一次稳定性提示，避免冗余。
- deltas 仅表示家长与学生的**认知差异**，不得写成能力高低。
- 禁用固定阈值分档与 Top/Bottom 排名；使用“更高/偏向/相对较强”等中性表述，避免“显著”。

【写作结构（严格 5 段）】
1) 抬头信息
   - 学生编号（meta.student_id）
   - 分析日期（优先 meta.generated_at；缺失则用当前日期）
   - 核心发现：写 2–3 个关键洞察（如某些学科均值突出、RIASEC/OCEAN 特征明显）
   - **必须点名一次 b5_C 与一次 b5_N，并说明作用（如 b5_C=尽责性→坚持学习，b5_N=低→情绪更稳定→更能应对考试压力）**
   - **本节末尾单独一行**：效度：有效/存疑/无效（依据 validity_section.checks 的事实）

2) 最优选科组合及适配度评分
   - **输出 2 套组合**（合法性自检，meta.mode 一致，学科存在于 payload.subjects）
   - 每套组合标注方向（偏文/偏理/偏工/偏艺）
   - 给出适配度评分（0–100 分，四舍五入）
   - 说明评分依据：遵循 payload.mode_support 权重口径；结合 subjects 与 riasec 均值加权；不得引入外部标准
   - 每套组合附“类别级别”的专业方向预览和未来职业方向预测，细分到具体路径（如工学→新能源/机电/环境工程）

3) 推荐理由（逐套组合）
   - 开头 **2 句**宏观引导：概述推荐方向的生涯价值，并**点名一次** mode_support 均值对比（如 3.67 vs 2.75）
   - 每套组合用 3–4 句给出理由：引用学科均值与 RIASEC/OCEAN（必要时提示 count≤2 的谨慎解读）
   - **必须再点名一次 b5_C 与一次 b5_N，并结合具体学科说明作用路径**
   - **每套组合理由部分必须有一句明确说明 1–2 个评分驱动因素**（必须同时包含学科数据 + 人格特质，如“物理均值 5.0 + b5_C=4.5 → 学科基础+学习坚持性”）
   - 若两套组合评分差 ≤3 分或维度差 <0.2，写明“差异有限”，并给出 3 类决策锚点（志趣/地域/培养体系）

4) 发展风险与机会成本（逐套组合）
   - **每套组合 2 条潜在挑战**（如薄弱学科、学习方式与人格风险，语言克制可执行）
   - **每套组合 2 条机会成本**（类别级别举例，如选择工科放弃文科→减少人文传播类路径）
   - 若 deltas |Δ| ≥ 1.0，必须点名量化（如“A 维 Δ≈1.2”），提出沟通/体验缓解措施，并声明为认知差异

5) 总结建议
   - **输出 2 条建议**，每条包含“五要素”：
     1. 行动：4–8 周可执行关键动作；
     2. 依据：引用本报告的学科 / RIASEC / OCEAN / mode_support 数据；
     3. 职业衔接：明确类别级别路径并提供 1 个可执行连接点（竞赛/实验/社团）；
     4. A/B 触发条件：写清去留判据；
     5. 责任与产出物标签（负责人/复核/产出物）
   - 每条建议末尾加时间锚点（如“第4周检查点：完成2次体验”）

【数据（JSON）】
${PAYLOAD}

【输出与自检】
- 输出 700–1200 字中文 Markdown 正文
- 文末加：“**数据来源**：request_id（meta.request_id） / 生成时间（meta.generated_at）”
- 自检清单：
  1) 所有组合与 meta.mode 一致，且学科均在 payload.subjects；
  2) “核心发现”与“推荐理由”各必须点名一次 b5_C 与一次 b5_N，且 b5_N 低=更稳定语义不得反转；
  3) 每套组合理由部分必须包含一句评分驱动因素说明（学科数据+人格特质）；
  4) deltas |Δ| ≥ 1.0 必须量化并提出缓解措施，并声明为认知差异；
  5) count≤2 必须提示“数据有限/谨慎解读”，每段最多出现一次；
  6) 禁用分档/排名/显著字样，统一中性表述；
  7) 职业衔接需具体到类别级别的细分路径。
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

// —— 对外测试入口 ——
// 在 main.go 的 switch 中添加： case "5": TestStep5(apiKey)
func TestStep5(apiKey string) {
	if err := step5(apiKey); err != nil {
		fmt.Println("Step5 执行失败:", err)
		return
	}
	fmt.Println("Step5 测试执行完成 ✅")
}
