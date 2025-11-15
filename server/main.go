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
)

func main() {
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

	go srv.Instance().StartServing()

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
