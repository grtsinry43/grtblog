package rss

import (
	"context"
	"testing"
	"time"

	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/sysconfig"
	"github.com/grtsinry43/grtblog-v2/server/internal/config"
	domainconfig "github.com/grtsinry43/grtblog-v2/server/internal/domain/config"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/content"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/identity"
	domainthinking "github.com/grtsinry43/grtblog-v2/server/internal/domain/thinking"
)

type fakeContentRepo struct {
	articles []*content.Article
	moments  []*content.Moment
	pages    []*content.Page
}

func (f *fakeContentRepo) ListArticles(_ context.Context, _ content.ArticleListOptionsInternal) ([]*content.Article, int64, error) {
	return f.articles, int64(len(f.articles)), nil
}

func (f *fakeContentRepo) ListMoments(_ context.Context, _ content.MomentListOptionsInternal) ([]*content.Moment, int64, error) {
	return f.moments, int64(len(f.moments)), nil
}

func (f *fakeContentRepo) ListPages(_ context.Context, _ content.PageListOptionsInternal) ([]*content.Page, int64, error) {
	return f.pages, int64(len(f.pages)), nil
}

type fakeThinkingRepo struct {
	items []*domainthinking.Thinking
}

func (f *fakeThinkingRepo) List(_ context.Context, _ int, _ int) ([]*domainthinking.Thinking, int64, error) {
	return f.items, int64(len(f.items)), nil
}

// fakeSysConfigRepo implements domainconfig.SysConfigRepository for testing.
type fakeSysConfigRepo struct {
	items map[string]domainconfig.SysConfig
}

func (f *fakeSysConfigRepo) GetByKey(_ context.Context, key string) (*domainconfig.SysConfig, error) {
	if f == nil || f.items == nil {
		return nil, domainconfig.ErrSysConfigNotFound
	}
	if cfg, ok := f.items[key]; ok {
		return &cfg, nil
	}
	return nil, domainconfig.ErrSysConfigNotFound
}

func (f *fakeSysConfigRepo) List(_ context.Context, keys []string) ([]domainconfig.SysConfig, error) {
	if f == nil || f.items == nil {
		return nil, nil
	}
	if keys == nil {
		result := make([]domainconfig.SysConfig, 0, len(f.items))
		for _, cfg := range f.items {
			result = append(result, cfg)
		}
		return result, nil
	}
	result := make([]domainconfig.SysConfig, 0, len(keys))
	for _, key := range keys {
		if cfg, ok := f.items[key]; ok {
			result = append(result, cfg)
		}
	}
	return result, nil
}

func (f *fakeSysConfigRepo) Upsert(_ context.Context, configs []domainconfig.SysConfig) error {
	if f.items == nil {
		f.items = make(map[string]domainconfig.SysConfig)
	}
	for _, cfg := range configs {
		f.items[cfg.Key] = cfg
	}
	return nil
}

// newTestSysCfg creates a *sysconfig.Service backed by in-memory site.* config entries.
// Pass bare keys (e.g., "website_name") — the site. prefix is added automatically.
func newTestSysCfg(entries map[string]string) *sysconfig.Service {
	items := make(map[string]domainconfig.SysConfig, len(entries))
	for k, v := range entries {
		fullKey := "site." + k
		items[fullKey] = domainconfig.SysConfig{Key: fullKey, Value: v}
	}
	return sysconfig.NewService(&fakeSysConfigRepo{items: items}, config.TurnstileConfig{}, appEvent.NopBus{})
}

type fakeIdentityRepo struct {
	users map[int64]*identity.User
}

func (f *fakeIdentityRepo) FindByID(_ context.Context, id int64) (*identity.User, error) {
	if f == nil || f.users == nil {
		return nil, identity.ErrUserNotFound
	}
	if u, ok := f.users[id]; ok {
		return u, nil
	}
	return nil, identity.ErrUserNotFound
}

func (f *fakeIdentityRepo) ListAdmins(_ context.Context) ([]identity.User, error) {
	if f == nil || f.users == nil {
		return nil, nil
	}
	out := make([]identity.User, 0, len(f.users))
	for _, u := range f.users {
		if u != nil && u.IsAdmin {
			out = append(out, *u)
		}
	}
	return out, nil
}

