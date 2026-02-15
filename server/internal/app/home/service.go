package home

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence/model"
)

const (
	defaultRangeDays = 365
	maxRangeDays     = 730
	githubBaseURL    = "https://api.github.com"
)

var (
	markdownImagePattern = regexp.MustCompile(`!\[[^\]]*\]\(([^)\s]+)(?:\s+"[^"]*")?\)`)
	htmlImagePattern     = regexp.MustCompile(`(?i)<img[^>]+src=["']([^"']+)["']`)
)

type Service struct {
	db         *gorm.DB
	now        func() time.Time
	httpClient *http.Client
}

type ActivityPulsePoint struct {
	Date    string `json:"date"`
	Posts   int64  `json:"posts"`
	Moments int64  `json:"moments"`
}

type ActivityPulse struct {
	Days         int                  `json:"days"`
	StartDate    string               `json:"startDate"`
	EndDate      string               `json:"endDate"`
	TotalPosts   int64                `json:"totalPosts"`
	TotalMoments int64                `json:"totalMoments"`
	StatusLabel  string               `json:"statusLabel"`
	Points       []ActivityPulsePoint `json:"points"`
}

type WordCountStats struct {
	Total     int64 `json:"total"`
	Articles  int64 `json:"articles"`
	Moments   int64 `json:"moments"`
	Pages     int64 `json:"pages"`
	Thinkings int64 `json:"thinkings"`
}

type GitHubStats struct {
	Username          string `json:"username"`
	ProfileURL        string `json:"profileUrl"`
	AvatarURL         string `json:"avatarUrl"`
	Followers         int64  `json:"followers"`
	PublicRepos       int64  `json:"publicRepos"`
	RecentPushCommits int64  `json:"recentPushCommits"`
	FetchedAt         string `json:"fetchedAt"`
}

type InspirationStats struct {
	Words       WordCountStats `json:"words"`
	GitHub      *GitHubStats   `json:"github,omitempty"`
	GitHubError string         `json:"githubError,omitempty"`
}

type TimelinePostItem struct {
	Title       string `json:"title"`
	ShortURL    string `json:"shortUrl"`
	URL         string `json:"url"`
	Cover       string `json:"cover,omitempty"`
	PublishedAt string `json:"publishedAt"`
}

type TimelineMomentItem struct {
	Title       string `json:"title"`
	ShortURL    string `json:"shortUrl"`
	URL         string `json:"url"`
	Image       string `json:"image,omitempty"`
	PublishedAt string `json:"publishedAt"`
}

type TimelineThinkingItem struct {
	Content     string `json:"content"`
	ShortURL    string `json:"shortUrl"`
	URL         string `json:"url"`
	PublishedAt string `json:"publishedAt"`
}

