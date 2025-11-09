package database

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"testing"

	_ "github.com/lib/pq"
)

var dbPassword = ""

func init() {
	flag.StringVar(&dbPassword, "dbpass", "", "PostgreSQL password for tweetcat user")
}

func TestQueryTestTable(t *testing.T) {
	flag.Parse()

	if dbPassword == "" {
		fmt.Println("⚠️ 未提供密码，可使用命令行参数: go test -v -args -dbpass=你的密码")
	}

	connStr := fmt.Sprintf(
		"host=localhost port=5432 user=tweetcat password=%s dbname=hyperorchid sslmode=disable",
		dbPassword,
	)
	// 连接数据库
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 检查连接是否可用
	err = db.Ping()
	if err != nil {
		t.Fatalf("无法连接到数据库: %v", err)
	}
	fmt.Println("✅ 成功连接到 PostgreSQL")

	// 查询 test 表
	rows, err := db.Query("SELECT id, name FROM test;")
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	defer rows.Close()

	fmt.Println("📋 查询结果:")
	for rows.Next() {
		var id int
		var name string
		err := rows.Scan(&id, &name)
		if err != nil {
			log.Printf("扫描行失败: %v", err)
			continue
		}
		fmt.Printf("id=%d, name=%s\n", id, name)
	}

	if err = rows.Err(); err != nil {
		t.Fatalf("遍历行时出错: %v", err)
	}
}
