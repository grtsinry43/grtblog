package identity

import "errors"

var (
	ErrUserExists         = errors.New("用户已存在")
	ErrUserNotFound       = errors.New("用户不存在")
	ErrInvalidCredentials = errors.New("用户名或密码不正确")
	ErrAdminTokenNotFound = errors.New("管理员令牌不存在")
	ErrAdminTokenExpired  = errors.New("管理员令牌已过期")
)
