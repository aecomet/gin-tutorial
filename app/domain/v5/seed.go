package v5

import (
	"log/slog"

	"gin-tutorial/app/db"
)

// RunSeed は articles テーブルにサンプルデータを投入する
// 既にデータが存在する場合はスキップする
func RunSeed() error {
	var count int64
	db.DB.Model(&Article{}).Count(&count)
	if count > 0 {
		slog.Info("seed skipped: articles already exist", slog.Int64("count", count))
		return nil
	}

	articles := []Article{
		{
			Title:  "GORMとGinで作るREST API",
			Body:   "GORMはGoで最も使われているORMライブラリです。Ginと組み合わせることで高速なREST APIを構築できます。",
			Author: "Alice",
		},
		{
			Title:  "Dockerでローカル開発環境を構築する",
			Body:   "Docker Composeを使うことで、MySQLなどのミドルウェアを含む開発環境を簡単に再現できます。",
			Author: "Bob",
		},
		{
			Title:  "GoのインターフェースとDIパターン",
			Body:   "Goのインターフェースを活用したDependency Injectionパターンについて解説します。",
			Author: "Charlie",
		},
	}

	if result := db.DB.Create(&articles); result.Error != nil {
		return result.Error
	}

	slog.Info("seed completed", slog.Int("articles", len(articles)))
	return nil
}
