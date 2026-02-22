package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/article"
	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/globalnotification"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/moment"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/page"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/content"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
)

type handlerFunc func(ctx context.Context, event appEvent.Event) error

func (h handlerFunc) Handle(ctx context.Context, event appEvent.Event) error {
	return h(ctx, event)
}

func RegisterArticleUpdateSubscriber(bus appEvent.Bus, manager *Manager) {
	if bus == nil || manager == nil {
		return
	}
	bus.Subscribe(article.ArticleUpdated{}.Name(), handlerFunc(func(ctx context.Context, event appEvent.Event) error {
		updated, ok := event.(article.ArticleUpdated)
		if !ok {
			return nil
		}
		payload := contract.ArticleContentPayload{
			ContentHash: updated.ContentHash,
			Title:       updated.Title,
			LeadIn:      updated.LeadIn,
			TOC:         mapTOCNodes(updated.TOC),
			Content:     updated.Content,
		}
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		manager.Broadcast(articleRoomKey(updated.ID), data)
		return nil
	}))
}

func RegisterMomentUpdateSubscriber(bus appEvent.Bus, manager *Manager) {
	if bus == nil || manager == nil {
		return
	}
	bus.Subscribe(moment.MomentUpdated{}.Name(), handlerFunc(func(ctx context.Context, event appEvent.Event) error {
		updated, ok := event.(moment.MomentUpdated)
		if !ok {
			return nil
		}
		payload := contract.MomentContentPayload{
			ContentHash: updated.ContentHash,
			Title:       updated.Title,
			Summary:     updated.Summary,
			TOC:         mapTOCNodes(updated.TOC),
			Content:     updated.Content,
		}
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		manager.Broadcast(momentRoomKey(updated.ID), data)
		return nil
	}))
}

func RegisterPageUpdateSubscriber(bus appEvent.Bus, manager *Manager) {
	if bus == nil || manager == nil {
		return
	}
	bus.Subscribe(page.PageUpdated{}.Name(), handlerFunc(func(ctx context.Context, event appEvent.Event) error {
		updated, ok := event.(page.PageUpdated)
		if !ok {
			return nil
		}
		payload := contract.PageContentPayload{
			ContentHash: updated.ContentHash,
			Title:       updated.Title,
			Description: updated.Description,
			TOC:         mapTOCNodes(updated.TOC),
			Content:     updated.Content,
		}
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		manager.Broadcast(pageRoomKey(updated.ID), data)
		return nil
	}))
}

func RegisterNotificationSubscriber(bus appEvent.Bus, manager *Manager) {
	if bus == nil || manager == nil {
		return
	}
	bus.Subscribe("admin.notification.created", handlerFunc(func(ctx context.Context, event appEvent.Event) error {
		generic, ok := event.(appEvent.Generic)
		if !ok || generic.Payload == nil {
			return nil
		}
		userID, ok := generic.Payload["UserID"].(int64)
		if !ok || userID <= 0 {
			return nil
		}
		payload := map[string]any{
			"id":         generic.Payload["ID"],
			"type":       generic.Payload["NotifType"],
			"title":      generic.Payload["Title"],
			"content":    generic.Payload["Content"],
			"payload":    generic.Payload["Payload"],
			"is_read":    false,
			"created_at": time.Now().UTC().Format(time.RFC3339),
		}
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		manager.Broadcast(NotificationRoomKey(userID), data)
		return nil
	}))
}

func RegisterGlobalNotificationSubscriber(bus appEvent.Bus, manager *Manager) {
	if bus == nil || manager == nil {
		return
	}

	bus.Subscribe(globalnotification.Created{}.Name(), handlerFunc(func(ctx context.Context, event appEvent.Event) error {
		created, ok := event.(globalnotification.Created)
		if !ok {
			return nil
		}
		payload := map[string]any{
			"type":       created.Name(),
			"id":         created.ID,
			"content":    created.Content,
			"publishAt":  created.PublishAt.UTC().Format(time.RFC3339),
			"expireAt":   created.ExpireAt.UTC().Format(time.RFC3339),
			"allowClose": created.AllowClose,
			"at":         created.At.UTC().Format(time.RFC3339),
		}
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		manager.Broadcast(RealtimeRoomKey(), data)
		return nil
	}))

	bus.Subscribe(globalnotification.Updated{}.Name(), handlerFunc(func(ctx context.Context, event appEvent.Event) error {
		updated, ok := event.(globalnotification.Updated)
		if !ok {
			return nil
		}
		payload := map[string]any{
			"type":       updated.Name(),
			"id":         updated.ID,
			"content":    updated.Content,
			"publishAt":  updated.PublishAt.UTC().Format(time.RFC3339),
			"expireAt":   updated.ExpireAt.UTC().Format(time.RFC3339),
			"allowClose": updated.AllowClose,
			"at":         updated.At.UTC().Format(time.RFC3339),
		}
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		manager.Broadcast(RealtimeRoomKey(), data)
		return nil
	}))

	bus.Subscribe(globalnotification.Deleted{}.Name(), handlerFunc(func(ctx context.Context, event appEvent.Event) error {
		deleted, ok := event.(globalnotification.Deleted)
		if !ok {
			return nil
		}
		payload := map[string]any{
			"type": deleted.Name(),
			"id":   deleted.ID,
		}
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		manager.Broadcast(RealtimeRoomKey(), data)
		return nil
	}))
}

func articleRoomKey(id int64) string {
	return fmt.Sprintf("article:%d", id)
}

func momentRoomKey(id int64) string {
	return fmt.Sprintf("moment:%d", id)
}

func pageRoomKey(id int64) string {
	return fmt.Sprintf("page:%d", id)
}

func NotificationRoomKey(userID int64) string {
	return fmt.Sprintf("notif:user:%d", userID)
}

func mapTOCNodes(nodes []content.TOCNode) []contract.TOCNode {
	result := make([]contract.TOCNode, len(nodes))
	for i, node := range nodes {
		result[i] = contract.TOCNode{
			Name:     node.Name,
			Anchor:   node.Anchor,
			Children: mapTOCNodes(node.Children),
		}
	}
	return result
}
