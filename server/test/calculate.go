package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Answer struct {
	ID    int `json:"id"`
	Score int `json:"score"`
}

type AnswerCase struct {
	CaseName string   `json:"case_name"` // liberalArts / science / conflict / invalid
	Variant  string   `json:"variant"`   // set1 / set2 / set3
	Student  []Answer `json:"student_answers"`
	Parent   []Answer `json:"parent_answers"`
}

var allAnswerCases = []AnswerCase{
	{
		CaseName: "liberalArts",
		Variant:  "set1",
		Student:  []Answer{ /* 从 answersLiberalArts */ },
		Parent:   []Answer{ /* 从 answersLiberalArts */ },
	},
	{
		CaseName: "liberalArts",
		Variant:  "set2",
		Student:  []Answer{ /* 从 liberalArtsAnswers struct */ },
		Parent:   []Answer{ /* 从 liberalArtsAnswers struct */ },
	},
	{
		CaseName: "liberalArts",
		Variant:  "set3",
		Student:  []Answer{ /* 从 const liberalArtsAnswers JSON */ },
		Parent:   []Answer{ /* 从 const liberalArtsAnswers JSON */ },
	},
	{
		CaseName: "science",
		Variant:  "set1",
		Student:  []Answer{ /* 从 scienceAnswers */ },
		Parent:   []Answer{ /* 从 parentAnswers_consistent_with_science */ },
	},
	{
		CaseName: "science",
		Variant:  "set2",
		Student:  []Answer{ /* 从 scienceAnswers struct */ },
		Parent:   []Answer{ /* 从 scienceAnswers struct */ },
	},
	{
		CaseName: "science",
		Variant:  "set3",
		Student:  []Answer{ /* 从 const scienceAnswers JSON */ },
		Parent:   []Answer{ /* 从 const scienceAnswers JSON */ },
	},
	{
		CaseName: "conflict",
		Variant:  "set1",
		Student:  []Answer{ /* 从 answersScienceConflict */ },
		Parent:   []Answer{ /* 从 answersScienceConflict */ },
	},
	{
		CaseName: "conflict",
		Variant:  "set2",
		Student:  []Answer{ /* 从 parentAnswers_inconsistent */ },
		Parent:   []Answer{ /* 从 parentAnswers_inconsistent */ },
	},
	{
		CaseName: "conflict",
		Variant:  "set3",
		Student:  []Answer{ /* 你定义的 scienceAnswers Student */ },
		Parent:   []Answer{ /* 你定义的 parentAnswers_inconsistent */ },
	},
	{
		CaseName: "invalid",
		Variant:  "set1",
		Student:  []Answer{ /* 从 answersInvalid */ },
		Parent:   []Answer{ /* 从 answersInvalid */ },
	},
	{
		CaseName: "invalid",
		Variant:  "set2",
		Student:  []Answer{ /* 从 invalidAnswers struct */ },
		Parent:   []Answer{ /* 从 invalidAnswers struct */ },
	},
	{
		CaseName: "invalid",
		Variant:  "set3",
		Student:  []Answer{ /* 从 const invalidAnswers JSON */ },
		Parent:   []Answer{ /* 从 const invalidAnswers JSON */ },
	},
}

func calculateQuota(requestID, studentID string) {
	//questionBytes, err := os.ReadFile("question.json")
	//if err != nil {
	//	fmt.Println("读取 question.json 出错:", err)
	//	return
	//}
	//// 假设固定答案（示例）
	//answers := []Answer{
	//	{ID: 1, Score: 4},
	//	{ID: 2, Score: 3},
	//	{ID: 3, Score: 5},
	//}

	// TODO: 根据 question.json 中的 type 做分组汇总
	quota := map[string]any{
		"request_id": requestID,
		"student_id": studentID,
		"RIASEC":     map[string]int{"R": 8, "I": 7, "A": 5, "S": 6, "E": 4, "C": 7},
		"OCEAN":      map[string]int{"O": 6, "C": 7, "E": 5, "A": 6, "N": 4},
		"Parents":    "家长观察：偏理科，重视实践",
	}
	bs, _ := json.MarshalIndent(quota, "", "  ")
	_ = os.WriteFile("quota.json", bs, 0644)
	fmt.Println("概要指标已保存到 quota.json")
}
