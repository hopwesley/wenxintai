package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/hopwesley/wenxintai/server/core"
)

func main() {
	if len(os.Args) < 2 {
		panic("missing stage parameter")
	}

	stage := os.Args[1]
	switch stage {
	case "question":
		if len(os.Args) < 4 {
			panic("usage: question <mode> <apiKey>")
		}
		mode, ok := core.ParseMode(os.Args[2])
		if !ok {
			panic("invalid mode")
		}
		apiKey := os.Args[3]
		hobby := core.StudentHobbies[rand.Intn(len(core.StudentHobbies))]
		fmt.Println("------>>>hobby:", hobby)
		if err := runGenerateQuestions(mode, apiKey, "男", "初三", hobby); err != nil {
			fmt.Println("generate questions error:", err)
		}

		hobby2 := core.StudentHobbies[rand.Intn(len(core.StudentHobbies))]
		fmt.Println("------>>>hobby:", hobby2)
		if err := runGenerateQuestions(mode, apiKey, "女", "高一", hobby2); err != nil {
			fmt.Println("generate questions error:", err)
		}
	case "answer":
		printSampleAnswers()
	case "demo":
		if len(os.Args) < 4 {
			panic("usage: demo <combo> <apiKey>")
		}
		combo := os.Args[2]
		idx := os.Args[3]
		runDemo(combo, idx)
	case "report":
		if len(os.Args) < 5 {
			panic("usage: report <session> <apiKey> <mode>")
		}
		payload := os.Args[2]
		apiKey := os.Args[3]
		mode, ok := core.ParseMode(os.Args[4])
		if !ok {
			panic("invalid mode")
		}
		if err := runReport(apiKey, payload, mode); err != nil {
			panic(err)
		}
	default:
		panic("unknown stage parameter")
	}
}

func runGenerateQuestions(mode core.Mode, apiKey, gender, grade, hobby string) error {
	res, err := core.GenerateQuestions(mode, apiKey, gender, grade, hobby)
	if err != nil {
		return err
	}

	ts := time.Now().Format("20060102_150405")
	for module, payload := range res.Modules {
		filename := fmt.Sprintf("questions_%s_%s_%s_%s.json", mode.String(), module, ts, hobby)
		if err := os.WriteFile(filename, payload, 0o644); err != nil {
			return err
		}
		fmt.Printf("问卷已保存：[%s] -> %s\n", module, filename)
	}
	return nil
}

func printSampleAnswers() {
	combo := core.ComboPHY_CHE_BIO
	answers := core.AllRIASECCombos[combo]
	data, _ := json.MarshalIndent(answers, "", "  ")
	fmt.Println(string(data))

	asc := core.AllASCCombos[combo]["aligned"]
	dataAsc, _ := json.MarshalIndent(asc, "", "  ")
	fmt.Println(string(dataAsc))
}

func runDemo(combo, idx string) {
	riasec := core.AllRIASECCombos[combo]

	ascAligned := core.AllASCCombos[combo]["aligned"]
	core.RunDemo33(riasec, ascAligned, 0, 0, 0, idx, "yes", combo)
	core.RunDemo312(riasec, ascAligned, 0, 0, 0, idx, "yes", combo)

	ascMismatch := core.AllASCCombos[combo]["mismatch"]
	core.RunDemo33(riasec, ascMismatch, 0, 0, 0, idx, "no", combo)
	core.RunDemo312(riasec, ascMismatch, 0, 0, 0, idx, "no", combo)

	fmt.Println("demo completed for", combo)
}

func runReport(apiKey, path string, mode core.Mode) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var param core.ParamForAIPrompt
	if err := json.Unmarshal(data, &param); err != nil {
		return err
	}

	report, err := core.TestUnifiedReport(apiKey, param, mode)
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("report_unified_v5_%s.json", mode.String())
	return os.WriteFile(filename, report, 0o644)
}
