// package redis は Redis クライアントのシングルトンを管理する。
// gRPC サーバーのストレージバックエンドとして使用する。
package redis

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

var client *redis.Client

// Init は環境変数から Redis 接続情報を読み取りクライアントを初期化する。
// 接続確認のため Ping を発行する。
func Init() error {
	host := getEnv("REDIS_HOST", "localhost")
	port := getEnv("REDIS_PORT", "6379")

	client = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", host, port),
	})

	// 接続確認
	if err := client.Ping(context.Background()).Err(); err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}
	return nil
}

// Client はシングルトンの Redis クライアントを返す。
func Client() *redis.Client {
	return client
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
