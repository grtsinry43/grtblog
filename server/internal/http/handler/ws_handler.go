package handler

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofiber/websocket/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/analytics"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/ownerstatus"
	"github.com/grtsinry43/grtblog-v2/server/internal/ws"
)

type WSHandler struct {
	manager      *ws.Manager
	analyticsSvc *analytics.Service
	presenceHub  *ws.PresenceHub
	ownerStatus  *ownerstatus.Service
}

func NewWSHandler(manager *ws.Manager, analyticsSvc *analytics.Service, presenceHub *ws.PresenceHub, ownerStatus *ownerstatus.Service) *WSHandler {
	return &WSHandler{manager: manager, analyticsSvc: analyticsSvc, presenceHub: presenceHub, ownerStatus: ownerStatus}
}

func (h *WSHandler) Handle(conn *websocket.Conn) {
	if h.manager == nil {
		return
	}
	roomKey, ok := conn.Locals("wsRoomKey").(string)
	if !ok || roomKey == "" {
		return
	}
	client, cached := h.manager.Join(roomKey, conn)
	if client == nil {
		return
	}
	if h.analyticsSvc != nil {
		_ = h.analyticsSvc.TrackOnlineSample(context.Background(), h.manager.CurrentConnections())
	}
	for _, payload := range cached {
		_ = client.Write(payload)
	}

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}

	h.manager.Leave(roomKey, client)
	if h.analyticsSvc != nil {
		_ = h.analyticsSvc.TrackOnlineSample(context.Background(), h.manager.CurrentConnections())
	}
}

func (h *WSHandler) HandleNotification(conn *websocket.Conn) {
	if h.manager == nil {
		return
	}

	userID := localUserID(conn)
	if userID <= 0 {
		closeWithCode(conn, websocket.ClosePolicyViolation, "unauthorized")
		return
	}

	roomKey := ws.NotificationRoomKey(userID)
	client, cached := h.manager.Join(roomKey, conn)
	if client == nil {
		return
	}
	if h.analyticsSvc != nil {
		_ = h.analyticsSvc.TrackOnlineSample(context.Background(), h.manager.CurrentConnections())
	}
	for _, payload := range cached {
		_ = client.Write(payload)
	}
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
	h.manager.Leave(roomKey, client)
	if h.analyticsSvc != nil {
		_ = h.analyticsSvc.TrackOnlineSample(context.Background(), h.manager.CurrentConnections())
	}
}

func (h *WSHandler) HandlePresence(conn *websocket.Conn) {
	if h.manager == nil || h.presenceHub == nil {
		return
	}

	roomKey := ws.PresenceRoomKey()
	client, cached := h.manager.Join(roomKey, conn)
	if client == nil {
		return
	}
	h.presenceHub.Register(client)
	if h.analyticsSvc != nil {
		_ = h.analyticsSvc.TrackOnlineSample(context.Background(), h.manager.CurrentConnections())
	}

	for _, payload := range cached {
		_ = client.Write(payload)
	}

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			break
		}
		var payload ws.PresenceClientPayload
		if err := json.Unmarshal(data, &payload); err != nil {
			continue
		}
		h.presenceHub.Update(client, payload)
	}

	h.presenceHub.Unregister(client)
	h.manager.Leave(roomKey, client)
	if h.analyticsSvc != nil {
		_ = h.analyticsSvc.TrackOnlineSample(context.Background(), h.manager.CurrentConnections())
	}
}

type realtimeInboundMessage struct {
	Type        string `json:"type"`
	ContentType string `json:"contentType"`
	ContentID   int64  `json:"contentId"`
	URL         string `json:"url"`
	VisitorID   string `json:"visitorId"`
}

func (h *WSHandler) HandleRealtime(conn *websocket.Conn) {
	if h.manager == nil || h.presenceHub == nil {
		return
	}
	isAdmin := localUserIsAdmin(conn)

	rootRoom := ws.RealtimeRoomKey()
	client, cachedRoot := h.manager.Join(rootRoom, conn)
	if client == nil {
		return
	}
	cachedPresence := h.manager.JoinClient(ws.PresenceRoomKey(), client)
	h.presenceHub.Register(client)
	if h.analyticsSvc != nil {
		_ = h.analyticsSvc.TrackOnlineSample(context.Background(), h.manager.CurrentConnections())
	}
	for _, payload := range cachedRoot {
		_ = client.Write(payload)
	}
	for _, payload := range cachedPresence {
		_ = client.Write(payload)
	}
	if isAdmin && h.ownerStatus != nil {
		h.ownerStatus.TouchAdminPanel()
	}

	currentContentRoom := ""
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msg realtimeInboundMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "presence.identify":
			h.presenceHub.Identify(client, msg.VisitorID)
		case "presence.report":
			h.presenceHub.Update(client, ws.PresenceClientPayload{
				ContentType: msg.ContentType,
				URL:         msg.URL,
				VisitorID:   msg.VisitorID,
			})
		case "owner.panel.ping":
			if isAdmin && h.ownerStatus != nil {
				h.ownerStatus.TouchAdminPanel()
			}
		case "content.subscribe":
			roomKey, ok := ws.ContentRoomKey(msg.ContentType, msg.ContentID)
			if !ok {
				continue
			}
			if currentContentRoom == roomKey {
				continue
			}

			if currentContentRoom != "" {
				h.manager.Leave(currentContentRoom, client)
			}
			cachedContent := h.manager.JoinClient(roomKey, client)
			currentContentRoom = roomKey
			for _, payload := range cachedContent {
				_ = client.Write(payload)
			}
		case "content.unsubscribe":
			if currentContentRoom == "" {
				continue
			}
			h.manager.Leave(currentContentRoom, client)
			currentContentRoom = ""
		}
	}

	if currentContentRoom != "" {
		h.manager.Leave(currentContentRoom, client)
	}
	h.presenceHub.Unregister(client)
	h.manager.Leave(ws.PresenceRoomKey(), client)
	h.manager.Leave(rootRoom, client)
	if h.analyticsSvc != nil {
		_ = h.analyticsSvc.TrackOnlineSample(context.Background(), h.manager.CurrentConnections())
	}
}

func localUserID(conn *websocket.Conn) int64 {
	switch v := conn.Locals("wsUserID").(type) {
	case int64:
		return v
	case int:
		return int64(v)
	default:
		return 0
	}
}

func localUserIsAdmin(conn *websocket.Conn) bool {
	switch v := conn.Locals("wsUserIsAdmin").(type) {
	case bool:
		return v
	default:
		return false
	}
}

func closeWithCode(conn *websocket.Conn, code int, reason string) {
	_ = conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(code, reason),
		time.Now().Add(time.Second),
	)
	_ = conn.Close()
}
