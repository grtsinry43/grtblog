package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type SystemUpdateInfo struct {
	Status              string             `json:"status"`
	Enabled             bool               `json:"enabled"`
	Repo                string             `json:"repo"`
	Channel             string             `json:"channel"`
	Source              string             `json:"source"`
	CurrentVersion      string             `json:"currentVersion"`
	CurrentPrerelease   bool               `json:"currentPrerelease"`
	HasUpdate           bool               `json:"hasUpdate"`
	Comparison          string             `json:"comparison"`
	UpgradeURL          string             `json:"upgradeUrl,omitempty"`
	ReleaseNotesURL     string             `json:"releaseNotesUrl,omitempty"`
	CheckedAt           time.Time          `json:"checkedAt"`
	Message             string             `json:"message,omitempty"`
	TargetRelease       *UpdateReleaseInfo `json:"targetRelease,omitempty"`
	LatestRelease       *UpdateReleaseInfo `json:"latestRelease,omitempty"`
	LatestStableRelease *UpdateReleaseInfo `json:"latestStableRelease,omitempty"`
}

type UpdateReleaseInfo struct {
	Tag         string    `json:"tag"`
	Name        string    `json:"name"`
	Body        string    `json:"body,omitempty"`
	Prerelease  bool      `json:"prerelease"`
	PublishedAt time.Time `json:"publishedAt"`
	URL         string    `json:"url"`
}

type githubRelease struct {
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	HTMLURL     string    `json:"html_url"`
	PublishedAt time.Time `json:"published_at"`
}

type githubTag struct {
	Name   string `json:"name"`
	Commit struct {
		SHA string `json:"sha"`
		URL string `json:"url"`
	} `json:"commit"`
}

type githubCommit struct {
	Commit struct {
		Committer struct {
			Date time.Time `json:"date"`
		} `json:"committer"`
		Author struct {
			Date time.Time `json:"date"`
		} `json:"author"`
	} `json:"commit"`
}

const updateCheckTTL = 30 * time.Minute

func (h *SystemHandler) getCachedUpdateCheck(ctx context.Context, force bool) SystemUpdateInfo {
	if !h.appCfg.UpdateCheckEnabled || strings.TrimSpace(h.appCfg.UpdateCheckRepo) == "" {
		return SystemUpdateInfo{
			Status:            "disabled",
			Enabled:           false,
			Repo:              strings.TrimSpace(h.appCfg.UpdateCheckRepo),
			Channel:           normalizeUpdateChannel(h.appCfg.UpdateCheckChannel),
			Source:            "disabled",
			CurrentVersion:    h.version,
			CurrentPrerelease: isPrereleaseTag(h.version),
			CheckedAt:         time.Now().UTC(),
			Message:           "update check disabled",
		}
	}

	ttl := updateCheckTTL

	h.updateMu.RLock()
	if !force && !h.lastUpdateCheck.IsZero() && time.Since(h.lastUpdateCheck) < ttl {
		cached := h.updateCache
		h.updateMu.RUnlock()
		return cached
	}
	h.updateMu.RUnlock()

	h.updateMu.Lock()
	defer h.updateMu.Unlock()
	if !force && !h.lastUpdateCheck.IsZero() && time.Since(h.lastUpdateCheck) < ttl {
		return h.updateCache
	}

	info, err := h.fetchUpdateCheck(ctx)
	if err != nil {
		info = SystemUpdateInfo{
			Status:            "error",
			Enabled:           true,
			Repo:              strings.TrimSpace(h.appCfg.UpdateCheckRepo),
			Channel:           normalizeUpdateChannel(h.appCfg.UpdateCheckChannel),
			Source:            "error",
			CurrentVersion:    h.version,
			CurrentPrerelease: isPrereleaseTag(h.version),
			CheckedAt:         time.Now().UTC(),
			Message:           err.Error(),
		}
	}
	h.updateCache = info
	h.lastUpdateCheck = time.Now()
	return info
}

func (h *SystemHandler) peekCachedUpdateCheck() SystemUpdateInfo {
	h.updateMu.RLock()
	defer h.updateMu.RUnlock()
	if h.lastUpdateCheck.IsZero() {
		return SystemUpdateInfo{
			Status:            "idle",
			Enabled:           h.appCfg.UpdateCheckEnabled,
			Repo:              strings.TrimSpace(h.appCfg.UpdateCheckRepo),
			Channel:           normalizeUpdateChannel(h.appCfg.UpdateCheckChannel),
			Source:            "idle",
			CurrentVersion:    h.version,
			CurrentPrerelease: isPrereleaseTag(h.version),
			CheckedAt:         time.Now().UTC(),
			Message:           "not checked yet",
		}
	}
	return h.updateCache
}

