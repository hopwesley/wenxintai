package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"
)

type RawItem struct {
	RequestID string `json:"request_id"`
	StudentID string `json:"student_id"`
	Source    string `json:"source"` // "student" | "parent"
	Qid       int    `json:"qid"`
	Type      string `json:"type"`
	Rev       bool   `json:"rev"`
	Pair      string `json:"pair,omitempty"`
	Score     int    `json:"score"`
	Text      string `json:"text"`
}

type CheckResult struct {
	AllSameScore      bool    `json:"all_same_score"`
	HighConcentration bool    `json:"high_concentration"`
	DominantValue     int     `json:"dominant_value,omitempty"`
	DominantRatio     float64 `json:"dominant_ratio,omitempty"`
}

type RawDataset struct {
	Meta struct {
		RequestID   string `json:"request_id"`
		StudentID   string `json:"student_id"`
		Mode        string `json:"mode"`
		GeneratedAt string `json:"generated_at"`
	} `json:"meta"`
	StudentItems []RawItem `json:"student_items"`
	ParentItems  []RawItem `json:"parent_items"`
	Checks       struct {
		Student CheckResult `json:"student"`
		Parent  CheckResult `json:"parent"`
	} `json:"checks"`
}

// --------- Step1 主要流程 ---------

func step1(combined Combined) error {
	// 初始化 RawDataset
	raw := RawDataset{}
	raw.Meta.RequestID = combined.RequestID
	raw.Meta.StudentID = combined.StudentID
	raw.Meta.Mode = combined.Mode
	raw.Meta.GeneratedAt = time.Now().Format(time.RFC3339)

	// 学生端
	for _, qa := range combined.Student {
		raw.StudentItems = append(raw.StudentItems, RawItem{
			RequestID: combined.RequestID,
			StudentID: combined.StudentID,
			Source:    "student",
			Qid:       qa.ID,
			Type:      qa.Type,
			Rev:       qa.Rev,
			Pair:      qa.Pair,
			Score:     qa.Score,
			Text:      qa.Text,
		})
	}

	// 家长端
	for _, qa := range combined.Parent {
		raw.ParentItems = append(raw.ParentItems, RawItem{
			RequestID: combined.RequestID,
			StudentID: combined.StudentID,
			Source:    "parent",
			Qid:       qa.ID,
			Type:      qa.Type,
			Rev:       qa.Rev,
			Pair:      qa.Pair,
			Score:     qa.Score,
			Text:      qa.Text,
		})
	}

	// 检查逻辑
	raw.Checks.Student = runChecks(raw.StudentItems)
	raw.Checks.Parent = runChecks(raw.ParentItems)

	// 控制台摘要打印
	printRawSummary(raw)

	// 落盘
	out, _ := json.MarshalIndent(raw, "", "  ")
	_ = os.WriteFile("items.raw.json", out, 0644)
	fmt.Println("Step1 原始数据已保存到 items.raw.json")

	return nil
}

// --------- 辅助函数：检查逻辑 ---------

func runChecks(items []RawItem) CheckResult {
	result := CheckResult{}
	if len(items) == 0 {
		return result
	}

	// 统计分布
	counts := map[int]int{}
	for _, it := range items {
		counts[it.Score]++
	}

	// 检查是否全同分
	if len(counts) == 1 {
		result.AllSameScore = true
	}

	// 检查是否有某个值超过85%
	total := len(items)
	for val, cnt := range counts {
		ratio := float64(cnt) / float64(total)
		if ratio >= 0.85 {
			result.HighConcentration = true
			result.DominantValue = val
			result.DominantRatio = ratio
			break
		}
	}

	return result
}

// --------- 辅助函数：摘要统计 ---------

func printRawSummary(raw RawDataset) {
	fmt.Println("=== Step1 摘要 ===")
	fmt.Printf("RequestID=%s StudentID=%s Mode=%s\n", raw.Meta.RequestID, raw.Meta.StudentID, raw.Meta.Mode)

	// 学生端
	fmt.Printf("[学生] count=%d 检查=%+v\n", len(raw.StudentItems), raw.Checks.Student)
	printScoreDistribution(raw.StudentItems)

	// 家长端
	fmt.Printf("[家长] count=%d 检查=%+v\n", len(raw.ParentItems), raw.Checks.Parent)
	printScoreDistribution(raw.ParentItems)

	fmt.Println("==================")
}

func printScoreDistribution(items []RawItem) {
	if len(items) == 0 {
		fmt.Println("  (无数据)")
		return
	}

	counts := map[int]int{}
	scores := []int{}
	for _, it := range items {
		counts[it.Score]++
		scores = append(scores, it.Score)
	}
	sort.Ints(scores)

	// 基本统计
	sum := 0
	for _, s := range scores {
		sum += s
	}
	mean := float64(sum) / float64(len(scores))
	median := float64(scores[len(scores)/2])
	if len(scores)%2 == 0 {
		median = float64(scores[len(scores)/2-1]+scores[len(scores)/2]) / 2.0
	}

	// 打印
	fmt.Printf("  分布: ")
	for i := 1; i <= 5; i++ {
		fmt.Printf("%d:%d ", i, counts[i])
	}
	fmt.Printf(" | 平均=%.2f 中位=%.1f 最小=%d 最大=%d\n",
		mean, median, scores[0], scores[len(scores)-1])
}

func TestStep1() {
	// 先从 step0_combined.json 读取 Combined
	data, err := os.ReadFile("step0_combined.json")
	if err != nil {
		fmt.Println("读取 step0_combined.json 失败:", err)
		return
	}
	var combined Combined
	if err := json.Unmarshal(data, &combined); err != nil {
		fmt.Println("解析 Combined 失败:", err)
		return
	}

	// 执行 Step1
	if err := step1(combined); err != nil {
		fmt.Println("Step1 执行失败:", err)
	}
}
