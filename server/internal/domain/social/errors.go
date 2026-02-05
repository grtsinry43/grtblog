package social

import "errors"

var ErrFriendLinkApplicationNotFound = errors.New("友链申请不存在")
var ErrFriendLinkNotFound = errors.New("友链不存在")
var ErrFriendLinkApplicationBlocked = errors.New("友链申请已被封禁")
var ErrGlobalNotificationNotFound = errors.New("全站通知不存在")
var ErrAdminNotificationNotFound = errors.New("站内通知不存在")
