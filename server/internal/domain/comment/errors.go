package comment

import "errors"

var ErrCommentNotFound = errors.New("评论不存在或已被删除")
var ErrCommentParentNotFound = errors.New("评论的父评论不存在")
var ErrCommentAreaNotFound = errors.New("评论区不存在")
var ErrCommentAreaClosed = errors.New("评论区已关闭")
var ErrCommentTooDeep = errors.New("评论层级过深")
var ErrCommentContentEmpty = errors.New("评论内容不能为空")
var ErrCommentContentTooLong = errors.New("评论内容过长")
var ErrCommentDisabled = errors.New("评论功能已关闭")
var ErrCommentBlocked = errors.New("评论身份已被封禁")
var ErrCommentStatusInvalid = errors.New("评论状态无效")
var ErrVisitorNotFound = errors.New("访客不存在")
var ErrCommentReplyDisabled = errors.New("该评论不允许本地回复")
