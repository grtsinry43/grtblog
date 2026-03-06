package ownerstatus

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/grtsinry43/grtblog-v2/server/internal/ws"
)

const (
	realtimeEventType      = "owner.status"
	ownerStatusTTL         = 5 * time.Minute
	adminPanelHeartbeatTTL = 90 * time.Second
	expireCheckInterval    = 10 * time.Second
)

type Media struct {
	Title     string `json:"title"`
	Artist    string `json:"artist"`
	Thumbnail string `json:"thumbnail"`
}

type Snapshot struct {
	OK               int    `json:"ok"`
	Process          string `json:"process,omitempty"`
	Extend           string `json:"extend,omitempty"`
	Media            *Media `json:"media,omitempty"`
	Timestamp        int64  `json:"timestamp,omitempty"`
	AdminPanelOnline bool   `json:"adminPanelOnline"`
}

type UpdateInput struct {
	OK        *int   `json:"ok"`
	Process   string `json:"process"`
	Extend    string `json:"extend"`
	Media     *Media `json:"media"`
	Timestamp *int64 `json:"timestamp"`
}

type Service struct {
	manager *ws.Manager

	mu                 sync.Mutex
	status             Snapshot
	lastOwnerUpdatedAt time.Time
	lastPanelBeatAt    time.Time
}

func NewService(manager *ws.Manager) *Service {
	svc := &Service{
		manager: manager,
		status: Snapshot{
			OK:               0,
			AdminPanelOnline: false,
		},
	}
	go svc.expireLoop()
	return svc
}

func (s *Service) Update(input UpdateInput) Snapshot {
	now := time.Now()

	s.mu.Lock()
	s.status.OK = normalizeOK(input.OK)
	s.status.Process = strings.TrimSpace(input.Process)
	s.status.Extend = strings.TrimSpace(input.Extend)
	s.status.Media = sanitizeMedia(input.Media)
	s.status.Timestamp = normalizeTimestamp(input.Timestamp, now)
	s.lastOwnerUpdatedAt = now

	if s.status.AdminPanelOnline && isPanelBeatExpired(now, s.lastPanelBeatAt) {
		s.status.AdminPanelOnline = false
	}
	snapshot := s.status
	s.mu.Unlock()

	s.broadcast(snapshot)
	return snapshot
}

func (s *Service) TouchAdminPanel() Snapshot {
	now := time.Now()

	s.mu.Lock()
	changed := false
	s.lastPanelBeatAt = now
	if !s.status.AdminPanelOnline {
		s.status.AdminPanelOnline = true
		changed = true
	}
	if s.status.OK == 1 && isOwnerExpired(now, s.lastOwnerUpdatedAt) {
		s.status.OK = 0
		s.status.Process = ""
		s.status.Extend = ""
		s.status.Media = nil
		changed = true
	}
	snapshot := s.status
	s.mu.Unlock()

	if changed {
		s.broadcast(snapshot)
	}
	return snapshot
}

func (s *Service) Get() Snapshot {
	now := time.Now()

	s.mu.Lock()
	changed := s.applyExpireLocked(now)
	snapshot := s.status
	s.mu.Unlock()

	if changed {
		s.broadcast(snapshot)
	}
	return snapshot
}

func (s *Service) expireLoop() {
	ticker := time.NewTicker(expireCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()

		s.mu.Lock()
		changed := s.applyExpireLocked(now)
		snapshot := s.status
		s.mu.Unlock()

		if changed {
			s.broadcast(snapshot)
		}
	}
}

func (s *Service) applyExpireLocked(now time.Time) bool {
	changed := false

	if s.status.OK == 1 && isOwnerExpired(now, s.lastOwnerUpdatedAt) {
		s.status.OK = 0
		s.status.Process = ""
		s.status.Extend = ""
		s.status.Media = nil
		changed = true
	}
	if s.status.AdminPanelOnline && isPanelBeatExpired(now, s.lastPanelBeatAt) {
		s.status.AdminPanelOnline = false
		changed = true
	}

	return changed
}

func (s *Service) broadcast(snapshot Snapshot) {
	if s == nil || s.manager == nil {
		return
	}

	payload := struct {
		Type string `json:"type"`
		Snapshot
	}{
		Type:     realtimeEventType,
		Snapshot: snapshot,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	s.manager.Broadcast(ws.RealtimeRoomKey(), data)
}

func normalizeOK(raw *int) int {
	if raw == nil {
		return 1
	}
	if *raw <= 0 {
		return 0
	}
	return 1
}

func normalizeTimestamp(raw *int64, now time.Time) int64 {
	if raw == nil || *raw <= 0 {
		return now.Unix()
	}
	return *raw
}

func sanitizeMedia(media *Media) *Media {
	if media == nil {
		return nil
	}
	next := &Media{
		Title:     strings.TrimSpace(media.Title),
		Artist:    strings.TrimSpace(media.Artist),
		Thumbnail: strings.TrimSpace(media.Thumbnail),
	}
	if next.Title == "" && next.Artist == "" && next.Thumbnail == "" {
		return nil
	}
	return next
}

func isOwnerExpired(now time.Time, updatedAt time.Time) bool {
	if updatedAt.IsZero() {
		return true
	}
	return now.Sub(updatedAt) >= ownerStatusTTL
}

func isPanelBeatExpired(now time.Time, beatAt time.Time) bool {
	if beatAt.IsZero() {
		return true
	}
	return now.Sub(beatAt) >= adminPanelHeartbeatTTL
}
