package service

import (
	"encoding/json"
	"net/http"
	"time"
)

// 用户信息结构
type UserInfo struct {
	Name      string `json:"name"`
	BirthYear int    `json:"birth_year"`
	Gender    string `json:"gender"`
}

func ParseUserInfo() *UserInfo {
	var userInfo service.UserInfo
	if err := json.NewDecoder(r.Body).Decode(&userInfo); err != nil {
		resp := ApiResponse{
			Success: false,
			Message: "无效的用户信息格式",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// 验证必填字段
	if userInfo.Name == "" || userInfo.BirthYear == 0 || userInfo.Gender == "" {
		resp := ApiResponse{
			Success: false,
			Message: "姓名、出生年份和性别为必填项",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// 验证出生年份合理性
	currentYear := time.Now().Year()
	if userInfo.BirthYear < 1900 || userInfo.BirthYear > currentYear {
		resp := ApiResponse{
			Success: false,
			Message: "出生年份应在1900年至当前年份之间",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}
}
