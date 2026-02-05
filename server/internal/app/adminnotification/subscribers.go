package adminnotification

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	appfed "github.com/grtsinry43/grtblog-v2/server/internal/app/federation"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/content"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/identity"
)

type handlerFunc func(ctx context.Context, event appEvent.Event) error

func (h handlerFunc) Handle(ctx context.Context, event appEvent.Event) error {
	return h(ctx, event)
}

func RegisterSubscribers(bus appEvent.Bus, svc *Service, contentRepo content.Repository, identityRepo identity.Repository) {
	if bus == nil || svc == nil {
		return
	}
	bus.Subscribe(appfed.DeliveryStatusChanged{}.Name(), handlerFunc(func(ctx context.Context, event appEvent.Event) error {
		payload, ok := event.(appfed.DeliveryStatusChanged)
		if !ok || payload.SourceArticleID == nil || contentRepo == nil {
			return nil
		}
		article, err := contentRepo.GetArticleByID(ctx, *payload.SourceArticleID)
		if err != nil || article == nil {
			return nil
		}
		title := "联合投递状态更新"
		contentText := fmt.Sprintf("文章《%s》的联合%s投递状态变更为 %s。", article.Title, payload.DeliveryType, payload.Status)
		if payload.ErrorMessage != nil && strings.TrimSpace(*payload.ErrorMessage) != "" {
			contentText += " 错误：" + strings.TrimSpace(*payload.ErrorMessage)
		}
		_, err = svc.Create(ctx, article.AuthorID, "federation.delivery.status", title, contentText, map[string]any{
			"deliveryId": payload.DeliveryID,
			"requestId":  payload.RequestID,
			"status":     payload.Status,
			"type":       payload.DeliveryType,
		})
		return err
	}))

	bus.Subscribe("federation.mention.received", handlerFunc(func(ctx context.Context, event appEvent.Event) error {
		if identityRepo == nil {
			return nil
		}
		generic, ok := event.(appEvent.Generic)
		if !ok {
			return nil
		}
		username, _ := generic.Payload["MentionedUser"].(string)
		username = strings.TrimSpace(username)
		if username == "" {
			return nil
		}
		user, err := identityRepo.FindByUsername(ctx, username)
		if err != nil || user == nil {
			return nil
		}
		source, _ := generic.Payload["SourceInstanceURL"].(string)
		title := "收到联合提及"
		contentText := fmt.Sprintf("你收到来自 %s 的联合提及通知。", strings.TrimSpace(source))
		_, err = svc.Create(ctx, user.ID, "federation.mention.received", title, contentText, generic.Payload)
		return err
	}))

	bus.Subscribe("federation.citation.received", handlerFunc(func(ctx context.Context, event appEvent.Event) error {
		if contentRepo == nil {
			return nil
		}
		generic, ok := event.(appEvent.Generic)
		if !ok {
			return nil
		}
		targetPostID, _ := generic.Payload["TargetPostID"].(string)
		targetPostID = strings.TrimSpace(targetPostID)
		if targetPostID == "" {
			return nil
		}
		article, err := resolveArticleByTargetID(ctx, contentRepo, targetPostID)
		if err != nil || article == nil {
			return nil
		}
		source, _ := generic.Payload["SourceInstanceURL"].(string)
		title := "收到联合引用"
		contentText := fmt.Sprintf("文章《%s》收到来自 %s 的联合引用请求。", article.Title, strings.TrimSpace(source))
		_, err = svc.Create(ctx, article.AuthorID, "federation.citation.received", title, contentText, generic.Payload)
		return err
	}))

	bus.Subscribe("federation.friendlink.received", handlerFunc(func(ctx context.Context, event appEvent.Event) error {
		if identityRepo == nil {
			return nil
		}
		admins, err := identityRepo.ListAdmins(ctx)
		if err != nil || len(admins) == 0 {
			return nil
		}
		generic, ok := event.(appEvent.Generic)
		if !ok {
			return nil
		}
		requester, _ := generic.Payload["RequesterURL"].(string)
		title := "收到联合友链申请"
		contentText := fmt.Sprintf("收到来自 %s 的联合友链申请。", strings.TrimSpace(requester))
		for _, admin := range admins {
			if _, err := svc.Create(ctx, admin.ID, "federation.friendlink.received", title, contentText, generic.Payload); err != nil {
				return err
			}
		}
		return nil
	}))
}

func resolveArticleByTargetID(ctx context.Context, repo content.Repository, target string) (*content.Article, error) {
	if id, err := strconv.ParseInt(target, 10, 64); err == nil {
		return repo.GetArticleByID(ctx, id)
	}
	return repo.GetArticleByShortURL(ctx, target)
}
