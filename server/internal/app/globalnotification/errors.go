package globalnotification

import "errors"

var ErrContentRequired = errors.New("通知内容不能为空")
var ErrInvalidPublishWindow = errors.New("发布时间必须早于过期时间")
var ErrInvalidNotificationID = errors.New("无效的通知ID")
