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

type AnalyticsVisitorView struct {
	VisitorID   string    `gorm:"column:visitor_id;size:255;primaryKey"`
	ContentType string    `gorm:"column:content_type;size:20;primaryKey"`
	ContentID   int64     `gorm:"column:content_id;primaryKey"`
	LastIP      string    `gorm:"column:last_ip;size:64"`
	Platform    string    `gorm:"column:platform;size:45"`
	Browser     string    `gorm:"column:browser;size:45"`
	Location    string    `gorm:"column:location;size:255"`
	FirstViewAt time.Time `gorm:"column:first_view_at;not null"`
	LastViewAt  time.Time `gorm:"column:last_view_at;not null"`
	ViewCount   int64     `gorm:"column:view_count;not null"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (AnalyticsVisitorView) TableName() string { return "analytics_visitor_view" }

type AnalyticsRSSAccessHourly struct {
	HourBucket  time.Time `gorm:"column:hour_bucket;primaryKey"`
	RequestPath string    `gorm:"column:request_path;size:64;primaryKey"`
	IP          string    `gorm:"column:ip;size:64;primaryKey"`
	ClientName  string    `gorm:"column:client_name;size:128;primaryKey"`
	ClientHint  string    `gorm:"column:client_hint;size:128"`
	UserAgent   string    `gorm:"column:user_agent;size:512"`
	Platform    string    `gorm:"column:platform;size:45"`
	Browser     string    `gorm:"column:browser;size:45"`
	Location    string    `gorm:"column:location;size:255"`
	Requests    int64     `gorm:"column:requests;not null"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (AnalyticsRSSAccessHourly) TableName() string { return "analytics_rss_access_hourly" }
