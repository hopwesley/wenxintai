package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// 读取 payload 时只为取 student_id，避免与既有类型重名
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

// —— 构建整段提示词（反引号原始字符串） ——
// 将 report_payload.json 的全文 JSON 作为 ${PAYLOAD} 注入到模板中。
func buildReportPrompt(payloadJSON string) string {
	template := `

你是一名兼具教育测评与生涯指导的分析师。请**仅依据提供的数据**生成《选科战略分析报告》。禁止编造外部事实或学校/专业名单；不要输出推理过程，只给出清晰结论与依据。

【写作要求】
- 篇幅：700–1200 字（中文）。
- 语气：中立、专业、可执行，避免术语堆砌与情绪化表述。
- 证据：关键处要“点名引用”数据来源（如 RIASEC/OCEAN 的维度名与均值、学科均值、mode_support 指标），但不展示计算公式或阈值细节。
- 若数据中提供了 *_ordered 视图（按维度/学科名排序的数组），请仅以 *_ordered 为准进行读取与引用，不依赖对象键的自然顺序。
- **在“核心发现/推荐理由”中，至少各点名一次 OCEAN 关键维度（如 b5_C、b5_N）与 mode_support 的总体均值对比（如 3.67 vs 2.75），不展示计算过程。**
- **措辞约束**：避免使用具有统计学含义的“显著”一词，统一使用“明显/更高/更低/更偏向”等中性表述。

【写作顺序与内容】（严格 5 段式）
1) 抬头信息
   - 学生编号（来自 meta.student_id）
   - 分析日期（**优先使用 meta.generated_at；如缺失再取当前日期**）
   - “核心发现”≤3 点（例：RIASEC Top2、学科 Top3、模式支持明显倾向；**须至少点名一次 OCEAN 关键维度**）
   - **在本部分末尾单独一行输出**：效度：有效/存疑/无效（基于 validity_section.checks 的事实简述理由）。

2) 最优选科组合及适配度评分
   - 给出 2–3 套组合，并标注方向标签（偏文/偏理/偏工/偏艺）。
   - 适配度评分：0–100 分，四舍五入为整数。
   - 评分依据：遵循 payload.mode_support 的 weight_scheme（一般学科均值权重 0.7 + RIASEC 提示 0.3），并结合 subjects 与 riasec 的均值高低做加权对比；不要引入自定义权重或外部标准。
   - 每套组合附：专业方向“**类别级别**”的预览（如：信息与计算类、管理与经济类、人文与传播类、生命与医药基础类等），以及“**类别级别**”的未来职业方向预测（不写学校名或具体专业代码）。

3) 推荐理由（逐套组合）
   - **本节开头 2–3 句的宏观引导**：仅针对“已被推荐的方向”（如偏理/偏工），简述其对国家社会与个人生涯的价值与路径，**不要描述未入选方向**；在这 2–3 句中**点名引用一次 mode_support 的总体或分向数值**（如理科向/历史向学科均值 3.67 vs 2.75）。
   - 随后对每套组合分别用 3–5 句话给出理由，引用数据（如“物理/化学均值 5.0/4.0，R/I 均值 3.5/3.0 → 支持理科向”）。
   - 若两套组合评分差 **≤3 分** 或维度均值差 **<0.2**，**明确说明“二选一差异有限”**，并给出**三类决策锚点**：**志趣方向**（如 生命科学 vs 资源环境） / **升学地域**（如 沿海 vs 内陆） / **培养体系**（如 研究导向 vs 工程导向）。若评分差 **≥4 分**，避免使用“差距有限”的措辞，直接给出“组合 X 更优”的判断。
   - **在本节中（除宏观引导外）至少再点名一次 OCEAN 关键维度（如 b5_C、b5_N），并再次引用一次 mode_support 的总体均值对比（如 3.67 vs 2.75）。**

4) 发展风险与机会成本（逐套组合）
   - 指出每套组合 2–3 条风险（如薄弱学科、与家长期望差异、OCEAN 某维低导致的学习方式风险），语言克制、可执行。
   - **家长-学生差异量化**：如 payload 存在 deltas（parent - student）且 |Δ| ≥ 1.0，需在本节**点名并量化**（如 “A、S 维 Δ≈1–1.5”），并提出**沟通/体验**类缓解动作（如一次跟岗/访谈/实验室体验），形成记录。
   - **机会成本（对称、具体到类别）**：选择 A 放弃 B 可能减少哪些方向的自然通道，请用**类别级别**举例补充一小句（如：“放弃资源环境路径，将减少地理信息/GIS 类岗位的自然入口”）。两套组合的机会成本陈述保持对称性，避免绝对化结论。

5) 总结建议（**对核心建议及其未来职业衔接进行深化阐述**）
   - **输出 ≤3 条“核心建议”**，每条采用**五要素**并各 1–2 句完成：
     1. **行动**：4–8 周内可执行的关键动作；
     2. **依据**：点名引用本报告中的学科 / RIASEC / OCEAN 或 mode_support 证据（如 “b5_C 高 + 理科向均值 3.67 支持 ××”）；
     3. **职业衔接**：说明该行动与**类别级别**专业/职业路径的对应关系（高中阶段可参与的竞赛/项目/实验/志愿/社团 → 本科方向 → 初期职业场景），给出 1 个**可执行连接点**（如“完成××比赛初赛/进入××实验室助理/完成 GIS 入门实践 1 次”）；
     4. **A/B 触发条件**：给出“去/留”条件（如 “若数学 4 周内周测 <70% → 优先偏工；若完成 1 次生命科学实验并达标 → 维持偏理”）；
     5. **责任与产出物标签**：**负责人**（学生/科任/班主任） / **复核**（家长或教师） / **产出物**（如 1 页对比表、访谈纪要、实验报告、学习记录卡）；
   - 每条建议在行动末尾加一个**时间锚点**（如“第 4 周检查点：提交对比表/完成 2 次体验/达成周测目标”）；**总数不超过 3 条**。

【数据（JSON）】
${PAYLOAD}

【术语对照（仅供理解，不要在报告中复述）】
- R/I/A/S/E/C：现实/研究/艺术/社会/企业/常规
- b5_O/C/E/A/N：开放/尽责/外向/宜人/情绪稳定（**注意：b5_N 分数越低越稳定**）

【输出格式】
- 只输出 Markdown 正文，不要多余前言或附录。
- 在文末以小号一行添加：“**数据来源**：request_id（来自 meta.request_id） / 生成时间（meta.generated_at）”。

【生成前自检（务必满足，否则请重写不合规段落）】
1) 任何“高于/低于/更强/更弱”的比较，先核对对应数值大小方向一致；
2) **推荐理由**中包含宏观引导 2–3 句，且在宏观引导与逐套理由部分**各出现一次** mode_support 的总体均值对比（如 3.67 vs 2.75）；
3) “核心发现”与“推荐理由”**至少各出现一次** OCEAN 关键维度（优先 b5_C、b5_N），且“b5_N 低=情绪更稳定”的语义**不得反转**；
4) “风险与机会成本”中如存在 deltas 且 |Δ| ≥ 1.0，**必须点名量化**并附沟通/体验的缓解动作；
5) “总结建议”严格输出 **≤3 条核心建议**，且每条包含**行动/依据/职业衔接/A-B 触发条件/责任与产出物标签**五要素，并带**时间锚点**；
6) 字数 700–1200；
7) 第 1 部分末尾单行输出“效度：……”；
8) 文末输出“数据来源：request_id / meta.generated_at”（无需展示计算过程）；
9) 全文避免使用“显著”字样，统一改为“明显/更高/更低/更偏向”等中性表述；
10) 如存在 *_ordered 视图，引用顺序以 *_ordered 为准。


`

	return strings.ReplaceAll(template, "${PAYLOAD}", payloadJSON)
}

// —— Step5 主流程 ——
// 读取 report_payload.json → 构建提示词 → 调用 callDeepSeek（流式）→ 落盘 report.md
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

	// 3) 构建提示词
	prompt := buildReportPrompt(payloadJSON)

	// 4) 选择模型（默认 chat；设置环境变量 DEEPSEEK_MODEL=deepseek-reason 可切换）
	model := "deepseek-chat"
	//model := "deepseek-reasoner"

	req := Request{
		Model:       model,
		Temperature: 0.1,
		MaxTokens:   8000,
		Stream:      true,
		Messages: []Message{
			{Role: "system", Content: "你是教育测评与生涯指导的分析师。"},
			{Role: "user", Content: prompt},
		},
		// 报告是文本，不用 json_object
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

// —— 对外的测试入口 ——
// 在 main.go 的 switch 中添加： case "5": TestStep5(apiKey)
func TestStep5(apiKey string) {
	if err := step5(apiKey); err != nil {
		fmt.Println("Step5 执行失败:", err)
		return
	}
	fmt.Println("Step5 测试执行完成 ✅")
}
