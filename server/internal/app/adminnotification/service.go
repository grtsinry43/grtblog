package adminnotification

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/social"
)

type Service struct {
	repo   social.AdminNotificationRepository
	events appEvent.Bus
}

func NewService(repo social.AdminNotificationRepository, events ...appEvent.Bus) *Service {
	var bus appEvent.Bus = appEvent.NopBus{}
	if len(events) > 0 && events[0] != nil {
		bus = events[0]
	}
	return &Service{repo: repo, events: bus}
}

func (s *Service) Create(ctx context.Context, userID int64, notifType, title, content string, payload any) (*social.AdminNotification, error) {
	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)
	if title == "" || content == "" || userID <= 0 {
		return nil, ErrInvalidNotification
	}
	raw := json.RawMessage("{}")
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		raw = encoded
	}
	item := &social.AdminNotification{
		UserID:    userID,
		NotifType: strings.TrimSpace(notifType),
		Title:     title,
		Content:   content,
		Payload:   raw,
		IsRead:    false,
	}
	if err := s.repo.Create(ctx, item); err != nil {
		return nil, err
	}
	_ = s.events.Publish(ctx, appEvent.Generic{
		EventName: "admin.notification.created",
		At:        time.Now().UTC(),
		Payload: map[string]any{
			"ID":        item.ID,
			"UserID":    item.UserID,
			"NotifType": item.NotifType,
			"Title":     item.Title,
			"Content":   item.Content,
			"Payload":   payload,
		},
	})
	return item, nil
}

func (s *Service) ListByUser(ctx context.Context, userID int64, unreadOnly bool, page, pageSize int) ([]social.AdminNotification, int64, error) {
	if userID <= 0 {
		return nil, 0, ErrInvalidNotification
	}
	return s.repo.ListByUser(ctx, userID, social.AdminNotificationListOptions{
		UnreadOnly: unreadOnly,
		Page:       page,
		PageSize:   pageSize,
	})
}

func (s *Service) MarkRead(ctx context.Context, userID, id int64) error {
	if userID <= 0 || id <= 0 {
		return ErrInvalidNotification
	}
	return s.repo.MarkRead(ctx, userID, id)
}

func (s *Service) MarkAllRead(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return ErrInvalidNotification
	}
	return s.repo.MarkAllRead(ctx, userID)
}
