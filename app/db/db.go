package db

import (
	"fmt"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB はアプリケーション全体で共有するGORMインスタンス
var DB *gorm.DB

// Init はデータベース接続を初期化する
func Init() error {
	return InitWithDialector(mysql.Open(BuildDSN()))
}

// InitWithDialector は任意のDialectorでDB接続を初期化する（テスト用途を含む）
func InitWithDialector(dialector gorm.Dialector) error {
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	DB = db
	return nil
}

// BuildDSN は環境変数からMySQL接続文字列を組み立てる
func BuildDSN() string {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "3306")
	user := getEnv("DB_USER", "root")
	pass := getEnv("DB_PASSWORD", "root")
	name := getEnv("DB_NAME", "gin_tutorial")
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, pass, host, port, name,
	)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
