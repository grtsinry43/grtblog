package isr

import (
	"context"
	"fmt"
	"strings"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/article"
	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/moment"
)

type handlerFunc func(ctx context.Context, event appEvent.Event) error

func (h handlerFunc) Handle(ctx context.Context, event appEvent.Event) error {
	return h(ctx, event)
}

func RegisterArticleSubscribers(bus appEvent.Bus, service *Service) {
	if bus == nil || service == nil {
		return
	}

	register := func(eventName string) {
		bus.Subscribe(eventName, handlerFunc(func(ctx context.Context, event appEvent.Event) error {
			articleID, shortURL := extractArticleEventPayload(event)
			if articleID <= 0 {
				return nil
			}

			deps := []string{
				"home:recent-posts",
				"home:activity-pulse",
				"home:inspiration-stats",
				"post:list:page:1",
				"post:list:page:2",
				"post:list:page:3",
				fmt.Sprintf("post:detail:%d", articleID),
			}
			urls := []string{
				"/",
				"/posts",
				"/posts/page/1",
				"/posts/page/2",
				"/posts/page/3",
			}
			if shortURL != "" {
				urls = append(urls, fmt.Sprintf("/posts/%s", shortURL))
			}
			return service.Invalidate(ctx, deps, urls)
		}))
	}

	register(article.ArticleCreated{}.Name())
	register(article.ArticleUpdated{}.Name())
	register(article.ArticlePublished{}.Name())
	register(article.ArticleUnpublished{}.Name())
	register(article.ArticleDeleted{}.Name())
}

func RegisterMomentSubscribers(bus appEvent.Bus, service *Service) {
	if bus == nil || service == nil {
		return
	}

	register := func(eventName string) {
		bus.Subscribe(eventName, handlerFunc(func(ctx context.Context, event appEvent.Event) error {
			momentID, shortURL := extractMomentEventPayload(event)
			if momentID <= 0 {
				return nil
			}

			deps := []string{
				"home:recent-moments",
				"home:activity-pulse",
				"home:inspiration-stats",
				fmt.Sprintf("moment:detail:%d", momentID),
			}
			_ = shortURL // moments detail URL uses date segments; dep invalidation handles tracked URLs.
			urls := []string{
				"/",
				"/moments",
			}
			return service.Invalidate(ctx, deps, urls)
		}))
	}

	register(moment.MomentCreated{}.Name())
	register(moment.MomentUpdated{}.Name())
	register(moment.MomentPublished{}.Name())
	register(moment.MomentUnpublished{}.Name())
	register(moment.MomentDeleted{}.Name())
}

func extractArticleEventPayload(event appEvent.Event) (articleID int64, shortURL string) {
	switch e := event.(type) {
	case article.ArticleCreated:
		return e.ID, strings.TrimSpace(e.ShortURL)
	case article.ArticleUpdated:
		return e.ID, strings.TrimSpace(e.ShortURL)
	case article.ArticlePublished:
		return e.ID, strings.TrimSpace(e.ShortURL)
	case article.ArticleUnpublished:
		return e.ID, strings.TrimSpace(e.ShortURL)
	case article.ArticleDeleted:
		return e.ID, strings.TrimSpace(e.ShortURL)
	default:
		return 0, ""
	}
}

func extractMomentEventPayload(event appEvent.Event) (momentID int64, shortURL string) {
	switch e := event.(type) {
	case moment.MomentCreated:
		return e.ID, strings.TrimSpace(e.ShortURL)
	case moment.MomentUpdated:
		return e.ID, strings.TrimSpace(e.ShortURL)
	case moment.MomentPublished:
		return e.ID, strings.TrimSpace(e.ShortURL)
	case moment.MomentUnpublished:
		return e.ID, strings.TrimSpace(e.ShortURL)
	case moment.MomentDeleted:
		return e.ID, strings.TrimSpace(e.ShortURL)
	default:
		return 0, ""
	}
}
