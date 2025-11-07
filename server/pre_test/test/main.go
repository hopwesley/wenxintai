package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run main.go <MODE:1|2|3> <API_KEY> <STUDENT_ID>")
		return
	}
	mode := os.Args[1]

	requestID := uuidLike()
	fmt.Println("生成的 request_id:", requestID)

	switch mode {
	case "1":

		apiKey := os.Args[2]
		studentID := os.Args[3]

		students := []struct {
			id, gender, grade, mode string
		}{
			{studentID, "男", "高一", "3+3"},
			{studentID + "1", "女", "初三", "3+1+2"},
		}

		for i, s := range students {
			rid := fmt.Sprintf("%s_%d", requestID, i)
			fetchQuestion(rid, s.id, s.gender, s.grade, s.mode, apiKey)
		}
	case "2":
		questionFile := os.Args[2]
		idx, err := strconv.Atoi(os.Args[3])
		if err != nil {
			panic(err)
		}
		TestStep0(questionFile, idx)

	case "3":
		TestStep1()

	case "4":
		TestStep2()

	case "5":
		TestStep3()

	case "6":
		TestStep4()

	case "7":
		apiKey := os.Args[2]
		TestStep5(apiKey)

	default:
		fmt.Println("无效模式: 请输入 1 | 2 | 3")
	}
}
