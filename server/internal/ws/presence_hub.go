package ws

import (
	"encoding/json"
	"sort"
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
	if h == nil || client == nil || h.resolver == nil {
		return
	}

	resolved, ok := h.resolver.Resolve(payload.ContentType, payload.URL)
	if !ok {
		return
	}

	h.mu.Lock()
	session, exists := h.sessions[client]
	if !exists {
		h.mu.Unlock()
		return
	}
	if session.ContentType == resolved.ContentType && session.URL == resolved.URL && session.Title == resolved.Title {
		h.mu.Unlock()
		return
	}

	h.sessions[client] = presenceSession{
		ContentType: resolved.ContentType,
		Title:       resolved.Title,
		URL:         resolved.URL,
		UpdatedAt:   time.Now(),
	}
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
		LastSeen    time.Time
	}

	groups := make(map[string]*aggregate)
	for _, session := range h.sessions {
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
			}
			groups[key] = current
		}
		current.Connections++
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
		Online: len(h.sessions),
		Pages:  pages,
	}
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
