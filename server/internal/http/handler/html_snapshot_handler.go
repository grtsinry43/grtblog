package handler

import (
	"log"

	"github.com/gofiber/fiber/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/htmlsnapshot"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/isr"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type HTMLSnapshotHandler struct {
	service *htmlsnapshot.Service
	isr     *isr.Service
}

func NewHTMLSnapshotHandler(service *htmlsnapshot.Service, isrSvc *isr.Service) *HTMLSnapshotHandler {
	return &HTMLSnapshotHandler{
		service: service,
		isr:     isrSvc,
	}
}

// RefreshPostsHTML godoc
// @Summary 触发 ISR bootstrap（兼容 refresh 路径）
// @Tags Public
// @Produce json
// @Success 200 {object} any
// @Router /public/html/posts/refresh [post]
func (h *HTMLSnapshotHandler) RefreshPostsHTML(c *fiber.Ctx) error {
	if h.isr != nil {
		if _, err := h.isr.Bootstrap(c.UserContext()); err != nil {
			log.Printf("[isr] bootstrap failed: %v", err)
		}
		return response.SuccessWithMessage[any](c, nil, "ok")
	}

	if h.service == nil {
		return response.SuccessWithMessage[any](c, nil, "service not initialized")
	}
	if err := h.service.RefreshPostsHTML(c.UserContext()); err != nil {
		log.Printf("[html-snapshot] generate posts html failed: %v", err)
	}

	return response.SuccessWithMessage[any](c, nil, "ok")
}
