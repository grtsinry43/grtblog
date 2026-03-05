package event

import "strings"

const (
	ChannelEmail   = "email"
	ChannelWebhook = "webhook"
)

type EventField struct {
	Name        string
	Type        string
	Required    bool
	Description string
}

type EventDescriptor struct {
	Name        string
	Title       string
	Category    string
	Description string
	PublicEmail bool
	Channels    []string
	Fields      []EventField
}

type EventGroup struct {
	Category string
	Events   []string
}

var catalog = []EventDescriptor{
	{Name: "article.created", Title: "文章创建", Category: "content", Description: "创建文章后触发", Channels: []string{ChannelEmail, ChannelWebhook}, Fields: []EventField{{Name: "ID", Type: "int64", Required: true, Description: "文章ID"}, {Name: "Title", Type: "string", Required: true, Description: "标题"}, {Name: "ShortURL", Type: "string", Required: true, Description: "短链"}, {Name: "Published", Type: "bool", Required: true, Description: "是否发布"}}},
	{Name: "article.updated", Title: "文章更新", Category: "content", Description: "更新文章后触发", Channels: []string{ChannelEmail, ChannelWebhook}, Fields: []EventField{{Name: "ID", Type: "int64", Required: true, Description: "文章ID"}, {Name: "Title", Type: "string", Required: true, Description: "标题"}, {Name: "ShortURL", Type: "string", Required: true, Description: "短链"}, {Name: "Published", Type: "bool", Required: true, Description: "是否发布"}}},
	{Name: "article.published", Title: "文章发布", Category: "content", Description: "文章发布后触发", PublicEmail: true, Channels: []string{ChannelEmail, ChannelWebhook}, Fields: []EventField{{Name: "ID", Type: "int64", Required: true, Description: "文章ID"}, {Name: "Title", Type: "string", Required: true, Description: "标题"}, {Name: "ShortURL", Type: "string", Required: true, Description: "短链"}}},
	{Name: "article.unpublished", Title: "文章下线", Category: "content", Description: "文章下线后触发", Channels: []string{ChannelEmail, ChannelWebhook}, Fields: []EventField{{Name: "ID", Type: "int64", Required: true, Description: "文章ID"}, {Name: "Title", Type: "string", Required: true, Description: "标题"}, {Name: "ShortURL", Type: "string", Required: true, Description: "短链"}}},
	{Name: "article.deleted", Title: "文章删除", Category: "content", Description: "删除文章后触发", Channels: []string{ChannelEmail, ChannelWebhook}, Fields: []EventField{{Name: "ID", Type: "int64", Required: true, Description: "文章ID"}, {Name: "Title", Type: "string", Required: true, Description: "标题"}, {Name: "ShortURL", Type: "string", Required: true, Description: "短链"}}},
	{Name: "moment.created", Title: "手记创建", Category: "content", Description: "创建手记后触发", Channels: []string{ChannelEmail, ChannelWebhook}, Fields: []EventField{{Name: "ID", Type: "int64", Required: true, Description: "手记ID"}, {Name: "Title", Type: "string", Required: true, Description: "标题"}, {Name: "ShortURL", Type: "string", Required: true, Description: "短链"}, {Name: "Published", Type: "bool", Required: true, Description: "是否发布"}}},
	{Name: "moment.updated", Title: "手记更新", Category: "content", Description: "更新手记后触发", Channels: []string{ChannelEmail, ChannelWebhook}, Fields: []EventField{{Name: "ID", Type: "int64", Required: true, Description: "手记ID"}, {Name: "Title", Type: "string", Required: true, Description: "标题"}, {Name: "ShortURL", Type: "string", Required: true, Description: "短链"}}},
	{Name: "moment.published", Title: "手记发布", Category: "content", Description: "手记发布后触发", PublicEmail: true, Channels: []string{ChannelEmail, ChannelWebhook}, Fields: []EventField{{Name: "ID", Type: "int64", Required: true, Description: "手记ID"}, {Name: "Title", Type: "string", Required: true, Description: "标题"}, {Name: "ShortURL", Type: "string", Required: true, Description: "短链"}}},
	{Name: "moment.unpublished", Title: "手记下线", Category: "content", Description: "手记下线后触发", Channels: []string{ChannelEmail, ChannelWebhook}, Fields: []EventField{{Name: "ID", Type: "int64", Required: true, Description: "手记ID"}, {Name: "Title", Type: "string", Required: true, Description: "标题"}, {Name: "ShortURL", Type: "string", Required: true, Description: "短链"}}},
	{Name: "moment.deleted", Title: "手记删除", Category: "content", Description: "删除手记后触发", Channels: []string{ChannelEmail, ChannelWebhook}, Fields: []EventField{{Name: "ID", Type: "int64", Required: true, Description: "手记ID"}, {Name: "Title", Type: "string", Required: true, Description: "标题"}}},
	{Name: "page.created", Title: "页面创建", Category: "content", Description: "创建页面后触发", Channels: []string{ChannelEmail, ChannelWebhook}, Fields: []EventField{{Name: "ID", Type: "int64", Required: true, Description: "页面ID"}, {Name: "Title", Type: "string", Required: true, Description: "标题"}, {Name: "ShortURL", Type: "string", Required: true, Description: "短链"}}},
	{Name: "page.updated", Title: "页面更新", Category: "content", Description: "更新页面后触发", Channels: []string{ChannelEmail, ChannelWebhook}, Fields: []EventField{{Name: "ID", Type: "int64", Required: true, Description: "页面ID"}, {Name: "Title", Type: "string", Required: true, Description: "标题"}, {Name: "ShortURL", Type: "string", Required: true, Description: "短链"}}},
	{Name: "page.deleted", Title: "页面删除", Category: "content", Description: "删除页面后触发", Channels: []string{ChannelEmail, ChannelWebhook}, Fields: []EventField{{Name: "ID", Type: "int64", Required: true, Description: "页面ID"}, {Name: "Title", Type: "string", Required: true, Description: "标题"}}},
	{Name: "thinking.created", Title: "思考创建", Category: "content", Description: "创建思考后触发", PublicEmail: true, Channels: []string{ChannelEmail, ChannelWebhook}, Fields: []EventField{{Name: "ID", Type: "int64", Required: true, Description: "思考ID"}, {Name: "Content", Type: "string", Required: true, Description: "内容"}}},
	{Name: "thinking.updated", Title: "思考更新", Category: "content", Description: "更新思考后触发", Channels: []string{ChannelEmail, ChannelWebhook}, Fields: []EventField{{Name: "ID", Type: "int64", Required: true, Description: "思考ID"}, {Name: "Content", Type: "string", Required: true, Description: "内容"}}},
	{Name: "thinking.deleted", Title: "思考删除", Category: "content", Description: "删除思考后触发", Channels: []string{ChannelEmail, ChannelWebhook}, Fields: []EventField{{Name: "ID", Type: "int64", Required: true, Description: "思考ID"}}},
	{Name: "comment.created", Title: "评论创建", Category: "comment", Description: "新评论创建后触发", Channels: []string{ChannelEmail, ChannelWebhook}, Fields: []EventField{{Name: "ID", Type: "int64", Required: true, Description: "评论ID"}, {Name: "AreaID", Type: "int64", Required: true, Description: "评论区ID"}, {Name: "Content", Type: "string", Required: true, Description: "内容"}, {Name: "Status", Type: "string", Required: true, Description: "状态"}}},
	{Name: "comment.reply", Title: "评论回复", Category: "comment", Description: "管理员回复评论后触发", Channels: []string{ChannelEmail, ChannelWebhook}, Fields: []EventField{{Name: "ID", Type: "int64", Required: true, Description: "回复评论ID"}, {Name: "ParentID", Type: "int64", Required: true, Description: "被回复评论ID"}, {Name: "AreaID", Type: "int64", Required: true, Description: "评论区ID"}, {Name: "ContentType", Type: "string", Required: false, Description: "所在内容类型（文章/手记/页面/思考）"}, {Name: "ContentTitle", Type: "string", Required: false, Description: "所在内容标题"}, {Name: "viewUrl", Type: "string", Required: false, Description: "查看原文链接（已按前端路由拼接）"}, {Name: "ParentContent", Type: "string", Required: false, Description: "被回复评论内容"}, {Name: "ReplyContent", Type: "string", Required: true, Description: "回复内容"}, {Name: "ParentNickName", Type: "string", Required: false, Description: "被回复昵称"}, {Name: "ReplyNickName", Type: "string", Required: false, Description: "回复人昵称"}, {Name: "recipientEmail", Type: "string", Required: false, Description: "目标收件邮箱"}}},
	{Name: "comment.updated", Title: "评论更新", Category: "comment", Description: "评论状态或属性更新后触发", Channels: []string{ChannelEmail, ChannelWebhook}, Fields: []EventField{{Name: "ID", Type: "int64", Required: true, Description: "评论ID"}, {Name: "Status", Type: "string", Required: true, Description: "状态"}}},
	{Name: "comment.deleted", Title: "评论删除", Category: "comment", Description: "评论删除后触发", Channels: []string{ChannelEmail, ChannelWebhook}, Fields: []EventField{{Name: "ID", Type: "int64", Required: true, Description: "评论ID"}, {Name: "AreaID", Type: "int64", Required: true, Description: "评论区ID"}}},
	{Name: "comment.blocked", Title: "评论屏蔽", Category: "comment", Description: "评论被屏蔽后触发", Channels: []string{ChannelEmail, ChannelWebhook}, Fields: []EventField{{Name: "ID", Type: "int64", Required: true, Description: "评论ID"}, {Name: "Status", Type: "string", Required: true, Description: "状态"}}},
	{Name: "media.uploaded", Title: "文件上传", Category: "media", Description: "上传文件后触发", Channels: []string{ChannelEmail, ChannelWebhook}, Fields: []EventField{{Name: "ID", Type: "int64", Required: true, Description: "文件ID"}, {Name: "Name", Type: "string", Required: true, Description: "文件名"}, {Name: "Path", Type: "string", Required: true, Description: "路径"}, {Name: "Type", Type: "string", Required: true, Description: "类型"}}},
	{Name: "media.deleted", Title: "文件删除", Category: "media", Description: "删除文件后触发", Channels: []string{ChannelEmail, ChannelWebhook}, Fields: []EventField{{Name: "ID", Type: "int64", Required: true, Description: "文件ID"}, {Name: "Name", Type: "string", Required: true, Description: "文件名"}, {Name: "Path", Type: "string", Required: true, Description: "路径"}}},
	{Name: "sysconfig.updated", Title: "系统配置更新", Category: "system", Description: "批量更新系统配置后触发", Channels: []string{ChannelEmail, ChannelWebhook}, Fields: []EventField{{Name: "Keys", Type: "[]string", Required: true, Description: "更新的配置键"}, {Name: "Count", Type: "int", Required: true, Description: "更新数量"}}},
	// websiteinfo.updated is now handled by sysconfig.updated with site.* prefix keys
	{Name: "global.notification.created", Title: "全站通知创建", Category: "system", Description: "创建全站通知后触发", Channels: []string{ChannelEmail, ChannelWebhook}, Fields: []EventField{{Name: "ID", Type: "int64", Required: true, Description: "通知ID"}, {Name: "Content", Type: "string", Required: true, Description: "通知内容"}, {Name: "PublishAt", Type: "time.Time", Required: true, Description: "生效时间"}, {Name: "ExpireAt", Type: "time.Time", Required: true, Description: "过期时间"}, {Name: "AllowClose", Type: "bool", Required: true, Description: "是否允许关闭"}}},
	{Name: "global.notification.updated", Title: "全站通知更新", Category: "system", Description: "更新全站通知后触发", Channels: []string{ChannelEmail, ChannelWebhook}, Fields: []EventField{{Name: "ID", Type: "int64", Required: true, Description: "通知ID"}, {Name: "Content", Type: "string", Required: true, Description: "通知内容"}, {Name: "PublishAt", Type: "time.Time", Required: true, Description: "生效时间"}, {Name: "ExpireAt", Type: "time.Time", Required: true, Description: "过期时间"}, {Name: "AllowClose", Type: "bool", Required: true, Description: "是否允许关闭"}}},
	{Name: "global.notification.deleted", Title: "全站通知删除", Category: "system", Description: "删除全站通知后触发", Channels: []string{ChannelEmail, ChannelWebhook}, Fields: []EventField{{Name: "ID", Type: "int64", Required: true, Description: "通知ID"}}},
	{Name: "friendlink.application.created", Title: "友链申请创建", Category: "friendlink", Description: "友链申请创建或重提后触发", Channels: []string{ChannelEmail, ChannelWebhook}, Fields: []EventField{{Name: "ID", Type: "int64", Required: true, Description: "申请ID"}, {Name: "URL", Type: "string", Required: true, Description: "站点URL"}, {Name: "Status", Type: "string", Required: true, Description: "状态"}}},
	{Name: "friendlink.application.approved", Title: "友链申请通过", Category: "friendlink", Description: "友链申请通过后触发", Channels: []string{ChannelEmail, ChannelWebhook}, Fields: []EventField{{Name: "ID", Type: "int64", Required: true, Description: "申请ID"}, {Name: "URL", Type: "string", Required: true, Description: "站点URL"}, {Name: "Status", Type: "string", Required: true, Description: "申请状态"}, {Name: "Name", Type: "string", Required: false, Description: "申请名称"}, {Name: "recipientEmail", Type: "string", Required: false, Description: "目标收件邮箱"}}},
	{Name: "friendlink.application.rejected", Title: "友链申请拒绝", Category: "friendlink", Description: "友链申请拒绝后触发", Channels: []string{ChannelEmail, ChannelWebhook}, Fields: []EventField{{Name: "ID", Type: "int64", Required: true, Description: "申请ID"}, {Name: "URL", Type: "string", Required: true, Description: "站点URL"}, {Name: "Status", Type: "string", Required: true, Description: "申请状态"}, {Name: "Name", Type: "string", Required: false, Description: "申请名称"}, {Name: "recipientEmail", Type: "string", Required: false, Description: "目标收件邮箱"}}},
	{Name: "friendlink.application.blocked", Title: "友链申请封禁", Category: "friendlink", Description: "友链申请封禁后触发", Channels: []string{ChannelEmail, ChannelWebhook}, Fields: []EventField{{Name: "ID", Type: "int64", Required: true, Description: "申请ID"}, {Name: "URL", Type: "string", Required: true, Description: "站点URL"}}},
	{Name: "federation.friendlink.requested", Title: "联邦友链触发", Category: "federation", Description: "本地向远端发起友链请求", Channels: []string{ChannelEmail, ChannelWebhook}, Fields: []EventField{{Name: "TargetURL", Type: "string", Required: true, Description: "远端地址"}, {Name: "StatusCode", Type: "int", Required: false, Description: "响应状态"}}},
	{Name: "federation.friendlink.received", Title: "联邦友链收到", Category: "federation", Description: "收到远端友链请求", Channels: []string{ChannelEmail, ChannelWebhook}, Fields: []EventField{{Name: "RequesterURL", Type: "string", Required: true, Description: "请求方"}, {Name: "ApplicationID", Type: "int64", Required: true, Description: "申请ID"}}},
	{Name: "federation.citation.requested", Title: "联邦引用触发", Category: "federation", Description: "本地向远端发起引用请求", Channels: []string{ChannelEmail, ChannelWebhook}, Fields: []EventField{{Name: "TargetInstanceURL", Type: "string", Required: true, Description: "远端地址"}, {Name: "TargetPostID", Type: "string", Required: true, Description: "目标文章ID"}}},
	{Name: "federation.citation.received", Title: "联邦引用收到", Category: "federation", Description: "收到远端引用请求", Channels: []string{ChannelEmail, ChannelWebhook}, Fields: []EventField{{Name: "CitationID", Type: "int64", Required: true, Description: "引用ID"}, {Name: "SourceInstanceURL", Type: "string", Required: true, Description: "来源实例"}, {Name: "Status", Type: "string", Required: true, Description: "状态"}}},
	{Name: "federation.mention.requested", Title: "联邦提及触发", Category: "federation", Description: "本地向远端发起提及通知", Channels: []string{ChannelEmail, ChannelWebhook}, Fields: []EventField{{Name: "TargetInstanceURL", Type: "string", Required: true, Description: "远端地址"}, {Name: "MentionedUser", Type: "string", Required: true, Description: "提及用户"}}},
	{Name: "federation.mention.received", Title: "联邦提及收到", Category: "federation", Description: "收到远端提及通知", Channels: []string{ChannelEmail, ChannelWebhook}, Fields: []EventField{{Name: "MentionID", Type: "int64", Required: true, Description: "提及ID"}, {Name: "SourceInstanceURL", Type: "string", Required: true, Description: "来源实例"}, {Name: "MentionedUser", Type: "string", Required: true, Description: "被提及用户"}}},
	{Name: "system.monitor.alert", Title: "系统监控异常", Category: "system", Description: "系统状态检测到异常", Channels: []string{ChannelEmail, ChannelWebhook}, Fields: []EventField{{Name: "DatabaseStatus", Type: "string", Required: true, Description: "数据库状态"}, {Name: "RedisStatus", Type: "string", Required: true, Description: "Redis状态"}}},
	{Name: "system.health.changed", Title: "健康状态变更", Category: "system", Description: "系统健康状态位变更时触发", Channels: []string{ChannelWebhook}, Fields: []EventField{{Name: "Prev", Type: "uint8", Required: true, Description: "变更前状态位"}, {Name: "Next", Type: "uint8", Required: true, Description: "变更后状态位"}, {Name: "Mode", Type: "string", Required: true, Description: "当前模式"}, {Name: "Maintenance", Type: "bool", Required: true, Description: "手动维护模式"}}},
}

