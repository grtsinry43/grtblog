package federation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	domainfed "github.com/grtsinry43/grtblog-v2/server/internal/domain/federation"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/social"
	fedinfra "github.com/grtsinry43/grtblog-v2/server/internal/infra/federation"
	"github.com/mmcdole/gofeed"
)

const (
	cacheSourceMethodTimeline    = "timeline"
	cacheSourceMethodRSS         = "rss"
	cacheSourceMethodRSSFallback = "rss_fallback"
)

type SyncWorker struct {
	instanceRepo domainfed.FederationInstanceRepository
	cacheRepo    domainfed.FederatedPostCacheRepository
	linkRepo     social.FriendLinkRepository
	syncJobRepo  social.FriendLinkSyncJobRepository
	resolver     *fedinfra.Resolver
	client       *http.Client
}

func NewSyncWorker(
	instanceRepo domainfed.FederationInstanceRepository,
	cacheRepo domainfed.FederatedPostCacheRepository,
	linkRepo social.FriendLinkRepository,
	syncJobRepo social.FriendLinkSyncJobRepository,
	resolver *fedinfra.Resolver,
) *SyncWorker {
	return &SyncWorker{
		instanceRepo: instanceRepo,
		cacheRepo:    cacheRepo,
		linkRepo:     linkRepo,
		syncJobRepo:  syncJobRepo,
		resolver:     resolver,
		client:       &http.Client{Timeout: 10 * time.Second},
	}
}

func (w *SyncWorker) Run(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 30 * time.Minute
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	w.SyncOnce(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.SyncOnce(ctx)
		}
	}
}

func (w *SyncWorker) SyncOnce(ctx context.Context) {
	if w == nil || w.cacheRepo == nil || w.linkRepo == nil {
		return
	}
	if w.syncJobRepo == nil {
		w.syncOnceDirect(ctx, time.Now().UTC())
		return
	}
	now := time.Now().UTC()
	_ = w.enqueueFriendLinkJobs(ctx, now)
	_ = w.processSyncJobs(ctx, now, 200)
	// Clean up old posts: keep only the 10 most recent posts per friend link
	_ = w.cacheRepo.CleanupOldPosts(ctx, 10)
}

func (w *SyncWorker) syncOnceDirect(ctx context.Context, now time.Time) {
	links, _, err := w.linkRepo.List(ctx, social.FriendLinkListOptions{
		IsActive: ptrBool(true),
		Page:     1,
		PageSize: 0,
	})
	if err != nil {
		return
	}
	for i := range links {
		if strings.EqualFold(strings.TrimSpace(links[i].Type), social.FriendLinkTypeNoRSS) {
			continue
		}
		if !shouldSyncFriendLink(links[i], now, 30*time.Minute) {
			continue
		}
		count, _, runErr := w.syncFriendLink(ctx, &links[i])
		_ = w.applyLinkSyncResult(ctx, &links[i], count, runErr)
	}
	// Clean up old posts: keep only the 10 most recent posts per friend link
	_ = w.cacheRepo.CleanupOldPosts(ctx, 10)
}

func (w *SyncWorker) syncFriendLink(ctx context.Context, link *social.FriendLink) (int, string, error) {
	if link == nil {
		return 0, social.FriendLinkSyncJobMethodRSS, nil
	}
	switch strings.ToLower(strings.TrimSpace(link.Type)) {
	case social.FriendLinkTypeFederation:
		return w.syncFederationFriendLink(ctx, link)
	case social.FriendLinkTypeRSS:
		rssURL := strings.TrimSpace(optionalStr(link.RSSURL))
		if rssURL == "" {
			return 0, social.FriendLinkSyncJobMethodRSS, fmt.Errorf("rss url is empty")
		}
		count, err := w.syncFromFeedURL(ctx, link.ID, nil, rssURL, cacheSourceMethodRSS)
		return count, social.FriendLinkSyncJobMethodRSS, err
	case social.FriendLinkTypeNoRSS:
		return 0, social.FriendLinkSyncJobMethodRSS, nil
	default:
		return 0, social.FriendLinkSyncJobMethodRSS, fmt.Errorf("unsupported friend link type: %s", strings.TrimSpace(link.Type))
	}
}

