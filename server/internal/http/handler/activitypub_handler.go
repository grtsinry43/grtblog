package handler

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"

	appap "github.com/grtsinry43/grtblog-v2/server/internal/app/activitypub"
)

type ActivityPubHandler struct {
	svc *appap.Service
}

func NewActivityPubHandler(svc *appap.Service) *ActivityPubHandler {
	return &ActivityPubHandler{svc: svc}
}

func (h *ActivityPubHandler) WebFinger(c *fiber.Ctx) error {
	if h.svc == nil {
		return c.SendStatus(fiber.StatusNotFound)
	}
	baseURL := resolveActivityPubBaseURL(c)
	doc, matched, err := h.svc.BuildWebFinger(c.Context(), baseURL, c.Query("resource"))
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}
	if !matched {
		return c.SendStatus(fiber.StatusNotFound)
	}
	return sendJSONWithContentType(c, fiber.StatusOK, "application/jrd+json", doc)
}

func (h *ActivityPubHandler) Actor(c *fiber.Ctx) error {
	if h.svc == nil {
		return c.SendStatus(fiber.StatusNotFound)
	}
	// Browser → redirect to homepage
	if wantsHTML(c.Get("Accept")) {
		return c.Redirect(resolveActivityPubBaseURL(c)+"/", fiber.StatusFound)
	}
	baseURL := resolveActivityPubBaseURL(c)
	doc, err := h.svc.ActorDocument(c.Context(), baseURL)
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}
	return sendJSONWithContentType(c, fiber.StatusOK, "application/activity+json", doc)
}

// wantsHTML returns true when the Accept header indicates a browser request
// (wants text/html) rather than an ActivityPub client request.
func wantsHTML(accept string) bool {
	if strings.Contains(accept, "application/activity+json") ||
		strings.Contains(accept, "application/ld+json") {
		return false
	}
	return strings.Contains(accept, "text/html")
}

func (h *ActivityPubHandler) Followers(c *fiber.Ctx) error {
	if h.svc == nil {
		return c.SendStatus(fiber.StatusNotFound)
	}
	baseURL := resolveActivityPubBaseURL(c)
	collection, err := h.svc.FollowersCollection(c.Context(), baseURL)
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}
	return sendJSONWithContentType(c, fiber.StatusOK, "application/activity+json", collection)
}

func (h *ActivityPubHandler) Outbox(c *fiber.Ctx) error {
	if h.svc == nil {
		return c.SendStatus(fiber.StatusNotFound)
	}
	baseURL := resolveActivityPubBaseURL(c)
	page := parseIntQuery(c, "page", 1)
	size := parseIntQuery(c, "per_page", 20)
	collection, err := h.svc.OutboxCollection(c.Context(), baseURL, page, size)
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}
	return sendJSONWithContentType(c, fiber.StatusOK, "application/activity+json", collection)
}

func (h *ActivityPubHandler) Object(c *fiber.Ctx) error {
	if h.svc == nil {
		return c.SendStatus(fiber.StatusNotFound)
	}
	baseURL := resolveActivityPubBaseURL(c)
	doc, err := h.svc.ObjectDocument(c.Context(), baseURL, c.Params("id"))
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}
	return sendJSONWithContentType(c, fiber.StatusOK, "application/activity+json", doc)
}

func (h *ActivityPubHandler) Inbox(c *fiber.Ctx) error {
	if h.svc == nil {
		return c.SendStatus(fiber.StatusNotFound)
	}
	req, err := parseFederationRequest(c)
	if err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	body := c.Body()
	baseURL := resolveActivityPubBaseURL(c)
	if err := h.svc.HandleInbox(c.Context(), baseURL, req, body); err != nil {
		log.Printf("[activitypub] inbox failed method=%s path=%s remote=%s err=%v", c.Method(), c.Path(), c.IP(), err)
		msg := strings.ToLower(strings.TrimSpace(err.Error()))
		switch {
		case strings.Contains(msg, "signature"):
			return c.SendStatus(fiber.StatusUnauthorized)
		case strings.Contains(msg, "digest"):
			return c.SendStatus(fiber.StatusUnauthorized)
		case strings.Contains(msg, "federation"):
			return c.SendStatus(fiber.StatusNotFound)
		default:
			return c.SendStatus(fiber.StatusBadRequest)
		}
	}
	return c.SendStatus(fiber.StatusAccepted)
}

func (h *ActivityPubHandler) NodeInfoDiscovery(c *fiber.Ctx) error {
	if h.svc == nil {
		return c.SendStatus(fiber.StatusNotFound)
	}
	baseURL := resolveActivityPubBaseURL(c)
	doc, err := h.svc.BuildNodeInfoDiscovery(c.Context(), baseURL)
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}
	return sendJSONWithContentType(c, fiber.StatusOK, "application/json", doc)
}

func (h *ActivityPubHandler) NodeInfo20(c *fiber.Ctx) error {
	if h.svc == nil {
		return c.SendStatus(fiber.StatusNotFound)
	}
	baseURL := resolveActivityPubBaseURL(c)
	doc, err := h.svc.BuildNodeInfo20(c.Context(), baseURL)
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}
	return sendJSONWithContentType(c, fiber.StatusOK, "application/json", doc)
}

func resolveActivityPubBaseURL(c *fiber.Ctx) string {
	scheme := "https"
	if strings.TrimSpace(c.Protocol()) != "" {
		scheme = c.Protocol()
	}
	host := strings.TrimSpace(string(c.Context().Host()))
	if host == "" {
		host = strings.TrimSpace(c.Hostname())
	}
	return strings.TrimRight(scheme+"://"+host, "/")
}

func sendJSONWithContentType(c *fiber.Ctx, status int, contentType string, payload any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	c.Status(status)
	c.Set("Content-Type", contentType)
	return c.Send(raw)
}
