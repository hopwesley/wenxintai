package main

import (
	"fmt"
	"os"
	"time"
)

func uuidLike() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

func main() {
	stage := os.Args[1]
	switch stage {
	case "question":
		mode := Mode(os.Args[2])
		apiKey := os.Args[3]
		_ = generateQuestions(mode, apiKey, "女", "高一")
		_ = generateQuestions(mode, apiKey, "男", "初三")
		break
	default:
		panic("unknown stage parameter")
	}
}