func Catalog() []EventDescriptor {
	return append([]EventDescriptor(nil), catalog...)
}

func CatalogByName(name string) (EventDescriptor, bool) {
	name = strings.TrimSpace(name)
	for _, item := range catalog {
		if item.Name == name {
			return item, true
		}
	}
	return EventDescriptor{}, false
}

func NamesByChannel(channel string) []string {
	channel = strings.TrimSpace(channel)
	result := make([]string, 0, len(catalog))
	for _, item := range catalog {
		if hasChannel(item, channel) {
			result = append(result, item.Name)
		}
	}
	return result
}

func GroupsByChannel(channel string) []EventGroup {
	channel = strings.TrimSpace(channel)
	order := make([]string, 0)
	index := make(map[string]int)
	for _, item := range catalog {
		if !hasChannel(item, channel) {
			continue
		}
		if _, ok := index[item.Category]; !ok {
			index[item.Category] = len(order)
			order = append(order, item.Category)
		}
	}
	groups := make([]EventGroup, len(order))
	for i, category := range order {
		groups[i] = EventGroup{Category: category, Events: []string{}}
	}
	for _, item := range catalog {
		if !hasChannel(item, channel) {
			continue
		}
		idx := index[item.Category]
		groups[idx].Events = append(groups[idx].Events, item.Name)
	}
	return groups
}

func hasChannel(item EventDescriptor, channel string) bool {
	if channel == "" {
		return true
	}
	for _, c := range item.Channels {
		if c == channel {
			return true
		}
	}
	return false
}
