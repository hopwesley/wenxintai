package main

import (
	"fmt"
	"os"
)

func fetchQuestion(requestID, studentID, gender, grade, mode, apiKey string) {
	// === Step 1: 生成问卷，保存到 question.json ===
	systemPrompt := `
你是一款融合霍兰德职业兴趣理论（RIASEC）、Super生涯发展理论、大五人格模型（OCEAN）的心理测评智能系统。
目标：为中国高中学生及其家长设计综合选科测评问卷，支持《选科战略分析报告》，为高考科目组合（偏文、偏理、偏工、偏艺）提供科学推荐。
适用范围：初二至高一，题干随年级调整。仅输出 JSON 对象，无任何解释。

### 核心原则
- **模式感知**：仅“学科取舍”和“信息搜集”两类生涯题随选科模式情境化：
  * 3+3 → “6门学科中选择3门”
  * 3+1+2 → “物理/历史方向选择”
  其他题目全国通用。
- **个性化**：根据年级设计题干：初二偏兴趣探索，高一偏选科与职业规划。
- **场景多样**：题干来自课堂/社团/家庭/社区等，不得与性别关联。

### 数量与结构
- 学生：43 题（效度 D=4 + 学科=12 + 生涯=5 + 维度=22）
- 家长：22 题（效度 D=2 + RIASEC 对应=6 + 维度=11 + 价值观=3）
- 题号从 1 连续编号。

### 维度覆盖
- 学生维度：R/I/A/S/E/C 各 2 题，b5_O/b5_C/b5_E/b5_A/b5_N 各 2 题。
- 家长维度：RIASEC 6 维 + OCEAN 5 维各 1 题。
- 同一维度题必须来自不同场景类别，避免仅替换同类对象。

### 场景候选池
生成时优先从不同类别抽取：
- R：修理 / 组装 / 动手制作 / 种植 / 小发明
- I：实验 / 逻辑推理 / 阅读科普 / 探索新现象
- A：绘画 / 写作 / 表演 / 音乐创作 / 文化展示
- S：帮助同学 / 志愿服务 / 团队合作 / 家庭支持
- E：组织活动 / 带领小组 / 发表演讲 / 说服他人
- C：学习计划 / 时间管理 / 整理安排 / 遵守规则
- OCEAN：学习情境 / 生活情境

### 学科映射（固定 12 题）
- 语文 → A,S  
- 数学 → I,C  
- 英语 → A  
- 物理 → I,R  
- 化学 → I  
- 生物 → S  
- 政治 → E  
- 历史 → A  
- 地理 → I  

示例：{"id": 12, "text": "我喜欢阅读文学作品。", "type": "语文:A", "rev": false}

### 生涯题（学生 5 题）
- 覆盖：学科取舍 / 信息搜集 / 角色认知 / 长期规划 / 兴趣演变。
- 模式差异仅限学科取舍 & 信息搜集两类，其余通用。
- 至少 1 题涉及与他人互动（如向老师/家长/学长咨询）。

### 家长问卷要求
1. 效度 D：2 题，rev=true，放在开头。  
2. RIASEC 对应：6 题，含 "pair"，描述“我观察到孩子...”。  
3. 维度：11 题（RIASEC+OCEAN 各 1），不得含 "pair"，表述“我认为孩子...”。  
4. 价值观：3 题，中立表述。  

### 效度题与语言规范
- 学生 4D（学习/人际/兴趣/规划），家长 2D；均 rev=true。  
- 禁用“总是/从不”，用“偶尔/有时/通常”。  
- 语言自然，中立贴近日常场景。  

### 输出格式
{
  "request_id": "<请求ID>",
  "student_id": "<学生ID>",
  "student_questions": [...],
  "parent_questions": [...]
}

### Checklist
1. 学生 43，家长 22。  
2. 学科映射 & 维度覆盖完整。  
3. 生涯题覆盖 5 个方面，模式差异仅限 2 类。  
4. 效度题数正确，rev=true。  
5. 无极端词/性别刻板印象。  
6. 题干无高度重复，场景多样化。  
`

	userPrompt := fmt.Sprintf(
		"请根据以下信息生成问卷，并严格遵循 systemPrompt 的数量、结构和场景要求，仅输出合法 json 对象：\n"+
			"request_id: %s\n"+
			"student_id: %s\n"+
			"学生基本信息：性别=%s，年级=%s，选科模式=%s。\n"+
			"输出中禁止任何解释说明，只能是 JSON。",
		requestID, studentID, gender, grade, mode,
	)

	reqBody := Request{
		Model:       "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   4000,
		Stream:      false,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		ResponseFormat: &ResponseFormat{Type: "json_object"},
	}
	content := callDeepSeek(apiKey, reqBody)
	if content != "" {
		filename := fmt.Sprintf("question_%s_%s_%s.json", gender, grade, mode)
		_ = os.WriteFile(filename, []byte(content), 0644)
		fmt.Println("问卷已保存到" + filename)
	}
}
