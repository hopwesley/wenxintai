package service

import (
	"errors"
	"time"
)

const (
	NoGender = iota
	Male
	Female
)

// 用户信息结构
type UserProfile struct {
	Name      string `json:"name"`
	BirthYear int    `json:"birth_year"`
	Gender    uint   `json:"gender"`
}

func (i *UserProfile) CheckUserData(reqType PromptType) error {
	// 验证必填字段
	if i.Name == "" || i.BirthYear == 0 {
		return errors.New("姓名、出生年份和性别为必填项")
	}

	age := time.Now().Year() - i.BirthYear

	var ageIsOk = false
	switch reqType {
	case PromptTypeChoseClass:
		ageIsOk = age >= 12 && age <= 19
		break
	case PromptTypePressureMiddleSchool:
		ageIsOk = age >= 12 && age <= 22
		break
	case PromptTypePressureUniversity:
		ageIsOk = age >= 14 && age <= 32
		break
	}
	if !ageIsOk {
		return errors.New("出生年份无法参与此测试")
	}
	return nil
}
