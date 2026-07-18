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

type updateBackupScheduleRequest struct {
	Enabled        bool `json:"enabled"`
	IntervalHours  int  `json:"intervalHours"`
	RetentionCount int  `json:"retentionCount"`
}

type updateBackupPinRequest struct {
	Pinned bool `json:"pinned"`
}

type requestBackupRestore struct {
	Confirmation string `json:"confirmation"`
}

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

func (h *BackupHandler) GetSchedule(c *fiber.Ctx) error {
	schedule, err := h.svc.GetSchedule(c.UserContext())
	if err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "获取备份计划失败", err)
	}
	return response.Success(c, schedule)
}

func (h *BackupHandler) UpdateSchedule(c *fiber.Ctx) error {
	var req updateBackupScheduleRequest
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	schedule, err := h.svc.UpdateSchedule(c.UserContext(), req.Enabled, req.IntervalHours, req.RetentionCount)
	if err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "备份计划参数无效", err)
	}
	return response.SuccessWithMessage(c, schedule, "备份计划已更新")
}

func (h *BackupHandler) SetPinned(c *fiber.Ctx) error {
	var req updateBackupPinRequest
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if err := h.svc.SetPinned(c.UserContext(), c.Params("id"), req.Pinned); err != nil {
		return mapBackupError(err, "更新备份固定状态失败")
	}
	return response.SuccessWithMessage(c, fiber.Map{"id": c.Params("id"), "pinned": req.Pinned}, "备份已更新")
}

func (h *BackupHandler) GetRestoreStatus(c *fiber.Ctx) error {
	status, err := h.svc.GetRestoreStatus()
	if err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "获取恢复状态失败", err)
	}
	return response.Success(c, status)
}

func (h *BackupHandler) RequestRestore(c *fiber.Ctx) error {
	var req requestBackupRestore
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	status, err := h.svc.RequestRestore(c.UserContext(), c.Params("id"), req.Confirmation)
	if err != nil {
		return mapBackupError(err, "创建恢复请求失败")
	}
	return response.SuccessWithMessage(c, status, "恢复请求已创建，服务即将重启")
}

func (h *BackupHandler) UploadAndRequestRestore(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("archive")
	if err != nil || fileHeader == nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "请选择 tar.gz 备份文件")
	}
	file, err := fileHeader.Open()
	if err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "读取备份文件失败", err)
	}
	defer file.Close()
	item, status, err := h.svc.ImportAndRequestRestore(c.UserContext(), file, c.FormValue("confirmation"))
	if err != nil {
		return mapBackupError(err, "导入备份文件失败")
	}
	return response.SuccessWithMessage(c, fiber.Map{"backup": item, "restore": status}, "备份已校验，服务即将重启")
}

func mapBackupError(err error, message string) error {
	switch {
	case errors.Is(err, backupdomain.ErrNotFound):
		return response.NewBizErrorWithMsg(response.NotFound, "备份不存在")
	case errors.Is(err, backupdomain.ErrBackupRunning):
		return response.NewBizErrorWithMsg(response.ParamsError, "备份任务仍在运行")
	case errors.Is(err, backupdomain.ErrRestorePending):
		return response.NewBizErrorWithMsg(response.ParamsError, "已有全站恢复等待执行")
	case errors.Is(err, backupdomain.ErrRestoreConfirmation):
		return response.NewBizErrorWithMsg(response.ParamsError, "请输入 OVERWRITE 确认覆盖恢复")
	default:
		return response.NewBizErrorWithCause(response.ServerError, message, err)
	}
}
