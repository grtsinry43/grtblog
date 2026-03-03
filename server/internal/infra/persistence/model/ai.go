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

type AITaskLog struct {
	ID            int64     `gorm:"column:id;primaryKey"`
	TaskType      string    `gorm:"column:task_type;size:40;not null"`
	ModelName     string    `gorm:"column:model_name;size:200;not null;default:''"`
	ProviderName  string    `gorm:"column:provider_name;size:100;not null;default:''"`
	Status        string    `gorm:"column:status;size:20;not null;default:'pending'"`
	InputText     string    `gorm:"column:input_text;type:text;not null;default:''"`
	OutputText    string    `gorm:"column:output_text;type:text;not null;default:''"`
	ErrorMessage  *string   `gorm:"column:error_message;type:text"`
	DurationMs    int       `gorm:"column:duration_ms;not null;default:0"`
	TriggerSource string    `gorm:"column:trigger_source;size:40;not null;default:'manual'"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (AITaskLog) TableName() string { return "ai_task_log" }
