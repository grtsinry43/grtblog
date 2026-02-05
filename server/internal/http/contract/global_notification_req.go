package contract

import "time"

type GlobalNotificationCreateReq struct {
	Content    string    `json:"content"`
	PublishAt  time.Time `json:"publishAt"`
	ExpireAt   time.Time `json:"expireAt"`
	AllowClose *bool     `json:"allowClose,omitempty"`
}

type GlobalNotificationUpdateReq struct {
	Content    string    `json:"content"`
	PublishAt  time.Time `json:"publishAt"`
	ExpireAt   time.Time `json:"expireAt"`
	AllowClose *bool     `json:"allowClose,omitempty"`
}
