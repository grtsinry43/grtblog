package handler

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/gamelookup"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type GameLookupHandler struct {
	svc *gamelookup.Service
}

func NewGameLookupHandler(svc *gamelookup.Service) *GameLookupHandler {
	return &GameLookupHandler{svc: svc}
}

// Lookup godoc
// @Summary 按名称查询游戏元数据（Steam 兜底源）
// @Tags OwnerStatus
// @Produce json
// @Param name query string true "游戏名称"
// @Success 200 {object} contract.GenericMessageEnvelope
// @Router /public/game-lookup [get]
func (h *GameLookupHandler) Lookup(c *fiber.Ctx) error {
	if h == nil || h.svc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "游戏查询服务未初始化")
	}
	name := strings.TrimSpace(c.Query("name"))
	if name == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "name 不能为空")
	}
	info, err := h.svc.Lookup(c.UserContext(), name)
	if err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "游戏查询失败", err)
	}
	return response.Success(c, info)
}
