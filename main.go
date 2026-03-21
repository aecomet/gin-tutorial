package main

import (
	"fmt"
	"log/slog"
	"os"

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

	r := router.New()

	slog.Info("server starting", slog.String("addr", ":8080"))
	if err := r.Run(":8080"); err != nil {
		slog.Error("server failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
