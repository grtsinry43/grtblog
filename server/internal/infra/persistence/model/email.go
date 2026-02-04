package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type EmailTemplate struct {
	ID              int64          `gorm:"column:id;primaryKey"`
	Code            string         `gorm:"column:code;size:80;not null;uniqueIndex:uq_email_template_code"`
	Name            string         `gorm:"column:name;size:120;not null"`
	EventName       string         `gorm:"column:event_name;size:120;not null"`
	SubjectTemplate string         `gorm:"column:subject_template;type:text;not null"`
	HTMLTemplate    string         `gorm:"column:html_template;type:text;not null"`
	TextTemplate    string         `gorm:"column:text_template;type:text;not null"`
	ToEmails        datatypes.JSON `gorm:"column:to_emails;type:jsonb;not null"`
	IsEnabled       bool           `gorm:"column:is_enabled;not null"`
	CreatedAt       time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (EmailTemplate) TableName() string { return "email_template" }

type EmailOutbox struct {
	ID           int64          `gorm:"column:id;primaryKey"`
	TemplateID   *int64         `gorm:"column:template_id"`
	TemplateCode string         `gorm:"column:template_code;size:80;not null"`
	EventName    string         `gorm:"column:event_name;size:120;not null"`
	ToEmails     datatypes.JSON `gorm:"column:to_emails;type:jsonb;not null"`
	Subject      string         `gorm:"column:subject;type:text;not null"`
	HTMLBody     string         `gorm:"column:html_body;type:text;not null"`
	TextBody     string         `gorm:"column:text_body;type:text;not null"`
	Status       string         `gorm:"column:status;size:20;not null"`
	RetryCount   int            `gorm:"column:retry_count;not null"`
	NextRetryAt  time.Time      `gorm:"column:next_retry_at;not null"`
	LastError    string         `gorm:"column:last_error;type:text"`
	SentAt       *time.Time     `gorm:"column:sent_at"`
	CreatedAt    time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;autoUpdateTime"`
}

func (EmailOutbox) TableName() string { return "email_outbox" }

type EmailSubscription struct {
	ID             int64      `gorm:"column:id;primaryKey"`
	Email          string     `gorm:"column:email;size:255;not null"`
	EventName      string     `gorm:"column:event_name;size:120;not null"`
	Status         string     `gorm:"column:status;size:20;not null"`
	Token          string     `gorm:"column:token;size:80;not null;uniqueIndex:uq_email_subscription_token"`
	SourceIP       string     `gorm:"column:source_ip;size:45"`
	UnsubscribedAt *time.Time `gorm:"column:unsubscribed_at"`
	CreatedAt      time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (EmailSubscription) TableName() string { return "email_subscription" }
