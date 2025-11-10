package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type databaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	User     string `json:"user"`
	Password string `json:"password"`
	SSLMode  string `json:"sslmode"`
}

func loadDatabaseConfig() (databaseConfig, error) {
	path, err := resolveDatabaseConfigPath()
	if err != nil {
		return databaseConfig{}, err
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return databaseConfig{}, fmt.Errorf("读取数据库配置失败: %w", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return databaseConfig{}, fmt.Errorf("解析数据库配置失败: %w", err)
	}

	cfg := databaseConfig{}
	cfg.Host = getString(payload, "host")
	cfg.Database = getString(payload, "database")
	cfg.User = getString(payload, "user")
	cfg.Password = getString(payload, "password")
	cfg.SSLMode = getString(payload, "sslmode")
	if cfg.SSLMode == "" {
		cfg.SSLMode = getString(payload, "ssl")
	}
	cfg.Port = getInt(payload, "port")

	applyEnvOverrides(&cfg)

	if err := cfg.validate(); err != nil {
		return databaseConfig{}, err
	}

	log.Printf("[config] database host=%s port=%d database=%s user=%s sslmode=%s", cfg.Host, cfg.Port, cfg.Database, cfg.User, cfg.SSLMode)
	return cfg, nil
}

func resolveDatabaseConfigPath() (string, error) {
	var candidates []string
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "config", "database.json"))
	}
	if wd, err := os.Getwd(); err == nil {
		path := filepath.Join(wd, "config", "database.json")
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
	return "", fmt.Errorf("未找到数据库配置文件，尝试路径: %s", strings.Join(candidates, ", "))
}

func applyEnvOverrides(cfg *databaseConfig) {
	if host := strings.TrimSpace(os.Getenv("DB_HOST")); host != "" {
		cfg.Host = host
	}
	if portValue := strings.TrimSpace(os.Getenv("DB_PORT")); portValue != "" {
		if port, err := strconv.Atoi(portValue); err == nil {
			cfg.Port = port
		}
	}
	if name := strings.TrimSpace(os.Getenv("DB_NAME")); name != "" {
		cfg.Database = name
	}
	if user := strings.TrimSpace(os.Getenv("DB_USER")); user != "" {
		cfg.User = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		cfg.Password = password
	}
	if sslmode := strings.TrimSpace(os.Getenv("DB_SSLMODE")); sslmode != "" {
		cfg.SSLMode = sslmode
	}
}

func (cfg databaseConfig) validate() error {
	var missing []string
	if strings.TrimSpace(cfg.Host) == "" {
		missing = append(missing, "host")
	}
	if cfg.Port <= 0 {
		missing = append(missing, "port")
	}
	if strings.TrimSpace(cfg.Database) == "" {
		missing = append(missing, "database")
	}
	if strings.TrimSpace(cfg.User) == "" {
		missing = append(missing, "user")
	}
	if strings.TrimSpace(cfg.Password) == "" {
		missing = append(missing, "password")
	}
	if strings.TrimSpace(cfg.SSLMode) == "" {
		missing = append(missing, "sslmode")
	}
	if len(missing) > 0 {
		return fmt.Errorf("数据库配置缺少字段: %s", strings.Join(missing, ", "))
	}

	return nil
}

func connectDatabase(cfg databaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s search_path=app,public", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("初始化数据库连接失败: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("数据库连接不可用: %w", err)
	}
	return db, nil
}

func getString(payload map[string]any, key string) string {
	raw, ok := payload[key]
	if !ok {
		return ""
	}
	switch value := raw.(type) {
	case string:
		return strings.TrimSpace(value)
	default:
		return ""
	}
}

func getInt(payload map[string]any, key string) int {
	raw, ok := payload[key]
	if !ok {
		return 0
	}
	switch value := raw.(type) {
	case float64:
		return int(value)
	case string:
		value = strings.TrimSpace(value)
		if value == "" {
			return 0
		}
		if n, err := strconv.Atoi(value); err == nil {
			return n
		}
		return 0
	default:
		return 0
	}
}
