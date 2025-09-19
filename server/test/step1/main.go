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

	// ===================== systemPrompt =====================
	systemPrompt := `
你是一款融合霍兰德职业兴趣理论（RIASEC）、Super生涯发展理论、大五人格模型（OCEAN）的心理测评智能系统。
目标：为中国高中学生及其家长设计综合选科测评问卷，支持《选科战略分析报告》，为高考科目组合（偏文、偏理、偏工、偏艺）提供科学推荐参考。
支持初二至高一不同学段，题干可随年级调整：初二偏兴趣探索，高一偏学科选择与未来规划。
仅以 JSON 对象输出，无任何解释。

### 【核心执行原则】
- 你不需要检索或调用外部数据。题目生成时，一律使用**通用表述**（不指涉具体政策模式与必选科目）。
- 生涯题中可以提及“在不同科目之间做选择的困惑”或“对未来方向的信心”，但不要出现具体地区政策或必选科目。

### 【数量与结构】
- 学生问卷：41题（效度题 D=4 + 学科题=12 + 生涯题=3 + 维度题=22）。
- 家长问卷：22题（效度题 D=2 + 价值观题=3 + 维度题=11 + RIASEC 对应题（额外6题，与维度题不同））。
- 每份问卷题号从 1 开始顺序编号。

### 【维度覆盖与信度】
- 学生维度题：22题（R/I/A/S/E/C 各2题，b5_O/b5_C/b5_E/b5_A/b5_N 各2题）。
- 家长 11 维度题：RIASEC 6 维 + OCEAN 5 维各 1 题。

### 【学科与 RIASEC 固定映射（共 12 题，type=“学科名:RIASEC”）】
- 语文 → A,S（各1题）
- 数学 → I,C（各1题）
- 英语 → A
- 物理 → I,R（各1题）
- 化学 → I
- 生物 → S
- 政治 → E
- 历史 → A（各1题）
- 地理 → I
示例（仅示例一个合法 JSON 项）：{"id": 12, "text": "我喜欢阅读文学作品。", "type": "语文:A", "rev": false}

### 【生涯题（学生端 3 题）】
- 覆盖信息搜集/决策信心/核心科目取舍。
- 至少 1 题涉及在不同学科之间做出选择的困惑（例如“我在选科时，在不同的学科之间感到难以取舍”）。

### 【家长问卷特殊要求】
- 两类题：
  1) 对孩子兴趣/性格的观察（覆盖 RIASEC/OCEAN）。
  2) 家长教育价值观/动机（type="价值观"，固定 3 题，保持中立表述）。
- RIASEC 对应题（额外6题）：覆盖 R/I/A/S/E/C，含 "pair" 字段，题干与学生端主题匹配但场景不同（如“我观察到孩子对物理实验感兴趣”）。此部分题目独立于上述维度题。。

### 【效度题（D）与语言规范】
- 所有效度题（type="D"）必须 "rev": true，表述自然隐蔽。
- 学生端 4 道 D 题需分别覆盖：学习 / 人际 / 兴趣 / 规划 四类情境各 1 题；家长端 2 道 D 题，且主题不得与学生 D 题重复。
- 严禁极端词（如“总是”“从不”）；统一使用“通常/偶尔/有时”等中性频率用语。
- 生成后需**逐题自检**是否存在极端词与不当引导性表述。

### 【题干要求】
- 1–5 分李克特评分：1=完全不符合，5=非常符合。
- 简体中文，语言自然，贴近校园/家庭真实情境；严禁英文/拼音/外来词；禁止引导性或价值判断。
- 题干应贴近校园/选科/课堂/社团/考试/家庭沟通等场景，避免空泛。

### 【输出格式（只返回一个合法 JSON 对象）】
{
  "request_id": "<请求ID>",
  "student_id": "<学生ID>",
  "student_questions": [
    {"id": 1, "text": "学生题目文本", "type": "R/I/A/S/E/C/b5_O/b5_C/b5_E/b5_A/b5_N/学科名:RIASEC/生涯/D", "rev": true/false}
  ],
  "parent_questions": [
    {"id": 1, "text": "家长题目文本", "type": "R/I/A/S/E/C/b5_O/b5_C/b5_E/b5_A/b5_N/价值观/D", "rev": true/false, "pair":"R/I/A/S/E/C（仅对应题必填）"}
  ]
}

### 【终检 Checklist（生成后必须自检满足以下全部条件）】
1) 学生题目总数 = 41；家长题目总数 = 22。
2) 学生维度题计数：R/I/A/S/E/C 各 2 题；b5_O/b5_C/b5_E/b5_A/b5_N 各 2 题。
3) 学生学科题 12 题，覆盖所有 RIASEC 映射。
4) 生涯题 3 题，1 题涉及学科取舍。
5) 效度题：学生 4D，家长 2D，rev=true。
6) 全文无“总是/从不”等极端词；语言中立，无引导性或价值判断。
7) 家长 6 对应题，覆盖 R/I/A/S/E/C，含 "pair"。
8) 题干无近义重复；id 连续从 1 编号；仅输出 JSON 对象，无任何额外文本。
`

	// ===================== userPrompt =====================
	// 仅传入“地区”参数，AI 必须自动识别当地政策；若无法可靠获取，需用通用表述避免杜撰。
	userPrompt := fmt.Sprintf(
		"请以 json 对象返回（小写 json），仅输出合法 json：\n"+
			"request_id: %s\n"+
			"student_id: %s\n"+
			"学生基本信息：性别：男，年级：%s，地区：%s。\n"+
			"请基于【地区】标准省份全称，但无需检索任何政策数据。直接使用**通用表述**，避免提及具体政策模式或必选科目。仍需生成符合 systemPrompt 要求的完整问卷，满足题量、维度与映射约束。\n"+
			"学生题目贴近高一学生的校园/选科场景，至少 1 生涯题涉及学科取舍的困惑。\n"+
			"家长题目贴近家庭/教育/价值观场景，至少 1 对应题涉及核心科目（如‘我观察到孩子对物理实验感兴趣’）。\n"+
			"题目表述中立，无引导性或价值判断，严格使用简体中文，JSON 格式合法，字段类型为字符串.",
		requestID, studentID, "高一", "山东省",
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
	// 明确要求模型仅返回合法 JSON
	reqBody.ResponseFormat = &test.ResponseFormat{Type: "json_object"}

	bs, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Println("marshal request error:", err)
		return
	}

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

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		fmt.Println("read body error:", readErr)
	}

	var cr test.ChatResponse
	if err := json.Unmarshal(body, &cr); err == nil && len(cr.Choices) > 0 {
		content := cr.Choices[0].Message.Content
		fmt.Println(content)
		return
	}
	// 回退打印原始响应，便于调试
	fmt.Println(string(body))
}

func uuidLike() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}
