package rss

import (
	"bytes"
	"context"
	"fmt"
	"html"
	"sort"
	"strings"
	"time"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/sysconfig"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/content"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/identity"
	domainthinking "github.com/grtsinry43/grtblog-v2/server/internal/domain/thinking"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

type ContentRepository interface {
	ListArticles(ctx context.Context, options content.ArticleListOptionsInternal) ([]*content.Article, int64, error)
	ListMoments(ctx context.Context, options content.MomentListOptionsInternal) ([]*content.Moment, int64, error)
	ListPages(ctx context.Context, options content.PageListOptionsInternal) ([]*content.Page, int64, error)
}

type ThinkingRepository interface {
	List(ctx context.Context, limit, offset int) ([]*domainthinking.Thinking, int64, error)
}

type IdentityRepository interface {
	FindByID(ctx context.Context, id int64) (*identity.User, error)
	ListAdmins(ctx context.Context) ([]identity.User, error)
}

type Service struct {
	contentRepo  ContentRepository
	thinkingRepo ThinkingRepository
	sysCfg       *sysconfig.Service
	identityRepo IdentityRepository
}

type Feed struct {
	Title         string
	Description   string
	Link          string
	ImageURL      string
	AuthorName    string
	AuthorEmail   string
	FollowFeedID  string
	FollowUserID  string
	Items         []Item
	LastBuildDate time.Time
}

type Item struct {
	Title       string
	Description string
	Link        string
	GUID        string
	Category    string
	AuthorName  string
	AuthorEmail string
	PublishedAt time.Time
}

var markdownRenderer = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
	),
)

func NewService(contentRepo ContentRepository, thinkingRepo ThinkingRepository, sysCfg *sysconfig.Service, identityRepo IdentityRepository) *Service {
	return &Service{
		contentRepo:  contentRepo,
		thinkingRepo: thinkingRepo,
		sysCfg:       sysCfg,
		identityRepo: identityRepo,
	}
}

func (s *Service) Build(ctx context.Context, requestBaseURL string, limit int) (*Feed, error) {
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	site := s.loadSiteMetadata(ctx)
	author := s.loadDefaultAuthor(ctx)
	authorCache := make(map[int64]authorMetadata)
	baseURL := normalizeBaseURL(site.publicURL)
	if baseURL == "" {
		baseURL = normalizeBaseURL(requestBaseURL)
	}

	items := make([]Item, 0, limit*4)
	if s.contentRepo != nil {
		published := true
		articles, _, err := s.contentRepo.ListArticles(ctx, content.ArticleListOptionsInternal{
			Page:      1,
			PageSize:  limit,
			Published: &published,
		})
		if err != nil {
			return nil, err
		}
		for _, article := range articles {
			if article == nil {
				continue
			}
			items = append(items, Item{
				Title:       article.Title,
				Description: buildRSSDescription(buildURL(baseURL, "/articles/short/"+article.ShortURL), renderHTML(article.Content)),
				Link:        buildURL(baseURL, "/articles/short/"+article.ShortURL),
				GUID:        fmt.Sprintf("article-%d", article.ID),
				Category:    "article",
				AuthorName:  s.resolveAuthorByUserID(ctx, article.AuthorID, authorCache).name,
				AuthorEmail: s.resolveAuthorByUserID(ctx, article.AuthorID, authorCache).email,
				PublishedAt: article.CreatedAt,
			})
		}

		moments, _, err := s.contentRepo.ListMoments(ctx, content.MomentListOptionsInternal{
			Page:      1,
			PageSize:  limit,
			Published: &published,
		})
		if err != nil {
			return nil, err
		}
		for _, moment := range moments {
			if moment == nil {
				continue
			}
			items = append(items, Item{
				Title:       moment.Title,
				Description: buildRSSDescription(buildURL(baseURL, "/moments/short/"+moment.ShortURL), renderHTML(moment.Content)),
				Link:        buildURL(baseURL, "/moments/short/"+moment.ShortURL),
				GUID:        fmt.Sprintf("moment-%d", moment.ID),
				Category:    "moment",
				AuthorName:  s.resolveAuthorByUserID(ctx, moment.AuthorID, authorCache).name,
				AuthorEmail: s.resolveAuthorByUserID(ctx, moment.AuthorID, authorCache).email,
				PublishedAt: moment.CreatedAt,
			})
		}

		enabled := true
		pages, _, err := s.contentRepo.ListPages(ctx, content.PageListOptionsInternal{
			Page:     1,
			PageSize: limit,
			Enabled:  &enabled,
		})
		if err != nil {
			return nil, err
		}
		for _, page := range pages {
			if page == nil {
				continue
			}
			items = append(items, Item{
				Title:       page.Title,
				Description: buildRSSDescription(buildURL(baseURL, "/pages/short/"+page.ShortURL), renderHTML(page.Content)),
				Link:        buildURL(baseURL, "/pages/short/"+page.ShortURL),
				GUID:        fmt.Sprintf("page-%d", page.ID),
				Category:    "page",
				PublishedAt: page.CreatedAt,
			})
		}
	}

	if s.thinkingRepo != nil {
		thinkings, _, err := s.thinkingRepo.List(ctx, limit, 0)
		if err != nil {
			return nil, err
		}
		for _, t := range thinkings {
			if t == nil {
				continue
			}
			contentText := strings.TrimSpace(t.Content)
			items = append(items, Item{
				Title:       buildThinkingTitle(contentText, t.ID),
				Description: buildRSSDescription(buildURL(baseURL, fmt.Sprintf("/thinkings#thinking-%d", t.ID)), renderHTML(contentText)),
				Link:        buildURL(baseURL, fmt.Sprintf("/thinkings#thinking-%d", t.ID)),
				GUID:        fmt.Sprintf("thinking-%d", t.ID),
				Category:    "thinking",
				AuthorName:  s.resolveAuthorByUserID(ctx, t.AuthorID, authorCache).name,
				AuthorEmail: s.resolveAuthorByUserID(ctx, t.AuthorID, authorCache).email,
				PublishedAt: t.CreatedAt,
			})
		}
	}

	sort.SliceStable(items, func(i, j int) bool {
		return items[i].PublishedAt.After(items[j].PublishedAt)
	})
	if len(items) > limit {
		items = items[:limit]
	}

	feed := &Feed{
		Title:         strings.TrimSpace(site.title),
		Description:   strings.TrimSpace(site.description),
		Link:          buildURL(baseURL, "/"),
		ImageURL:      resolveMediaURL(baseURL, firstNonEmpty(site.imageURL, author.avatar)),
		AuthorName:    author.name,
		AuthorEmail:   author.email,
		FollowFeedID:  strings.TrimSpace(site.followFeedID),
		FollowUserID:  strings.TrimSpace(site.followUserID),
		Items:         items,
		LastBuildDate: time.Now().UTC(),
	}
	if feed.Title == "" {
		feed.Title = "RSS Feed"
	}
	if feed.Description == "" {
		feed.Description = "Aggregated updates"
	}
	if len(items) > 0 {
		feed.LastBuildDate = items[0].PublishedAt.UTC()
	}
	return feed, nil
}

