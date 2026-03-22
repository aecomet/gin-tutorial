package logger_test

import (
	"os"
	"path/filepath"
	"testing"

	"gin-tutorial/app/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit_CreatesLogFile(t *testing.T) {
	// t.Chdir はテスト終了時に元のディレクトリへ自動で戻る (Go 1.24+)
	t.Chdir(t.TempDir())

	cleanup, err := logger.Init()
	require.NoError(t, err)
	require.NotNil(t, cleanup)
	defer cleanup()

	_, statErr := os.Stat(filepath.Join("logs", "app.log"))
	assert.NoError(t, statErr, "logs/app.log が作成されていること")
}

func TestInit_CleanupClosesFile(t *testing.T) {
	t.Chdir(t.TempDir())

	cleanup, err := logger.Init()
	require.NoError(t, err)

	// cleanup を2回呼んでもパニックしないこと
	assert.NotPanics(t, func() {
		cleanup()
		cleanup()
	})
}
