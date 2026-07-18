package handler

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"

	backupapp "github.com/grtsinry43/grtblog-v2/server/internal/app/backup"
	backupdomain "github.com/grtsinry43/grtblog-v2/server/internal/domain/backup"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type BackupHandler struct{ svc *backupapp.Service }

func NewBackupHandler(svc *backupapp.Service) *BackupHandler { return &BackupHandler{svc: svc} }

func (h *BackupHandler) List(c *fiber.Ctx) error {
	items, err := h.svc.List(c.UserContext())
	if err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "获取备份列表失败", err)
	}
	return response.Success(c, items)
}

func (h *BackupHandler) Create(c *fiber.Ctx) error {
	item, err := h.svc.CreateManual(c.UserContext())
	if err != nil {
		if errors.Is(err, backupdomain.ErrBackupRunning) {
			return response.NewBizErrorWithMsg(response.ParamsError, "已有备份任务正在运行")
		}
		return response.NewBizErrorWithCause(response.ServerError, "创建备份失败", err)
	}
	return response.SuccessWithMessage(c, item, "备份任务已创建")
}

func (h *BackupHandler) Get(c *fiber.Ctx) error {
	item, err := h.svc.Get(c.UserContext(), c.Params("id"))
	if err != nil {
		return mapBackupError(err, "获取备份详情失败")
	}
	return response.Success(c, item)
}

func (h *BackupHandler) Delete(c *fiber.Ctx) error {
	if err := h.svc.Delete(c.UserContext(), c.Params("id")); err != nil {
		return mapBackupError(err, "删除备份失败")
	}
	return response.SuccessWithMessage(c, fiber.Map{"id": c.Params("id")}, "备份已删除")
}

func (h *BackupHandler) IssueDownloadTicket(c *fiber.Ctx) error {
	token, expiresAt, err := h.svc.IssueDownloadTicket(c.UserContext(), c.Params("id"))
	if err != nil {
		return mapBackupError(err, "生成下载链接失败")
	}
	path := "/api/v2/backups/download?ticket=" + url.QueryEscape(token)
	return response.Success(c, fiber.Map{"url": path, "expiresAt": expiresAt})
}

func (h *BackupHandler) Download(c *fiber.Ctx) error {
	token := strings.TrimSpace(c.Query("ticket"))
	if token == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "下载凭证不能为空")
	}
	item, path, err := h.svc.ResolveDownload(c.UserContext(), token)
	if err != nil {
		if errors.Is(err, backupdomain.ErrInvalidTicket) {
			return response.NewBizErrorWithMsg(response.Unauthorized, "下载链接无效或已过期")
		}
		return response.NewBizErrorWithCause(response.ServerError, "读取备份文件失败", err)
	}
	c.Set(fiber.HeaderContentType, "application/gzip")
	c.Set(fiber.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%q", item.Filename))
	c.Set(fiber.HeaderCacheControl, "private, no-store")
	return c.SendFile(path)
}

func mapBackupError(err error, message string) error {
	switch {
	case errors.Is(err, backupdomain.ErrNotFound):
		return response.NewBizErrorWithMsg(response.NotFound, "备份不存在")
	case errors.Is(err, backupdomain.ErrBackupRunning):
		return response.NewBizErrorWithMsg(response.ParamsError, "备份任务仍在运行")
	default:
		return response.NewBizErrorWithCause(response.ServerError, message, err)
	}
}
