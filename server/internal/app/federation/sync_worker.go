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
	if w == nil || w.instanceRepo == nil || w.cacheRepo == nil || w.resolver == nil {
		return
	}
	if w.syncJobRepo == nil {
		w.syncOnceDirect(ctx)
		return
	}
	now := time.Now().UTC()
	_ = w.enqueueInstanceJobs(ctx)
	_ = w.enqueueRSSFriendLinkJobs(ctx, now)
	_ = w.processSyncJobs(ctx, now, 200)
}

func (w *SyncWorker) syncOnceDirect(ctx context.Context) {
	instances, err := w.instanceRepo.ListActive(ctx)
	if err != nil {
		return
	}
	for _, instance := range instances {
		_, _, _ = w.syncInstance(ctx, instance)
	}
	_ = w.syncRSSFriendLinks(ctx, time.Now().UTC())
}

func (w *SyncWorker) syncInstance(ctx context.Context, instance domainfed.FederationInstance) (int, string, error) {
	baseURL := strings.TrimRight(instance.BaseURL, "/")
	if baseURL == "" {
		return 0, social.FriendLinkSyncJobMethodTimeline, nil
	}
	endpoints, err := w.resolver.FetchEndpoints(ctx, baseURL)
	if err == nil && endpoints != nil {
		if posts, err := w.fetchTimelinePosts(ctx, instance.ID, endpoints); err == nil && len(posts) > 0 {
			if err := w.cacheRepo.UpsertBatch(ctx, posts); err != nil {
				return 0, social.FriendLinkSyncJobMethodTimeline, err
			}
			return len(posts), social.FriendLinkSyncJobMethodTimeline, nil
		}
	}
	count, err := w.syncFromRSS(ctx, instance.ID, baseURL)
	return count, social.FriendLinkSyncJobMethodRSS, err
}

func (w *SyncWorker) fetchTimelinePosts(ctx context.Context, instanceID int64, endpoints *fedinfra.EndpointsDoc) ([]domainfed.FederatedPostCache, error) {
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
	baseURL := strings.TrimSpace(endpoints.BaseURL)
	if baseURL == "" {
		return nil, fmt.Errorf("endpoints.base_url is empty")
	}
	u, err := joinURL(baseURL, path)
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
		posts = append(posts, domainfed.FederatedPostCache{
			InstanceID:     instanceID,
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
			CachedAt:       time.Now().UTC(),
		})
	}
	return posts, nil
}

func (w *SyncWorker) syncFromRSS(ctx context.Context, instanceID int64, baseURL string) (int, error) {
	manifest, err := w.resolver.FetchManifest(ctx, baseURL)
	if err != nil || manifest == nil || len(manifest.RSSFeeds) == 0 {
		return 0, err
	}
	feedURL := strings.TrimSpace(manifest.RSSFeeds[0].URL)
	if feedURL == "" {
		return 0, nil
	}
	return w.syncFromFeedURL(ctx, instanceID, feedURL)
}

func (w *SyncWorker) syncFromFeedURL(ctx context.Context, instanceID int64, feedURL string) (int, error) {
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
	posts := parseFeedItems(feed, instanceID)
	if len(posts) == 0 {
		return 0, nil
	}
	if err := w.cacheRepo.UpsertBatch(ctx, posts); err != nil {
		return 0, err
	}
	return len(posts), nil
}

func (w *SyncWorker) syncRSSFriendLinks(ctx context.Context, now time.Time) error {
	if w.linkRepo == nil || w.instanceRepo == nil {
		return nil
	}
	page := 1
	pageSize := 100
	for {
		links, total, err := w.linkRepo.List(ctx, social.FriendLinkListOptions{
			IsActive: ptrBool(true),
			SyncMode: social.FriendLinkSyncModeRSS,
			Page:     page,
			PageSize: pageSize,
		})
		if err != nil {
			return err
		}
		for i := range links {
			rssURL := strings.TrimSpace(optionalStr(links[i].RSSURL))
			if rssURL == "" {
				continue
			}
			if !shouldSyncRSSFriendLink(links[i], now, 30*time.Minute) {
				continue
			}
			instance, err := w.ensureRSSInstance(ctx, &links[i])
			if err != nil || instance == nil {
				continue
			}
			if _, err := w.syncFromFeedURL(ctx, instance.ID, rssURL); err == nil {
				now := time.Now().UTC()
				ok := "ok"
				links[i].LastSyncAt = &now
				links[i].LastSyncStatus = &ok
				if posts, err := w.cacheRepo.ListByInstance(ctx, instance.ID, nil, 0); err == nil {
					links[i].TotalPostsCached = len(posts)
				}
				_ = w.linkRepo.Update(ctx, &links[i])
			} else {
				failed := "failed"
				links[i].LastSyncStatus = &failed
				_ = w.linkRepo.Update(ctx, &links[i])
			}
		}
		if int64(page*pageSize) >= total || len(links) == 0 {
			break
		}
		page++
	}
	return nil
}

func (w *SyncWorker) enqueueInstanceJobs(ctx context.Context) error {
	if w.syncJobRepo == nil || w.instanceRepo == nil {
		return nil
	}
	instances, err := w.instanceRepo.ListActive(ctx)
	if err != nil {
		return err
	}
	for i := range instances {
		targetURL := strings.TrimSpace(instances[i].BaseURL)
		if targetURL == "" {
			continue
		}
		instanceID := instances[i].ID
		job := &social.FriendLinkSyncJob{
			TargetType:    social.FriendLinkSyncJobTargetFederationInstance,
			SyncMethod:    social.FriendLinkSyncJobMethodTimeline,
			InstanceID:    &instanceID,
			TargetURL:     targetURL,
			Status:        social.FriendLinkSyncJobStatusQueued,
			MaxAttempts:   1,
			TriggerSource: "scheduler",
		}
		_ = w.syncJobRepo.Create(ctx, job)
	}
	return nil
}

