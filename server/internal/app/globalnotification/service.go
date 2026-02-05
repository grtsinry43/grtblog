package globalnotification

import (
	"context"
	"strings"
	"time"

	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/social"
)

type Service struct {
	repo   social.GlobalNotificationRepository
	events appEvent.Bus
}

func NewService(repo social.GlobalNotificationRepository, events appEvent.Bus) *Service {
	if events == nil {
		events = appEvent.NopBus{}
	}
	return &Service{repo: repo, events: events}
}

func (s *Service) Create(ctx context.Context, cmd CreateCmd) (*social.GlobalNotification, error) {
	content := strings.TrimSpace(cmd.Content)
	if content == "" {
		return nil, ErrContentRequired
	}
	if !cmd.PublishAt.Before(cmd.ExpireAt) {
		return nil, ErrInvalidPublishWindow
	}

	allowClose := true
	if cmd.AllowClose != nil {
		allowClose = *cmd.AllowClose
	}

	entity := &social.GlobalNotification{
		Content:    content,
		PublishAt:  cmd.PublishAt,
		ExpireAt:   cmd.ExpireAt,
		AllowClose: allowClose,
	}
	if err := s.repo.Create(ctx, entity); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	_ = s.events.Publish(ctx, Created{
		ID:         entity.ID,
		Content:    entity.Content,
		PublishAt:  entity.PublishAt,
		ExpireAt:   entity.ExpireAt,
		AllowClose: entity.AllowClose,
		At:         now,
	})
	return entity, nil
}

func (s *Service) Update(ctx context.Context, cmd UpdateCmd) (*social.GlobalNotification, error) {
	entity, err := s.repo.GetByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	content := strings.TrimSpace(cmd.Content)
	if content == "" {
		return nil, ErrContentRequired
	}
	if !cmd.PublishAt.Before(cmd.ExpireAt) {
		return nil, ErrInvalidPublishWindow
	}

	entity.Content = content
	entity.PublishAt = cmd.PublishAt
	entity.ExpireAt = cmd.ExpireAt
	if cmd.AllowClose != nil {
		entity.AllowClose = *cmd.AllowClose
	}
	if err := s.repo.Update(ctx, entity); err != nil {
		return nil, err
	}

	updated, err := s.repo.GetByID(ctx, entity.ID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	_ = s.events.Publish(ctx, Updated{
		ID:         updated.ID,
		Content:    updated.Content,
		PublishAt:  updated.PublishAt,
		ExpireAt:   updated.ExpireAt,
		AllowClose: updated.AllowClose,
		At:         now,
	})
	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrInvalidNotificationID
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	_ = s.events.Publish(ctx, Deleted{ID: id, At: time.Now().UTC()})
	return nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*social.GlobalNotification, error) {
	if id <= 0 {
		return nil, ErrInvalidNotificationID
	}
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, options ListOptions) ([]social.GlobalNotification, int64, error) {
	page := options.Page
	if page < 1 {
		page = 1
	}
	size := options.PageSize
	if size < 1 {
		size = 10
	}
	if size > 100 {
		size = 100
	}
	return s.repo.List(ctx, social.GlobalNotificationListOptions{
		Status:   options.Status,
		Page:     page,
		PageSize: size,
	})
}

func (s *Service) ListActive(ctx context.Context) ([]social.GlobalNotification, error) {
	return s.repo.ListActive(ctx, time.Now().UTC())
}
