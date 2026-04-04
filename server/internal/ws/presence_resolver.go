package ws

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/sysconfig"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/content"
	domainthinking "github.com/grtsinry43/grtblog-v2/server/internal/domain/thinking"
)

const (
	presenceTypeArticle  = "article"
	presenceTypeMoment   = "moment"
	presenceTypePage     = "page"
	presenceTypeThinking = "thinking"
)

type PresenceTitleResolver struct {
	contentRepo  content.Repository
	thinkingRepo domainthinking.ThinkingRepository
	sysCfg       *sysconfig.Service
}

func NewPresenceTitleResolver(contentRepo content.Repository, thinkingRepo domainthinking.ThinkingRepository, sysCfg *sysconfig.Service) *PresenceTitleResolver {
	return &PresenceTitleResolver{
		contentRepo:  contentRepo,
		thinkingRepo: thinkingRepo,
		sysCfg:       sysCfg,
	}
}

func (r *PresenceTitleResolver) Resolve(contentType string, rawURL string) (PresenceResolvedView, bool) {
	normalizedType := normalizePresenceType(contentType)
	if normalizedType == "" {
		return PresenceResolvedView{}, false
	}

	normalizedPath := normalizePresencePath(rawURL)
	switch normalizedType {
	case presenceTypeArticle:
		return r.resolveArticle(normalizedPath), true
	case presenceTypeMoment:
		return r.resolveMoment(normalizedPath), true
	case presenceTypePage:
		return r.resolvePage(normalizedPath), true
	case presenceTypeThinking:
		return r.resolveThinking(normalizedPath), true
	default:
		return PresenceResolvedView{}, false
	}
}

func normalizePresenceType(contentType string) string {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case presenceTypeArticle:
		return presenceTypeArticle
	case presenceTypeMoment:
		return presenceTypeMoment
	case presenceTypePage:
		return presenceTypePage
	case presenceTypeThinking:
		return presenceTypeThinking
	default:
		return ""
	}
}

func normalizePresencePath(rawURL string) string {
	value := strings.TrimSpace(rawURL)
	if value == "" {
		return "/"
	}

	parsed, err := url.Parse(value)
	if err == nil && parsed.Path != "" {
		value = parsed.Path
	}

	if !strings.HasPrefix(value, "/") {
		value = "/" + value
	}

	normalized := path.Clean(value)
	if normalized == "." {
		return "/"
	}
	return normalized
}

func splitPathSegments(pathname string) []string {
	trimmed := strings.Trim(pathname, "/")
	if trimmed == "" {
		return nil
	}
	parts := strings.Split(trimmed, "/")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		decoded, err := url.PathUnescape(part)
		if err == nil {
			out = append(out, decoded)
			continue
		}
		out = append(out, part)
	}
	return out
}

func (r *PresenceTitleResolver) resolveArticle(pathname string) PresenceResolvedView {
	view := PresenceResolvedView{
		ContentType: presenceTypeArticle,
		Title:       "文章",
		URL:         pathname,
	}

	parts := splitPathSegments(pathname)
	if len(parts) < 2 || parts[0] != "posts" {
		return view
	}

	slug := parts[1]
	view.URL = "/posts/" + url.PathEscape(slug)
	if r.contentRepo == nil {
		return view
	}

	article, err := r.contentRepo.GetArticleByShortURL(context.Background(), slug)
	if err != nil || article == nil || !article.IsPublished {
		return view
	}

	view.Title = strings.TrimSpace(article.Title)
	if view.Title == "" {
		view.Title = "文章"
	}
	view.URL = "/posts/" + url.PathEscape(article.ShortURL)
	return view
}

func (r *PresenceTitleResolver) resolveMoment(pathname string) PresenceResolvedView {
	view := PresenceResolvedView{
		ContentType: presenceTypeMoment,
		Title:       "手记",
		URL:         pathname,
	}

	parts := splitPathSegments(pathname)
	if len(parts) < 2 || parts[0] != "moments" {
		return view
	}

	slug := parts[len(parts)-1]
	if slug == "" || slug == "moments" {
		return view
	}

	if r.contentRepo == nil {
		return view
	}

	item, err := r.contentRepo.GetMomentByShortURL(context.Background(), slug)
	if err != nil || item == nil || !item.IsPublished {
		return view
	}

	view.Title = strings.TrimSpace(item.Title)
	if view.Title == "" {
		view.Title = "手记"
	}
	siteTZ := r.sysCfg.Timezone(context.Background())
	view.URL = buildMomentPath(item.ShortURL, item.CreatedAt.In(siteTZ))
	return view
}

func (r *PresenceTitleResolver) resolvePage(pathname string) PresenceResolvedView {
	view := PresenceResolvedView{
		ContentType: presenceTypePage,
		Title:       "页面",
		URL:         pathname,
	}

	if pathname == "/" {
		view.Title = "首页"
		view.URL = "/"
		return view
	}

	parts := splitPathSegments(pathname)
	if len(parts) != 1 {
		return view
	}

	slug := parts[0]
	view.URL = "/" + url.PathEscape(slug)
	if r.contentRepo == nil {
		return view
	}

	item, err := r.contentRepo.GetPageByShortURL(context.Background(), slug)
	if err != nil || item == nil || !item.IsEnabled {
		return view
	}

	view.Title = strings.TrimSpace(item.Title)
	if view.Title == "" {
		view.Title = "页面"
	}
	view.URL = "/" + url.PathEscape(item.ShortURL)
	return view
}

func (r *PresenceTitleResolver) resolveThinking(pathname string) PresenceResolvedView {
	view := PresenceResolvedView{
		ContentType: presenceTypeThinking,
		Title:       "思考",
		URL:         "/thinkings",
	}

	parts := splitPathSegments(pathname)
	if len(parts) == 0 {
		return view
	}
	if parts[0] != "thinkings" {
		view.URL = pathname
		return view
	}

	if len(parts) < 2 || r.thinkingRepo == nil {
		return view
	}

	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || id <= 0 {
		return view
	}

	item, err := r.thinkingRepo.FindByID(context.Background(), id)
	if err != nil || item == nil {
		return view
	}

	view.Title = formatThinkingTitle(id, item.Content)
	view.URL = fmt.Sprintf("/thinkings/%d", id)
	return view
}

func buildMomentPath(slug string, createdAt interface{ Format(string) string }) string {
	date := createdAt.Format("2006/01/02")
	return "/moments/" + date + "/" + url.PathEscape(slug)
}

func formatThinkingTitle(id int64, content string) string {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return fmt.Sprintf("思考 #%d", id)
	}

	runes := []rune(trimmed)
	if len(runes) > 24 {
		return string(runes[:24]) + "..."
	}
	return string(runes)
}
