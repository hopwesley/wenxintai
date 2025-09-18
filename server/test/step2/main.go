package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("用法: go run stage2.go <API_KEY> <REQUEST_ID> <STUDENT_ID>")
		return
	}
	apiKey := os.Args[1]
	requestID := os.Args[2]
	studentID := os.Args[3]

	url := "https://api.deepseek.com/chat/completions"
	method := "POST"

	// ⚠️ 示例答题数据（需要替换成真实提交的答案）
	studentAnswers := `[
	  {"id":1,"score":4},{"id":2,"score":3},{"id":3,"score":3},{"id":4,"score":4},{"id":5,"score":4},
	  {"id":6,"score":4},{"id":7,"score":3},{"id":8,"score":2},{"id":9,"score":5},{"id":10,"score":1},
	  {"id":11,"score":4},{"id":12,"score":3},{"id":13,"score":2},{"id":14,"score":4},{"id":15,"score":3},
	  {"id":16,"score":5},{"id":17,"score":2},{"id":18,"score":5},{"id":19,"score":4},{"id":20,"score":3}
	]`

	parentAnswers := `[
	  {"id":1,"score":3},{"id":2,"score":4},{"id":3,"score":3},{"id":4,"score":4},{"id":5,"score":3},
	  {"id":6,"score":4},{"id":7,"score":2},{"id":8,"score":4},{"id":9,"score":3},{"id":10,"score":3},
	  {"id":11,"score":3},{"id":12,"score":4},{"id":13,"score":4},{"id":14,"score":3},{"id":15,"score":5},
	  {"id":16,"score":2}
	]`

	payload := fmt.Sprintf(`{
  "messages": [
    {
      "role": "system",
      "content": "你是心理与教育测评专家，熟悉高考选科政策。你必须严格依据阶段 1 的问卷题目进行计分：按题目 id 逐一匹配学生与家长答案，并使用题目中的 type（RIASEC / OCEAN / D）与 rev（反向计分：1↔5, 2↔4, 3=3）进行计算。综合霍兰德职业兴趣 (RIASEC)、大五人格 (OCEAN)、舒伯生涯发展理论进行分析。\n\n输出要求：\n- 严格返回合法 JSON 对象，不包含任何解释、推理或 Markdown。\n- JSON 必须包含字段：request_id, student_id, analysis_date, core_findings, optimal_combinations, future_pathways, recommendations, risk_warning, summary。\n- 字段细则：\n  - analysis_date：YYYY-MM-DD\n  - optimal_combinations：数组（2–3 项），每项含 combo（如“偏文/偏理/偏工/偏艺”）、score（0–100）、preview（简短说明）\n  - future_pathways：数组，每项含 direction、related_majors（数组）、potential_careers（数组）\n- 字数限制：core_findings + recommendations + risk_warning + summary 四个文本字段合计 500–800 字（仅统计这四个字段）。"
    },
    {
      "role": "user",
      "content": "request_id: %s\nstudent_id: %s\n性别: 男, 年级: 高一\n学生答案（JSON 数组，元素 {id, score}）：%s\n家长答案（JSON 数组，元素 {id, score}）：%s\n请按上述约束生成《选科战略分析报告》，只返回 JSON。"
    }
  ],
  "model": "deepseek-reasoner",
  "temperature": 0.5,
  "max_tokens": 1800,
  "response_format": { "type": "json_object" },
  "stream": false
}`, requestID, studentID, studentAnswers, parentAnswers)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, strings.NewReader(payload))
	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+apiKey)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	fmt.Println(string(body))
}
