package webhook

import (
	"time"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/article"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/comment"
	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/globalnotification"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/moment"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/page"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/thinking"
)

var AvailableEventNames = appEvent.NamesByChannel(appEvent.ChannelWebhook)

func IsValidEventName(name string) bool {
	for _, item := range AvailableEventNames {
		if item == name {
			return true
		}
	}
	return false
}

func SampleEvent(name string) (appEvent.Event, error) {
	now := time.Now()
	switch name {
	case article.ArticleCreated{}.Name():
		return article.ArticleCreated{ID: 1, AuthorID: 1, Title: "Sample Article", ShortURL: "sample-article", Published: true, At: now}, nil
	case article.ArticleUpdated{}.Name():
		return article.ArticleUpdated{ID: 1, AuthorID: 1, Title: "Sample Article", ShortURL: "sample-article", Published: true, ContentHash: "hash", LeadIn: nil, TOC: nil, Content: "Sample", At: now}, nil
	case article.ArticlePublished{}.Name():
		return article.ArticlePublished{ID: 1, AuthorID: 1, Title: "Sample Article", ShortURL: "sample-article", At: now}, nil
	case article.ArticleUnpublished{}.Name():
		return article.ArticleUnpublished{ID: 1, AuthorID: 1, Title: "Sample Article", ShortURL: "sample-article", At: now}, nil
	case article.ArticleDeleted{}.Name():
		return article.ArticleDeleted{ID: 1, AuthorID: 1, Title: "Sample Article", ShortURL: "sample-article", At: now}, nil
	case moment.MomentCreated{}.Name():
		return moment.MomentCreated{ID: 1, AuthorID: 1, Title: "Sample Moment", ShortURL: "sample-moment", Published: true, At: now}, nil
	case moment.MomentUpdated{}.Name():
		return moment.MomentUpdated{ID: 1, AuthorID: 1, Title: "Sample Moment", ShortURL: "sample-moment", Published: true, ContentHash: "hash", Summary: "Sample", TOC: nil, Content: "Sample", At: now}, nil
	case moment.MomentPublished{}.Name():
		return moment.MomentPublished{ID: 1, AuthorID: 1, Title: "Sample Moment", ShortURL: "sample-moment", At: now}, nil
	case moment.MomentUnpublished{}.Name():
		return moment.MomentUnpublished{ID: 1, AuthorID: 1, Title: "Sample Moment", ShortURL: "sample-moment", At: now}, nil
	case moment.MomentDeleted{}.Name():
		return moment.MomentDeleted{ID: 1, AuthorID: 1, Title: "Sample Moment", ShortURL: "sample-moment", At: now}, nil
	case page.PageCreated{}.Name():
		return page.PageCreated{ID: 1, Title: "Sample Page", ShortURL: "sample-page", Enabled: true, At: now}, nil
	case page.PageUpdated{}.Name():
		return page.PageUpdated{ID: 1, Title: "Sample Page", ShortURL: "sample-page", Enabled: true, ContentHash: "hash", Description: nil, TOC: nil, Content: "Sample", At: now}, nil
	case page.PageDeleted{}.Name():
		return page.PageDeleted{ID: 1, Title: "Sample Page", ShortURL: "sample-page", At: now}, nil
	case globalnotification.Created{}.Name():
		return globalnotification.Created{ID: 1, Content: "Sample Global Notification", PublishAt: now.Add(-time.Hour), ExpireAt: now.Add(24 * time.Hour), AllowClose: true, At: now}, nil
	case globalnotification.Updated{}.Name():
		return globalnotification.Updated{ID: 1, Content: "Sample Global Notification Updated", PublishAt: now.Add(-time.Hour), ExpireAt: now.Add(24 * time.Hour), AllowClose: false, At: now}, nil
	case globalnotification.Deleted{}.Name():
		return globalnotification.Deleted{ID: 1, At: now}, nil

	// --- Comment events ---
	case comment.CommentCreated{}.Name():
		return comment.CommentCreated{ID: 1, AreaID: 42, NickName: "示例访客", Email: "visitor@example.com", Content: "这是一条示例评论", Status: "pending", At: now}, nil
	case "comment.reply":
		return appEvent.Generic{EventName: name, At: now, Payload: map[string]any{
			"ID": int64(2), "ParentID": int64(1), "AreaID": int64(42),
			"ContentType": "article", "ContentTitle": "示例文章标题",
			"viewUrl": "https://example.com/posts/sample-article",
			"ParentContent": "这是被回复的评论内容", "ReplyContent": "这是回复内容",
			"ParentNickName": "示例访客", "ReplyNickName": "博主",
			"recipientEmail": "visitor@example.com", "Status": "approved",
		}}, nil
	case "comment.updated":
		return appEvent.Generic{EventName: name, At: now, Payload: map[string]any{
			"ID": int64(1), "Status": "approved",
		}}, nil
	case "comment.deleted":
		return appEvent.Generic{EventName: name, At: now, Payload: map[string]any{
			"ID": int64(1), "AreaID": int64(42),
		}}, nil
	case "comment.blocked":
		return appEvent.Generic{EventName: name, At: now, Payload: map[string]any{
			"ID": int64(1), "Status": "blocked",
		}}, nil

	// --- Thinking events ---
	case thinking.ThinkingCreated{}.Name():
		return thinking.ThinkingCreated{ID: 1, AuthorID: 1, Content: "这是一条示例思考", At: now}, nil
	case "thinking.updated":
		return appEvent.Generic{EventName: name, At: now, Payload: map[string]any{
			"ID": int64(1), "Content": "这是更新后的思考内容",
		}}, nil
	case "thinking.deleted":
		return appEvent.Generic{EventName: name, At: now, Payload: map[string]any{
			"ID": int64(1),
		}}, nil

	// --- Media events ---
	case "media.uploaded":
		return appEvent.Generic{EventName: name, At: now, Payload: map[string]any{
			"ID": int64(1), "Name": "photo.jpg", "Path": "/uploads/2026/03/photo.jpg", "Type": "image/jpeg",
		}}, nil
	case "media.deleted":
		return appEvent.Generic{EventName: name, At: now, Payload: map[string]any{
			"ID": int64(1), "Name": "photo.jpg", "Path": "/uploads/2026/03/photo.jpg",
		}}, nil

	// --- System events ---
	case "sysconfig.updated":
		return appEvent.Generic{EventName: name, At: now, Payload: map[string]any{
			"Keys": []string{"site.name", "site.description"}, "Count": 2,
		}}, nil
	case "system.monitor.alert":
		return appEvent.Generic{EventName: name, At: now, Payload: map[string]any{
			"DatabaseStatus": "healthy", "RedisStatus": "healthy",
		}}, nil
	case "system.health.changed":
		return appEvent.Generic{EventName: name, At: now, Payload: map[string]any{
			"Prev": uint8(0), "Next": uint8(1), "Mode": "normal", "Maintenance": false,
		}}, nil

	// --- Friend link events ---
	case "friendlink.application.created":
		return appEvent.Generic{EventName: name, At: now, Payload: map[string]any{
			"ID": int64(1), "URL": "https://friend-blog.example.com", "Name": "示例博客", "Status": "pending",
		}}, nil
	case "friendlink.application.approved":
		return appEvent.Generic{EventName: name, At: now, Payload: map[string]any{
			"ID": int64(1), "URL": "https://friend-blog.example.com", "Name": "示例博客", "Status": "approved", "recipientEmail": "friend@example.com",
		}}, nil
	case "friendlink.application.rejected":
		return appEvent.Generic{EventName: name, At: now, Payload: map[string]any{
			"ID": int64(1), "URL": "https://friend-blog.example.com", "Name": "示例博客", "Status": "rejected", "recipientEmail": "friend@example.com",
		}}, nil
	case "friendlink.application.blocked":
		return appEvent.Generic{EventName: name, At: now, Payload: map[string]any{
			"ID": int64(1), "URL": "https://friend-blog.example.com",
		}}, nil

	// --- Federation events ---
	case "federation.friendlink.requested":
		return appEvent.Generic{EventName: name, At: now, Payload: map[string]any{
			"TargetURL": "https://remote-blog.example.com", "StatusCode": 200,
		}}, nil
	case "federation.friendlink.received":
		return appEvent.Generic{EventName: name, At: now, Payload: map[string]any{
			"RequesterURL": "https://remote-blog.example.com", "ApplicationID": int64(1),
		}}, nil
	case "federation.citation.requested":
		return appEvent.Generic{EventName: name, At: now, Payload: map[string]any{
			"TargetInstanceURL": "https://remote-blog.example.com", "TargetPostID": "post-123",
		}}, nil
	case "federation.citation.received":
		return appEvent.Generic{EventName: name, At: now, Payload: map[string]any{
			"CitationID": int64(1), "SourceInstanceURL": "https://remote-blog.example.com", "Status": "pending",
		}}, nil
	case "federation.mention.requested":
		return appEvent.Generic{EventName: name, At: now, Payload: map[string]any{
			"TargetInstanceURL": "https://remote-blog.example.com", "MentionedUser": "admin",
		}}, nil
	case "federation.mention.received":
		return appEvent.Generic{EventName: name, At: now, Payload: map[string]any{
			"MentionID": int64(1), "SourceInstanceURL": "https://remote-blog.example.com", "MentionedUser": "admin",
		}}, nil

	default:
		return appEvent.Generic{
			EventName: name,
			At:        now,
			Payload: map[string]any{
				"eventName": name,
				"sample":    true,
			},
		}, nil
	}
}
