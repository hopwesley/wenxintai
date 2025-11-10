package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

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
