package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("用法: go run main.go <MODE:1|2|3> <API_KEY> <STUDENT_ID>")
		return
	}
	mode := os.Args[1]
	apiKey := os.Args[2]
	studentID := os.Args[3]

	requestID := uuidLike()
	fmt.Println("生成的 request_id:", requestID)

	switch mode {
	case "1":
		fetchQuestion(requestID, studentID, "山东省", apiKey)
	case "2":
		calculateQuota(requestID, studentID)
	case "3":
		// === Step 3: quota.json + 提示词，调用 DeepSeek 生成报告 ===
		quotaBytes, err := os.ReadFile("quota.json")
		if err != nil {
			fmt.Println("读取 quota.json 出错:", err)
			return
		}

		systemPrompt := `
你是一款融合霍兰德职业兴趣理论、Super生涯发展理论和大五人格模型的心理测评智能系统。
任务：根据测评概要指标，生成《选科战略分析报告》。
报告格式：500~800字，JSON格式，包含：
1) 学生编号、分析日期、核心发现
2) 2~3个最优选科组合及适配度评分，包含专业方向、可报考大学专业群、未来职业方向
3) 推荐理由
4) 发展风险提示与机会成本
5) 总结建议
语言风格：简体中文，专业且偏建议性，避免极端词。
仅输出合法 JSON 对象。`

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