type siteMetadata struct {
	title        string
	description  string
	publicURL    string
	imageURL     string
	followFeedID string
	followUserID string
}

func (s *Service) loadSiteMetadata(ctx context.Context) siteMetadata {
	meta := siteMetadata{}
	if s.sysCfg == nil {
		return meta
	}
	info, err := s.sysCfg.WebsiteInfo(ctx)
	if err != nil {
		return meta
	}
	for key, value := range info {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		switch key {
		case "website_name":
			meta.title = value
		case "og_site_name", "og_title":
			if meta.title == "" {
				meta.title = value
			}
		case "description":
			meta.description = value
		case "og_description":
			if meta.description == "" {
				meta.description = value
			}
		case "public_url":
			meta.publicURL = value
		case "rss_image_url":
			meta.imageURL = value
		case "favicon":
			if meta.imageURL == "" {
				meta.imageURL = value
			}
		case "og_image":
			if meta.imageURL == "" {
				meta.imageURL = value
			}
		case "rss_follow_feed_id":
			meta.followFeedID = value
		case "rss_follow_user_id":
			meta.followUserID = value
		}
	}
	return meta
}

type authorMetadata struct {
	name   string
	email  string
	avatar string
}

func (s *Service) loadDefaultAuthor(ctx context.Context) authorMetadata {
	if s.identityRepo == nil {
		return authorMetadata{}
	}
	admins, err := s.identityRepo.ListAdmins(ctx)
	if err != nil || len(admins) == 0 {
		return authorMetadata{}
	}
	selected := admins[0]
	for i := 1; i < len(admins); i++ {
		if admins[i].ID < selected.ID {
			selected = admins[i]
		}
	}
	name := strings.TrimSpace(selected.Nickname)
	if name == "" {
		name = strings.TrimSpace(selected.Username)
	}
	return authorMetadata{
		name:   name,
		email:  strings.TrimSpace(selected.Email),
		avatar: strings.TrimSpace(selected.Avatar),
	}
}

func (s *Service) resolveAuthorByUserID(ctx context.Context, userID int64, cache map[int64]authorMetadata) authorMetadata {
	if userID <= 0 || s.identityRepo == nil {
		return authorMetadata{}
	}
	if v, ok := cache[userID]; ok {
		return v
	}
	user, err := s.identityRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		cache[userID] = authorMetadata{}
		return authorMetadata{}
	}
	name := strings.TrimSpace(user.Nickname)
	if name == "" {
		name = strings.TrimSpace(user.Username)
	}
	result := authorMetadata{
		name:   name,
		email:  strings.TrimSpace(user.Email),
		avatar: strings.TrimSpace(user.Avatar),
	}
	cache[userID] = result
	return result
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		v := strings.TrimSpace(value)
		if v != "" {
			return v
		}
	}
	return ""
}

func resolveMediaURL(baseURL string, raw string) string {
	v := strings.TrimSpace(raw)
	if v == "" {
		return ""
	}
	lower := strings.ToLower(v)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return v
	}
	return buildURL(baseURL, v)
}

func normalizeBaseURL(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	return strings.TrimRight(trimmed, "/")
}

func buildURL(baseURL, path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if baseURL == "" {
		return path
	}
	return strings.TrimRight(baseURL, "/") + path
}

func buildThinkingTitle(content string, id int64) string {
	if content == "" {
		return fmt.Sprintf("思考 #%d", id)
	}
	runes := []rune(content)
	if len(runes) > 40 {
		return string(runes[:40]) + "..."
	}
	return content
}

func renderHTML(markdown string) string {
	markdown = strings.TrimSpace(markdown)
	if markdown == "" {
		return ""
	}
	var buf bytes.Buffer
	if err := markdownRenderer.Convert([]byte(markdown), &buf); err != nil {
		return markdown
	}
	return strings.TrimSpace(buf.String())
}

func buildRSSDescription(link string, contentHTML string) string {
	link = strings.TrimSpace(link)
	contentHTML = strings.TrimSpace(contentHTML)
	if link == "" {
		return contentHTML
	}
	intro := `<blockquote><p>该内容由 RSS 渲染生成，最佳阅读体验请前往：<a href="` +
		html.EscapeString(link) + `">` + html.EscapeString(link) + `</a></p></blockquote>`
	if contentHTML == "" {
		return intro
	}
	return intro + contentHTML
}
