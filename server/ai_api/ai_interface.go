package ai_api

import (
	"context"
)

type QuestParam struct {
	Qid   string `json:"qid"`
	Grade string `json:"grade"`
}
type TokenHandler func(string) error

type AIApi interface {
	Init(api *Cfg) error
	GenerateQuestion(ctx context.Context, basicInfo *BasicInfo, tt TestTyp, callback TokenHandler) (string, error)
	GenerateUnifiedReport(ctx context.Context, common *CommonSection, param interface{}, mode Mode, callback TokenHandler) (string, error)
}
