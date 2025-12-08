package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hopwesley/wenxintai/server/ai_api"
	"github.com/hopwesley/wenxintai/server/comm"
	"github.com/hopwesley/wenxintai/server/dbSrv"
	"github.com/hopwesley/wenxintai/server/srv"
	"golang.org/x/sys/unix"
)

func bumpRlimitNoFile() {
	var r unix.Rlimit
	if err := unix.Getrlimit(unix.RLIMIT_NOFILE, &r); err != nil {
		fmt.Printf("Getrlimit failed: %v", err)
		return
	}

	// 把软限制拉到硬限制（如果你有权限）
	r.Cur = r.Max

	if err := unix.Setrlimit(unix.RLIMIT_NOFILE, &r); err != nil {
		fmt.Printf("Setrlimit failed: %v", err)
		return
	}

	fmt.Printf("RLIMIT_NOFILE set to %d", r.Cur)
}

func main() {
	bumpRlimitNoFile()

	env := "prod"
	if len(os.Args) > 1 && os.Args[1] == "dev" {
		env = "dev"
	}
	fmt.Printf("\n[main] starting server with env=%s\n", env)

	cfg, err := loadAppConfig(env)
	if err != nil {
		panic(fmt.Sprintf("load config: %v", err))
	}

	comm.SetLogLevel(cfg.DebugLevel)

	err = dbSrv.Instance().Init(cfg.Database)
	if err != nil {
		panic(fmt.Sprintf("create database: %v", err))
	}

	err = srv.Instance().Init(cfg.Server, cfg.PaymentCfg, cfg.MiniAppCfg)
	if err != nil {
		panic(fmt.Sprintf("create http service: %v", err))
	}

	err = ai_api.Instance().Init(cfg.AIApi)
	if err != nil {
		panic(fmt.Sprintf("ai api service: %v", err))
	}

	srv.Instance().StartServing()

	waitShutDown()
}

func waitShutDown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Instance().Shutdown(ctx); err != nil {
		comm.LogInst().Err(err).Msg("server shutdown error")
	}

	if err := dbSrv.Instance().Shutdown(ctx); err != nil {
		comm.LogInst().Err(err).Msg("database shutdown error")
	}
	comm.LogInst().Info().Msg("server stopped")
}
