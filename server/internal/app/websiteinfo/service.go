package websiteinfo

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/config"
)

// Service 编排 WebsiteInfo 相关用例。
type Service struct {
	repo   config.WebsiteInfoRepository
	events appEvent.Bus
}

func NewService(repo config.WebsiteInfoRepository, events appEvent.Bus) *Service {
	if events == nil {
		events = appEvent.NopBus{}
	}
	return &Service{repo: repo, events: events}
}

type CreateCmd struct {
	Key      string
	Name     *string
	Value    *string
	InfoJSON *json.RawMessage
}

type UpdateCmd struct {
	Key      string
	Name     *string
	Value    *string
	InfoJSON *json.RawMessage
}

func (s *Service) List(ctx context.Context) ([]config.WebsiteInfo, error) {
	return s.repo.List(ctx)
}

func (s *Service) Create(ctx context.Context, cmd CreateCmd) (*config.WebsiteInfo, error) {
	info := &config.WebsiteInfo{
		Key:      strings.TrimSpace(cmd.Key),
		Name:     trimPtr(cmd.Name),
		Value:    cmd.Value,
		InfoJSON: rawMessageOrNil(cmd.InfoJSON),
	}
	if err := s.repo.Create(ctx, info); err != nil {
		return nil, err
	}
	s.publishUpdated(ctx, info)
	return info, nil
}

func (s *Service) Update(ctx context.Context, cmd UpdateCmd) (*config.WebsiteInfo, error) {
	key := strings.TrimSpace(cmd.Key)
	info, err := s.repo.GetByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	info.Key = key
	if cmd.Name != nil {
		info.Name = trimPtr(cmd.Name)
	}
	info.Value = cmd.Value
	info.InfoJSON = rawMessageOrNil(cmd.InfoJSON)
	if err := s.repo.Update(ctx, info); err != nil {
		return nil, err
	}
	s.publishUpdated(ctx, info)
	return info, nil
}

func (s *Service) Delete(ctx context.Context, key string) error {
	return s.repo.Delete(ctx, strings.TrimSpace(key))
}

func (s *Service) Get(ctx context.Context, key string) (*config.WebsiteInfo, error) {
	return s.repo.GetByKey(ctx, strings.TrimSpace(key))
}

func trimPtr(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	return &trimmed
}

func rawMessageOrNil(value *json.RawMessage) json.RawMessage {
	if value == nil {
		return nil
	}
	copied := make(json.RawMessage, len(*value))
	copy(copied, *value)
	return copied
}

func (s *Service) publishUpdated(ctx context.Context, info *config.WebsiteInfo) {
	if info == nil {
		return
	}
	_ = s.events.Publish(ctx, appEvent.Generic{
		EventName: "websiteinfo.updated",
		At:        time.Now(),
		Payload: map[string]any{
			"Key":      info.Key,
			"Name":     toString(info.Name),
			"Value":    toString(info.Value),
			"InfoJSON": string(info.InfoJSON),
		},
	})
}

func toString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
