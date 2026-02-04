package model

import "time"

type AnalyticsContentHourly struct {
	ContentType string    `gorm:"column:content_type;size:20;primaryKey"`
	ContentID   int64     `gorm:"column:content_id;primaryKey"`
	HourBucket  time.Time `gorm:"column:hour_bucket;primaryKey"`
	PV          int64     `gorm:"column:pv;not null"`
	UV          int64     `gorm:"column:uv;not null"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (AnalyticsContentHourly) TableName() string { return "analytics_content_hourly" }

type AnalyticsOnlineHourly struct {
	HourBucket  time.Time `gorm:"column:hour_bucket;primaryKey"`
	PeakOnline  int64     `gorm:"column:peak_online;not null"`
	SampleTotal int64     `gorm:"column:sample_total;not null"`
	SampleCount int64     `gorm:"column:sample_count;not null"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (AnalyticsOnlineHourly) TableName() string { return "analytics_online_hourly" }
