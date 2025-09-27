package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// ScoreResult：维度/学科/价值观的得分结果
type ScoreResult struct {
	Mean  float64 `json:"mean"`
	Sum   int     `json:"sum"`
	Count int     `json:"count"`
}

// QuotaDataset：Step3 输出结果
type QuotaDataset struct {
	Meta struct {
		RequestID   string `json:"request_id"`
		StudentID   string `json:"student_id"`
		Mode        string `json:"mode"`
		GeneratedAt string `json:"generated_at"`
	} `json:"meta"`
	StudentScores struct {
		RIASEC   map[string]ScoreResult `json:"RIASEC"`
		OCEAN    map[string]ScoreResult `json:"OCEAN"`
		Subjects map[string]ScoreResult `json:"subjects"`
		Career   map[string]int         `json:"career"`
	} `json:"student_scores"`
	ParentScores struct {
		RIASEC map[string]ScoreResult `json:"RIASEC"`
		OCEAN  map[string]ScoreResult `json:"OCEAN"`
		Values map[string]ScoreResult `json:"values"`
	} `json:"parent_scores"`
}

// Step3 主流程
func step3(cleaned CleanedDataset) error {
	quota := QuotaDataset{}
	quota.Meta.RequestID = cleaned.Meta.RequestID
	quota.Meta.StudentID = cleaned.Meta.StudentID
	quota.Meta.Mode = cleaned.Meta.Mode
	quota.Meta.GeneratedAt = time.Now().Format(time.RFC3339)

	// 初始化 map
	quota.StudentScores.RIASEC = make(map[string]ScoreResult)
	quota.StudentScores.OCEAN = make(map[string]ScoreResult)
	quota.StudentScores.Subjects = make(map[string]ScoreResult)
	quota.StudentScores.Career = make(map[string]int)

	quota.ParentScores.RIASEC = make(map[string]ScoreResult)
	quota.ParentScores.OCEAN = make(map[string]ScoreResult)
	quota.ParentScores.Values = make(map[string]ScoreResult)

	// 学生端
	for _, it := range cleaned.StudentCleaned {
		switch it.Category {
		case "RIASEC":
			updateScore(quota.StudentScores.RIASEC, it.Type, it.ScoreFinal)
		case "OCEAN":
			updateScore(quota.StudentScores.OCEAN, it.Type, it.ScoreFinal)
		case "subject":
			// 学科类: type= "语文:A" 取学科名
			subject := extractSubject(it.Type)
			updateScore(quota.StudentScores.Subjects, subject, it.ScoreFinal)
		case "career":
			// 生涯题直接保留原始分数
			quota.StudentScores.Career[it.Type] = it.ScoreFinal
		}
	}

	// 家长端
	for _, it := range cleaned.ParentCleaned {
		switch it.Category {
		case "RIASEC":
			updateScore(quota.ParentScores.RIASEC, it.Type, it.ScoreFinal)
		case "OCEAN":
			updateScore(quota.ParentScores.OCEAN, it.Type, it.ScoreFinal)
		case "value":
			updateScore(quota.ParentScores.Values, it.Type, it.ScoreFinal)
		}
	}

	// 计算均值
	finalizeScores(quota.StudentScores.RIASEC)
	finalizeScores(quota.StudentScores.OCEAN)
	finalizeScores(quota.StudentScores.Subjects)

	finalizeScores(quota.ParentScores.RIASEC)
	finalizeScores(quota.ParentScores.OCEAN)
	finalizeScores(quota.ParentScores.Values)

	// 控制台摘要
	printQuotaSummary(quota)

	// 落盘
	out, _ := json.MarshalIndent(quota, "", "  ")
	_ = os.WriteFile("quota.json", out, 0644)
	fmt.Println("Step3 维度分数已保存到 quota.json")

	return nil
}

// --------- 辅助函数 ---------

// 更新分数字典
func updateScore(m map[string]ScoreResult, key string, score int) {
	if val, ok := m[key]; ok {
		val.Sum += score
		val.Count++
		m[key] = val
	} else {
		m[key] = ScoreResult{Sum: score, Count: 1}
	}
}

// 计算均值
func finalizeScores(m map[string]ScoreResult) {
	for k, v := range m {
		if v.Count > 0 {
			v.Mean = float64(v.Sum) / float64(v.Count)
			m[k] = v
		}
	}
}

// 从 type 中提取学科名（例如 "语文:A" → "语文"）
func extractSubject(t string) string {
	for i, r := range t {
		if r == ':' {
			return t[:i]
		}
	}
	return t
}

// 打印摘要
func printQuotaSummary(q QuotaDataset) {
	fmt.Println("=== Step3 摘要 ===")
	fmt.Printf("RequestID=%s StudentID=%s Mode=%s\n", q.Meta.RequestID, q.Meta.StudentID, q.Meta.Mode)

	fmt.Println("[学生端] RIASEC:", q.StudentScores.RIASEC)
	fmt.Println("[学生端] OCEAN:", q.StudentScores.OCEAN)
	fmt.Println("[学生端] 学科:", q.StudentScores.Subjects)
	fmt.Println("[学生端] 生涯:", q.StudentScores.Career)

	fmt.Println("[家长端] RIASEC:", q.ParentScores.RIASEC)
	fmt.Println("[家长端] OCEAN:", q.ParentScores.OCEAN)
	fmt.Println("[家长端] 价值观:", q.ParentScores.Values)

	fmt.Println("==================")
}

func TestStep3() {
	data, err := os.ReadFile("items.cleaned.json")
	if err != nil {
		fmt.Println("读取 items.cleaned.json 失败:", err)
		return
	}

	var cleaned CleanedDataset
	if err := json.Unmarshal(data, &cleaned); err != nil {
		fmt.Println("解析 CleanedDataset 失败:", err)
		return
	}

	if err := step3(cleaned); err != nil {
		fmt.Println("Step3 执行失败:", err)
		return
	}

	fmt.Println("Step3 测试执行完成 ✅")
}
