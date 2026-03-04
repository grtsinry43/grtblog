package friendlink

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/social"
)

type Service struct {
	repo   social.FriendLinkApplicationRepository
	events appEvent.Bus
}

func NewService(repo social.FriendLinkApplicationRepository, events appEvent.Bus) *Service {
	if events == nil {
		events = appEvent.NopBus{}
	}
	return &Service{repo: repo, events: events}
}

type SubmitCmd struct {
	Name        string
	URL         string
	Logo        string
	Description string
	Message     string
	RSSURL      string
	UserID      *int64
}

type SubmitResult struct {
	Application social.FriendLinkApplication
	Created     bool
}

func (s *Service) Submit(ctx context.Context, cmd SubmitCmd) (*SubmitResult, error) {
	url := strings.TrimSpace(cmd.URL)
	existing, err := s.repo.FindByURL(ctx, url)
	if err != nil && !errors.Is(err, social.ErrFriendLinkApplicationNotFound) {
		return nil, err
	}

	if existing == nil {
		app := &social.FriendLinkApplication{
			Name:              toOptionalString(cmd.Name),
			URL:               url,
			Logo:              toOptionalString(cmd.Logo),
			Description:       toOptionalString(cmd.Description),
			ApplyChannel:      social.FriendLinkApplyChannelUser,
			RequestedSyncMode: requestedSyncMode(cmd.RSSURL),
			RSSURL:            toOptionalString(cmd.RSSURL),
			Manifest:          json.RawMessage("{}"),
			UserID:            cmd.UserID,
			Message:           toOptionalString(cmd.Message),
			Status:            social.FriendLinkAppStatusPending,
		}
		if err := s.repo.Create(ctx, app); err != nil {
			return nil, err
		}
		s.publishApplicationEvent(ctx, "friendlink.application.created", app)
		return &SubmitResult{Application: *app, Created: true}, nil
	}

	if existing.Status == social.FriendLinkAppStatusBlocked {
		return nil, social.ErrFriendLinkApplicationBlocked
	}

	existing.Name = toOptionalString(cmd.Name)
	existing.Logo = toOptionalString(cmd.Logo)
	existing.Description = toOptionalString(cmd.Description)
	existing.ApplyChannel = social.FriendLinkApplyChannelUser
	existing.RequestedSyncMode = requestedSyncMode(cmd.RSSURL)
	existing.RSSURL = toOptionalString(cmd.RSSURL)
	if len(existing.Manifest) == 0 {
		existing.Manifest = json.RawMessage("{}")
	}
	existing.Message = toOptionalString(cmd.Message)
	existing.UserID = cmd.UserID
	existing.Status = social.FriendLinkAppStatusPending

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	s.publishApplicationEvent(ctx, "friendlink.application.created", existing)
	return &SubmitResult{Application: *existing, Created: false}, nil
}

func (s *Service) publishApplicationEvent(ctx context.Context, name string, app *social.FriendLinkApplication) {
	if app == nil {
		return
	}
	_ = s.events.Publish(ctx, appEvent.Generic{
		EventName: name,
		At:        time.Now(),
		Payload: map[string]any{
			"ID":     app.ID,
			"URL":    app.URL,
			"Name":   derefString(app.Name),
			"Status": app.Status,
		},
	})
}

func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func toOptionalString(val string) *string {
	trimmed := strings.TrimSpace(val)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func requestedSyncMode(rssURL string) string {
	if strings.TrimSpace(rssURL) == "" {
		return "none"
	}
	return "rss"
}
