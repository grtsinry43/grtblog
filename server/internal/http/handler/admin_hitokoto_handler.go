package handler

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/hitokoto"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type AdminHitokotoHandler struct {
	svc *hitokoto.Service
}

func NewAdminHitokotoHandler(svc *hitokoto.Service) *AdminHitokotoHandler {
	return &AdminHitokotoHandler{svc: svc}
}

type AdminHitokotoEnvelope struct {
	Code   int             `json:"code"`
	BizErr string          `json:"bizErr"`
	Msg    string          `json:"msg"`
	Data   hitokoto.Result `json:"data"`
	Meta   response.Meta   `json:"meta"`
}

// GetSentence godoc
// @Summary 获取一言（管理员）
// @Description 从 hitokoto 官方接口获取随机一言，并使用 Redis 做短期缓存。
// @Tags Admin-Stats
// @Produce json
// @Param c query string false "分类，支持单个或逗号分隔多个(a-l)"
// @Param min_length query int false "最小长度"
// @Param max_length query int false "最大长度"
// @Param charset query string false "返回编码"
// @Success 200 {object} AdminHitokotoEnvelope
// @Security BearerAuth
// @Router /admin/hitokoto [get]
func (h *AdminHitokotoHandler) GetSentence(c *fiber.Ctx) error {
	cats := parseCategories(c.Query("c"))
	res, err := h.svc.GetSentence(c.UserContext(), hitokoto.Query{
		Categories: cats,
		MinLength:  c.Query("min_length"),
		MaxLength:  c.Query("max_length"),
		Charset:    c.Query("charset"),
	})
	if err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "获取一言失败", err)
	}
	return response.Success(c, res)
}

func parseCategories(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}
