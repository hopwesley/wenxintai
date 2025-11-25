package main

import (
	"encoding/json"
	"fmt"
	"os"

	core "github.com/hopwesley/wenxintai/server/ai_api"
)

// StudentHobbies 提供前端可选的兴趣标签列表。
var StudentHobbies = []string{
	// 体育类
	"篮球",
	"足球",
	"羽毛球",
	"跑步",
	"游泳",
	"乒乓球",
	"健身",

	// 艺术类
	"音乐",
	"绘画",
	"舞蹈",
	"摄影",
	"书法",
	"写作",

	// 科技类
	"编程",
	"机器人",
	"科学实验",
	"电子制作",
	"下棋（象棋/围棋/国际象棋）",

	// 生活方式类
	"旅行",
	"美食",
	"志愿活动",
	"阅读",
	"看电影",
	"园艺",
}

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
		mode := os.Args[4]
		if err := runReport(apiKey, payload, core.Mode(mode)); err != nil {
			panic(err)
		}
	default:
		panic("unknown stage parameter")
	}
}

func printSampleAnswers() {
	combo := core.ComboPHY_CHE_BIO
	answers := AllRIASECCombos[combo]
	data, _ := json.MarshalIndent(answers, "", "  ")
	fmt.Println(string(data))

	asc := AllASCCombos[combo]["aligned"]
	dataAsc, _ := json.MarshalIndent(asc, "", "  ")
	fmt.Println(string(dataAsc))
}

func runDemo(combo, idx string) {
	riasec := AllRIASECCombos[combo]

	ascAligned := AllASCCombos[combo]["aligned"]
	core.RunDemo33(riasec, ascAligned, 0, 0, 0, idx, "yes", combo)
	core.RunDemo312(riasec, ascAligned, 0, 0, 0, idx, "yes", combo)

	ascMismatch := AllASCCombos[combo]["mismatch"]
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

	filename := fmt.Sprintf("report_unified_v5_%s.json", mode)
	return os.WriteFile(filename, report, 0o644)
}
