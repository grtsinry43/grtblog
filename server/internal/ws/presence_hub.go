package ws

import (
	"encoding/json"
	"sort"
	"strings"
	"sync"
	"time"
)

const presenceRoomKey = "presence:lobby"

type PresenceResolver interface {
	Resolve(contentType string, rawURL string) (PresenceResolvedView, bool)
}

type PresenceResolvedView struct {
	ContentType string
	Title       string
	URL         string
}

type PresenceClientPayload struct {
	ContentType string `json:"contentType"`
	URL         string `json:"url"`
	VisitorID   string `json:"visitorId"`
}

type PresencePageItem struct {
	ContentType string `json:"contentType"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Connections int    `json:"connections"`
}

type PresenceSnapshotPayload struct {
	Type   string             `json:"type"`
	Online int                `json:"online"`
	Pages  []PresencePageItem `json:"pages"`
}

type presenceSession struct {
	ContentType string
	Title       string
	URL         string
	VisitorID   string
	UpdatedAt   time.Time
}

type PresenceHub struct {
	manager  *Manager
	resolver PresenceResolver

	mu       sync.Mutex
	sessions map[*Client]presenceSession
}

func NewPresenceHub(manager *Manager, resolver PresenceResolver) *PresenceHub {
	return &PresenceHub{
		manager:  manager,
		resolver: resolver,
		sessions: make(map[*Client]presenceSession),
	}
}

func PresenceRoomKey() string {
	return presenceRoomKey
}

func (h *PresenceHub) Register(client *Client) {
	if h == nil || client == nil {
		return
	}

	h.mu.Lock()
	h.sessions[client] = presenceSession{
		UpdatedAt: time.Now(),
	}
	snapshot := h.snapshotLocked()
	h.mu.Unlock()

	h.broadcast(snapshot)
}

func (h *PresenceHub) Unregister(client *Client) {
	if h == nil || client == nil {
		return
	}

	h.mu.Lock()
	delete(h.sessions, client)
	snapshot := h.snapshotLocked()
	h.mu.Unlock()

	h.broadcast(snapshot)
}

func (h *PresenceHub) Update(client *Client, payload PresenceClientPayload) {
	if h == nil || client == nil {
		return
	}

	h.mu.Lock()
	session, exists := h.sessions[client]
	if !exists {
		h.mu.Unlock()
		return
	}
	updated := false
	visitorID := normalizePresenceVisitorID(payload.VisitorID)
	if visitorID != "" && session.VisitorID != visitorID {
		session.VisitorID = visitorID
		updated = true
	}

	if h.resolver != nil {
		resolved, ok := h.resolver.Resolve(payload.ContentType, payload.URL)
		if ok && (session.ContentType != resolved.ContentType || session.URL != resolved.URL || session.Title != resolved.Title) {
			session.ContentType = resolved.ContentType
			session.Title = resolved.Title
			session.URL = resolved.URL
			updated = true
		}
	}

	if !updated {
		h.mu.Unlock()
		return
	}

	session.UpdatedAt = time.Now()
	h.sessions[client] = session
	snapshot := h.snapshotLocked()
	h.mu.Unlock()

	h.broadcast(snapshot)
}

func (h *PresenceHub) Identify(client *Client, visitorID string) {
	if h == nil || client == nil {
		return
	}

	normalized := normalizePresenceVisitorID(visitorID)
	if normalized == "" {
		return
	}

	h.mu.Lock()
	session, exists := h.sessions[client]
	if !exists {
		h.mu.Unlock()
		return
	}
	if session.VisitorID == normalized {
		h.mu.Unlock()
		return
	}

	session.VisitorID = normalized
	session.UpdatedAt = time.Now()
	h.sessions[client] = session
	snapshot := h.snapshotLocked()
	h.mu.Unlock()

	h.broadcast(snapshot)
}

func (h *PresenceHub) snapshotLocked() PresenceSnapshotPayload {
	type aggregate struct {
		ContentType string
		Title       string
		URL         string
		Connections int
		VisitorIDs  map[string]struct{}
		LastSeen    time.Time
	}

	groups := make(map[string]*aggregate)
	online := 0
	onlineVisitors := make(map[string]struct{})
	for _, session := range h.sessions {
		visitorID := normalizePresenceVisitorID(session.VisitorID)
		if visitorID == "" {
			online++
		} else if _, exists := onlineVisitors[visitorID]; !exists {
			onlineVisitors[visitorID] = struct{}{}
			online++
		}

		if session.URL == "" || session.ContentType == "" {
			continue
		}

		key := session.ContentType + "|" + session.URL
		current := groups[key]
		if current == nil {
			current = &aggregate{
				ContentType: session.ContentType,
				Title:       session.Title,
				URL:         session.URL,
				VisitorIDs:  make(map[string]struct{}),
			}
			groups[key] = current
		}
		if visitorID == "" {
			current.Connections++
		} else if _, exists := current.VisitorIDs[visitorID]; !exists {
			current.VisitorIDs[visitorID] = struct{}{}
			current.Connections++
		}
		if session.UpdatedAt.After(current.LastSeen) {
			current.LastSeen = session.UpdatedAt
			if session.Title != "" {
				current.Title = session.Title
			}
		}
	}

	pages := make([]PresencePageItem, 0, len(groups))
	for _, group := range groups {
		pages = append(pages, PresencePageItem{
			ContentType: group.ContentType,
			Title:       group.Title,
			URL:         group.URL,
			Connections: group.Connections,
		})
	}

	sort.Slice(pages, func(i, j int) bool {
		if pages[i].Connections != pages[j].Connections {
			return pages[i].Connections > pages[j].Connections
		}
		if pages[i].ContentType != pages[j].ContentType {
			return pages[i].ContentType < pages[j].ContentType
		}
		return pages[i].URL < pages[j].URL
	})

	return PresenceSnapshotPayload{
		Type:   "presence.snapshot",
		Online: online,
		Pages:  pages,
	}
}

func normalizePresenceVisitorID(raw string) string {
	id := strings.TrimSpace(raw)
	if id == "" {
		return ""
	}
	if len(id) > 255 {
		return id[:255]
	}
	return id
}

func (h *PresenceHub) broadcast(snapshot PresenceSnapshotPayload) {
	if h == nil || h.manager == nil {
		return
	}

	payload, err := json.Marshal(snapshot)
	if err != nil {
		return
	}
	h.manager.Broadcast(PresenceRoomKey(), payload)
}
