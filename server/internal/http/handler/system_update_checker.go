package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type SystemUpdateInfo struct {
	Status              string             `json:"status"`
	Enabled             bool               `json:"enabled"`
	Repo                string             `json:"repo"`
	Channel             string             `json:"channel"`
	CurrentVersion      string             `json:"currentVersion"`
	CurrentPrerelease   bool               `json:"currentPrerelease"`
	HasUpdate           bool               `json:"hasUpdate"`
	Comparison          string             `json:"comparison"`
	UpgradeURL          string             `json:"upgradeUrl,omitempty"`
	CheckedAt           time.Time          `json:"checkedAt"`
	Message             string             `json:"message,omitempty"`
	TargetRelease       *UpdateReleaseInfo `json:"targetRelease,omitempty"`
	LatestRelease       *UpdateReleaseInfo `json:"latestRelease,omitempty"`
	LatestStableRelease *UpdateReleaseInfo `json:"latestStableRelease,omitempty"`
}

type UpdateReleaseInfo struct {
	Tag         string    `json:"tag"`
	Name        string    `json:"name"`
	Prerelease  bool      `json:"prerelease"`
	PublishedAt time.Time `json:"publishedAt"`
	URL         string    `json:"url"`
}

type githubRelease struct {
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	HTMLURL     string    `json:"html_url"`
	PublishedAt time.Time `json:"published_at"`
}

const updateCheckTTL = 30 * time.Minute

func (h *SystemHandler) getCachedUpdateCheck(ctx context.Context, force bool) SystemUpdateInfo {
	if !h.appCfg.UpdateCheckEnabled || strings.TrimSpace(h.appCfg.UpdateCheckRepo) == "" {
		return SystemUpdateInfo{
			Status:            "disabled",
			Enabled:           false,
			Repo:              strings.TrimSpace(h.appCfg.UpdateCheckRepo),
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
	if latest == nil {
		return SystemUpdateInfo{
			Status:            "ok",
			Enabled:           true,
			Repo:              repo,
			Channel:           "stable",
			CurrentVersion:    h.version,
			CurrentPrerelease: isPrereleaseTag(h.version),
			HasUpdate:         false,
			Comparison:        "unknown",
			CheckedAt:         time.Now().UTC(),
			Message:           "no GitHub releases found",
		}, nil
	}

	currentVersion := h.version
	currentPrerelease := isPrereleaseTag(currentVersion)
	channel := "stable"
	target := latestStable
	if currentPrerelease {
		channel = "prerelease"
		target = latest
	}
	if target == nil {
		target = latest
	}

	comparison, hasUpdate, message := compareVersionForUpdate(currentVersion, target.Tag)

	result := SystemUpdateInfo{
		Status:              "ok",
		Enabled:             true,
		Repo:                repo,
		Channel:             channel,
		CurrentVersion:      currentVersion,
		CurrentPrerelease:   currentPrerelease,
		HasUpdate:           hasUpdate,
		Comparison:          comparison,
		CheckedAt:           time.Now().UTC(),
		Message:             message,
		TargetRelease:       target,
		LatestRelease:       latest,
		LatestStableRelease: latestStable,
	}
	if target.URL != "" {
		result.UpgradeURL = target.URL
	}
	if target.Prerelease {
		if result.Message == "" {
			result.Message = "latest target is a prerelease (testing build)"
		} else {
			result.Message += "; target is prerelease (testing build)"
		}
	}
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

func toUpdateRelease(item githubRelease) UpdateReleaseInfo {
	name := strings.TrimSpace(item.Name)
	if name == "" {
		name = item.TagName
	}
	return UpdateReleaseInfo{
		Tag:         strings.TrimSpace(item.TagName),
		Name:        name,
		Prerelease:  item.Prerelease,
		PublishedAt: item.PublishedAt.UTC(),
		URL:         strings.TrimSpace(item.HTMLURL),
	}
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