type TimelineYearBucket struct {
	YearSummary *TimelinePostItem      `json:"yearSummary,omitempty"`
	Posts       []TimelinePostItem     `json:"posts"`
	Moments     []TimelineMomentItem   `json:"moments"`
	Thinkings   []TimelineThinkingItem `json:"thinkings"`
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		db:  db,
		now: time.Now,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (s *Service) GetActivityPulse(ctx context.Context, days int) (*ActivityPulse, error) {
	today := s.now().UTC().Truncate(24 * time.Hour)
	endExclusive := today.Add(24 * time.Hour)
	start := today

	switch {
	case days == -1:
		earliest, found, err := s.findEarliestPublishedCreatedDate(ctx)
		if err != nil {
			return nil, err
		}
		if found {
			start = earliest
		}
		days = int(today.Sub(start).Hours()/24) + 1
		if days <= 0 {
			days = 1
			start = today
		}
	default:
		if days <= 0 {
			days = defaultRangeDays
		}
		if days > maxRangeDays {
			days = maxRangeDays
		}
		start = today.AddDate(0, 0, -(days - 1))
	}

	points := make([]ActivityPulsePoint, 0, days)
	byDate := make(map[string]int, days)
	for i := 0; i < days; i++ {
		date := start.AddDate(0, 0, i).Format("2006-01-02")
		byDate[date] = len(points)
		points = append(points, ActivityPulsePoint{
			Date: date,
		})
	}

	type countRow struct {
		Date  string `gorm:"column:date"`
		Count int64  `gorm:"column:count"`
	}

	var articleRows []countRow
	if err := s.db.WithContext(ctx).
		Model(&model.Article{}).
		Select("DATE(created_at) AS date, COUNT(*) AS count").
		Where("is_published = ?", true).
		Where("created_at >= ? AND created_at < ?", start, endExclusive).
		Group("DATE(created_at)").
		Scan(&articleRows).Error; err != nil {
		return nil, err
	}

	var momentRows []countRow
	if err := s.db.WithContext(ctx).
		Model(&model.Moment{}).
		Select("DATE(created_at) AS date, COUNT(*) AS count").
		Where("is_published = ?", true).
		Where("created_at >= ? AND created_at < ?", start, endExclusive).
		Group("DATE(created_at)").
		Scan(&momentRows).Error; err != nil {
		return nil, err
	}

	var totalPosts int64
	for _, row := range articleRows {
		idx, ok := byDate[normalizeDateKey(row.Date)]
		if !ok {
			continue
		}
		points[idx].Posts = row.Count
		totalPosts += row.Count
	}

	var totalMoments int64
	for _, row := range momentRows {
		idx, ok := byDate[normalizeDateKey(row.Date)]
		if !ok {
			continue
		}
		points[idx].Moments = row.Count
		totalMoments += row.Count
	}

	return &ActivityPulse{
		Days:         days,
		StartDate:    start.Format("2006-01-02"),
		EndDate:      today.Format("2006-01-02"),
		TotalPosts:   totalPosts,
		TotalMoments: totalMoments,
		StatusLabel:  buildStatusLabel(totalPosts, totalMoments, days),
		Points:       points,
	}, nil
}

func (s *Service) findEarliestPublishedCreatedDate(ctx context.Context) (time.Time, bool, error) {
	type minRow struct {
		MinAt sql.NullTime `gorm:"column:min_at"`
	}

	var articleRow minRow
	if err := s.db.WithContext(ctx).
		Model(&model.Article{}).
		Select("MIN(created_at) AS min_at").
		Where("is_published = ?", true).
		Scan(&articleRow).Error; err != nil {
		return time.Time{}, false, err
	}

	var momentRow minRow
	if err := s.db.WithContext(ctx).
		Model(&model.Moment{}).
		Select("MIN(created_at) AS min_at").
		Where("is_published = ?", true).
		Scan(&momentRow).Error; err != nil {
		return time.Time{}, false, err
	}

	var earliest time.Time
	found := false
	if articleRow.MinAt.Valid {
		earliest = articleRow.MinAt.Time.UTC().Truncate(24 * time.Hour)
		found = true
	}
	if momentRow.MinAt.Valid {
		candidate := momentRow.MinAt.Time.UTC().Truncate(24 * time.Hour)
		if !found || candidate.Before(earliest) {
			earliest = candidate
			found = true
		}
	}
	return earliest, found, nil
}

func buildStatusLabel(totalPosts, totalMoments int64, days int) string {
	if days <= 0 {
		return "Quiet"
	}
	avg := float64(totalPosts+totalMoments) / float64(days)
	switch {
	case avg >= 2:
		return "Prolific"
	case avg >= 0.9:
		return "Steady"
	case avg >= 0.2:
		return "Active"
	default:
		return "Quiet"
	}
}

func (s *Service) GetInspirationStats(ctx context.Context, githubUsername string) (*InspirationStats, error) {
	words, err := s.queryWordStats(ctx)
	if err != nil {
		return nil, err
	}

	stats := &InspirationStats{Words: words}
	username := strings.TrimSpace(githubUsername)
	if username == "" {
		return stats, nil
	}

	githubStats, err := s.fetchGitHubStats(ctx, username)
	if err != nil {
		// GitHub 是增强信息，不应该阻断首页渲染。
		stats.GitHubError = err.Error()
		log.Printf("[home] github stats fetch failed username=%s err=%v", username, err)
		return stats, nil
	}
	stats.GitHub = githubStats
	return stats, nil
}

func (s *Service) GetTimelineByYear(ctx context.Context) (map[string]TimelineYearBucket, error) {
	timeline := make(map[string]TimelineYearBucket)

	type articleRow struct {
		Title     string    `gorm:"column:title"`
		ShortURL  string    `gorm:"column:short_url"`
		Cover     *string   `gorm:"column:cover"`
		ExtInfo   []byte    `gorm:"column:ext_info"`
		CreatedAt time.Time `gorm:"column:created_at"`
	}
	var articles []articleRow
	if err := s.db.WithContext(ctx).
		Model(&model.Article{}).
		Select("title, short_url, cover, ext_info, created_at").
		Where("is_published = ?", true).
		Order("created_at DESC").
		Scan(&articles).Error; err != nil {
		return nil, err
	}

	type momentRow struct {
		Title     string    `gorm:"column:title"`
		ShortURL  string    `gorm:"column:short_url"`
		Image     *string   `gorm:"column:img"`
		Content   string    `gorm:"column:content"`
		CreatedAt time.Time `gorm:"column:created_at"`
	}
	var moments []momentRow
	if err := s.db.WithContext(ctx).
		Model(&model.Moment{}).
		Select("title, short_url, img, content, created_at").
		Where("is_published = ?", true).
		Order("created_at DESC").
		Scan(&moments).Error; err != nil {
		return nil, err
	}

	type thinkingRow struct {
		ID        int64     `gorm:"column:id"`
		Content   string    `gorm:"column:content"`
		CreatedAt time.Time `gorm:"column:created_at"`
	}
	var thinkings []thinkingRow
	if err := s.db.WithContext(ctx).
		Model(&model.Thinking{}).
		Select("id, content, created_at").
		Order("created_at DESC").
		Scan(&thinkings).Error; err != nil {
		return nil, err
	}

	for _, item := range articles {
		yearKey, publishedAt := toYearKeyAndPublishedAt(item.CreatedAt)
		bucket := ensureTimelineBucket(timeline, yearKey)

		post := TimelinePostItem{
			Title:       strings.TrimSpace(item.Title),
			ShortURL:    strings.TrimSpace(item.ShortURL),
			URL:         buildPostURL(item.ShortURL),
			Cover:       optionalString(item.Cover),
			PublishedAt: publishedAt,
		}

		yearSummaryMark, marked := parseYearSummaryYear(item.ExtInfo)
		if marked && yearSummaryMark == item.CreatedAt.UTC().Year() && bucket.YearSummary == nil {
			bucket.YearSummary = &post
		} else {
			bucket.Posts = append(bucket.Posts, post)
		}
		timeline[yearKey] = bucket
	}

	for _, item := range moments {
		yearKey, publishedAt := toYearKeyAndPublishedAt(item.CreatedAt)
		bucket := ensureTimelineBucket(timeline, yearKey)
		image := optionalString(item.Image)
		if image == "" {
			image = extractFirstImageURL(item.Content)
		}
		bucket.Moments = append(bucket.Moments, TimelineMomentItem{
			Title:       strings.TrimSpace(item.Title),
			ShortURL:    strings.TrimSpace(item.ShortURL),
			URL:         buildMomentURL(item.ShortURL, item.CreatedAt),
			Image:       image,
			PublishedAt: publishedAt,
		})
		timeline[yearKey] = bucket
	}

	for _, item := range thinkings {
		yearKey, publishedAt := toYearKeyAndPublishedAt(item.CreatedAt)
		bucket := ensureTimelineBucket(timeline, yearKey)
		shortURL := fmt.Sprintf("#%d", item.ID)
		bucket.Thinkings = append(bucket.Thinkings, TimelineThinkingItem{
			Content:     strings.TrimSpace(item.Content),
			ShortURL:    shortURL,
			URL:         fmt.Sprintf("/thinkings#%d", item.ID),
			PublishedAt: publishedAt,
		})
		timeline[yearKey] = bucket
	}

	return timeline, nil
}

func (s *Service) queryWordStats(ctx context.Context) (WordCountStats, error) {
	var out WordCountStats
	var err error
	if out.Articles, err = s.sumContentLength(ctx, "article", "content"); err != nil {
		return out, err
	}
	if out.Moments, err = s.sumContentLength(ctx, "moment", "content"); err != nil {
		return out, err
	}
	if out.Pages, err = s.sumContentLength(ctx, "page", "content"); err != nil {
		return out, err
	}
	if out.Thinkings, err = s.sumContentLength(ctx, "thinking", "content"); err != nil {
		return out, err
	}
	out.Total = out.Articles + out.Moments + out.Pages + out.Thinkings
	return out, nil
}

func (s *Service) sumContentLength(ctx context.Context, tableName, columnName string) (int64, error) {
	type row struct {
		Val int64 `gorm:"column:val"`
	}
	var out row
	query := fmt.Sprintf("COALESCE(SUM(CHAR_LENGTH(%s)), 0) AS val", columnName)
	err := s.db.WithContext(ctx).Table(tableName).Select(query).Scan(&out).Error
	return out.Val, err
}

func (s *Service) fetchGitHubStats(ctx context.Context, username string) (*GitHubStats, error) {
	type githubUser struct {
		Login       string `json:"login"`
		HTMLURL     string `json:"html_url"`
		AvatarURL   string `json:"avatar_url"`
		Followers   int64  `json:"followers"`
		PublicRepos int64  `json:"public_repos"`
	}
	type githubCommitSearch struct {
		TotalCount int64 `json:"total_count"`
	}

	escaped := url.PathEscape(username)
	var user githubUser
	if err := s.fetchGitHubJSON(ctx, githubBaseURL+"/users/"+escaped, &user); err != nil {
		return nil, err
	}

	now := s.now().UTC()
	startDate := now.AddDate(-1, 0, 0).Format("2006-01-02")
	endDate := now.Format("2006-01-02")
	searchURL := fmt.Sprintf(
		"%s/search/commits?q=author:%s+author-date:%s..%s&per_page=1",
		githubBaseURL,
		url.QueryEscape(username),
		startDate,
		endDate,
	)
	var commitSearch githubCommitSearch
	if err := s.fetchGitHubJSON(ctx, searchURL, &commitSearch); err != nil {
		return nil, err
	}

	resolvedUsername := strings.TrimSpace(user.Login)
	if resolvedUsername == "" {
		resolvedUsername = username
	}

	return &GitHubStats{
		Username:          resolvedUsername,
		ProfileURL:        strings.TrimSpace(user.HTMLURL),
		AvatarURL:         strings.TrimSpace(user.AvatarURL),
		Followers:         user.Followers,
		PublicRepos:       user.PublicRepos,
		RecentPushCommits: commitSearch.TotalCount,
		FetchedAt:         s.now().UTC().Format(time.RFC3339),
	}, nil
}

func (s *Service) fetchGitHubJSON(ctx context.Context, endpoint string, target any) error {
	client := s.httpClient
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "grtblog-v2-home")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if resp.StatusCode == http.StatusNotFound {
			return errors.New("github 用户不存在")
		}
		if resp.StatusCode == http.StatusForbidden {
			return errors.New("github API 速率限制，请稍后再试")
		}
		return fmt.Errorf("github API error: status=%d", resp.StatusCode)
	}

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(target); err != nil {
		return err
	}
	return nil
}

