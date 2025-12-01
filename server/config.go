package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hopwesley/wenxintai/server/ai_api"
	"github.com/hopwesley/wenxintai/server/dbSrv"
	"github.com/hopwesley/wenxintai/server/srv"
)

type appConfig struct {
	DebugLevel string            `json:"debug_level,omitempty"`
	Server     *srv.Config       `json:"server"`
	Database   *dbSrv.PSDBConfig `json:"database"`
	AIApi      *ai_api.Cfg       `json:"ai_api"`
}

// 根据 env 选择不同的配置文件名：
// env == "dev"  -> conf_dev.json (测试)
// 其他           -> conf.json     (生产)
func resolveDatabaseConfigPath(env string) (string, error) {
	fileName := "conf.json"
	if env == "dev" {
		fileName = "conf_dev.json"
	}

	var candidates []string
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "config", fileName))
	}
	if wd, err := os.Getwd(); err == nil {
		path := filepath.Join(wd, "config", fileName)
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
			fmt.Printf("[config] use config file: %s (env=%s)\n", candidate, env)
			return candidate, nil
		}
	}
	return "", fmt.Errorf("未找到配置文件(环境: %s)，尝试路径: %s", env, strings.Join(candidates, ", "))
}

func loadAppConfig(env string) (*appConfig, error) {
	path, err := resolveDatabaseConfigPath(env)
	if err != nil {
		return nil, err
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg appConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("解析 %s 失败: %w", filepath.Base(path), err)
	}

	if err := cfg.Database.Validate(); err != nil {
		return nil, err
	}

	if err := cfg.Server.Validate(); err != nil {
		return nil, err
	}

	if err := cfg.AIApi.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}
