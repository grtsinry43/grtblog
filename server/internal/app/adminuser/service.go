package adminuser

import (
	"context"
	"errors"
	"strings"

	"github.com/grtsinry43/grtblog-v2/server/internal/domain/identity"
)

var ErrLastAdminMutation = errors.New("至少保留一个可用管理员账号")
var ErrSelfAdminMutation = errors.New("当前登录用户不能修改自己的状态")

type Service struct {
	repo identity.Repository
}

func NewService(repo identity.Repository) *Service {
	return &Service{repo: repo}
}

type ListUsersCmd struct {
	Keyword    string
	OnlyAdmin  *bool
	OnlyActive *bool
	Page       int
	PageSize   int
}

type UpdateUserCmd struct {
	OperatorID int64
	UserID     int64
	Nickname   string
	Email      string
	IsActive   bool
	IsAdmin    bool
}

func (s *Service) ListUsers(ctx context.Context, cmd ListUsersCmd) ([]identity.User, int64, error) {
	return s.repo.ListUsers(ctx, identity.UserListOptions{
		Keyword:    strings.TrimSpace(cmd.Keyword),
		OnlyAdmin:  cmd.OnlyAdmin,
		OnlyActive: cmd.OnlyActive,
		Page:       cmd.Page,
		PageSize:   cmd.PageSize,
	})
}

func (s *Service) UpdateUser(ctx context.Context, cmd UpdateUserCmd) (*identity.User, error) {
	if cmd.UserID <= 0 {
		return nil, identity.ErrUserNotFound
	}

	current, err := s.repo.FindByID(ctx, cmd.UserID)
	if err != nil {
		return nil, err
	}

	nickname := strings.TrimSpace(cmd.Nickname)
	if nickname == "" {
		nickname = current.Nickname
	}
	email := strings.TrimSpace(cmd.Email)

	if current.IsAdmin && (!cmd.IsActive || !cmd.IsAdmin) {
		activeAdmins, err := s.repo.CountActiveAdmins(ctx)
		if err != nil {
			return nil, err
		}
		if activeAdmins <= 1 {
			return nil, ErrLastAdminMutation
		}
	}

	if cmd.UserID == cmd.OperatorID && (cmd.IsActive != current.IsActive || cmd.IsAdmin != current.IsAdmin) {
		return nil, ErrSelfAdminMutation
	}

	updated, err := s.repo.UpdateAdminUser(ctx, cmd.UserID, nickname, email, cmd.IsActive, cmd.IsAdmin)
	if err != nil {
		return nil, err
	}
	updated.Password = ""
	return updated, nil
}
