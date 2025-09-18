package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/hopwesley/wenxintai/server/test"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("用法: go run main.go <API_KEY> <STUDENT_ID>")
		return
	}
	apiKey := os.Args[1]
	studentID := os.Args[2]

	requestID := uuidLike()
	fmt.Println("生成的 request_id:", requestID)

	// 重要：在提示词中显式包含“小写 json”关键字，满足 response_format 的校验
	systemPrompt := `你是心理测评专家(RIASEC/Super/OCEAN)。
请为山东省高一学生及其家长设计综合测评问卷，并严格以 json 对象输出（注意此处为小写 json），只返回一个合法的 json。
【结构】
- 学生问卷20题，其中效度题(type='D')恰好4道，且不允许连续≥3道效度题；
- 家长问卷16题，其中效度题(type='D')恰好1道；
- 所有效度题rev必须为true，且表述需自然隐蔽，避免“总是/从不”等极端词。
【维度覆盖】
- 学生与家长问卷各自覆盖R/I/A/S/E/C/O/N八维度（效度D单独计，不计入覆盖）。
【题目要求】
- Likert 1-5分：1=完全不符合，5=非常符合；
- 题干用流畅简体中文，贴近校园/家庭场景；禁止英文/拼音/引导性/价值判断；
- type∈{R,I,A,S,E,C,O,N,D}，rev=true表示反向(1↔5,2↔4,3=3)。
【输出格式】
- 仅输出一个合法的 json 对象，且仅包含：request_id, student_id, student_questions, parent_questions；
- student_questions/parent_questions 为数组，元素为{id,text,type,rev}。
【强制校验】
- 输出前自检效度题数量和总题数；若不满足（学生题≠20或D≠4，家长题≠16或D≠1）须修正后再输出；
- 不得返回除json以外的任何内容，题干中禁止出现任何英文。`

	userPrompt := fmt.Sprintf(
		"请以 json 对象返回（小写 json），仅输出合法 json：request_id: %s\nstudent_id: %s\n用户基本信息：性别：男，年级：高一。生成符合要求的问卷；题干需贴近学习/社团/同学关系等场景，避免引导性或价值判断。",
		requestID, studentID,
	)

	reqBody := test.Request{
		Model:       "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   4000,
		Stream:      false,
		Messages: []test.Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}
	// 必须启用 response_format
	reqBody.ResponseFormat = &test.ResponseFormat{Type: "json_object"}

	bs, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Println("marshal request error:", err)
		return
	}

	fmt.Println("完整请求体:")
	fmt.Println(string(bs))

	client := &http.Client{Timeout: 120 * time.Second}
	httpReq, _ := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewReader(bs))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Accept", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		fmt.Println("request error:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("HTTP status:", resp.Status)
	fmt.Println("---- Response headers ----")
	for k, v := range resp.Header {
		fmt.Println(k+":", v)
	}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		fmt.Println("read body error:", readErr)
	}
	fmt.Println("Body length:", len(body))
	fmt.Println("完整响应:", string(body))

	var cr test.ChatResponse
	if err := json.Unmarshal(body, &cr); err == nil && len(cr.Choices) > 0 {
		content := cr.Choices[0].Message.Content
		fmt.Println(content)
		checkBasic(content)
		return
	}
	fmt.Println(string(body))
}

func uuidLike() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

func checkBasic(content string) {
	var out test.Out
	if err := json.Unmarshal([]byte(content), &out); err != nil {
		fmt.Println("[check] 无法解析assistant内容为JSON：", err)
		return
	}
	fmt.Println("[check] 学生题数/家长题数：", len(out.StudentQuestions), len(out.ParentQuestions))

	countD := func(arr []test.Item) int {
		n := 0
		for _, it := range arr {
			if it.Type == "D" {
				n++
			}
		}
		return n
	}
	okRevForD := func(arr []test.Item) bool {
		for _, it := range arr {
			if it.Type == "D" && !it.Rev {
				return false
			}
		}
		return true
	}
	coverAll := func(arr []test.Item) bool {
		need := map[string]bool{"R": true, "I": true, "A": true, "S": true, "E": true, "C": true, "O": true, "N": true}
		for _, it := range arr {
			if need[it.Type] {
				delete(need, it.Type)
			}
		}
		return len(need) == 0
	}
	no3ConsecutiveD := func(arr []test.Item) bool {
		run := 0
		for _, it := range arr {
			if it.Type == "D" {
				run++
				if run >= 3 {
					return false
				}
			} else {
				run = 0
			}
		}
		return true
	}

	fmt.Printf("[check] 学生端 D=%d (需=4), rev(D)=OK? %v, 维度齐全? %v, D不>=3连? %v\n",
		countD(out.StudentQuestions), okRevForD(out.StudentQuestions), coverAll(out.StudentQuestions), no3ConsecutiveD(out.StudentQuestions))
	fmt.Printf("[check] 家长端 D=%d (需=1), rev(D)=OK? %v, 维度齐全? %v\n",
		countD(out.ParentQuestions), okRevForD(out.ParentQuestions), coverAll(out.ParentQuestions))
}
