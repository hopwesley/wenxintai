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
	systemPrompt := `
你是心理测评专家（RIASEC/Super/OCEAN）。现在需要为山东省高一学生及其家长设计一份综合测评问卷。请严格按照以下要求生成问卷，并严格以 json 对象输出（注意“json”为小写），仅返回一个合法的 json。

【数量要求】
- 学生问卷：总题数 = 20。效度题（type = "D"）= 4。其他非效度题 = 16。不得出现超过 2 道效度题连续排列。
- 家长问卷：总题数 = 16。效度题（type = "D"）= 1。其他非效度题 = 15。
- 请把题目数量要求当成小学数学的加减法题，务必优先确保数字准确。

【效度题要求】
- 所有效度题的 rev 必须为 true。
- 效度题表述要自然隐蔽，可以涉及日常小习惯或细节（如“有时会忘记带水杯”“偶尔会忽略小细节”），但禁止使用“总是”“从不”“一定”“必须”等极端绝对表述。
- 效度题不能太直白，不要集中在迟到、忘记作业、误解老师等明显场景。

【维度覆盖要求】
- 学生问卷与家长问卷的非效度题必须分别覆盖 R、I、A、S、E、C、O、N 八个维度，每个维度至少一道题。
- 不同维度的题目内容需多样化，避免重复表达。

【题干内容要求】
- 所有题目均使用 1～5 分的李克特五级评分：1=完全不符合，5=非常符合。
- 题干必须为流畅的简体中文，贴近校园（学生问卷）或家庭（家长问卷）的真实场景。
- 严禁出现英文、拼音或外来词，违者输出不合格。
- 禁止引导性或价值判断用语。

【输出格式】
- 仅输出一个合法的 json 对象（小写 json）。
- 字段：request_id、student_id、student_questions、parent_questions。
- student_questions 和 parent_questions 均为数组，每道题为对象，包含：id（题目编号，从1开始顺序编号）、text（题干）、type（题目类型）、rev（是否反向计分）。

【检查与修正】
- 在输出前请自行检查：学生题数是否 = 20 且效度题 = 4；家长题数是否 = 16 且效度题 = 1；所有效度题是否 rev = true；非效度题是否覆盖了 8 个维度；是否禁止出现英文或拼音。
- 若发现不符合，请先修正，再输出结果。
- 最终仅返回合法 json，不得包含任何额外解释或说明。`

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