func (h *SystemHandler) fetchUpdateCheck(ctx context.Context) (SystemUpdateInfo, error) {
	repo := strings.TrimSpace(h.appCfg.UpdateCheckRepo)
	channel := normalizeUpdateChannel(h.appCfg.UpdateCheckChannel)
	currentVersion := h.version
	currentPrerelease := isPrereleaseTag(currentVersion)

	if channel == "preview" {
		return h.fetchPreviewUpdateCheck(ctx, repo, currentVersion, currentPrerelease)
	}
	return h.fetchStableUpdateCheck(ctx, repo, currentVersion, currentPrerelease)
}

func (h *SystemHandler) fetchStableUpdateCheck(ctx context.Context, repo, currentVersion string, currentPrerelease bool) (SystemUpdateInfo, error) {
	releases, err := h.fetchGitHubReleases(ctx, repo)
	if err != nil {
		return SystemUpdateInfo{}, err
	}

	var latest *UpdateReleaseInfo
	var latestStable *UpdateReleaseInfo
	for _, release := range releases {
		item := toUpdateRelease(release)
		if latest == nil {
			latest = &item
		}
		if !release.Prerelease && latestStable == nil {
			latestStable = &item
		}
		if latest != nil && latestStable != nil {
			break
		}
	}

	result := SystemUpdateInfo{
		Status:              "ok",
		Enabled:             true,
		Repo:                repo,
		Channel:             "stable",
		Source:              "github_release",
		CurrentVersion:      currentVersion,
		CurrentPrerelease:   currentPrerelease,
		HasUpdate:           false,
		Comparison:          "unknown",
		CheckedAt:           time.Now().UTC(),
		LatestRelease:       latest,
		LatestStableRelease: latestStable,
		Message:             "version source: GitHub Releases",
	}
	if latestStable == nil {
		result.Message = "no stable GitHub releases found"
		return result, nil
	}

	comparison, hasUpdate, message := compareVersionForUpdate(currentVersion, latestStable.Tag)
	result.TargetRelease = latestStable
	result.HasUpdate = hasUpdate
	result.Comparison = comparison
	if latestStable.URL != "" {
		result.UpgradeURL = latestStable.URL
		result.ReleaseNotesURL = latestStable.URL
	}
	if message != "" {
		result.Message = fmt.Sprintf("version source: GitHub Releases; %s", message)
	}
	return result, nil
}

func (h *SystemHandler) fetchPreviewUpdateCheck(ctx context.Context, repo, currentVersion string, currentPrerelease bool) (SystemUpdateInfo, error) {
	tags, err := h.fetchGitHubTags(ctx, repo)
	if err != nil {
		return SystemUpdateInfo{}, err
	}

	latestStable, _ := h.fetchLatestStableRelease(ctx, repo)

	latestTagName, targetTagName, channelMessage := selectPreviewTag(tags, currentVersion)
	var latest *UpdateReleaseInfo
	if latestTagName != "" {
		item, err := h.fetchTagReleaseInfo(ctx, repo, latestTagName)
		if err == nil {
			latest = &item
		}
	}

	result := SystemUpdateInfo{
		Status:              "ok",
		Enabled:             true,
		Repo:                repo,
		Channel:             "preview",
		Source:              "github_tag",
		CurrentVersion:      currentVersion,
		CurrentPrerelease:   currentPrerelease,
		HasUpdate:           false,
		Comparison:          "unknown",
		CheckedAt:           time.Now().UTC(),
		LatestRelease:       latest,
		LatestStableRelease: latestStable,
		Message:             strings.TrimSpace(channelMessage),
	}

	if targetTagName == "" {
		if result.Message == "" {
			result.Message = "preview channel found no prerelease tag for the current major"
		}
		return result, nil
	}

	target, err := h.fetchTagReleaseInfo(ctx, repo, targetTagName)
	if err != nil {
		return SystemUpdateInfo{}, err
	}
	comparison, hasUpdate, compareMessage := compareVersionForUpdate(currentVersion, target.Tag)

	result.TargetRelease = &target
	result.HasUpdate = hasUpdate
	result.Comparison = comparison
	if target.URL != "" {
		result.UpgradeURL = target.URL
	}
	if target.URL != "" {
		result.ReleaseNotesURL = target.URL
	}

	parts := make([]string, 0, 3)
	if result.Message != "" {
		parts = append(parts, result.Message)
	}
	parts = append(parts, "version source: Git tags")
	if compareMessage != "" {
		parts = append(parts, compareMessage)
	}
	result.Message = strings.Join(parts, "; ")
	return result, nil
}

