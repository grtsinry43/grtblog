package handler

import (
	"github.com/gofiber/fiber/v2"

	backupapp "github.com/grtsinry43/grtblog-v2/server/internal/app/backup"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/setupstate"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type SetupRestoreHandler struct {
	setup  *setupstate.Service
	backup *backupapp.Service
}

func NewSetupRestoreHandler(setup *setupstate.Service, backup *backupapp.Service) *SetupRestoreHandler {
	return &SetupRestoreHandler{setup: setup, backup: backup}
}

func (h *SetupRestoreHandler) Upload(c *fiber.Ctx) error {
	state, err := h.setup.Evaluate(c.UserContext())
	if err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "检查初始化状态失败", err)
	}
	if state.HasUser || state.HasAdmin {
		return response.NewBizErrorWithMsg(response.Unauthorized, "站点已存在用户，请登录后从设置中恢复")
	}
	fileHeader, err := c.FormFile("archive")
	if err != nil || fileHeader == nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "请选择 tar.gz 备份文件")
	}
	file, err := fileHeader.Open()
	if err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "读取备份文件失败", err)
	}
	defer file.Close()
	item, status, err := h.backup.ImportAndRequestRestore(c.UserContext(), file, c.FormValue("confirmation"))
	if err != nil {
		return mapBackupError(err, "导入初始化备份失败")
	}
	return response.SuccessWithMessage(c, fiber.Map{"backup": item, "restore": status}, "备份已校验，服务即将重启恢复")
}
