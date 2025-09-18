package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("用法: go run stage1.go <API_KEY> <STUDENT_ID>")
		return
	}
	apiKey := os.Args[1]
	studentID := os.Args[2]

	// 生成 request_id
	requestID := uuid.New().String()
	fmt.Println("生成的 request_id:", requestID)

	url := "https://api.deepseek.com/chat/completions"
	method := "POST"

	// 构造 payload
	payload := fmt.Sprintf(`{
	  "messages": [
	    {
	      "role": "system",
	      "content": "你是一个心理测评专家，熟练掌握霍兰德职业兴趣理论 (HollandTheory)、舒伯生涯发展理论 (SuperTheory)、大五人格模型 (BigFive) 等心理学原理。请为山东省高一学生及其家长设计一套综合测评问卷：学生问卷必须包含 恰好 20 题，其中效度题必须恰好为 4 道，编号分布在 20 道学生题中，且不得全部集中在最后。家长问卷必须包含恰好 16 道题，其中 1 道为效度题，用于检测答卷态度。所有题目均采用 1–5 分 Likert 记分，1 表示完全不符合，5 表示非常符合。输出必须严格为合法JSON对象，必须包含字段 request_id, student_id, student_questions, parent_questions。其中 student_questions 与 parent_questions 必须是数组，每个题目对象必须包含字段 id, text, type, rev。不要输出任何额外解释或markdown，只返回JSON。type 字段仅允许使用 {R, I, A, S, E, C, O, N, D}，其中 RIASEC 对应霍兰德六维度，OCEAN 对应大五人格，D 对应效度题。禁止混用或新造缩写。"
	    },
	    {
	      "role": "user",
	      "content": "request_id: %s\nstudent_id: %s\n用户基本信息：性别：男，年级：高一。请生成符合要求的问卷。所有题干必须用流畅的简体中文，禁止输出英文或拼音。题干必须贴近中国高中生和家长的日常学习与生活场景，例如学习、社团、同学关系，而不是抽象表述。避免引导性或价值判断用语，保持中性。"
	    }
	  ],
	  "model": "deepseek-chat",
	  "temperature": 0.7,
	  "max_tokens": 2000,
	  "response_format": { "type": "json_object" },
	  "stream": false
	}`, requestID, studentID)

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
