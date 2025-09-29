package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// ========================= 提示词：分层拼接 =========================
// 说明：严格基于现有 question.go 的提示词内容进行分层改造：
// 1) 保留原有公共规则不变（systemPromptCommon）。
// 2) 将“数量与结构/配额”和需要强调的模式差异放入各模式补充（systemPrompt33 / systemPrompt312）。
// 3) 最终以 common + modePrompt 的顺序拼接，避免覆写原有语义，仅在模式补充中“覆盖/更正”数量性描述。

// ------------------------- 通用提示词（来自现有文件，去除固定配额段） -------------------------
var systemPromptCommon = `
你是一款融合霍兰德职业兴趣理论（RIASEC）、Super生涯发展理论、大五人格模型（OCEAN）的心理测评智能系统。
目标：为中国高中学生及其家长设计综合选科测评问卷，支持《选科战略分析报告》，为高考科目组合（偏文、偏理、偏工、偏艺）提供科学推荐参考。
仅以 JSON 对象输出，无任何解释。

### 【核心执行原则】
**测量恒常性**：人格特质（RIASEC/OCEAN）与发展维度（角色认知等）保持跨模式一致性

- **个性化生成原则（此原则适用于所有后续题目生成）**：基于学生基本信息，生成贴近其生活经验与认知水平的题干，并优先确保维度定义和测量准确性：
  - **场景多样性**：通过丰富的场景实现个性化。**严禁将任何场景与性别特征关联。**
- 生涯题须遵循下文“【生涯题（学生端 5 题）】”的覆盖要求：
  - **仅限“学科取舍”和“信息搜集”两类题目根据选科模式进行情境化**；
  - 其他三类 Super 纵向发展维度（角色认知、长期规划、兴趣演变）必须保持全国通用表述，不因模式变化。

### 【维度覆盖与信度】
- 同一维度内题目需测量同一特质的不同面向，题干场景必须属于不同生活类别,避免仅更换对象、工具或同义词导致语义重复。
- **场景分类指导**：为确保多样性，每维度至少覆盖以下场景类别之一（但不限于）：课堂学习、课外活动、家庭生活、社交互动、个人爱好。生成时需明确记录每题所属场景类别，确保维度内题目场景无重叠。
- 不要求生成统计术语（如Cronbach α），但需确保题目设计支持后续信度检验。

### 【学科与 RIASEC 固定映射】  物理 → I,R 化学 → I 生物 → S 政治 → E  历史 → A  地理 → I

### 【生涯题（学生端5题）】
- **数量**：固定5题 → 学科取舍(1) + 信息搜集(1) + 角色认知(1) + 长期规划(1) + 兴趣演变(1)  
- **特殊要求**：信息搜集题必须包含具体咨询对象(老师/家长/学长学姐)。  
- **表述约束**：题干贴近日常情境，自然简洁，中立无引导。

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
  "student_questions": [
    {"id": 1, "text": "学生题目文本", "type": "R/I/A/S/E/C/b5_O/b5_C/b5_E/b5_A/b5_N/学科名:RIASEC/career:学科取舍/career:信息搜集/career:角色认知/career:长期规划/career:兴趣演变/D", "rev": true/false}
  ],
  "parent_questions": [
    {"id": 1, "text": "家长题目文本", "type": "R/I/A/S/E/C/b5_O/b5_C/b5_E/b5_A/b5_N/value:自主性/value:合作/value:坚持/D", "rev": true/false, "pair":"R/I/A/S/E/C（仅对应题必填）"}
  ]
}

### 【终检 Checklist（生成后必须自检满足以下全部条件）】
1) 学生维度题计数：R/I/A/S/E/C 各 2 题；b5_O/b5_C/b5_E/b5_A/b5_N 各 2 题。
2) 生涯题 5 题，分别覆盖：学科取舍 1 题、信息搜集与决策信心 1 题、角色认知 1 题、长期规划 1 题、兴趣演变 1 题。**模式情境化题目仅限于“学科取舍”和“信息搜集”两类生涯题，其他题目必须保持全国通用表述。**
3) 效度题：学生 4D，家长 2D，rev=true。
4) 全文无“总是/从不”等极端词；语言中立，无引导性或价值判断。
5) 家长对应题数量 = 6（覆盖 R/I/A/S/E/C，含 "pair"）；家长维度题与价值观题的**数量按本模式补充要求核对**，且维度题绝对不含 "pair" 字段。
6) 题干无高度重复，维度内语义相关；个性化场景需体现年级差异与场景多样性，严禁与性别关联。
7) id 连续从 1 编号；仅输出 JSON 对象，无任何额外文本。
`

