package v5

import (
	"time"

	"gorm.io/gorm"
)

// Article はブログ記事を表すGORMモデル
type Article struct {
	ID        uint           `gorm:"primarykey"        json:"id"`
	Title     string         `gorm:"size:255;not null" json:"title"`
	Body      string         `gorm:"type:text"         json:"body"`
	Author    string         `gorm:"size:100"          json:"author"`
	CreatedAt time.Time      `                         json:"created_at"`
	UpdatedAt time.Time      `                         json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"             json:"-"`
}

// CreateArticleInput はPOSTリクエストのバインド用
type CreateArticleInput struct {
	Title  string `json:"title"  binding:"required,max=255"`
	Body   string `json:"body"`
	Author string `json:"author" binding:"max=100"`
}

// UpdateArticleInput はPUTリクエストのバインド用
type UpdateArticleInput struct {
	Title  *string `json:"title"  binding:"omitempty,min=1,max=255"`
	Body   *string `json:"body"`
	Author *string `json:"author" binding:"omitempty,min=1,max=100"`
}
