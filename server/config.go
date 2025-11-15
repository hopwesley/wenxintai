package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hopwesley/wenxintai/server/dbSrv"
	"github.com/hopwesley/wenxintai/server/srv"
)

type appConfig struct {
	DebugLevel string            `json:"debug_level,omitempty"`
	Server     *srv.Config       `json:"server"`
	Database   *dbSrv.PSDBConfig `json:"database"`
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

	if err := cfg.Database.Validate(); err != nil {
		return nil, err
	}

	if err := cfg.Server.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}
