package service

type PromptType uint8

const (
	_ PromptType = iota
	PromptTypeChoseClass
	PromptTypePressureMiddleSchool
	PromptTypePressureUniversity
)

type Prompt struct {
	PT              PromptType `json:"prompt_type"`
	SystemPromptTxt []string   `json:"system_prompt_txt"`
	UserPromptTxt   []string   `json:"user_prompt_txt"`
}
