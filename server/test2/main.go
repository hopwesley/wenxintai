package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

func uuidLike() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

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
	stage := os.Args[1]
	switch stage {
	case "question":
		mode := Mode(os.Args[2])
		apiKey := os.Args[3]
		//_ = generateQuestions(mode, apiKey, "女", "高一")
		hobby := StudentHobbies[rand.Intn(len(StudentHobbies))]
		_ = generateQuestions(mode, apiKey, "男", "初三", hobby)

		hobby2 := StudentHobbies[rand.Intn(len(StudentHobbies))]
		_ = generateQuestions(mode, apiKey, "女", "高一", hobby2)
		//_ = generateQuestions(mode, apiKey, "男", "初三", "无")
		break
	case "answer":
		TestRIASECAnswer()
		TestASCAnswer()
	case "demo":
		// 1) 选择一个组合与能力场景
		combo := ComboPHY_CHE_BIO
		//category := "aligned" // 或 "mismatch"
		category := "mismatch" // 或 "mismatch"

		// 2) 拿到样本答案
		riasec := AllRIASECCombos[combo]
		asc := AllASCCombos[combo][category]

		RunDemo33(riasec, asc, 0, 0, 0)
		RunDemo312(riasec, asc, 0, 0, 0)

	case "report":
		filePath := os.Args[2]
		apiKey := os.Args[3]
		err := TestReport(apiKey, filePath)
		if err != nil {
			panic(err)
		}

	default:
		panic("unknown stage parameter")
	}
}