func normalizeDateKey(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if len(trimmed) >= 10 && trimmed[4] == '-' && trimmed[7] == '-' {
		return trimmed[:10]
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05 -0700 MST",
		"2006-01-02 15:04:05 -0700",
		"2006-01-02 15:04:05",
	}
	for _, layout := range layouts {
		t, err := time.Parse(layout, trimmed)
		if err == nil {
			return t.Format("2006-01-02")
		}
	}
	return trimmed
}

func ensureTimelineBucket(m map[string]TimelineYearBucket, yearKey string) TimelineYearBucket {
	bucket, ok := m[yearKey]
	if !ok {
		return TimelineYearBucket{
			Posts:     make([]TimelinePostItem, 0),
			Moments:   make([]TimelineMomentItem, 0),
			Thinkings: make([]TimelineThinkingItem, 0),
		}
	}
	if bucket.Posts == nil {
		bucket.Posts = make([]TimelinePostItem, 0)
	}
	if bucket.Moments == nil {
		bucket.Moments = make([]TimelineMomentItem, 0)
	}
	if bucket.Thinkings == nil {
		bucket.Thinkings = make([]TimelineThinkingItem, 0)
	}
	return bucket
}

func toYearKeyAndPublishedAt(t time.Time) (string, string) {
	utc := t.UTC()
	return strconv.Itoa(utc.Year()), utc.Format(time.RFC3339)
}