func (w *SyncWorker) syncFederationFriendLink(ctx context.Context, link *social.FriendLink) (int, string, error) {
	if link == nil {
		return 0, social.FriendLinkSyncJobMethodTimeline, nil
	}
	if link.InstanceID == nil || *link.InstanceID <= 0 {
		return 0, social.FriendLinkSyncJobMethodTimeline, fmt.Errorf("federation friend link missing instance_id")
	}
	if w.instanceRepo == nil {
		return 0, social.FriendLinkSyncJobMethodTimeline, fmt.Errorf("instance repository not configured")
	}
	if w.resolver == nil {
		return 0, social.FriendLinkSyncJobMethodTimeline, fmt.Errorf("resolver not configured")
	}
	instance, err := w.instanceRepo.GetByID(ctx, *link.InstanceID)
	if err != nil {
		return 0, social.FriendLinkSyncJobMethodTimeline, err
	}
	baseURL := strings.TrimRight(strings.TrimSpace(instance.BaseURL), "/")
	if baseURL == "" {
		return 0, social.FriendLinkSyncJobMethodTimeline, fmt.Errorf("instance base url is empty")
	}

	posts, timelineErr := w.fetchTimelinePosts(ctx, link.ID, instance.ID, baseURL)
	if timelineErr == nil && len(posts) > 0 {
		if err := w.cacheRepo.UpsertBatch(ctx, posts); err != nil {
			return 0, social.FriendLinkSyncJobMethodTimeline, err
		}
		return len(posts), social.FriendLinkSyncJobMethodTimeline, nil
	}

	rssURL := strings.TrimSpace(optionalStr(link.RSSURL))
	if rssURL == "" {
		if timelineErr != nil {
			return 0, social.FriendLinkSyncJobMethodTimeline, timelineErr
		}
		return 0, social.FriendLinkSyncJobMethodTimeline, nil
	}

	count, rssErr := w.syncFromFeedURL(ctx, link.ID, &instance.ID, rssURL, cacheSourceMethodRSSFallback)
	if rssErr != nil {
		if timelineErr != nil {
			return 0, social.FriendLinkSyncJobMethodRSS, fmt.Errorf("timeline sync failed: %w; rss fallback failed: %v", timelineErr, rssErr)
		}
		return 0, social.FriendLinkSyncJobMethodRSS, rssErr
	}
	return count, social.FriendLinkSyncJobMethodRSS, nil
}

