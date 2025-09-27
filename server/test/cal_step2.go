package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// CleanedItem：清洗后的单题记录
type CleanedItem struct {
	Qid        int    `json:"qid"`
	Type       string `json:"type"`
	ScoreRaw   int    `json:"score_raw"`
	ScoreFinal int    `json:"score_final"`
	Rev        bool   `json:"rev"`
	Category   string `json:"category"`
}

// ValidityItem：效度题，保留原始分数
type ValidityItem struct {
	Source   string `json:"source"` // student 或 parent
	Qid      int    `json:"qid"`
	ScoreRaw int    `json:"score_raw"`
	Rev      bool   `json:"rev"`
}

// CleanedDataset：Step2 输出结构
type CleanedDataset struct {
	Meta struct {
		RequestID   string `json:"request_id"`
		StudentID   string `json:"student_id"`
		Mode        string `json:"mode"`
		GeneratedAt string `json:"generated_at"`
	} `json:"meta"`
	StudentCleaned  []CleanedItem `json:"student_cleaned"`
	ParentCleaned   []CleanedItem `json:"parent_cleaned"`
	ValiditySection struct {
		ValidityItems []ValidityItem `json:"validity_items"`
		Checks        struct {
			Student CheckResult `json:"student"`
			Parent  CheckResult `json:"parent"`
		} `json:"checks"`
	} `json:"validity_section"`
}

// --------- Step2 主流程 ---------

func step2(raw RawDataset) error {
	cleaned := CleanedDataset{}
	cleaned.Meta.RequestID = raw.Meta.RequestID
	cleaned.Meta.StudentID = raw.Meta.StudentID
	cleaned.Meta.Mode = raw.Meta.Mode
	cleaned.Meta.GeneratedAt = time.Now().Format(time.RFC3339)

	// 学生端
	for _, it := range raw.StudentItems {
		if it.Type == "D" {
			cleaned.ValiditySection.ValidityItems = append(cleaned.ValiditySection.ValidityItems, ValidityItem{
				Source:   "student",
				Qid:      it.Qid,
				ScoreRaw: it.Score,
				Rev:      it.Rev,
			})
			continue
		}

		cleaned.StudentCleaned = append(cleaned.StudentCleaned, CleanedItem{
			Qid:        it.Qid,
			Type:       it.Type,
			ScoreRaw:   it.Score,
			ScoreFinal: processScore(it.Score, it.Rev),
			Rev:        it.Rev,
			Category:   classifyType(it.Type),
		})
	}

	// 家长端
	for _, it := range raw.ParentItems {
		// 家长端效度题
		if it.Type == "D" {
			cleaned.ValiditySection.ValidityItems = append(cleaned.ValiditySection.ValidityItems, ValidityItem{
				Source:   "parent",
				Qid:      it.Qid,
				ScoreRaw: it.Score,
				Rev:      it.Rev,
			})
			continue
		}

		cleaned.ParentCleaned = append(cleaned.ParentCleaned, CleanedItem{
			Qid:        it.Qid,
			Type:       it.Type,
			ScoreRaw:   it.Score,
			ScoreFinal: processScore(it.Score, it.Rev),
			Rev:        it.Rev,
			Category:   classifyType(it.Type),
		})
	}

	// 效度检查结果直接从 Step1 复制过来
	cleaned.ValiditySection.Checks.Student = raw.Checks.Student
	cleaned.ValiditySection.Checks.Parent = raw.Checks.Parent

	// 控制台摘要
	printCleanedSummary(cleaned)

	// 落盘
	out, _ := json.MarshalIndent(cleaned, "", "  ")
	_ = os.WriteFile("items.cleaned.json", out, 0644)
	fmt.Println("Step2 清洗数据已保存到 items.cleaned.json")

	return nil
}

// --------- 辅助函数 ---------

// 处理反向计分：非 D 且 rev=true 时反向
func processScore(score int, rev bool) int {
	if rev {
		return 6 - score
	}
	return score
}

// 根据 type 分类
func classifyType(t string) string {
	switch t {
	case "R", "I", "A", "S", "E", "C":
		return "RIASEC"
	case "b5_O", "b5_C", "b5_E", "b5_A", "b5_N":
		return "OCEAN"
	case "生涯":
		return "career"
	case "价值观":
		return "value"
	case "D":
		return "validity"
	default:
		// 学科类: "语文:A" / "数学:I" ...
		return "subject"
	}
}

// 打印摘要
func printCleanedSummary(c CleanedDataset) {
	fmt.Println("=== Step2 摘要 ===")
	fmt.Printf("RequestID=%s StudentID=%s Mode=%s\n", c.Meta.RequestID, c.Meta.StudentID, c.Meta.Mode)

	fmt.Printf("[学生] 有效题=%d\n", len(c.StudentCleaned))
	fmt.Printf("[家长] 有效题=%d\n", len(c.ParentCleaned))
	fmt.Printf("[效度题] 共=%d (学生+家长)\n", len(c.ValiditySection.ValidityItems))

	fmt.Println("==================")
}

// TestStep2 封装 Step2 的测试执行流程
func TestStep2() {
	data, err := os.ReadFile("items.raw.json")
	if err != nil {
		fmt.Println("读取 items.raw.json 失败:", err)
		return
	}

	var raw RawDataset
	if err := json.Unmarshal(data, &raw); err != nil {
		fmt.Println("解析 RawDataset 失败:", err)
		return
	}

	if err := step2(raw); err != nil {
		fmt.Println("Step2 执行失败:", err)
		return
	}

	fmt.Println("Step2 测试执行完成 ✅")
}
