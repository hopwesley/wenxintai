package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hopwesley/wenxintai/server/comm"
	_ "github.com/lib/pq"
)

// ================== 命令行参数 ==================

var (
	dbPassword = flag.String("pwd", "", "database password for user wenxintai")
	dbHost     = flag.String("host", "", "database host for user wenxintai")
	flagN      = flag.Int("no", 10, "number of invite codes to create")
)

// ================== 固定参数 ==================

const (
	tier       = "B" // B=基础版, P=专业版, C=校园版
	expireDays = 30  // 0 表示不过期
)

// ================== main ==================

func main() {
	flag.Parse()

	if *dbPassword == "" {
		log.Fatal("❌ 必须提供数据库密码：-pwd=xxxx")
	}

	fmt.Println("🚀 Starting invite bootstrap...")
	fmt.Printf("   count=%d, tier=%s, expireDays=%d\n", *flagN, tier, expireDays)

	if err := BootstrapInvites(*dbPassword, *dbHost, *flagN); err != nil {
		log.Fatalf("❌ 执行失败: %v", err)
	}

	fmt.Println("✅ 完成")
}

// ================== 普通函数版本 ==================

func QueryTestTable(dbPassword string) error {
	connStr := fmt.Sprintf(
		"host=localhost port=5432 user=wesley password=%s dbname=hyperorchid sslmode=disable",
		dbPassword,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("sql.Open: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("db.Ping: %w", err)
	}

	rows, err := db.Query("SELECT id, name FROM test;")
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	fmt.Println("📋 test 表数据：")
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return err
		}
		fmt.Printf("id=%d, name=%s\n", id, name)
	}
	return rows.Err()
}

// ================== 邀请码主逻辑 ==================

func BootstrapInvites(dbPassword, dbHost string, count int) error {
	db, err := openDB(dbPassword, dbHost)
	if err != nil {
		return err
	}
	defer db.Close()

	var expiresAt *time.Time
	x := time.Now().Add(time.Duration(expireDays) * 24 * time.Hour)
	expiresAt = &x

	ctx := context.Background()
	inserted := 0
	start := time.Now()

	for inserted < count {
		code, err := comm.MakeInviteCode()
		if err != nil {
			return err
		}

		const q = `
			INSERT INTO app.invites (code, expires_at)
			VALUES ($1, $2)
			ON CONFLICT (code) DO NOTHING
		`
		res, err := db.ExecContext(ctx, q, code, expiresAt)
		if err != nil {
			return err
		}

		aff, _ := res.RowsAffected()
		if aff == 1 {
			inserted++
			fmt.Printf("[%d/%d] %s\n", inserted, count, code)
		} else {
			// 极小概率撞码，继续
			if strings.HasPrefix(code, tier+"-") {
				_ = code
			}
		}
	}

	fmt.Printf(
		"🎉 done: %d invites inserted in %s (tier=%s)\n",
		inserted,
		time.Since(start),
		tier,
	)

	return nil
}

// ================== DB ==================

func openDB(pwd, host string) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=5432 user=wesley password=%s dbname=wenxintai sslmode=disable",
		host, pwd,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}
	return db, nil
}
