package v5

import (
	"fmt"

	"gin-tutorial/app/db"
)

// RunMigrations はarticlesテーブルを自動マイグレーションする
func RunMigrations() error {
	if err := db.DB.AutoMigrate(&Article{}); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	return nil
}
