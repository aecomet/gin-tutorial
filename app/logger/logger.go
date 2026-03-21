package logger

import (
	"fmt"
	"log/slog"
	"os"
)

// Init はJSON形式のstructured loggerを logs/app.log に設定する。
// 戻り値のcleanup関数はプロセス終了前に呼び出すこと。
func Init() (cleanup func(), err error) {
	if err = os.MkdirAll("logs", 0o755); err != nil {
		return nil, fmt.Errorf("logger: failed to create logs dir: %w", err)
	}

	f, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("logger: failed to open log file: %w", err)
	}

	h := slog.NewJSONHandler(f, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(h))

	return func() { _ = f.Close() }, nil
}
