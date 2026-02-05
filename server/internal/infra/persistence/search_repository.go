package persistence

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	domainsearch "github.com/grtsinry43/grtblog-v2/server/internal/domain/search"
	"gorm.io/gorm"
)

type SearchRepository struct {
	db     *gorm.DB
	driver string
}

type searchRow struct {
	Kind       string    `gorm:"column:kind"`
	ID         int64     `gorm:"column:id"`
	Title      string    `gorm:"column:title"`
	Summary    string    `gorm:"column:summary"`
	SourceText string    `gorm:"column:source_text"`
	ShortURL   *string   `gorm:"column:short_url"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	Score      float64   `gorm:"column:score"`
}

func NewSearchRepository(db *gorm.DB) *SearchRepository {
	driver := ""
	if db != nil && db.Dialector != nil {
		driver = strings.ToLower(db.Dialector.Name())
	}
	return &SearchRepository{
		db:     db,
		driver: driver,
	}
}

func (r *SearchRepository) SearchSite(ctx context.Context, keyword string, limitPerGroup int) ([]domainsearch.Hit, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return []domainsearch.Hit{}, nil
	}
	if r.db == nil {
		return nil, fmt.Errorf("search repository db is nil")
	}
	if r.driver != "postgres" {
		return nil, fmt.Errorf("site search requires postgres, current driver: %s", r.driver)
	}

	rows, err := r.searchSitePostgres(ctx, keyword, limitPerGroup)
	if err != nil {
		return nil, err
	}

	result := make([]domainsearch.Hit, 0, len(rows))
	for _, row := range rows {
		result = append(result, domainsearch.Hit{
			Kind:      domainsearch.Kind(row.Kind),
			ID:        row.ID,
			Title:     row.Title,
			Summary:   row.Summary,
			Snippet:   buildSnippet(row.SourceText, keyword),
			ShortURL:  row.ShortURL,
			CreatedAt: row.CreatedAt,
			Score:     row.Score,
		})
	}
	return result, nil
}

func (r *SearchRepository) searchSitePostgres(ctx context.Context, keyword string, limitPerGroup int) ([]searchRow, error) {
	sql := `
WITH q AS (
    SELECT plainto_tsquery('simple', ?) AS tsq, ?::text AS kw
),
base AS (
    SELECT
        'article'::text AS kind,
        a.id,
        a.title,
        a.summary,
        concat_ws(' ', coalesce(a.title, ''), coalesce(a.summary, ''), coalesce(a.content, '')) AS source_text,
        a.short_url,
        a.created_at,
        ts_rank(
            setweight(to_tsvector('simple', coalesce(a.title, '')), 'A') ||
            setweight(to_tsvector('simple', coalesce(a.summary, '')), 'B') ||
            setweight(to_tsvector('simple', coalesce(a.content, '')), 'C'),
            q.tsq
        ) AS score
    FROM article a, q
    WHERE a.deleted_at IS NULL
      AND a.is_published = TRUE
      AND (
        (
            setweight(to_tsvector('simple', coalesce(a.title, '')), 'A') ||
            setweight(to_tsvector('simple', coalesce(a.summary, '')), 'B') ||
            setweight(to_tsvector('simple', coalesce(a.content, '')), 'C')
        ) @@ q.tsq
        OR a.title ILIKE '%' || q.kw || '%'
        OR a.summary ILIKE '%' || q.kw || '%'
        OR a.content ILIKE '%' || q.kw || '%'
      )

    UNION ALL

    SELECT
        'moment'::text AS kind,
        m.id,
        m.title,
        m.summary,
        concat_ws(' ', coalesce(m.title, ''), coalesce(m.summary, ''), coalesce(m.content, '')) AS source_text,
        m.short_url,
        m.created_at,
        ts_rank(
            setweight(to_tsvector('simple', coalesce(m.title, '')), 'A') ||
            setweight(to_tsvector('simple', coalesce(m.summary, '')), 'B') ||
            setweight(to_tsvector('simple', coalesce(m.content, '')), 'C'),
            q.tsq
        ) AS score
    FROM moment m, q
    WHERE m.deleted_at IS NULL
      AND m.is_published = TRUE
      AND (
        (
            setweight(to_tsvector('simple', coalesce(m.title, '')), 'A') ||
            setweight(to_tsvector('simple', coalesce(m.summary, '')), 'B') ||
            setweight(to_tsvector('simple', coalesce(m.content, '')), 'C')
        ) @@ q.tsq
        OR m.title ILIKE '%' || q.kw || '%'
        OR m.summary ILIKE '%' || q.kw || '%'
        OR m.content ILIKE '%' || q.kw || '%'
      )

    UNION ALL

    SELECT
        'page'::text AS kind,
        p.id,
        p.title,
        coalesce(p.description, '') AS summary,
        concat_ws(' ', coalesce(p.title, ''), coalesce(p.description, ''), coalesce(p.content, '')) AS source_text,
        p.short_url,
        p.created_at,
        ts_rank(
            setweight(to_tsvector('simple', coalesce(p.title, '')), 'A') ||
            setweight(to_tsvector('simple', coalesce(p.description, '')), 'B') ||
            setweight(to_tsvector('simple', coalesce(p.content, '')), 'C'),
            q.tsq
        ) AS score
    FROM page p, q
    WHERE p.deleted_at IS NULL
      AND p.is_enabled = TRUE
      AND (
        (
            setweight(to_tsvector('simple', coalesce(p.title, '')), 'A') ||
            setweight(to_tsvector('simple', coalesce(p.description, '')), 'B') ||
            setweight(to_tsvector('simple', coalesce(p.content, '')), 'C')
        ) @@ q.tsq
        OR p.title ILIKE '%' || q.kw || '%'
        OR coalesce(p.description, '') ILIKE '%' || q.kw || '%'
        OR p.content ILIKE '%' || q.kw || '%'
      )

    UNION ALL

    SELECT
        'thinking'::text AS kind,
        t.id,
        left(t.content, 42) AS title,
        left(t.content, 180) AS summary,
        coalesce(t.content, '') AS source_text,
        NULL::text AS short_url,
        t.created_at,
        ts_rank(to_tsvector('simple', coalesce(t.content, '')), q.tsq) AS score
    FROM thinking t, q
    WHERE (
        to_tsvector('simple', coalesce(t.content, '')) @@ q.tsq
        OR t.content ILIKE '%' || q.kw || '%'
    )
),
ranked AS (
    SELECT
        kind,
        id,
        title,
        summary,
        source_text,
        short_url,
        created_at,
        score,
        row_number() OVER (PARTITION BY kind ORDER BY score DESC, created_at DESC) AS rn
    FROM base
)
SELECT kind, id, title, summary, source_text, short_url, created_at, score
FROM ranked
WHERE rn <= ?
ORDER BY kind ASC, score DESC, created_at DESC;
`
	rows := make([]searchRow, 0)
	err := r.db.WithContext(ctx).Raw(sql, keyword, keyword, limitPerGroup).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func buildSnippet(text, keyword string) string {
	cleanText := strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
	if cleanText == "" {
		return ""
	}

	match := strings.TrimSpace(keyword)
	if idx, matchLen, ok := locateMatch(cleanText, match); ok {
		return windowSnippet(cleanText, idx, matchLen, 24, 34)
	}
	for _, token := range strings.Fields(match) {
		if idx, matchLen, ok := locateMatch(cleanText, token); ok {
			return windowSnippet(cleanText, idx, matchLen, 24, 34)
		}
	}
	return windowSnippet(cleanText, 0, 0, 0, 58)
}

func locateMatch(text, needle string) (int, int, bool) {
	needle = strings.TrimSpace(needle)
	if needle == "" {
		return 0, 0, false
	}
	lowerText := strings.ToLower(text)
	lowerNeedle := strings.ToLower(needle)
	byteIdx := strings.Index(lowerText, lowerNeedle)
	if byteIdx < 0 {
		return 0, 0, false
	}
	startRune := utf8.RuneCountInString(text[:byteIdx])
	matchRuneLen := utf8.RuneCountInString(needle)
	return startRune, matchRuneLen, true
}

func windowSnippet(text string, start, matchLen, before, after int) string {
	runes := []rune(text)
	if len(runes) == 0 {
		return ""
	}
	if start < 0 {
		start = 0
	}
	if start > len(runes) {
		start = len(runes)
	}
	endMatch := start + matchLen
	if endMatch > len(runes) {
		endMatch = len(runes)
	}

	left := start - before
	if left < 0 {
		left = 0
	}
	right := endMatch + after
	if matchLen == 0 {
		right = after
	}
	if right > len(runes) {
		right = len(runes)
	}
	if right < left {
		right = left
	}
	snippet := string(runes[left:right])
	if left > 0 {
		snippet = "..." + snippet
	}
	if right < len(runes) {
		snippet = snippet + "..."
	}
	return snippet
}
