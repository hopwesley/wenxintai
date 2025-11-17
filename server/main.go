package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	cfg, err := loadAppConfig()
	if err != nil {
		panic(fmt.Sprintf("load config: %v", err))
	}

	comm.SetLogLevel(cfg.DebugLevel)

	err = dbSrv.Instance().Init(cfg.Database)
	if err != nil {
		panic(fmt.Sprintf("create database: %v", err))
	}

	err = srv.Instance().Init(cfg.Server)
	if err != nil {
		panic(fmt.Sprintf("create http service: %v", err))
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
