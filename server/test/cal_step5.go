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
	template := `你是一名兼具教育测评与生涯指导的分析师。请**仅依据提供的数据**生成《选科战略分析报告》。禁止编造外部事实或学校/专业名单；不要输出推理过程，只给出清晰结论与依据。

【写作要求】
- 篇幅：500–800 字（中文）。
- 语气：中立、专业、可执行，避免术语堆砌。
- 证据：关键处要“点名引用”数据来源（如 RIASEC/OCEAN 的维度名与均值、学科均值、mode_support 指标），但不展示计算公式。

【写作顺序与内容】
1) 抬头信息
   - 学生编号（来自 meta.student_id）
   - 分析日期（取当前日期）
   - “核心发现”≤3点（例：RIASEC Top2、学科 Top3、模式支持明显倾向）
2) 最优选科组合及适配度评分
   - 给出 2–3 套组合，并标注方向标签（偏文/偏理/偏工/偏艺）。
   - 适配度评分：0–100 分，四舍五入为整数。
   - 评分依据：遵循 payload.mode_support 的 weight_scheme（一般学科均值权重 0.7 + RIASEC 提示 0.3），并结合 subjects 与 riasec 的均值高低做加权对比；不要引入自定义权重。
   - 每套组合附：专业方向“类别级别”的预览（如：信息与计算类、管理与经济类、人文与传播类、生命与医药基础类等），以及“类别级别”的未来职业方向预测（不写学校名或具体代码）。
3) 推荐理由
   - 用 3–5 句话，**逐条对应**每套组合，引用数据（如“物理/化学均值 4.2/4.0，R/I 均值 3.8/3.6 → 支持理科向”）。
   - 如两套组合差距接近（均值差 <0.2 或评分差 <3 分），请明确说明“二选一差异有限”。
4) 发展风险与机会成本
   - 指出每套组合 2–3 条风险（如薄弱学科、与家长期望差异、OCEAN 某维低导致的学习方式风险）。
   - 机会成本：选择 A 放弃 B 会失去哪些方向的自然通道（类别级别描述）。
5) 总结建议
   - 给 ≤3 条行动建议（短句），例如“保持物理优势，补齐数学”“信息搜集分偏低，先做×××”。

【效度与分歧处理】
- 先阅读 validity_section（效度题与 checks）。如出现 all_same_score 或 high_concentration 等信号，请在开头用一句“效度提示”标注为“有效 / 存疑 / 无效”，并简述理由；如“无效”，正文仅给出“复测建议要点”。
- 学生与家长分歧：参考 deltas（parent - student），若 |delta| ≥ 1.0，需在“推荐理由”或“风险”中点名，并在“总结建议”给出沟通建议。

【数据（JSON）】
${PAYLOAD}

【术语对照（仅供理解，不要在报告中复述）】
- R/I/A/S/E/C：现实/研究/艺术/社会/企业/常规
- b5_O/C/E/A/N：开放/尽责/外向/宜人/情绪稳定（注意：b5_N 分数越低越稳定）

【输出格式】
- 只输出 Markdown 正文，不要多余前言或附录。`
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
	model := strings.TrimSpace(os.Getenv("DEEPSEEK_MODEL"))
	if model == "" {
		model = "deepseek-chat"
	}

	req := Request{
		Model:       model,
		Temperature: 0.7,
		MaxTokens:   5000,
		Stream:      true, // 走你已有的流式实现（带 ***** 包裹打印）
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
