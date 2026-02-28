package model

import "time"

type AIProvider struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	Name      string    `gorm:"column:name;size:100;not null"`
	Type      string    `gorm:"column:type;size:20;not null"`
	APIURL    string    `gorm:"column:api_url;type:text;not null;default:''"`
	APIKey    string    `gorm:"column:api_key;type:text;not null;default:''"`
	IsActive  bool      `gorm:"column:is_active;not null;default:true"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (AIProvider) TableName() string { return "ai_provider" }

type AIModel struct {
	ID         int64     `gorm:"column:id;primaryKey"`
	ProviderID int64     `gorm:"column:provider_id;not null"`
	Name       string    `gorm:"column:name;size:100;not null"`
	ModelID    string    `gorm:"column:model_id;size:200;not null"`
	IsActive   bool      `gorm:"column:is_active;not null;default:true"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt  time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (AIModel) TableName() string { return "ai_model" }
