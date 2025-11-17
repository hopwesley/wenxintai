package ai_api

import (
	"context"
	"encoding/json"
)

type QuestParam struct {
	Qid   string `json:"qid"`
	Grade string `json:"grade"`
}
type TokenHandler func(string) error

type AIApi interface {
	Init(api *Cfg) error
	GenerateQuestion(ctx context.Context, basicInfo *BasicInfo, tt TestTyp, callback TokenHandler) (json.RawMessage, error)
}
