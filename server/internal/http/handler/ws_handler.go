package handler

import (
	"context"
	"time"

	"github.com/gofiber/websocket/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/analytics"
	"github.com/grtsinry43/grtblog-v2/server/internal/ws"
)

type WSHandler struct {
	manager      *ws.Manager
	analyticsSvc *analytics.Service
}

func NewWSHandler(manager *ws.Manager, analyticsSvc *analytics.Service) *WSHandler {
	return &WSHandler{manager: manager, analyticsSvc: analyticsSvc}
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

func closeWithCode(conn *websocket.Conn, code int, reason string) {
	_ = conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(code, reason),
		time.Now().Add(time.Second),
	)
	_ = conn.Close()
}
