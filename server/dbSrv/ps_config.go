package dbSrv

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type PSDBConfig struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Database    string `json:"database"`
	User        string `json:"user"`
	Password    string `json:"password"`
	SSLMode     string `json:"sslmode"`
	MaxOpenConn int    `json:"max_open_conn,omitempty"`
	MaxIdleConn int    `json:"max_idle_conn,omitempty"`
	MaxLifeTime int    `json:"max_life_time,omitempty"`
	ConnTimeOut int    `json:"conn_time_out,omitempty"`
}

func (cfg *PSDBConfig) connDB() (*sql.DB, error) {
	dsn := cfg.buildDSN()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConn)
	db.SetMaxIdleConns(cfg.MaxIdleConn)
	db.SetConnMaxLifetime(time.Duration(cfg.MaxLifeTime) * time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ConnTimeOut)*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}
	return db, nil
}

func (cfg *PSDBConfig) buildDSN() string {
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

func (cfg *PSDBConfig) Validate() error {
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

	if cfg.MaxOpenConn <= 0 {
		cfg.MaxOpenConn = 10
	}
	if cfg.MaxIdleConn <= 0 {
		cfg.MaxIdleConn = 5
	}
	if cfg.MaxLifeTime <= 0 {
		cfg.MaxLifeTime = 30
	}
	if cfg.ConnTimeOut <= 0 {
		cfg.ConnTimeOut = 10
	}

	if len(missing) > 0 {
		return fmt.Errorf("数据库配置缺少字段: %s", strings.Join(missing, ", "))
	}
	return nil
}
