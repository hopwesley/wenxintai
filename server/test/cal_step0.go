package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// Step0Result 用于存放载入后的数据
type Step0Result struct {
	Question  Question  `json:"question"`
	AnswerSet AnswerSet `json:"answer_set"`
}

// loadAnswerSetByIndex 从 AllAnswerSets 中按下标获取答案集
func loadAnswerSetByIndex(index int) (AnswerSet, error) {
	if index < 0 || index >= len(AllAnswerSets) {
		return AnswerSet{}, fmt.Errorf("答案集 index 越界: %d (总数=%d)", index, len(AllAnswerSets))
	}
	return AllAnswerSets[index], nil
}

// loadQuestion 从 JSON 文件加载问卷
func loadQuestion(filePath string) (Question, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return Question{}, fmt.Errorf("读取问卷文件失败: %w", err)
	}
	var q Question
	if err := json.Unmarshal(data, &q); err != nil {
		return Question{}, fmt.Errorf("解析问卷 JSON 失败: %w", err)
	}
	return q, nil
}

func TestStep0(questionFile string, answerIndex int) error {
	// Step 1: 读取问卷
	q, err := loadQuestion(questionFile)
	if err != nil {
		return err
	}
	fmt.Printf("问卷加载完成: request_id=%s, student_id=%s\n", q.RequestID, q.StudentID)

	// Step 2: 读取答案集（按 index）
	aset, err := loadAnswerSetByIndex(answerIndex)
	if err != nil {
		return err
	}
	fmt.Printf("答案集加载完成: index=%d, name=%s, mode=%s, 学生答案数=%d, 家长答案数=%d\n",
		answerIndex, aset.Name, aset.Mode, len(aset.StudentAnswers), len(aset.ParentAnswers))

	combined := combineQuestionAndAnswer(q, aset)

	out, _ := json.MarshalIndent(combined, "", "  ")
	_ = os.WriteFile("step0_combined.json", out, 0644)

	return nil
}

// 对齐后的单题结构（已去掉 per-item 的 Mode）
type AlignedQA struct {
	ID     int    `json:"id"`
	Source string `json:"source"` // "student" | "parent"
	Text   string `json:"text"`
	Type   string `json:"type"`
	Rev    bool   `json:"rev"`
	Pair   string `json:"pair,omitempty"`
	Score  int    `json:"score"` // 原始分(1-5)，未做反向/清洗
}

// 汇总统计
type CombineStats struct {
	Matched  int      `json:"matched"`
	Total    int      `json:"total"`
	Missing  []int    `json:"missing"`  // 问卷有题但无答案
	Extra    []int    `json:"extra"`    // 答案里多出的题号
	Warnings []string `json:"warnings"` // 分数越界、重复题号等
}

// 对齐后的总结构：仅在这里记录 Mode（来自 AnswerSet）
type Combined struct {
	RequestID string       `json:"request_id"`
	StudentID string       `json:"student_id"`
	Mode      string       `json:"mode"` // 来自 AnswerSet.Mode
	Student   []AlignedQA  `json:"student"`
	Parent    []AlignedQA  `json:"parent"`
	SStats    CombineStats `json:"student_stats"`
	PStats    CombineStats `json:"parent_stats"`
}

// ---------- 工具函数 ----------

// 索引题目
func indexItems(items []Item) map[int]Item {
	m := make(map[int]Item)
	for _, it := range items {
		m[it.ID] = it
	}
	return m
}

// 索引答案，同时检查重复
func indexAnswers(ans []Answer) (map[int]Answer, []string) {
	m := make(map[int]Answer)
	warnings := []string{}
	for _, a := range ans {
		if _, exists := m[a.ID]; exists {
			warnings = append(warnings, fmt.Sprintf("重复答案 Qid=%d", a.ID))
		}
		m[a.ID] = a
	}
	return m, warnings
}

// 对齐一侧（学生或家长）
func alignOneSide(items map[int]Item, answers map[int]Answer, source string) ([]AlignedQA, CombineStats) {
	aligned := []AlignedQA{}
	stats := CombineStats{
		Matched:  0,
		Total:    len(items),
		Missing:  []int{},
		Extra:    []int{},
		Warnings: []string{},
	}

	// 遍历题目，找答案
	for id, it := range items {
		if ans, ok := answers[id]; ok {
			stats.Matched++
			if ans.Score < 1 || ans.Score > 5 {
				stats.Warnings = append(stats.Warnings,
					fmt.Sprintf("分数越界 Qid=%d 得分=%d", id, ans.Score))
			}
			aligned = append(aligned, AlignedQA{
				ID:     id,
				Source: source,
				Text:   it.Text,
				Type:   it.Type,
				Rev:    it.Rev,
				Pair:   it.Pair,
				Score:  ans.Score,
			})
		} else {
			stats.Missing = append(stats.Missing, id)
		}
	}

	// 遍历答案，找多余
	for id := range answers {
		if _, ok := items[id]; !ok {
			stats.Extra = append(stats.Extra, id)
		}
	}

	return aligned, stats
}

// 打印概览
func printCombineSummary(c Combined) {
	fmt.Println("=== combineQuestionAndAnswer 结果 ===")
	fmt.Printf("RequestID=%s StudentID=%s Mode=%s\n", c.RequestID, c.StudentID, c.Mode)
	fmt.Printf("[学生] total=%d matched=%d missing=%v extra=%v warnings=%v\n",
		c.SStats.Total, c.SStats.Matched, c.SStats.Missing, c.SStats.Extra, c.SStats.Warnings)
	fmt.Printf("[家长] total=%d matched=%d missing=%v extra=%v warnings=%v\n",
		c.PStats.Total, c.PStats.Matched, c.PStats.Missing, c.PStats.Extra, c.PStats.Warnings)
	fmt.Println("====================================")
}

// ---------- 主函数 ----------

// 仅在 Combined 上记录一次 Mode（来自 AnswerSet.Mode）
func combineQuestionAndAnswer(q Question, aset AnswerSet) Combined {
	// 建立索引
	sItems := indexItems(q.StudentQuestions)
	pItems := indexItems(q.ParentQuestions)
	sAnswers, sWarn := indexAnswers(aset.StudentAnswers)
	pAnswers, pWarn := indexAnswers(aset.ParentAnswers)

	// 学生端对齐
	sAligned, sStats := alignOneSide(sItems, sAnswers, "student")
	sStats.Warnings = append(sStats.Warnings, sWarn...)

	// 家长端对齐
	pAligned, pStats := alignOneSide(pItems, pAnswers, "parent")
	pStats.Warnings = append(pStats.Warnings, pWarn...)

	// 汇总
	combined := Combined{
		RequestID: q.RequestID,
		StudentID: q.StudentID,
		Mode:      aset.Mode,
		Student:   sAligned,
		Parent:    pAligned,
		SStats:    sStats,
		PStats:    pStats,
	}

	// 打印概览
	printCombineSummary(combined)
	return combined
}
