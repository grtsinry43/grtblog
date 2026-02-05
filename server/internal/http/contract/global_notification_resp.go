package contract

import (
	"time"

	"github.com/grtsinry43/grtblog-v2/server/internal/domain/social"
)

type GlobalNotificationResp struct {
	ID         int64     `json:"id"`
	Content    string    `json:"content"`
	PublishAt  time.Time `json:"publishAt"`
	ExpireAt   time.Time `json:"expireAt"`
	AllowClose bool      `json:"allowClose"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type GlobalNotificationListResp struct {
	Items []GlobalNotificationResp `json:"items"`
	Total int64                    `json:"total"`
	Page  int                      `json:"page"`
	Size  int                      `json:"size"`
}

func ToGlobalNotificationResp(item social.GlobalNotification) GlobalNotificationResp {
	return GlobalNotificationResp{
		ID:         item.ID,
		Content:    item.Content,
		PublishAt:  item.PublishAt,
		ExpireAt:   item.ExpireAt,
		AllowClose: item.AllowClose,
		CreatedAt:  item.CreatedAt,
		UpdatedAt:  item.UpdatedAt,
	}
}
