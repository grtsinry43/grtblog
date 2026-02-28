package email

import "errors"

var ErrEmailTemplateNotFound = errors.New("邮件模板不存在")
var ErrEmailTemplateCodeExists = errors.New("邮件模板编码已存在")
var ErrEmailTemplateEventInvalid = errors.New("邮件模板事件无效")
var ErrEmailTemplateRenderFailed = errors.New("邮件模板渲染失败")
var ErrEmailTemplateInternalLocked = errors.New("内置邮件模板不允许删除")
var ErrEmailNoRecipient = errors.New("邮件收件人为空")
var ErrEmailDisabled = errors.New("邮件发送未启用")
var ErrEmailConfigInvalid = errors.New("邮件配置无效")
var ErrEmailSendFailed = errors.New("邮件发送失败")
var ErrEmailSubscriptionInvalid = errors.New("邮件订阅参数无效")
var ErrEmailSubscriptionEventInvalid = errors.New("邮件订阅事件无效")
var ErrEmailSubscriptionNotFound = errors.New("邮件订阅不存在")
var ErrEmailSubscriptionStatusInvalid = errors.New("邮件订阅状态无效")
var ErrEmailOutboxNotFound = errors.New("邮件出站记录不存在")
