package friendlink

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	appfed "github.com/grtsinry43/grtblog-v2/server/internal/app/federation"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/federation"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/social"
	fedinfra "github.com/grtsinry43/grtblog-v2/server/internal/infra/federation"
)

type handlerFunc func(ctx context.Context, event appEvent.Event) error

func (h handlerFunc) Handle(ctx context.Context, event appEvent.Event) error {
	return h(ctx, event)
}

// RegisterFederationSubscribers 订阅 federation.friendlink.approved 事件，
// 当我方出站友链请求被远端审批通过后，自动创建 Instance + FriendLink 并触发同步。
func RegisterFederationSubscribers(
	bus appEvent.Bus,
	instanceRepo federation.FederationInstanceRepository,
	linkRepo social.FriendLinkRepository,
	resolver *fedinfra.Resolver,
	syncWorker *appfed.SyncWorker,
) {
	if bus == nil {
		return
	}
	bus.Subscribe("federation.friendlink.approved", handlerFunc(func(ctx context.Context, event appEvent.Event) error {
		generic, ok := event.(appEvent.Generic)
		if !ok {
			return nil
		}
		targetURL, _ := generic.Payload["TargetInstanceURL"].(string)
		targetURL = strings.TrimSpace(targetURL)
		if targetURL == "" {
			return nil
		}

		instance, err := ensureActiveInstance(ctx, targetURL, instanceRepo, resolver)
		if err != nil {
			log.Printf("[federation] 友链审批事件处理失败：创建实例 target=%s err=%v", targetURL, err)
			return nil
		}

		if err := ensureFederationFriendLink(ctx, instance, linkRepo); err != nil {
			log.Printf("[federation] 友链审批事件处理失败：创建友链 target=%s err=%v", targetURL, err)
			return nil
		}

		log.Printf("[federation] 友链审批事件处理完成：instance=%s(id=%d) status=active", targetURL, instance.ID)

		// 触发即时同步
		if syncWorker != nil {
			go syncWorker.SyncFriendLinkByURL(context.Background(), targetURL)
		}

		return nil
	}))
}

// ensureActiveInstance 拉取远端 manifest 并创建/激活联合实例。
func ensureActiveInstance(
	ctx context.Context,
	baseURL string,
	instanceRepo federation.FederationInstanceRepository,
	resolver *fedinfra.Resolver,
) (*federation.FederationInstance, error) {
	if instanceRepo == nil {
		return nil, errors.New("instance repository not configured")
	}
	baseURL = strings.TrimRight(baseURL, "/")

	// 尝试从 resolver 拉取远端文档
	var manifest *fedinfra.Manifest
	var endpoints *fedinfra.EndpointsDoc
	var keyDoc *fedinfra.PublicKeyDoc
	if resolver != nil {
		var err error
		manifest, err = resolver.FetchManifest(ctx, baseURL)
		if err != nil {
			log.Printf("[federation] 拉取 manifest 失败 target=%s err=%v", baseURL, err)
		}
		endpoints, err = resolver.FetchEndpoints(ctx, baseURL)
		if err != nil {
			log.Printf("[federation] 拉取 endpoints 失败 target=%s err=%v", baseURL, err)
		}
		keyDoc, err = resolver.FetchPublicKey(ctx, baseURL)
		if err != nil {
			log.Printf("[federation] 拉取 public-key 失败 target=%s err=%v", baseURL, err)
		}
	}

	instance, err := instanceRepo.GetByBaseURL(ctx, baseURL)
	if err != nil {
		if !errors.Is(err, federation.ErrFederationInstanceNotFound) {
			return nil, err
		}
		// 创建新实例
		instance = &federation.FederationInstance{
			BaseURL:    baseURL,
			Status:     "active",
			LastSeenAt: timePtr(time.Now().UTC()),
		}
		applyManifestToInstance(instance, manifest, endpoints, keyDoc)
		if err := instanceRepo.Create(ctx, instance); err != nil {
			return nil, err
		}
		return instance, nil
	}

	// 更新已有实例
	instance.Status = "active"
	instance.LastSeenAt = timePtr(time.Now().UTC())
	applyManifestToInstance(instance, manifest, endpoints, keyDoc)
	if err := instanceRepo.Update(ctx, instance); err != nil {
		return nil, err
	}
	return instance, nil
}

func applyManifestToInstance(instance *federation.FederationInstance, manifest *fedinfra.Manifest, endpoints *fedinfra.EndpointsDoc, keyDoc *fedinfra.PublicKeyDoc) {
	if manifest != nil {
		instance.Name = optStr(manifest.Instance.Name)
		instance.Description = optStr(manifest.Instance.Description)
		instance.ProtocolVersion = optStr(manifest.ProtocolVersion)
		instance.Features = marshalJSON(manifest.Features)
		instance.Policies = marshalJSON(manifest.Policies)
	}
	if endpoints != nil {
		instance.Endpoints = marshalJSON(endpoints)
	}
	if keyDoc != nil {
		instance.PublicKey = optStr(keyDoc.PublicKey)
		instance.KeyID = optStr(keyDoc.KeyID)
	}
}

// ensureFederationFriendLink 创建或激活联合友链。
func ensureFederationFriendLink(ctx context.Context, instance *federation.FederationInstance, linkRepo social.FriendLinkRepository) error {
	if linkRepo == nil || instance == nil {
		return errors.New("repository or instance is nil")
	}

	link, err := linkRepo.FindByURL(ctx, instance.BaseURL)
	if err != nil {
		if !errors.Is(err, social.ErrFriendLinkNotFound) {
			return err
		}
		// 创建新友链
		name := instance.BaseURL
		if instance.Name != nil && strings.TrimSpace(*instance.Name) != "" {
			name = *instance.Name
		}
		link = &social.FriendLink{
			Name:             name,
			URL:              instance.BaseURL,
			Description:      instance.Description,
			Type:             social.FriendLinkTypeFederation,
			InstanceID:       &instance.ID,
			IsActive:         true,
			TotalPostsCached: 0,
		}
		return linkRepo.Create(ctx, link)
	}

	// 已存在则激活并绑定实例
	link.Type = social.FriendLinkTypeFederation
	link.InstanceID = &instance.ID
	link.IsActive = true
	return linkRepo.Update(ctx, link)
}

func optStr(val string) *string {
	trimmed := strings.TrimSpace(val)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func marshalJSON(value any) json.RawMessage {
	payload, err := json.Marshal(value)
	if err != nil {
		return json.RawMessage("{}")
	}
	return payload
}

func timePtr(t time.Time) *time.Time {
	return &t
}