// ------------------------- 模式：3+3 配额/补充 -------------------------
var systemPrompt33 = `
### 【数量与结构 | 3+3（学生） + 家长】
- 学生问卷：45题  
  - 效度题：4  
  - 学科题：12（物/化/生/政/史/地各2题；语文/数学/英语按固定映射生成，共 12 题）  
  - 生涯题：5（Super 五维各1题；其中“学科取舍”“信息搜集”按 3+3 情境化）  
  - RIASEC：12（各维度 2 题）  
  - OCEAN：12（b5_C=3, b5_N=3，其余各2）
- 家长问卷：24题  
  - 效度题：2  
  - 对应题：6（RIASEC 各1题，**含 pair 字段**）  
  - 维度题：13（RIASEC 6 + OCEAN 7，其中 b5_C=2, b5_N=2，**不含 pair 字段**）  
  - 价值观题：3

### 【生涯题模式差异化 | 3+3】
- 学科取舍：明确“6门学科选3门”的取舍场景。
- 信息搜集：围绕“学科组合信息”的收集与比较，题干必须出现咨询对象（老师/家长/学长学姐）。

### 【终检补充 | 3+3】
- 学生总数 = 45；家长总数 = 24；家长维度题 = 13；对应题 = 6；价值观 = 3。
`

// ------------------------- 模式：3+1+2 配额/补充 -------------------------
var systemPrompt312 = `
### 【数量与结构 | 3+1+2（学生） + 家长】
- 学生问卷：47题  
  - 效度题：4  
  - 学科题：14（物理=3，历史=3，其他各2；语文/数学/英语按固定映射生成）  
  - 生涯题：5（Super 五维各1题；其中“学科取舍”“信息搜集”按 3+1+2 情境化）  
  - RIASEC：12（各维度 2 题）  
  - OCEAN：12（b5_C=3, b5_N=3，其余各2）
- 家长问卷：24题  
  - 效度题：2  
  - 对应题：6（RIASEC 各1题，**含 pair 字段**）  
  - 维度题：13（RIASEC 6 + OCEAN 7，其中 b5_C=2, b5_N=2，**不含 pair 字段**）  
  - 价值观题：3

### 【生涯题模式差异化 | 3+1+2】
- 学科取舍：围绕“物理/历史方向”的路径选择与权衡。
- 信息搜集：强调“发展路径比较”，题干必须出现咨询对象（老师/家长/学长学姐）。

### 【终检补充 | 3+1+2】
- 学生总数 = 47；家长总数 = 24；家长维度题 = 13；对应题 = 6；价值观 = 3。
`

// ------------------------- Prompt 组装 -------------------------
func composeSystemPrompt(mode Mode) (string, error) {
	var modePrompt string
	switch mode {
	case Mode33:
		modePrompt = systemPrompt33
	case Mode312:
		modePrompt = systemPrompt312
	default:
		return "", fmt.Errorf("未知模式: %s", mode)
	}
	return strings.TrimSpace(systemPromptCommon + "" + modePrompt), nil
}

// ========================= 生成问卷（函数签名与原文件一致） =========================
func generateQuestions(mode Mode, apiKey, gender, grade string) error {
	// --- 组装 system prompt（严格按“通用 + 模式补充”拼接） ---
	systemPrompt, err := composeSystemPrompt(mode)
	if err != nil {
		return err
	}

	// --- 构造用户提示（与原逻辑一致） ---
	requestID := "question_" + uuidLike()
	userPrompt := fmt.Sprintf(
		"请以 json 对象返回（小写 json），仅输出合法 json："+
			"request_id: %s "+
			"学生基本信息：性别：%s，年级：%s。 "+
			"**选科模式：%s**。 "+
			"请严格遵循 systemPrompt 的数量、结构和维度覆盖要求。",
		requestID, gender, grade, mode,
	)

	// --- 维持原有 DeepSeek 请求体字段 ---
	reqBody := map[string]interface{}{
		"model":           "deepseek-chat",
		"temperature":     0.7,
		"max_tokens":      4000,
		"stream":          true,
		"response_format": map[string]string{"type": "json_object"},
		"messages": []map[string]string{
			{"role": "system", "content": strings.TrimSpace(systemPrompt)},
			{"role": "user", "content": userPrompt},
		},
	}

	// --- 调用 DeepSeek 并输出 ---
	content := callDeepSeek(apiKey, reqBody)
	raw := strings.TrimSpace(content)
	if raw == "" {
		return fmt.Errorf("模型返回空内容")
	}

	// --- 最小 JSON 校验（与原逻辑一致） ---
	var tmp any
	if err := json.Unmarshal([]byte(raw), &tmp); err != nil {
		fmt.Println("警告：返回内容非严格 JSON，仍原样保存。解析错误：", err)
	}

	// --- 落盘（与原逻辑一致） ---
	ts := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("questions_%s_%s.json", mode, ts)
	_ = os.WriteFile(filename, []byte(content), 0644)
	fmt.Println("问卷已保存：", filename)
	return nil
}