func optionalString(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func buildPostURL(shortURL string) string {
	return "/posts/" + url.PathEscape(strings.TrimSpace(shortURL))
}

func buildMomentURL(shortURL string, createdAt time.Time) string {
	utc := createdAt.UTC()
	return fmt.Sprintf(
		"/moments/%04d/%02d/%02d/%s",
		utc.Year(),
		utc.Month(),
		utc.Day(),
		url.PathEscape(strings.TrimSpace(shortURL)),
	)
}

func parseYearSummaryYear(extInfo []byte) (int, bool) {
	if len(extInfo) == 0 {
		return 0, false
	}
	obj := make(map[string]any)
	if err := json.Unmarshal(extInfo, &obj); err != nil {
		return 0, false
	}
	raw, ok := obj["is_year_summary"]
	if !ok {
		return 0, false
	}
	switch v := raw.(type) {
	case float64:
		year := int(v)
		if year > 0 {
			return year, true
		}
	case string:
		year, err := strconv.Atoi(strings.TrimSpace(v))
		if err == nil && year > 0 {
			return year, true
		}
	}
	return 0, false
}

func extractFirstImageURL(content string) string {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return ""
	}
	if matched := markdownImagePattern.FindStringSubmatch(trimmed); len(matched) >= 2 {
		return sanitizeImageURL(matched[1])
	}
	if matched := htmlImagePattern.FindStringSubmatch(trimmed); len(matched) >= 2 {
		return sanitizeImageURL(matched[1])
	}
	return ""
}

func sanitizeImageURL(raw string) string {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.TrimPrefix(trimmed, "<")
	trimmed = strings.TrimSuffix(trimmed, ">")
	return trimmed
}
