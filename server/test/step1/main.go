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
	systemPrompt := `你是心理测评专家（RIASEC/Super/OCEAN）。现在需要为一名山东省高一学生及其家长设计一份综合测评问卷。请严格按照以下要求生成问卷，并严格以 JSON 对象输出（注意“json”为小写），仅返回一个合法的 JSON。

【结构要求】

学生问卷：共20题。其中效度题（type = 'D'）恰好4题，其余16题为非效度题。注意效度题不允许连续出现3题及以上。

家长问卷：共16题。其中效度题（type = 'D'）恰好1题，其余15题为非效度题。

所有效度题（D）的 rev 字段必须为 true，且效度题的表述要自然隐蔽，避免使用“总是”“从不”等极端词语。

重要提示：请将以上数量要求视为简单的加减计算题，务必首先确保每份问卷的总题数和效度题数量准确无误，然后再满足下面的维度覆盖要求。

【维度覆盖要求】

学生问卷和家长问卷中，所有非效度题必须分别覆盖 R、I、A、S、E、C、O、N 八种类型维度，每种类型的题目至少各有一道。（效度题 D 独立成类，不计入维度覆盖范围）

【题目内容要求】

使用1～5分的李克特五级评分：1分表示“完全不符合”，5分表示“非常符合”。

题干语言需使用流畅的简体中文，贴近校园（学生问卷）或家庭（家长问卷）的真实场景。不得出现任何英文或拼音词语，且避免明显的引导性或价值判断。

问卷题目的类型 type 必须是上述九种取值之一：{R, I, A, S, E, C, O, N, D}。其中 R/I/A/S/E/C/O/N 分别代表问卷内容的八个维度类型，D 表示效度题。rev = true 表示该题为反向计分题（选项评分需反转：1↔5，2↔4，3不变）。

【输出格式】

最终只输出一个 JSON 对象（注意使用小写 “json”），包含以下字段：request_id、student_id、student_questions、parent_questions。

其中 student_questions 和 parent_questions 为数组，分别列出学生问卷和家长问卷的所有题目。每个题目用一个 JSON 对象表示，包含字段：id（题目序号），text（题干文本），type（题目类型），rev（是否反向计分）。请分别在学生问卷和家长问卷中对题目 id 从1开始顺序编号。

【检查与修正】

在给出最终答案前，请自行检查生成的问卷是否满足上述所有要求，尤其是题目总数和效度题数量是否正确无误。如果发现不符合要求，请修正后再输出结果。

最终答案只需输出符合指定格式要求的 JSON，对问卷内容不作任何额外解释或评论，不得出现除 JSON 以外的其他文本。`

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