func (h *SystemHandler) fetchGitHubReleases(ctx context.Context, repo string) ([]githubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases?per_page=10", repo)

	client := &http.Client{Timeout: 4 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "grtblog-update-check")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api %s", resp.Status)
	}

	var releases []githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}

	filtered := make([]githubRelease, 0, len(releases))
	for _, item := range releases {
		if item.Draft {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered, nil
}

func (h *SystemHandler) fetchGitHubTags(ctx context.Context, repo string) ([]githubTag, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/tags?per_page=100", repo)

	client := &http.Client{Timeout: 4 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "grtblog-update-check")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api %s", resp.Status)
	}

	var tags []githubTag
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, err
	}
	return tags, nil
}

func (h *SystemHandler) fetchLatestStableRelease(ctx context.Context, repo string) (*UpdateReleaseInfo, error) {
	releases, err := h.fetchGitHubReleases(ctx, repo)
	if err != nil {
		return nil, err
	}
	for _, release := range releases {
		if release.Draft || release.Prerelease {
			continue
		}
		item := toUpdateRelease(release)
		return &item, nil
	}
	return nil, nil
}

func (h *SystemHandler) fetchTagReleaseInfo(ctx context.Context, repo, tag string) (UpdateReleaseInfo, error) {
	item := UpdateReleaseInfo{
		Tag:        strings.TrimSpace(tag),
		Name:       strings.TrimSpace(tag),
		Prerelease: isPrereleaseTag(tag),
		URL:        fmt.Sprintf("https://github.com/%s/tree/%s", repo, tag),
	}

	if notesBody, notesURL, err := h.fetchTagReleaseNotes(ctx, repo, tag); err == nil {
		item.Body = notesBody
		if notesURL != "" {
			item.URL = notesURL
		}
	}

	if commitTime, err := h.fetchTagCommitTime(ctx, repo, tag); err == nil && !commitTime.IsZero() {
		item.PublishedAt = commitTime.UTC()
	}
	return item, nil
}

