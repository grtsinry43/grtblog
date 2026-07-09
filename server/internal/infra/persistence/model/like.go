package model

import "time"

type ContentLike struct {
	ID         int64     `gorm:"column:id;primaryKey"`
	TargetType string    `gorm:"column:target_type;size:32;not null"`
	TargetID   int64     `gorm:"column:target_id;not null"`
	UserID     *int64    `gorm:"column:user_id"`
	VisitorID  string    `gorm:"column:visitor_id;size:255"`
	ClientFP   string    `gorm:"column:client_fp;size:64"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (ContentLike) TableName() string { return "content_like" }
