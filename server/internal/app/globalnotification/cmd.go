package globalnotification

import "time"

type CreateCmd struct {
	Content    string
	PublishAt  time.Time
	ExpireAt   time.Time
	AllowClose *bool
}

type UpdateCmd struct {
	ID         int64
	Content    string
	PublishAt  time.Time
	ExpireAt   time.Time
	AllowClose *bool
}

type ListOptions struct {
	Status   string
	Page     int
	PageSize int
}
