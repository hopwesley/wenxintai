package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type databaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	User     string `json:"user"`
	Password string `json:"password"`
	SSLMode  string `json:"sslmode"`
}
type serverConfig struct {
	Port          string `json:"port"`
	StaticDir     string `json:"static_dir"`
	DefaultAPIKey string `json:"default_api_key"`
}
type appConfig struct {
	Server   serverConfig   `json:"server"`
	Database databaseConfig `json:"database"`
}

func resolveDatabaseConfigPath() (string, error) {
	var candidates []string
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "config", "conf.json"))
	}
	if wd, err := os.Getwd(); err == nil {
		path := filepath.Join(wd, "config", "conf.json")
		exists := false
		for _, c := range candidates {
			if c == path {
				exists = true
				break
			}
		}
		if !exists {
			candidates = append(candidates, path)
		}
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("未找到配置文件，尝试路径: %s", strings.Join(candidates, ", "))
}

func loadAppConfig() (*appConfig, error) {

	path, err := resolveDatabaseConfigPath()
	if err != nil {
		return nil, err
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg appConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("解析 conf.json 失败: %w", err)
	}

	if err := cfg.Database.validate(); err != nil {
		return nil, err
	}
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080"
	}
	if cfg.Server.StaticDir == "" {
		cfg.Server.StaticDir = "./static"
	}

	return &cfg, nil
}