func (h *SystemHandler) fetchTagReleaseNotes(ctx context.Context, repo, tag string) (string, string, error) {
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/docs/releases/%s.md", repo, tag, tag)

	client := &http.Client{Timeout: 4 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", "grtblog-update-check")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("release notes %s", resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
	if err != nil {
		return "", "", err
	}

	content := strings.TrimSpace(string(body))
	if content == "" {
		return "", "", fmt.Errorf("release notes empty")
	}
	htmlURL := fmt.Sprintf("https://github.com/%s/blob/%s/docs/releases/%s.md", repo, tag, tag)
	return content, htmlURL, nil
}

func (h *SystemHandler) fetchTagCommitTime(ctx context.Context, repo, tag string) (time.Time, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/commits/%s", repo, tag)

	client := &http.Client{Timeout: 4 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return time.Time{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "grtblog-update-check")

	resp, err := client.Do(req)
	if err != nil {
		return time.Time{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return time.Time{}, fmt.Errorf("github api %s", resp.Status)
	}

	var commit githubCommit
	if err := json.NewDecoder(resp.Body).Decode(&commit); err != nil {
		return time.Time{}, err
	}
	if !commit.Commit.Committer.Date.IsZero() {
		return commit.Commit.Committer.Date, nil
	}
	return commit.Commit.Author.Date, nil
}

func toUpdateRelease(item githubRelease) UpdateReleaseInfo {
	name := strings.TrimSpace(item.Name)
	if name == "" {
		name = item.TagName
	}
	return UpdateReleaseInfo{
		Tag:         strings.TrimSpace(item.TagName),
		Name:        name,
		Body:        strings.TrimSpace(item.Body),
		Prerelease:  item.Prerelease,
		PublishedAt: item.PublishedAt.UTC(),
		URL:         strings.TrimSpace(item.HTMLURL),
	}
}

func normalizeUpdateChannel(channel string) string {
	switch strings.ToLower(strings.TrimSpace(channel)) {
	case "preview":
		return "preview"
	default:
		return "stable"
	}
}

func selectPreviewTag(tags []githubTag, currentVersion string) (latest string, target string, message string) {
	candidates := make([]semverTag, 0, len(tags))
	for _, item := range tags {
		parsed, ok := parseSemver(item.Name)
		if !ok || len(parsed.pre) == 0 {
			continue
		}
		candidates = append(candidates, semverTag{
			tag:    strings.TrimSpace(item.Name),
			parsed: parsed,
		})
	}
	if len(candidates) == 0 {
		return "", "", "preview channel found no prerelease tags"
	}

	sort.Slice(candidates, func(i, j int) bool {
		return compareSemver(candidates[i].parsed, candidates[j].parsed) > 0
	})
	latest = candidates[0].tag

	current, ok := parseSemver(currentVersion)
	if !ok {
		return latest, latest, "preview channel follows the latest prerelease tag"
	}

	sameMajor := make([]semverTag, 0, len(candidates))
	for _, item := range candidates {
		if item.parsed.major == current.major {
			sameMajor = append(sameMajor, item)
		}
	}
	if len(sameMajor) == 0 {
		return latest, "", fmt.Sprintf("preview channel only targets current major %d; next available preview is %s", current.major, latest)
	}
	return latest, sameMajor[0].tag, fmt.Sprintf("preview channel tracks prerelease tags in current major %d", current.major)
}

type semverTag struct {
	tag    string
	parsed semverValue
}

func isPrereleaseTag(version string) bool {
	parsed, ok := parseSemver(strings.TrimSpace(version))
	if !ok {
		return false
	}
	return len(parsed.pre) > 0
}

func compareVersionForUpdate(current, latest string) (comparison string, hasUpdate bool, message string) {
	cur, okCur := parseSemver(current)
	lat, okLat := parseSemver(latest)
	if okCur && okLat {
		cmp := compareSemver(cur, lat)
		switch {
		case cmp < 0:
			return "older", true, ""
		case cmp == 0:
			return "equal", false, ""
		default:
			return "newer", false, "current build is newer than target release"
		}
	}

	if strings.TrimSpace(current) != "" && strings.TrimSpace(current) == strings.TrimSpace(latest) {
		return "equal", false, ""
	}
	return "unknown", false, "current version is a non-semver build (likely commit hash), comparison is manual"
}

var semverPattern = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z.-]+))?(?:\+[0-9A-Za-z.-]+)?$`)

type semverValue struct {
	major int
	minor int
	patch int
	pre   []semverID
}

type semverID struct {
	isNum bool
	num   int
	str   string
}

func parseSemver(raw string) (semverValue, bool) {
	input := strings.TrimSpace(raw)
	matched := semverPattern.FindStringSubmatch(input)
	if len(matched) != 5 {
		return semverValue{}, false
	}

	major, err := strconv.Atoi(matched[1])
	if err != nil {
		return semverValue{}, false
	}
	minor, err := strconv.Atoi(matched[2])
	if err != nil {
		return semverValue{}, false
	}
	patch, err := strconv.Atoi(matched[3])
	if err != nil {
		return semverValue{}, false
	}
	result := semverValue{major: major, minor: minor, patch: patch}
	if matched[4] == "" {
		return result, true
	}

	parts := strings.Split(matched[4], ".")
	result.pre = make([]semverID, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			return semverValue{}, false
		}
		if isAllDigits(part) {
			n, err := strconv.Atoi(part)
			if err != nil {
				return semverValue{}, false
			}
			result.pre = append(result.pre, semverID{isNum: true, num: n, str: part})
			continue
		}
		result.pre = append(result.pre, semverID{isNum: false, str: part})
	}
	return result, true
}

func compareSemver(a, b semverValue) int {
	switch {
	case a.major != b.major:
		if a.major < b.major {
			return -1
		}
		return 1
	case a.minor != b.minor:
		if a.minor < b.minor {
			return -1
		}
		return 1
	case a.patch != b.patch:
		if a.patch < b.patch {
			return -1
		}
		return 1
	}

	aPre := len(a.pre) > 0
	bPre := len(b.pre) > 0
	if !aPre && !bPre {
		return 0
	}
	if !aPre && bPre {
		return 1
	}
	if aPre && !bPre {
		return -1
	}

	maxLen := len(a.pre)
	if len(b.pre) > maxLen {
		maxLen = len(b.pre)
	}
	for i := 0; i < maxLen; i++ {
		if i >= len(a.pre) {
			return -1
		}
		if i >= len(b.pre) {
			return 1
		}
		cmp := compareSemverID(a.pre[i], b.pre[i])
		if cmp != 0 {
			return cmp
		}
	}
	return 0
}

func compareSemverID(a, b semverID) int {
	if a.isNum && b.isNum {
		switch {
		case a.num < b.num:
			return -1
		case a.num > b.num:
			return 1
		default:
			return 0
		}
	}
	if a.isNum && !b.isNum {
		return -1
	}
	if !a.isNum && b.isNum {
		return 1
	}
	if a.str < b.str {
		return -1
	}
	if a.str > b.str {
		return 1
	}
	return 0
}

func isAllDigits(input string) bool {
	for _, ch := range input {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return input != ""
}
