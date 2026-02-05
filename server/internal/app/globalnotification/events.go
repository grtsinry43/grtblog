package globalnotification

import "time"

type Created struct {
	ID         int64
	Content    string
	PublishAt  time.Time
	ExpireAt   time.Time
	AllowClose bool
	At         time.Time
}

func (e Created) Name() string { return "global.notification.created" }
func (e Created) OccurredAt() time.Time {
	return e.At
}

type Updated struct {
	ID         int64
	Content    string
	PublishAt  time.Time
	ExpireAt   time.Time
	AllowClose bool
	At         time.Time
}

func (e Updated) Name() string { return "global.notification.updated" }
func (e Updated) OccurredAt() time.Time {
	return e.At
}

type Deleted struct {
	ID int64
	At time.Time
}

func (e Deleted) Name() string { return "global.notification.deleted" }
func (e Deleted) OccurredAt() time.Time {
	return e.At
}
