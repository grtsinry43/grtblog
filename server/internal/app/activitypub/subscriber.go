package activitypub

import (
	"context"
	"log"
	"strings"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/article"
	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/moment"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/thinking"
)

type handlerFunc func(ctx context.Context, event appEvent.Event) error

func (h handlerFunc) Handle(ctx context.Context, event appEvent.Event) error {
	return h(ctx, event)
}

// RegisterSubscribers wires activitypub auto-publish handlers to content publish events.
func RegisterSubscribers(bus appEvent.Bus, svc *Service) {
	if bus == nil || svc == nil {
		return
	}

	register := func(eventName string, sourceType string, extractID func(appEvent.Event) int64) {
		bus.Subscribe(eventName, handlerFunc(func(ctx context.Context, event appEvent.Event) error {
			sourceID := extractID(event)
			if sourceID <= 0 {
				return nil
			}
			_, err := svc.Publish(ctx, "", PublishCmd{
				SourceType:    sourceType,
				SourceID:      sourceID,
				TriggerSource: "auto",
			})
			if err == nil || isSkippableAutoPublishError(err) {
				return nil
			}
			log.Printf("[activitypub] auto publish failed event=%s source=%s:%d err=%v", eventName, sourceType, sourceID, err)
			return nil
		}))
	}

	register(article.ArticlePublished{}.Name(), "article", func(event appEvent.Event) int64 {
		payload, ok := event.(article.ArticlePublished)
		if !ok {
			return 0
		}
		return payload.ID
	})
	register(moment.MomentPublished{}.Name(), "moment", func(event appEvent.Event) int64 {
		payload, ok := event.(moment.MomentPublished)
		if !ok {
			return 0
		}
		return payload.ID
	})
	register(thinking.ThinkingCreated{}.Name(), "thinking", func(event appEvent.Event) int64 {
		payload, ok := event.(thinking.ThinkingCreated)
		if !ok {
			return 0
		}
		return payload.ID
	})
}

func isSkippableAutoPublishError(err error) bool {
	if err == nil {
		return true
	}
	msg := strings.ToLower(strings.TrimSpace(err.Error()))
	if msg == "" {
		return true
	}
	skips := []string{
		"activitypub disabled",
		"activitypub outbound disabled",
		"source type is not allowed by activitypub.publishtypes",
		"instance url is empty",
	}
	for _, item := range skips {
		if strings.Contains(msg, item) {
			return true
		}
	}
	return false
}