func (w *SyncWorker) fetchTimelinePosts(ctx context.Context, friendLinkID, instanceID int64, baseURL string) ([]domainfed.FederatedPostCache, error) {
	endpoints, err := w.resolver.FetchEndpoints(ctx, baseURL)
	if err != nil {
		return nil, err
	}
	if endpoints == nil {
		return nil, fmt.Errorf("endpoints is nil")
	}
	if endpoints.Endpoints == nil {
		return nil, fmt.Errorf("endpoints map is nil")
	}
	path := strings.TrimSpace(endpoints.Endpoints["timeline"])
	if path == "" {
		return nil, fmt.Errorf("endpoints.timeline is empty")
	}
	endpointBaseURL := strings.TrimSpace(endpoints.BaseURL)
	if endpointBaseURL == "" {
		return nil, fmt.Errorf("endpoints.base_url is empty")
	}
	u, err := joinURL(endpointBaseURL, path)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("page", "1")
	q.Set("per_page", "50")
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := w.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("timeline status %d", resp.StatusCode)
	}
	var envelope struct {
		Data struct {
			Items []struct {
				ID             string     `json:"id"`
				URL            string     `json:"url"`
				Title          string     `json:"title"`
				Summary        string     `json:"summary"`
				ContentPreview *string    `json:"content_preview"`
				Author         any        `json:"author"`
				PublishedAt    time.Time  `json:"published_at"`
				UpdatedAt      *time.Time `json:"updated_at"`
				CoverImage     *string    `json:"cover_image"`
				Language       *string    `json:"language"`
				AllowCitation  bool       `json:"allow_citation"`
				AllowComment   bool       `json:"allow_comment"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, err
	}
	posts := make([]domainfed.FederatedPostCache, 0, len(envelope.Data.Items))
	for _, item := range envelope.Data.Items {
		if strings.TrimSpace(item.URL) == "" || strings.TrimSpace(item.Title) == "" || strings.TrimSpace(item.Summary) == "" {
			continue
		}
		authorRaw, _ := json.Marshal(item.Author)
		id := strings.TrimSpace(item.ID)
		if id == "" {
			id = item.URL
		}
		instanceIDCopy := instanceID
		posts = append(posts, domainfed.FederatedPostCache{
			FriendLinkID:   friendLinkID,
			InstanceID:     &instanceIDCopy,
			RemotePostID:   &id,
			URL:            item.URL,
			Title:          item.Title,
			Summary:        item.Summary,
			ContentPreview: item.ContentPreview,
			Author:         authorRaw,
			Tags:           json.RawMessage("[]"),
			Categories:     json.RawMessage("[]"),
			PublishedAt:    item.PublishedAt,
			UpdatedAt:      item.UpdatedAt,
			CoverImage:     item.CoverImage,
			Language:       item.Language,
			AllowCitation:  item.AllowCitation,
			AllowComment:   item.AllowComment,
			SourceMethod:   cacheSourceMethodTimeline,
			CachedAt:       time.Now().UTC(),
		})
	}
	return posts, nil
}

func (w *SyncWorker) syncFromFeedURL(ctx context.Context, friendLinkID int64, instanceID *int64, feedURL string, sourceMethod string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return 0, err
	}
	resp, err := w.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("rss status %d", resp.StatusCode)
	}
	parser := gofeed.NewParser()
	feed, err := parser.Parse(resp.Body)
	if err != nil {
		return 0, err
	}
	posts := parseFeedItems(feed, friendLinkID, instanceID, sourceMethod)
	if len(posts) == 0 {
		return 0, nil
	}
	if err := w.cacheRepo.UpsertBatch(ctx, posts); err != nil {
		return 0, err
	}
	return len(posts), nil
}

func (w *SyncWorker) enqueueFriendLinkJobs(ctx context.Context, now time.Time) error {
	if w.syncJobRepo == nil || w.linkRepo == nil {
		return nil
	}
	links, _, err := w.linkRepo.List(ctx, social.FriendLinkListOptions{
		IsActive: ptrBool(true),
		Page:     1,
		PageSize: 0,
	})
	if err != nil {
		return err
	}
	for i := range links {
		if !shouldSyncFriendLink(links[i], now, 30*time.Minute) {
			continue
		}
		switch strings.ToLower(strings.TrimSpace(links[i].Type)) {
		case social.FriendLinkTypeNoRSS:
			continue
		case social.FriendLinkTypeFederation, social.FriendLinkTypeRSS:
		default:
			continue
		}
		friendLinkID := links[i].ID
		method := social.FriendLinkSyncJobMethodRSS
		if links[i].Type == social.FriendLinkTypeFederation {
			method = social.FriendLinkSyncJobMethodTimeline
		}
		job := &social.FriendLinkSyncJob{
			TargetType:    social.FriendLinkSyncJobTargetFriendLink,
			SyncMethod:    method,
			FriendLinkID:  &friendLinkID,
			InstanceID:    links[i].InstanceID,
			TargetURL:     strings.TrimSpace(links[i].URL),
			FeedURL:       links[i].RSSURL,
			Status:        social.FriendLinkSyncJobStatusQueued,
			MaxAttempts:   1,
			TriggerSource: "scheduler",
		}
		_ = w.syncJobRepo.Create(ctx, job)
	}
	return nil
}

func (w *SyncWorker) processSyncJobs(ctx context.Context, now time.Time, limit int) error {
	if w.syncJobRepo == nil {
		return nil
	}
	jobs, err := w.syncJobRepo.ListProcessable(ctx, now, limit)
	if err != nil {
		return err
	}
	for i := range jobs {
		_ = w.runSyncJob(ctx, &jobs[i])
	}
	return nil
}

func (w *SyncWorker) runSyncJob(ctx context.Context, job *social.FriendLinkSyncJob) error {
	if job == nil || w.syncJobRepo == nil {
		return nil
	}
	startedAt := time.Now().UTC()
	job.Status = social.FriendLinkSyncJobStatusRunning
	job.AttemptCount++
	job.StartedAt = &startedAt
	job.FinishedAt = nil
	job.DurationMS = nil
	job.ErrorMessage = nil
	_ = w.syncJobRepo.Update(ctx, job)

	pulledCount, method, runErr := w.executeSyncJob(ctx, job)
	finishedAt := time.Now().UTC()
	durationMS := finishedAt.Sub(startedAt).Milliseconds()

	job.SyncMethod = method
	job.PulledCount = pulledCount
	job.FinishedAt = &finishedAt
	job.DurationMS = &durationMS
	if runErr != nil {
		job.Status = social.FriendLinkSyncJobStatusFailed
		job.ErrorMessage = toSyncJobErrorMessage(runErr)
	} else {
		job.Status = social.FriendLinkSyncJobStatusSuccess
		job.ErrorMessage = nil
	}
	return w.syncJobRepo.Update(ctx, job)
}

func (w *SyncWorker) executeSyncJob(ctx context.Context, job *social.FriendLinkSyncJob) (int, string, error) {
	if job == nil {
		return 0, social.FriendLinkSyncJobMethodRSS, nil
	}
	if job.TargetType != social.FriendLinkSyncJobTargetFriendLink {
		return 0, social.FriendLinkSyncJobMethodRSS, fmt.Errorf("unsupported sync target type: %s", strings.TrimSpace(job.TargetType))
	}

	var (
		link *social.FriendLink
		err  error
	)
	if job.FriendLinkID != nil && *job.FriendLinkID > 0 {
		link, err = w.linkRepo.GetByID(ctx, *job.FriendLinkID)
	} else {
		link, err = w.linkRepo.FindByURL(ctx, strings.TrimSpace(job.TargetURL))
	}
	if err != nil {
		return 0, social.FriendLinkSyncJobMethodRSS, err
	}
	if !link.IsActive {
		return 0, social.FriendLinkSyncJobMethodRSS, fmt.Errorf("friend link is inactive")
	}

	pulledCount, method, runErr := w.syncFriendLink(ctx, link)
	_ = w.applyLinkSyncResult(ctx, link, pulledCount, runErr)
	return pulledCount, method, runErr
}

func (w *SyncWorker) applyLinkSyncResult(ctx context.Context, link *social.FriendLink, pulledCount int, runErr error) error {
	if link == nil || w.linkRepo == nil {
		return nil
	}
	now := time.Now().UTC()
	link.LastSyncAt = &now
	if runErr != nil {
		failed := "failed"
		link.LastSyncStatus = &failed
		return w.linkRepo.Update(ctx, link)
	}
	ok := "ok"
	link.LastSyncStatus = &ok
	if count, err := w.cacheRepo.CountByFriendLink(ctx, link.ID); err == nil {
		link.TotalPostsCached = count
	} else {
		link.TotalPostsCached = pulledCount
	}
	return w.linkRepo.Update(ctx, link)
}

func parseFeedItems(feed *gofeed.Feed, friendLinkID int64, instanceID *int64, sourceMethod string) []domainfed.FederatedPostCache {
	if feed == nil || len(feed.Items) == 0 {
		return nil
	}
	posts := make([]domainfed.FederatedPostCache, 0, len(feed.Items))
	for _, item := range feed.Items {
		if item == nil {
			continue
		}
		link := strings.TrimSpace(item.Link)
		title := strings.TrimSpace(item.Title)
		if link == "" || title == "" {
			continue
		}
		summary := strings.TrimSpace(item.Description)
		if summary == "" {
			summary = strings.TrimSpace(item.Content)
		}
		if summary == "" {
			summary = title
		}
		publishedAt := time.Now().UTC()
		switch {
		case item.PublishedParsed != nil:
			publishedAt = item.PublishedParsed.UTC()
		case item.UpdatedParsed != nil:
			publishedAt = item.UpdatedParsed.UTC()
		}
		authorPayload := map[string]any{}
		if item.Author != nil && strings.TrimSpace(item.Author.Name) != "" {
			authorPayload["name"] = strings.TrimSpace(item.Author.Name)
		}
		authorRaw, _ := json.Marshal(authorPayload)
		id := strings.TrimSpace(item.GUID)
		if id == "" {
			id = link
		}
		var updatedAt *time.Time
		if item.UpdatedParsed != nil {
			u := item.UpdatedParsed.UTC()
			updatedAt = &u
		}
		posts = append(posts, domainfed.FederatedPostCache{
			FriendLinkID:   friendLinkID,
			InstanceID:     instanceID,
			RemotePostID:   &id,
			URL:            link,
			Title:          title,
			Summary:        summary,
			ContentPreview: nil,
			Author:         json.RawMessage(authorRaw),
			Tags:           json.RawMessage("[]"),
			Categories:     json.RawMessage("[]"),
			PublishedAt:    publishedAt,
			UpdatedAt:      updatedAt,
			AllowCitation:  true,
			AllowComment:   true,
			SourceMethod:   sourceMethod,
			CachedAt:       time.Now().UTC(),
		})
	}
	return posts
}

func shouldSyncFriendLink(link social.FriendLink, now time.Time, fallbackInterval time.Duration) bool {
	interval := fallbackInterval
	if link.SyncInterval != nil && *link.SyncInterval > 0 {
		interval = time.Duration(*link.SyncInterval) * time.Minute
	}
	if interval <= 0 {
		interval = 30 * time.Minute
	}
	if link.LastSyncAt == nil {
		return true
	}
	next := link.LastSyncAt.Add(interval)
	return !next.After(now)
}

func toSyncJobErrorMessage(err error) *string {
	if err == nil {
		return nil
	}
	msg := strings.TrimSpace(err.Error())
	if msg == "" {
		return nil
	}
	const maxLen = 2000
	if len(msg) > maxLen {
		msg = msg[:maxLen]
	}
	return &msg
}

func optionalStr(v *string) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(*v)
}

func ptrBool(v bool) *bool { return &v }

func joinURL(base, p string) (*url.URL, error) {
	if strings.TrimSpace(base) == "" {
		return nil, fmt.Errorf("empty base url")
	}
	parsed, err := url.Parse(strings.TrimSpace(base))
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(p, "http://") || strings.HasPrefix(p, "https://") {
		return url.Parse(p)
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/") + p
	return parsed, nil
}