func TestBuildAggregatesAndSortsItems(t *testing.T) {
	now := time.Date(2026, 2, 6, 12, 0, 0, 0, time.UTC)
	desc := "Page Desc"
	svc := NewService(
		&fakeContentRepo{
			articles: []*content.Article{{ID: 1, AuthorID: 101, Title: "A", Summary: "AS", Content: "# Article", ShortURL: "a", CreatedAt: now.Add(-3 * time.Hour)}},
			moments:  []*content.Moment{{ID: 2, AuthorID: 102, Title: "M", Summary: "MS", Content: "Moment **Body**", ShortURL: "m", CreatedAt: now.Add(-1 * time.Hour)}},
			pages:    []*content.Page{{ID: 3, Title: "P", Description: &desc, Content: "Page Body", ShortURL: "p", CreatedAt: now.Add(-2 * time.Hour)}},
		},
		&fakeThinkingRepo{
			items: []*domainthinking.Thinking{{ID: 4, AuthorID: 103, Content: "thinking **content**", CreatedAt: now}},
		},
		newTestSysCfg(map[string]string{
			"website_name": "My Site",
			"public_url":   "https://example.com",
			"description":  "Site Desc",
		}),
		&fakeIdentityRepo{
			users: map[int64]*identity.User{
				101: {ID: 101, Username: "a1", Nickname: "Author A", Email: "a@example.com"},
				102: {ID: 102, Username: "m1", Nickname: "Author M", Email: "m@example.com"},
				103: {ID: 103, Username: "t1", Nickname: "Author T", Email: "t@example.com"},
			},
		},
	)

	feed, err := svc.Build(context.Background(), "http://localhost:8080", 10)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	if feed.Title != "My Site" {
		t.Fatalf("unexpected feed title: %s", feed.Title)
	}
	if feed.Link != "https://example.com/" {
		t.Fatalf("unexpected feed link: %s", feed.Link)
	}
	if len(feed.Items) != 4 {
		t.Fatalf("unexpected items count: %d", len(feed.Items))
	}
	if feed.Items[0].Category != "thinking" {
		t.Fatalf("expected newest item to be thinking, got %s", feed.Items[0].Category)
	}
	if feed.Items[1].Category != "moment" {
		t.Fatalf("expected second item to be moment, got %s", feed.Items[1].Category)
	}
	if feed.Items[2].Category != "page" {
		t.Fatalf("expected third item to be page, got %s", feed.Items[2].Category)
	}
	if feed.Items[3].Category != "article" {
		t.Fatalf("expected fourth item to be article, got %s", feed.Items[3].Category)
	}
	if feed.Items[0].Description == "thinking **content**" {
		t.Fatalf("expected html description, got raw markdown")
	}
	if feed.Items[0].Description == "" {
		t.Fatalf("expected non-empty description")
	}
	if feed.Items[0].AuthorName == "" {
		t.Fatalf("expected item author")
	}
}

func TestBuildRespectsLimit(t *testing.T) {
	now := time.Date(2026, 2, 6, 12, 0, 0, 0, time.UTC)
	svc := NewService(
		&fakeContentRepo{
			articles: []*content.Article{{ID: 1, Title: "A", Summary: "S", ShortURL: "a", CreatedAt: now.Add(-2 * time.Hour)}},
			moments:  []*content.Moment{{ID: 2, Title: "M", Summary: "S", ShortURL: "m", CreatedAt: now.Add(-1 * time.Hour)}},
		},
		&fakeThinkingRepo{items: []*domainthinking.Thinking{{ID: 3, Content: "T", CreatedAt: now}}},
		nil,
		nil,
	)

	feed, err := svc.Build(context.Background(), "http://localhost:8080", 2)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if len(feed.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(feed.Items))
	}
	if feed.Items[0].Category != "thinking" || feed.Items[1].Category != "moment" {
		t.Fatalf("unexpected categories after limit: %s, %s", feed.Items[0].Category, feed.Items[1].Category)
	}
}

func TestBuildUsesOGDescriptionFallback(t *testing.T) {
	now := time.Date(2026, 2, 6, 12, 0, 0, 0, time.UTC)
	svc := NewService(
		&fakeContentRepo{
			articles: []*content.Article{{ID: 1, Title: "A", Content: "A", ShortURL: "a", CreatedAt: now}},
		},
		nil,
		newTestSysCfg(map[string]string{
			"website_name":   "My Site",
			"description":    "",
			"og_description": "OG Desc",
			"public_url":     "https://example.com",
		}),
		nil,
	)

	feed, err := svc.Build(context.Background(), "http://localhost:8080", 10)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if feed.Description != "OG Desc" {
		t.Fatalf("expected OG Desc fallback, got %q", feed.Description)
	}
}

func TestBuildIncludesItemCoverImage(t *testing.T) {
	now := time.Date(2026, 2, 6, 12, 0, 0, 0, time.UTC)
	articleCover := "/uploads/article-cover.jpg"
	momentImages := "https://cdn.example.com/m1.png, https://cdn.example.com/m2.png"

	svc := NewService(
		&fakeContentRepo{
			articles: []*content.Article{{
				ID:        1,
				Title:     "A",
				Content:   "A",
				ShortURL:  "a",
				Cover:     &articleCover,
				CreatedAt: now.Add(-1 * time.Hour),
			}},
			moments: []*content.Moment{{
				ID:        2,
				Title:     "M",
				Content:   "M",
				ShortURL:  "m",
				Image:     &momentImages,
				CreatedAt: now,
			}},
		},
		nil,
		newTestSysCfg(map[string]string{
			"website_name": "My Site",
			"public_url":   "https://example.com",
		}),
		nil,
	)

	feed, err := svc.Build(context.Background(), "http://localhost:8080", 10)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	var articleImage string
	var momentImage string
	for _, item := range feed.Items {
		switch item.Category {
		case "article":
			articleImage = item.ImageURL
		case "moment":
			momentImage = item.ImageURL
		}
	}

	if articleImage != "https://example.com/uploads/article-cover.jpg" {
		t.Fatalf("unexpected article image url: %q", articleImage)
	}
	if momentImage != "https://cdn.example.com/m1.png" {
		t.Fatalf("unexpected moment image url: %q", momentImage)
	}
}
