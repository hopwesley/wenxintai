package database

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

var dbPassword = flag.String("pwd", "", "database password for user wenxintai")
var flagN = flag.Int("no", 10, "number of invite codes to create")

func TestQueryTestTable(t *testing.T) {
	flag.Parse()

	if *dbPassword == "" {
		fmt.Println("⚠️ 未提供密码，可使用命令行参数: go test -v -args -dbpass=你的密码")
	}

	connStr := fmt.Sprintf(
		"host=localhost port=5432 user=tweetcat password=%s dbname=hyperorchid sslmode=disable",
		*dbPassword,
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

// ==== 固定参数（按需改动即可） ====
const (
	tier       = "B"                                // B=基础版, P=专业版, C=校园版（写死）
	expireDays = 30                                 // 过期天数；0 表示不过期（写死）
	alphabet   = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Crockford Base32，无 I,O,L,0,1
)

// 仅密码通过 flag 输入；其余连接参数固定

// ---- 工具函数 ----

func randBase32(n int) (string, error) {
	out := make([]byte, n)
	for i := 0; i < n; i++ {
		// 用 5 bit 随机值映射到 32 个字符
		var b [1]byte
		if _, err := rand.Read(b[:]); err != nil {
			return "", err
		}
		out[i] = alphabet[int(b[0])&31]
	}
	return string(out), nil
}

func checksumChar(s string) byte {
	h := sha1.Sum([]byte(s))
	return alphabet[int(h[0])&31] // 取 5 bit 作为校验位
}

// 生成：TIER-XXXX-XXXX-XXXX-C
func makeInviteCode() (string, error) {
	b1, err := randBase32(4)
	if err != nil {
		return "", err
	}
	b2, err := randBase32(4)
	if err != nil {
		return "", err
	}
	b3, err := randBase32(4)
	if err != nil {
		return "", err
	}
	body := fmt.Sprintf("%s-%s-%s-%s", tier, b1, b2, b3)
	c := checksumChar(body)
	return fmt.Sprintf("%s-%c", body, c), nil
}

func openDB(t *testing.T, pwd string) *sql.DB {
	t.Helper()
	if pwd == "" {
		t.Fatal("missing -pwd (database password for user wenxintai)")
	}
	// 固定连接参数（仅密码可变）
	dsn := fmt.Sprintf(
		"host=127.0.0.1 port=5432 user=wenxintai password=%s dbname=wenxintai sslmode=disable",
		pwd,
	)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("db.Ping: %v", err)
	}
	return db
}

// ---- 主测试：生成并写入 ----

func TestBootstrapInvites(t *testing.T) {
	flag.Parse()

	db := openDB(t, *dbPassword)
	defer db.Close()

	var expiresAt *time.Time
	if expireDays > 0 {
		x := time.Now().Add(time.Duration(expireDays) * 24 * time.Hour)
		expiresAt = &x
	}

	ctx := context.Background()
	inserted := 0
	start := time.Now()

	for inserted < *flagN {
		code, err := makeInviteCode()
		if err != nil {
			t.Fatalf("makeInviteCode: %v", err)
		}

		// 依赖 PK(code) 去重；撞码则跳过继续，直到插满 targetCount
		const q = `INSERT INTO app.invites (code, expires_at) VALUES ($1, $2)
		           ON CONFLICT (code) DO NOTHING`
		res, err := db.ExecContext(ctx, q, code, expiresAt)
		if err != nil {
			// 遇到权限/表不存在等问题直报错；duplicate 已由 ON CONFLICT 处理
			t.Fatalf("insert: %v", err)
		}
		aff, _ := res.RowsAffected()
		if aff == 1 {
			inserted++
			t.Logf("[%d/%d] %s", inserted, *flagN, code)
		} else {
			// 未插入（极小概率撞码），继续生成下一个
			if strings.HasPrefix(code, tier+"-") {
				_ = code // 仅避免编译器“未使用变量”的警告
			}
		}
	}

	t.Logf("done: %d invites inserted in %s (tier=%s, expireDays=%d)",
		inserted, time.Since(start), tier, expireDays)
}