func (w *SyncWorker) enqueueRSSFriendLinkJobs(ctx context.Context, now time.Time) error {
	if w.syncJobRepo == nil || w.linkRepo == nil {
		return nil
	}
	page := 1
	pageSize := 100
	for {
		links, total, err := w.linkRepo.List(ctx, social.FriendLinkListOptions{
			IsActive: ptrBool(true),
			SyncMode: social.FriendLinkSyncModeRSS,
			Page:     page,
			PageSize: pageSize,
		})
		if err != nil {
			return err
		}
		for i := range links {
			rssURL := strings.TrimSpace(optionalStr(links[i].RSSURL))
			if rssURL == "" {
				continue
			}
			if !shouldSyncRSSFriendLink(links[i], now, 30*time.Minute) {
				continue
			}
			friendLinkID := links[i].ID
			job := &social.FriendLinkSyncJob{
				TargetType:    social.FriendLinkSyncJobTargetFriendLink,
				SyncMethod:    social.FriendLinkSyncJobMethodRSS,
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
		if int64(page*pageSize) >= total || len(links) == 0 {
			break
		}
		page++
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
	switch job.TargetType {
	case social.FriendLinkSyncJobTargetFederationInstance:
		var instance *domainfed.FederationInstance
		var err error
		if job.InstanceID != nil && *job.InstanceID > 0 {
			instance, err = w.instanceRepo.GetByID(ctx, *job.InstanceID)
		} else {
			instance, err = w.instanceRepo.GetByBaseURL(ctx, strings.TrimSpace(job.TargetURL))
		}
		if err != nil {
			return 0, social.FriendLinkSyncJobMethodTimeline, err
		}
		return w.syncInstance(ctx, *instance)
	case social.FriendLinkSyncJobTargetFriendLink:
		var link *social.FriendLink
		var err error
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
		rssURL := strings.TrimSpace(optionalStr(link.RSSURL))
		if rssURL == "" {
			return 0, social.FriendLinkSyncJobMethodRSS, fmt.Errorf("rss url is empty")
		}
		instance, err := w.ensureRSSInstance(ctx, link)
		if err != nil || instance == nil {
			if err == nil {
				err = fmt.Errorf("rss instance not resolved")
			}
			failed := "failed"
			link.LastSyncStatus = &failed
			_ = w.linkRepo.Update(ctx, link)
			return 0, social.FriendLinkSyncJobMethodRSS, err
		}
		count, err := w.syncFromFeedURL(ctx, instance.ID, rssURL)
		if err != nil {
			failed := "failed"
			link.LastSyncStatus = &failed
			_ = w.linkRepo.Update(ctx, link)
			return 0, social.FriendLinkSyncJobMethodRSS, err
		}
		now := time.Now().UTC()
		ok := "ok"
		link.LastSyncAt = &now
		link.LastSyncStatus = &ok
		if posts, err := w.cacheRepo.ListByInstance(ctx, instance.ID, nil, 0); err == nil {
			link.TotalPostsCached = len(posts)
		}
		_ = w.linkRepo.Update(ctx, link)
		return count, social.FriendLinkSyncJobMethodRSS, nil
	default:
		return 0, social.FriendLinkSyncJobMethodRSS, fmt.Errorf("unsupported sync target type: %s", strings.TrimSpace(job.TargetType))
	}
}

func (w *SyncWorker) ensureRSSInstance(ctx context.Context, link *social.FriendLink) (*domainfed.FederationInstance, error) {
	if link == nil || w.instanceRepo == nil {
		return nil, nil
	}
	if link.InstanceID != nil && *link.InstanceID > 0 {
		return w.instanceRepo.GetByID(ctx, *link.InstanceID)
	}
	baseURL := strings.TrimRight(strings.TrimSpace(link.URL), "/")
	if baseURL == "" {
		return nil, nil
	}
	instance, err := w.instanceRepo.GetByBaseURL(ctx, baseURL)
	if err == nil && instance != nil {
		link.InstanceID = &instance.ID
		_ = w.linkRepo.Update(ctx, link)
		return instance, nil
	}
	status := "active"
	now := time.Now().UTC()
	newInstance := &domainfed.FederationInstance{
		BaseURL:    baseURL,
		Name:       ptrString(strings.TrimSpace(link.Name)),
		Status:     status,
		Features:   json.RawMessage(`["rss"]`),
		Policies:   json.RawMessage(`{}`),
		Endpoints:  json.RawMessage(`{}`),
		LastSeenAt: &now,
	}
	if err := w.instanceRepo.Create(ctx, newInstance); err != nil {
		return nil, err
	}
	link.InstanceID = &newInstance.ID
	_ = w.linkRepo.Update(ctx, link)
	return newInstance, nil
}

func parseFeedItems(feed *gofeed.Feed, instanceID int64) []domainfed.FederatedPostCache {
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
			CachedAt:       time.Now().UTC(),
		})
	}
	return posts
}

func shouldSyncRSSFriendLink(link social.FriendLink, now time.Time, fallbackInterval time.Duration) bool {
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
func ptrString(v string) *string {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	s := strings.TrimSpace(v)
	return &s
}

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
