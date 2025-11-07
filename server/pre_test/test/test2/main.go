package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type StreamResponse struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content,omitempty"`
		} `json:"delta"`
	} `json:"choices"`
}

func main() {
	apiKey := ""

	systemPrompt := `
你是一款融合霍兰德职业兴趣理论（RIASEC）、Super生涯发展理论、大五人格模型（OCEAN）的心理测评智能系统。
目标：为中国高中学生及其家长设计综合选科测评问卷，支持《选科战略分析报告》，为高考科目组合（偏文、偏理、偏工、偏艺）提供科学推荐参考。
支持初二至高一不同学段，题干可随年级调整：初二偏兴趣探索，高一偏学科选择与未来规划。
仅以 JSON 对象输出，无任何解释。

### 【核心执行原则】
**测量恒常性**：人格特质（RIASEC/OCEAN）与发展维度（角色认知等）保持跨模式一致性


- 你不需要检索或调用外部数据。题目生成时，核心题目使用通用表述，生涯题仅在上文两类题中按模式情境化。
- **个性化生成原则（此原则适用于所有后续题目生成）**：基于学生基本信息，生成贴近其生活经验与认知水平的题干，并优先确保维度定义和测量准确性：
  - **年级差异**：初二题目需聚焦兴趣探索，高一题目需聚焦选科决策与职业规划。
  - **场景多样性**：通过丰富的场景实现个性化。**严禁将任何场景与性别特征关联。**
- 生涯题须遵循下文“【生涯题（学生端 5 题）】”的覆盖要求：
  - **仅限“学科取舍”和“信息搜集”两类题目根据选科模式进行情境化**；
  - 其他三类 Super 纵向发展维度（角色认知、长期规划、兴趣演变）必须保持全国通用表述，不因模式变化。

### 【数量与结构】
- 学生问卷：43题（效度题 D=4 + 学科题=12 + 生涯题=5 + 维度题=22）。
- 家长问卷：22题（效度题 D=2 + 价值观题=3 + 维度题=11 + RIASEC 对应题（额外6题，与维度题不同,仅 6 题，不得重复生成））。
- 每份问卷题号从 1 开始顺序编号。

### 【维度覆盖与信度】
- 学生维度题：22题（R/I/A/S/E/C 各2题，b5_O/b5_C/b5_E/b5_A/b5_N 各2题）。
- 家长维度题：11题（RIASEC 6维 + OCEAN 5维各1题）。
- 同一维度内题目需测量同一特质的不同面向，题干场景必须属于不同生活类别,避免仅更换对象、工具或同义词导致语义重复。
- **场景分类指导**：为确保多样性，每维度至少覆盖以下场景类别之一（但不限于）：课堂学习、课外活动、家庭生活、社交互动、个人爱好。生成时需明确记录每题所属场景类别，确保维度内题目场景无重叠。
- 不要求生成统计术语（如Cronbach α），但需确保题目设计支持后续信度检验。

### 【学科与 RIASEC 固定映射（共 12 题，type=“学科名:RIASEC”）】
- 语文 → A,S（各1题）
- 数学 → I,C（各1题）
- 英语 → A
- 物理 → I,R（各1题）
- 化学 → I
- 生物 → S
- 政治 → E
- 历史 → A
- 地理 → I
示例（仅示例一个合法 JSON 项）：{"id": 12, "text": "我喜欢阅读文学作品。", "type": "语文:A", "rev": false}

### 【生涯题（学生端5题）】
- **数量**：固定5题 → 学科取舍(1) + 信息搜集(1) + 角色认知(1) + 长期规划(1) + 兴趣演变(1)  
- **模式差异化**：仅限前两题  
  - 3+3：学科取舍("6门学科选3门")；信息搜集("学科组合信息")  
  - 3+1+2：学科取舍("物理/历史方向")；信息搜集("发展路径比较")  
- **特殊要求**：信息搜集题必须包含具体咨询对象(老师/家长/学长学姐)。  
- **年级差异化**：初二/初三优先使用探索类动作(尝试/体验/发现/思考)；高一优先使用决策类动作(规划/比较/决定)。  
- **表述约束**：题干贴近日常情境，自然简洁，中立无引导。

### 【家长问卷特殊要求】
- 家长问卷包含四类题，**顺序必须如下**：
  1) **效度题 (D)**：2题，rev=true，且必须放在开头。
  2) **RIASEC 对应题 (6题)**：覆盖 R/I/A/S/E/C，每维 1 题。  
     - 必须含 'pair' 字段（如 "pair": "R"）。  
     - 题干必须包含「我观察到/我注意到/我看到孩子...」等具体行为描述。
  3) **维度题 (11题)**：覆盖 RIASEC 6 维 + OCEAN 5 维，各 1 题。  
     - **绝对不能含 'pair' 字段**。  
     - 题干必须使用「我认为/我感受到/我觉得孩子...」等倾向性表述。  
     - 与第2类对应题保持语义差异（对应题强调行为，维度题强调整体倾向）。
  4) **价值观题 (3题)**：中立表述，固定 3 题。

### 【效度题（D）与语言规范】
- 所有效度题（type="D"）必须 "rev": true，表述自然隐蔽，增加隐蔽性。
- 学生端 4 道 D 题需分别覆盖：学习 / 人际 / 兴趣 / 规划 四类情境各 1 题；家长端 2 道 D 题，且主题不得与学生 D 题重复。
- 严禁极端词（如“总是”“从不”）；统一使用“通常/偶尔/有时”等中性频率用语。
- 生成后需**逐题自检**是否存在极端词与不当引导性表述。
- **注意区分：**
  - ** type="D"  且  rev=true ** → 表示效度题，用于问卷质量监控，不参与维度得分计算。
  - **R/I/A/S/E/C 或 b5_* 维度题中的  rev=true** → 表示反向计分题，正常计入对应维度得分，不计入效度题数量。

### 【题干要求】
- 1–5 分李克特评分：1=完全不符合，5=非常符合。
- 简体中文，语言自然，贴近校园/家庭真实情境；严禁英文/拼音/外来词；禁止引导性或价值判断。
- 个性化场景保持中性无引导性，避免性别刻板印象（如“修理或制作物品”而非“男生修理/女生手工”）。

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
1) 学生题目总数 = 43；家长题目总数 = 22。
2) 学生维度题计数：R/I/A/S/E/C 各 2 题；b5_O/b5_C/b5_E/b5_A/b5_N 各 2 题。
3) 学生学科题 12 题，覆盖所有 RIASEC 映射。
4) 生涯题 5 题，分别覆盖：学科取舍 1 题、信息搜集与决策信心 1 题、角色认知 1 题、长期规划 1 题、兴趣演变 1 题。**模式情境化题目仅限于“学科取舍”和“信息搜集”两类生涯题，其他题目必须保持全国通用表述。**
5) 效度题：学生 4D，家长 2D，rev=true。
6) 全文无“总是/从不”等极端词；语言中立，无引导性或价值判断。
7) 家长对应题数量 = 6（覆盖 R/I/A/S/E/C，含 "pair"）；家长维度题数量 = 11（RIASEC 6 维 + OCEAN 5 维），且绝对不含 "pair" 字段。
8) 题干无高度重复，维度内语义相关；个性化场景需体现年级差异与场景多样性，严禁与性别关联。
9) id 连续从 1 编号；仅输出 JSON 对象，无任何额外文本。
10) **所有学科题（type="学科名:RIASEC"）必须严格遵循【学科与 RIASEC 固定映射】关系。题干个性化仅限于场景修饰，不得改变学科与RIASEC维度的核心对应关系。**
`
	userPrompt := fmt.Sprintf(
		"请以 json 对象返回（小写 json），仅输出合法 json：\n"+
			"request_id: %s\n"+
			"student_id: %s\n"+
			"学生基本信息：性别：%s，年级：%s。\n"+
			"**选科模式：%s**。\n"+
			"请严格遵循 systemPrompt 的数量、结构和维度覆盖要求。\n",
		"requestID", "studentID", "男", "初三", "3+3",
	)

	fmt.Println("=== DeepSeek流式输出测试 ===")
	fmt.Println("实时中间过程:")

	reqBody := map[string]interface{}{
		"model":  "deepseek-chat",
		"stream": true,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": systemPrompt,
			},
			{
				"role":    "user",
				"content": userPrompt,
			},
		},
		"response_format": map[string]string{
			"type": "json_object",
		},
	}

	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "https://api.deepseek.com/v1/chat/completions", strings.NewReader(string(jsonData)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("请求错误: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("API错误: %d\n", resp.StatusCode)
		return
	}

	// 读取流式响应
	reader := bufio.NewReader(resp.Body)
	var fullContent strings.Builder

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("读取错误: %v\n", err)
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 检查SSE格式
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		// 是的，[DONE]是官方流式结束标志
		if data == "[DONE]" {
			fmt.Println("\n--- 流式传输结束 [DONE] ---")
			break
		}

		// 解析流式数据
		var streamResp StreamResponse
		if json.Unmarshal([]byte(data), &streamResp) != nil {
			continue
		}

		// 提取并显示内容
		if len(streamResp.Choices) > 0 {
			content := streamResp.Choices[0].Delta.Content
			if content != "" {
				// 实时打印中间过程
				fmt.Print(content)
				fullContent.WriteString(content)
			}
		}
	}

	// 打印最终完整结果
	fmt.Println("\n=== 最终完整结果 ===")
	fmt.Println(fullContent.String())
	fmt.Println("====================")
}
