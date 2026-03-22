package main

import (
	"fmt"
	"log/slog"
	"os"

	"gin-tutorial/app/db"
	v5 "gin-tutorial/app/domain/v5"
	"gin-tutorial/app/logger"
	"gin-tutorial/app/router"
)

func main() {
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

	r := router.New()

	slog.Info("server starting", slog.String("addr", ":8080"))
	if err := r.Run(":8080"); err != nil {
		slog.Error("server failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
