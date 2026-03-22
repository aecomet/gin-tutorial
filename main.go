package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"

	"gin-tutorial/app/db"
	v5 "gin-tutorial/app/domain/v5"
	grpcserver "gin-tutorial/app/grpc/server"
	"gin-tutorial/app/logger"
	rdb "gin-tutorial/app/redis"
	"gin-tutorial/app/router"
)

func main() {
	// .env ファイルが存在する場合は環境変数に読み込む。
	// ファイルが存在しない場合はスキップする（本番環境では直接環境変数を設定する）。
	// すでに設定済みの環境変数は上書きしない。
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "warning: failed to load .env: %v\n", err)
	}
	cleanup, err := logger.Init()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	if err := db.Init(); err != nil {
		slog.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	slog.Info("database connected")

	if err := v5.RunMigrations(); err != nil {
		slog.Error("migration failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	slog.Info("migration completed")

	if os.Getenv("DB_SEED") == "true" {
		if err := v5.RunSeed(); err != nil {
			slog.Error("seed failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}

	// Redis に接続する（gRPC サーバーのストレージとして使用）
	if err := rdb.Init(); err != nil {
		slog.Error("failed to connect to redis", slog.String("error", err.Error()))
		os.Exit(1)
	}
	slog.Info("redis connected")

	// gRPC サーバーを別 goroutine で起動する。
	// Gin（HTTP）と gRPC は独立したポートで並行動作する。
	go func() {
		if err := grpcserver.Start("50051"); err != nil {
			slog.Error("gRPC server failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	r := router.New()

	slog.Info("server starting", slog.String("addr", ":8080"))
	if err := r.Run(":8080"); err != nil {
		slog.Error("server failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
