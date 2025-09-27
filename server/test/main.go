package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run main.go <MODE:1|2|3> <API_KEY> <STUDENT_ID>")
		return
	}
	mode := os.Args[1]

	requestID := uuidLike()
	fmt.Println("生成的 request_id:", requestID)

	switch mode {
	case "1":

		apiKey := os.Args[2]
		studentID := os.Args[3]

		students := []struct {
			id, gender, grade, mode string
		}{
			{studentID, "男", "高一", "3+3"},
			{studentID + "1", "女", "初三", "3+1+2"},
		}

		for i, s := range students {
			rid := fmt.Sprintf("%s_%d", requestID, i)
			fetchQuestion(rid, s.id, s.gender, s.grade, s.mode, apiKey)
		}
	case "2":
		questionFile := os.Args[2]
		idx, err := strconv.Atoi(os.Args[3])
		if err != nil {
			panic(err)
		}
		calculateQuota(questionFile, "", idx)

	case "3":
		TestStep1()

	case "4":
		TestStep2()

	case "5":
		TestStep3()

	case "6":
		// === Step 3: quota.json + 提示词，调用 DeepSeek 生成报告 ===
		quotaBytes, err := os.ReadFile("quota.json")
		if err != nil {
			fmt.Println("读取 quota.json 出错:", err)
			return
		}

		systemPrompt := `
你是一款融合霍兰德职业兴趣理论、Super生涯发展理论和大五人格模型的心理测评智能系统。
任务：根据测评概要指标，生成《选科战略分析报告》。
你将收到一份 quota.json，里面包含学生的测评结果。
字段含义如下：
- R/I/A/S/E/C：霍兰德职业兴趣六维度
  - R=现实型（喜欢动手、实物操作）
  - I=研究型（喜欢探索、逻辑推理）
  - A=艺术型（喜欢创作、表达）
  - S=社会型（喜欢助人、交流）
  - E=企业型（喜欢组织、说服）
  - C=常规型（喜欢秩序、计划）
- b5_O/b5_C/b5_E/b5_A/b5_N：大五人格五维度
  - O=开放性
  - C=尽责性
  - E=外向性
  - A=宜人性
  - N=情绪稳定性（分数越低越稳定）
- 学科字段：是对应学科兴趣得分，例如 "语文":4.5 (高)，"数学":2.0 (低)。
- career_items：生涯发展题目的加总分（角色认知、规划能力、兴趣演变等）。
- validity_check：问卷效度检查结果。

报告格式：500~800字，JSON格式，包含：
1) 学生编号、分析日期、核心发现
2) 2~3个最优选科组合及适配度评分，包含专业方向、可报考大学专业群、未来职业方向
3) 推荐理由
4) 发展风险提示与机会成本
5) 总结建议
语言风格：简体中文，专业且偏建议性，避免极端词。
仅输出合法 JSON 对象。`

		apiKey := os.Args[2]
		studentID := os.Args[3]

		userPrompt := fmt.Sprintf("学生ID:%s\n概要指标如下:\n%s\n请生成分析报告。", studentID, string(quotaBytes))

		reqBody := Request{
			Model:     "deepseek-reasoner",
			MaxTokens: 1800,
			Stream:    false,
			Messages: []Message{
				{Role: "system", Content: systemPrompt},
				{Role: "user", Content: userPrompt},
			},
			ResponseFormat: &ResponseFormat{Type: "json_object"},
		}
		content := callDeepSeek(apiKey, reqBody)
		if content != "" {
			fmt.Println("报告输出:")
			fmt.Println(content)
		}

	default:
		fmt.Println("无效模式: 请输入 1 | 2 | 3")
	}
}
