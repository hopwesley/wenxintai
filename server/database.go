package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

func connectDatabase(cfg databaseConfig) (*sql.DB, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	dsn := buildDSN(cfg)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}
	return db, nil
}

func buildDSN(cfg databaseConfig) string {
	parts := []string{
		fmt.Sprintf("host=%s", cfg.Host),
		fmt.Sprintf("port=%d", cfg.Port),
		fmt.Sprintf("user=%s", cfg.User),
		fmt.Sprintf("password=%s", cfg.Password),
		fmt.Sprintf("dbname=%s", cfg.Database),
		fmt.Sprintf("sslmode=%s", cfg.SSLMode),
		"search_path=app,public",
	}
	return strings.Join(parts, " ")
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
