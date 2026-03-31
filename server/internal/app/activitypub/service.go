package activitypub

import (
	"bytes"
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"html"
	htmltemplate "html/template"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"code.superseriousbusiness.org/httpsig"
	"github.com/google/uuid"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/adminnotification"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/sysconfig"
	domainap "github.com/grtsinry43/grtblog-v2/server/internal/domain/activitypub"
	domaincomment "github.com/grtsinry43/grtblog-v2/server/internal/domain/comment"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/content"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/identity"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/thinking"
	fedinfra "github.com/grtsinry43/grtblog-v2/server/internal/infra/federation"
)

const (
	publicCollection = "https://www.w3.org/ns/activitystreams#Public"
	activityContext  = "https://www.w3.org/ns/activitystreams"
	securityContext  = "https://w3id.org/security/v1"
)

type Service struct {
	cfgSvc       *sysconfig.Service
	followers    domainap.FollowerRepository
	outbox       domainap.OutboxRepository
	contentRepo  content.Repository
	thinkingRepo thinking.ThinkingRepository
	commentRepo  domaincomment.CommentRepository
	identityRepo identity.Repository
	notifSvc     *adminnotification.Service
	httpClient   *http.Client
}

type PublishCmd struct {
	SourceType    string
	SourceID      int64
	Summary       string
	TriggerSource string
}

type PublishResult struct {
	Item          domainap.OutboxItem
	Deliveries    int
	SuccessCount  int
	FailureCount  int
	FailedTargets []string
}

type sendResult struct {
	HTTPStatus int
	Error      error
}

type ActorDocument struct {
	Context           []string       `json:"@context"`
	ID                string         `json:"id"`
	Type              string         `json:"type"`
	PreferredUsername string         `json:"preferredUsername"`
	Name              string         `json:"name,omitempty"`
	Summary           string         `json:"summary,omitempty"`
	URL               string         `json:"url,omitempty"`
	Icon              *actorImageRef `json:"icon,omitempty"`
	Image             *actorImageRef `json:"image,omitempty"`
	Inbox             string         `json:"inbox"`
	Outbox            string         `json:"outbox"`
	Followers         string         `json:"followers"`
	PublicKey         actorPublicKey `json:"publicKey"`
}

type actorImageRef struct {
	Type      string `json:"type"`
	MediaType string `json:"mediaType,omitempty"`
	URL       string `json:"url"`
}

type actorPublicKey struct {
	ID           string `json:"id"`
	Owner        string `json:"owner"`
	PublicKeyPEM string `json:"publicKeyPem"`
}

type WebFingerDocument struct {
	Subject string              `json:"subject"`
	Links   []webFingerLinkItem `json:"links"`
}

type webFingerLinkItem struct {
	Rel  string `json:"rel"`
	Type string `json:"type,omitempty"`
	Href string `json:"href,omitempty"`
}

type NodeInfoDiscoveryDocument struct {
	Links []NodeInfoDiscoveryLink `json:"links"`
}

type NodeInfoDiscoveryLink struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}

type OrderedCollection struct {
	Context      string `json:"@context"`
	ID           string `json:"id"`
	Type         string `json:"type"`
	TotalItems   int64  `json:"totalItems"`
	OrderedItems []any  `json:"orderedItems,omitempty"`
}

type activityEnvelope struct {
	Context json.RawMessage `json:"@context,omitempty"`
	ID      string          `json:"id,omitempty"`
	Type    string          `json:"type"`
	Actor   json.RawMessage `json:"actor,omitempty"`
	Object  json.RawMessage `json:"object,omitempty"`
	To      stringList      `json:"to,omitempty"`
	CC      stringList      `json:"cc,omitempty"`
}

type noteObject struct {
	ID           string        `json:"id,omitempty"`
	Type         string        `json:"type"`
	AttributedTo string        `json:"attributedTo,omitempty"`
	Content      string        `json:"content,omitempty"`
	InReplyTo    string        `json:"inReplyTo,omitempty"`
	URL          string        `json:"url,omitempty"`
	Published    string        `json:"published,omitempty"`
	To           stringList    `json:"to,omitempty"`
	CC           stringList    `json:"cc,omitempty"`
	Tag          []noteTagItem `json:"tag,omitempty"`
}

type noteTagItem struct {
	Type string `json:"type,omitempty"`
	Href string `json:"href,omitempty"`
	Name string `json:"name,omitempty"`
}

type stringList []string

func (s *stringList) UnmarshalJSON(raw []byte) error {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		*s = nil
		return nil
	}

	var many []string
	if err := json.Unmarshal(raw, &many); err == nil {
		out := make([]string, 0, len(many))
		for _, item := range many {
			item = strings.TrimSpace(item)
			if item != "" {
				out = append(out, item)
			}
		}
		*s = out
		return nil
	}

	var one string
	if err := json.Unmarshal(raw, &one); err == nil {
		one = strings.TrimSpace(one)
		if one == "" {
			*s = nil
			return nil
		}
		*s = []string{one}
		return nil
	}

	*s = nil
	return nil
}

type remoteActor struct {
	ID                string
	PreferredUsername string
	Name              string
	Icon              *remoteMediaRef
	Image             *remoteMediaRef
	Inbox             string
	Endpoints         struct {
		SharedInbox string
	}
	PublicKey struct {
		ID           string
		PublicKeyPEM string
	}
}

type remoteMediaRef struct {
	Type      string
	MediaType string
	URL       string
}

type commentTarget struct {
	AreaID int64
}

type localActorProfile struct {
	Summary  string
	URL      string
	IconURL  string
	ImageURL string
}

func NewService(
	cfgSvc *sysconfig.Service,
	followers domainap.FollowerRepository,
	outbox domainap.OutboxRepository,
	contentRepo content.Repository,
	thinkingRepo thinking.ThinkingRepository,
	commentRepo domaincomment.CommentRepository,
	identityRepo identity.Repository,
	notifSvc *adminnotification.Service,
) *Service {
	return &Service{
		cfgSvc:       cfgSvc,
		followers:    followers,
		outbox:       outbox,
		contentRepo:  contentRepo,
		thinkingRepo: thinkingRepo,
		commentRepo:  commentRepo,
		identityRepo: identityRepo,
		notifSvc:     notifSvc,
		httpClient:   &http.Client{Timeout: 12 * time.Second},
	}
}

func (s *Service) ResolveBaseURL(ctx context.Context, fallbackBaseURL string) (string, sysconfig.ActivityPubSettings, error) {
	if s.cfgSvc == nil {
		return "", sysconfig.ActivityPubSettings{}, errors.New("activitypub config service not configured")
	}
	settings, err := s.cfgSvc.ActivityPubSettings(ctx)
	if err != nil {
		return "", sysconfig.ActivityPubSettings{}, err
	}
	if !settings.Enabled {
		return "", settings, errors.New("activitypub disabled")
	}
	baseURL := strings.TrimRight(strings.TrimSpace(settings.InstanceURL), "/")
	if baseURL == "" {
		baseURL = strings.TrimRight(strings.TrimSpace(fallbackBaseURL), "/")
	}
	if baseURL == "" {
		return "", settings, errors.New("instance url is empty")
	}
	return baseURL, settings, nil
}

func (s *Service) ActorDocument(ctx context.Context, baseURL string) (*ActorDocument, error) {
	baseURL, settings, err := s.ResolveBaseURL(ctx, baseURL)
	if err != nil {
		return nil, err
	}
	actorID := actorURL(baseURL)
	name := strings.TrimSpace(settings.InstanceName)
	if name == "" {
		name = "grtblog"
	}
	profile := s.resolveLocalActorProfile(ctx, baseURL)
	// Dedicated header image config takes precedence over auto-resolved og_image
	if headerImg := strings.TrimSpace(settings.ActorHeaderImage); headerImg != "" {
		profile.ImageURL = resolveRelativeURL(baseURL, headerImg)
	}
	pubKey := strings.TrimSpace(settings.PublicKey)
	if pubKey == "" {
		return nil, errors.New("public key not configured")
	}
	doc := &ActorDocument{
		Context:           []string{activityContext, securityContext},
		ID:                actorID,
		Type:              "Person",
		PreferredUsername: preferredUsername(settings),
		Name:              name,
		Summary:           profile.Summary,
		URL:               profile.URL,
		Icon:              buildActorImageRef(profile.IconURL),
		Image:             buildActorImageRef(profile.ImageURL),
		Inbox:             inboxURL(baseURL),
		Outbox:            outboxURL(baseURL),
		Followers:         followersURL(baseURL),
		PublicKey: actorPublicKey{
			ID:           actorKeyID(baseURL),
			Owner:        actorID,
			PublicKeyPEM: pubKey,
		},
	}
	return doc, nil
}

func (s *Service) BuildWebFinger(ctx context.Context, baseURL, resource string) (*WebFingerDocument, bool, error) {
	baseURL, settings, err := s.ResolveBaseURL(ctx, baseURL)
	if err != nil {
		return nil, false, err
	}
	resource = strings.TrimSpace(resource)
	if resource == "" {
		return nil, false, nil
	}
	actorID := actorURL(baseURL)
	acct := acctURI(baseURL, preferredUsername(settings))
	if !strings.EqualFold(resource, acct) && !strings.EqualFold(strings.TrimRight(resource, "/"), actorID) {
		return nil, false, nil
	}
	return &WebFingerDocument{
		Subject: acct,
		Links: []webFingerLinkItem{{
			Rel:  "self",
			Type: "application/activity+json",
			Href: actorID,
		}},
	}, true, nil
}

func (s *Service) BuildNodeInfoDiscovery(ctx context.Context, baseURL string) (*NodeInfoDiscoveryDocument, error) {
	baseURL, _, err := s.ResolveBaseURL(ctx, baseURL)
	if err != nil {
		return nil, err
	}
	return &NodeInfoDiscoveryDocument{
		Links: []NodeInfoDiscoveryLink{{
			Rel:  "http://nodeinfo.diaspora.software/ns/schema/2.0",
			Href: strings.TrimRight(baseURL, "/") + "/nodeinfo/2.0",
		}},
	}, nil
}

func (s *Service) BuildNodeInfo20(ctx context.Context, baseURL string) (map[string]any, error) {
	baseURL, _, err := s.ResolveBaseURL(ctx, baseURL)
	if err != nil {
		return nil, err
	}
	usageUsers := map[string]any{
		"total":          1,
		"activeHalfyear": 1,
		"activeMonth":    1,
	}
	var localPosts int64
	if s.contentRepo != nil {
		if _, total, err := s.contentRepo.ListPublicArticlesForFederation(ctx, nil, nil, 1, 1); err == nil {
			localPosts = total
		}
	}
	return map[string]any{
		"version": "2.0",
		"software": map[string]any{
			"name":    "grtblog",
			"version": "2",
		},
		"protocols":         []string{"activitypub"},
		"services":          map[string]any{"inbound": []string{}, "outbound": []string{}},
		"openRegistrations": false,
		"usage": map[string]any{
			"users":      usageUsers,
			"localPosts": localPosts,
		},
		"metadata": map[string]any{
			"homepage": baseURL,
		},
	}, nil
}

func (s *Service) FollowersCollection(ctx context.Context, baseURL string) (*OrderedCollection, error) {
	baseURL, _, err := s.ResolveBaseURL(ctx, baseURL)
	if err != nil {
		return nil, err
	}
	if s.followers == nil {
		return &OrderedCollection{Context: activityContext, ID: followersURL(baseURL), Type: "Collection", TotalItems: 0}, nil
	}
	_, total, err := s.followers.List(ctx, "active", 1, 1)
	if err != nil {
		return nil, err
	}
	return &OrderedCollection{
		Context:    activityContext,
		ID:         followersURL(baseURL),
		Type:       "Collection",
		TotalItems: total,
	}, nil
}

func (s *Service) OutboxCollection(ctx context.Context, baseURL string, page, pageSize int) (*OrderedCollection, error) {
	baseURL, _, err := s.ResolveBaseURL(ctx, baseURL)
	if err != nil {
		return nil, err
	}
	if s.outbox == nil {
		return &OrderedCollection{Context: activityContext, ID: outboxURL(baseURL), Type: "OrderedCollection", TotalItems: 0, OrderedItems: []any{}}, nil
	}
	items, total, err := s.outbox.List(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}
	ordered := make([]any, 0, len(items))
	for _, item := range items {
		var payload any
		if err := json.Unmarshal(item.Activity, &payload); err != nil {
			continue
		}
		ordered = append(ordered, payload)
	}
	return &OrderedCollection{
		Context:      activityContext,
		ID:           outboxURL(baseURL),
		Type:         "OrderedCollection",
		TotalItems:   total,
		OrderedItems: ordered,
	}, nil
}

func (s *Service) ObjectDocument(ctx context.Context, baseURL string, objectToken string) (map[string]any, error) {
	baseURL, settings, err := s.ResolveBaseURL(ctx, baseURL)
	if err != nil {
		return nil, err
	}
	token := strings.Trim(strings.TrimSpace(objectToken), "/")
	if token == "" {
		return nil, errors.New("object id is empty")
	}
	parts := strings.SplitN(token, "-", 2)
	if len(parts) != 2 {
		return nil, errors.New("invalid object id")
	}
	sourceType := strings.TrimSpace(parts[0])
	sourceID, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
	if err != nil || sourceID <= 0 {
		return nil, errors.New("invalid object source id")
	}
	object, err := s.buildObjectForSource(ctx, baseURL, sourceType, sourceID, "", settings.PublishTemplate, false)
	if err != nil {
		return nil, err
	}
	return object, nil
}

func (s *Service) ListFollowers(ctx context.Context, page, pageSize int) ([]domainap.Follower, int64, error) {
	if s.followers == nil {
		return nil, 0, nil
	}
	return s.followers.List(ctx, "", page, pageSize)
}

func (s *Service) ListOutbox(ctx context.Context, opts domainap.OutboxListOptions) ([]domainap.OutboxItem, int64, error) {
	if s.outbox == nil {
		return nil, 0, errors.New("activitypub outbox repository not configured")
	}
	return s.outbox.ListWithOptions(ctx, opts)
}

func (s *Service) GetOutbox(ctx context.Context, id int64) (*domainap.OutboxItem, error) {
	if s.outbox == nil {
		return nil, errors.New("activitypub outbox repository not configured")
	}
	if id <= 0 {
		return nil, domainap.ErrOutboxItemNotFound
	}
	return s.outbox.GetByID(ctx, id)
}

func (s *Service) RetryFailedDeliveries(ctx context.Context, baseURL string, id int64) (*domainap.OutboxItem, error) {
	if s.outbox == nil {
		return nil, errors.New("activitypub outbox repository not configured")
	}
	if id <= 0 {
		return nil, domainap.ErrOutboxItemNotFound
	}
	baseURL, settings, err := s.ResolveBaseURL(ctx, baseURL)
	if err != nil {
		return nil, err
	}
	if !settings.AllowOutbound {
		return nil, errors.New("activitypub outbound disabled")
	}
	item, err := s.outbox.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if item.Status != domainap.OutboxStatusFailed && item.Status != domainap.OutboxStatusPartial {
		return nil, domainap.ErrOutboxItemNotRetryable
	}
	started := time.Now().UTC()
	item.Status = domainap.OutboxStatusSending
	item.StartedAt = &started
	_ = s.outbox.UpdateDeliveryResult(ctx, item)

	raw := []byte(item.Activity)
	failedTargets := 0
	successCount := 0
	failureCount := 0
	for i := range item.Deliveries {
		detail := &item.Deliveries[i]
		if detail.Status != "failed" {
			if detail.Status == "success" {
				successCount++
			} else {
				failureCount++
			}
			continue
		}
		failedTargets++
		if strings.TrimSpace(detail.Inbox) == "" {
			failureCount++
			continue
		}
		res := s.sendActivityWithResult(ctx, settings, baseURL, detail.Inbox, raw)
		if res.HTTPStatus > 0 {
			httpStatus := res.HTTPStatus
			detail.HTTPStatus = &httpStatus
		}
		nowAt := time.Now().UTC()
		detail.DeliveredAt = &nowAt
		if res.Error != nil {
			detail.Error = res.Error.Error()
			detail.Status = "failed"
			failureCount++
			continue
		}
		detail.Status = "success"
		detail.Error = ""
		successCount++
	}
	if failedTargets == 0 {
		return nil, domainap.ErrOutboxItemNotRetryable
	}
	finished := time.Now().UTC()
	duration := finished.Sub(started).Milliseconds()
	item.SuccessCount = successCount
	item.FailureCount = failureCount
	item.TotalTargets = len(item.Deliveries)
	item.StartedAt = &started
	item.FinishedAt = &finished
	item.DurationMs = &duration
	if item.TotalTargets == 0 || item.SuccessCount == item.TotalTargets {
		item.Status = domainap.OutboxStatusCompleted
	} else if item.SuccessCount > 0 {
		item.Status = domainap.OutboxStatusPartial
	} else {
		item.Status = domainap.OutboxStatusFailed
	}
	if err := s.outbox.UpdateDeliveryResult(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) Publish(ctx context.Context, baseURL string, cmd PublishCmd) (*PublishResult, error) {
	baseURL, settings, err := s.ResolveBaseURL(ctx, baseURL)
	if err != nil {
		return nil, err
	}
	if !settings.AllowOutbound {
		return nil, errors.New("activitypub outbound disabled")
	}
	if s.followers == nil || s.outbox == nil {
		return nil, errors.New("activitypub repositories not configured")
	}
	if cmd.SourceID <= 0 {
		return nil, errors.New("source id is invalid")
	}

	sourceType := strings.ToLower(strings.TrimSpace(cmd.SourceType))
	if sourceType == "note" {
		sourceType = "article"
	}
	if sourceType == "moments" {
		sourceType = "moment"
	}
	if !allowPublishType(settings.PublishTypes, sourceType) {
		return nil, errors.New("source type is not allowed by activitypub.publishTypes")
	}
	now := time.Now().UTC()
	object, err := s.buildObjectForSource(ctx, baseURL, sourceType, cmd.SourceID, cmd.Summary, settings.PublishTemplate, true)
	if err != nil {
		return nil, err
	}
	sourceURL := strings.TrimSpace(valueAsString(object["url"]))
	objectID := strings.TrimSpace(valueAsString(object["id"]))
	contentHTML := strings.TrimSpace(valueAsString(object["content"]))
	if objectID == "" || sourceURL == "" || contentHTML == "" {
		return nil, errors.New("invalid object payload")
	}

	activityID := strings.TrimRight(baseURL, "/") + "/ap/activities/" + strings.ReplaceAll(uuid.NewString(), "-", "")
	actorID := actorURL(baseURL)
	object["published"] = now.Format(time.RFC3339)
	object["attributedTo"] = actorID
	object["to"] = []string{publicCollection}
	object["cc"] = []string{followersURL(baseURL)}
	activity := map[string]any{
		"@context": []string{activityContext},
		"id":       activityID,
		"type":     "Create",
		"actor":    actorID,
		"to":       []string{publicCollection},
		"cc":       []string{followersURL(baseURL)},
		"object":   object,
	}
	raw, err := json.Marshal(activity)
	if err != nil {
		return nil, err
	}
	triggerSource := strings.TrimSpace(cmd.TriggerSource)
	if triggerSource == "" {
		triggerSource = "auto"
	}

	outboxItem := &domainap.OutboxItem{
		ActivityID:    activityID,
		ObjectID:      objectID,
		SourceType:    sourceType,
		SourceID:      cmd.SourceID,
		SourceURL:     sourceURL,
		Summary:       strings.TrimSpace(stripHTML(contentHTML)),
		Activity:      raw,
		Status:        domainap.OutboxStatusQueued,
		TriggerSource: triggerSource,
		Deliveries:    make([]domainap.DeliveryDetail, 0),
		PublishedAt:   now,
	}
	if err := s.outbox.Create(ctx, outboxItem); err != nil {
		return nil, err
	}

	followers, err := s.followers.ListActive(ctx)
	if err != nil {
		return nil, err
	}

	result := &PublishResult{Item: *outboxItem, Deliveries: len(followers)}
	if len(followers) == 0 {
		outboxItem.TotalTargets = 0
		outboxItem.SuccessCount = 0
		outboxItem.FailureCount = 0
		outboxItem.Status = domainap.OutboxStatusCompleted
		finished := time.Now().UTC()
		outboxItem.StartedAt = &finished
		outboxItem.FinishedAt = &finished
		zero := int64(0)
		outboxItem.DurationMs = &zero
		_ = s.outbox.UpdateDeliveryResult(ctx, outboxItem)
		result.Item = *outboxItem
		return result, nil
	}

	started := time.Now().UTC()
	outboxItem.StartedAt = &started
	outboxItem.Status = domainap.OutboxStatusSending
	_ = s.outbox.UpdateDeliveryResult(ctx, outboxItem)

	details := make([]domainap.DeliveryDetail, 0, len(followers))
	for _, follower := range followers {
		target := strings.TrimSpace(firstNonEmpty(ptrValue(follower.SharedInboxURL), follower.InboxURL))
		detail := domainap.DeliveryDetail{Inbox: target, ActorID: follower.ActorID, Status: "failed"}
		if target == "" {
			detail.Error = "missing inbox target"
			result.FailureCount++
			result.FailedTargets = append(result.FailedTargets, follower.ActorID)
			details = append(details, detail)
			continue
		}
		sendRes := s.sendActivityWithResult(ctx, settings, baseURL, target, raw)
		if sendRes.HTTPStatus > 0 {
			httpStatus := sendRes.HTTPStatus
			detail.HTTPStatus = &httpStatus
		}
		nowAt := time.Now().UTC()
		detail.DeliveredAt = &nowAt
		if sendRes.Error != nil {
			detail.Error = sendRes.Error.Error()
			result.FailureCount++
			result.FailedTargets = append(result.FailedTargets, target)
			details = append(details, detail)
			continue
		}
		detail.Status = "success"
		detail.Error = ""
		result.SuccessCount++
		details = append(details, detail)
	}

	finished := time.Now().UTC()
	duration := finished.Sub(started).Milliseconds()
	outboxItem.Deliveries = details
	outboxItem.TotalTargets = len(followers)
	outboxItem.SuccessCount = result.SuccessCount
	outboxItem.FailureCount = result.FailureCount
	outboxItem.StartedAt = &started
	outboxItem.FinishedAt = &finished
	outboxItem.DurationMs = &duration
	if outboxItem.TotalTargets == 0 || outboxItem.SuccessCount == outboxItem.TotalTargets {
		outboxItem.Status = domainap.OutboxStatusCompleted
	} else if outboxItem.SuccessCount > 0 {
		outboxItem.Status = domainap.OutboxStatusPartial
	} else {
		outboxItem.Status = domainap.OutboxStatusFailed
	}
	if err := s.outbox.UpdateDeliveryResult(ctx, outboxItem); err != nil {
		return nil, err
	}
	result.Item = *outboxItem
	return result, nil
}

func (s *Service) HandleInbox(ctx context.Context, baseURL string, req *http.Request, body []byte) error {
	baseURL, settings, err := s.ResolveBaseURL(ctx, baseURL)
	if err != nil {
		return err
	}
	if !settings.AllowInbound {
		return errors.New("activitypub inbound disabled")
	}
	if err := s.verifyRequestSignature(ctx, req, body); err != nil {
		return err
	}

	var activity activityEnvelope
	if err := json.Unmarshal(body, &activity); err != nil {
		return err
	}
	activityType := strings.ToLower(strings.TrimSpace(activity.Type))
	switch activityType {
	case "follow":
		return s.handleFollow(ctx, baseURL, settings, activity, body)
	case "create":
		return s.handleCreate(ctx, baseURL, settings, activity)
	case "undo":
		return s.handleUndo(ctx, activity)
	default:
		return nil
	}
}

func (s *Service) handleFollow(ctx context.Context, baseURL string, settings sysconfig.ActivityPubSettings, activity activityEnvelope, raw []byte) error {
	if s.followers == nil {
		return errors.New("follower repository not configured")
	}
	actorID := parseActorID(activity.Actor)
	if actorID == "" {
		return errors.New("follow actor is empty")
	}
	objID := strings.TrimSpace(parseObjectID(activity.Object))
	if objID == "" {
		return errors.New("follow object is empty")
	}
	if !sameURL(objID, actorURL(baseURL)) {
		return nil
	}
	remote, err := s.fetchRemoteActor(ctx, actorID)
	if err != nil {
		return err
	}
	followedAt := time.Now().UTC()
	follower := &domainap.Follower{
		ActorID:           actorID,
		InboxURL:          strings.TrimSpace(remote.Inbox),
		SharedInboxURL:    strPtr(strings.TrimSpace(remote.Endpoints.SharedInbox)),
		PreferredUsername: strPtr(strings.TrimSpace(remote.PreferredUsername)),
		DisplayName:       strPtr(strings.TrimSpace(remote.Name)),
		Status:            "active",
		FollowedAt:        followedAt,
		LastSeenAt:        &followedAt,
	}
	if follower.InboxURL == "" {
		return errors.New("remote actor inbox is empty")
	}
	if err := s.followers.Upsert(ctx, follower); err != nil {
		return err
	}

	if !settings.AutoAcceptFollow {
		return nil
	}
	accept := map[string]any{
		"@context": []string{activityContext},
		"id":       strings.TrimRight(baseURL, "/") + "/ap/activities/" + strings.ReplaceAll(uuid.NewString(), "-", ""),
		"type":     "Accept",
		"actor":    actorURL(baseURL),
		"object":   json.RawMessage(raw),
	}
	payload, err := json.Marshal(accept)
	if err != nil {
		return err
	}
	return s.sendActivity(ctx, settings, baseURL, follower.InboxURL, payload)
}

func (s *Service) handleUndo(ctx context.Context, activity activityEnvelope) error {
	if s.followers == nil {
		return nil
	}
	actorID := parseActorID(activity.Actor)
	if actorID == "" {
		return nil
	}
	// Parse the inner object to determine what is being undone.
	var inner activityEnvelope
	if err := json.Unmarshal(activity.Object, &inner); err != nil {
		return nil
	}
	if !strings.EqualFold(strings.TrimSpace(inner.Type), "Follow") {
		return nil
	}
	follower, err := s.followers.GetByActorID(ctx, actorID)
	if err != nil || follower == nil {
		return nil
	}
	follower.Status = "inactive"
	return s.followers.Upsert(ctx, follower)
}

func (s *Service) handleCreate(ctx context.Context, baseURL string, settings sysconfig.ActivityPubSettings, activity activityEnvelope) error {
	actorID := parseActorID(activity.Actor)
	if actorID == "" {
		return errors.New("create actor is empty")
	}
	if len(activity.Object) == 0 {
		return nil
	}
	var note noteObject
	if err := json.Unmarshal(activity.Object, &note); err != nil {
		return nil
	}
	if !strings.EqualFold(strings.TrimSpace(note.Type), "Note") {
		return nil
	}
	if strings.TrimSpace(note.AttributedTo) == "" {
		note.AttributedTo = actorID
	}
	if settings.AcceptInboundComment {
		if err := s.handleCreateAsComment(ctx, baseURL, actorID, note); err != nil {
			return err
		}
	}
	if settings.MentionToAdmin && isMentionToLocal(baseURL, actorURL(baseURL), note, preferredUsername(settings)) {
		_ = s.notifyAdminsMention(ctx, actorID, note)
	}
	return nil
}

func (s *Service) handleCreateAsComment(ctx context.Context, baseURL string, actorID string, note noteObject) error {
	if s.commentRepo == nil || s.contentRepo == nil {
		return nil
	}
	inReplyTo := strings.TrimSpace(note.InReplyTo)
	if inReplyTo == "" {
		return nil
	}

	objectID := strings.TrimSpace(firstNonEmpty(note.ID, note.URL))
	if objectID != "" {
		if _, err := s.commentRepo.FindByFederatedObjectID(ctx, objectID); err == nil {
			return nil
		}
	}

	target, parent, err := s.resolveCommentTarget(ctx, baseURL, inReplyTo)
	if err != nil || target == nil {
		return nil
	}

	nick := extractDisplayNameFromActor(actorID)
	avatar := ""
	if remote, err := s.fetchRemoteActor(ctx, actorID); err == nil && remote != nil {
		if strings.TrimSpace(remote.PreferredUsername) != "" {
			nick = strings.TrimSpace(remote.PreferredUsername)
		}
		if strings.TrimSpace(remote.Name) != "" && nick == "" {
			nick = strings.TrimSpace(remote.Name)
		}
		avatarURL := remote.avatarURL()
		if avatarURL != "" {
			avatar = avatarURL
		}
	}
	if nick == "" {
		nick = "federated"
	}
	contentText := strings.TrimSpace(stripHTML(note.Content))
	if contentText == "" {
		return nil
	}

	entity := &domaincomment.Comment{
		AreaID:            target.AreaID,
		Content:           contentText,
		AuthorID:          nil,
		VisitorID:         strPtr(actorID),
		NickName:          strPtr(nick),
		Email:             nil,
		Website:           strPtr(strings.TrimSpace(actorID)),
		Avatar:            strPtr(strings.TrimSpace(avatar)),
		IsOwner:           false,
		IsFriend:          false,
		IsAuthor:          false,
		IsViewed:          false,
		IsTop:             false,
		IsMy:              false,
		IsFederated:       true,
		FederatedProtocol: strPtr("activitypub"),
		FederatedActor:    strPtr(actorID),
		FederatedObjectID: strPtr(objectID),
		CanReply:          false,
		Status:            domaincomment.CommentStatusApproved,
	}
	if parent != nil {
		entity.ParentID = &parent.ID
	}
	return s.commentRepo.Create(ctx, entity)
}

func (s *Service) resolveCommentTarget(ctx context.Context, baseURL string, inReplyTo string) (*commentTarget, *domaincomment.Comment, error) {
	if s.commentRepo != nil {
		if parent, err := s.commentRepo.FindByFederatedObjectID(ctx, inReplyTo); err == nil && parent != nil {
			area, err := s.commentRepo.GetAreaByID(ctx, parent.AreaID)
			if err != nil || area == nil {
				return nil, nil, err
			}
			return &commentTarget{AreaID: area.ID}, parent, nil
		}
	}

	if article, err := s.contentRepo.GetArticleByActivityPubObjectID(ctx, inReplyTo); err == nil {
		if article != nil && article.CommentID != nil {
			return &commentTarget{AreaID: *article.CommentID}, nil, nil
		}
	}
	if moment, err := s.contentRepo.GetMomentByActivityPubObjectID(ctx, inReplyTo); err == nil {
		if moment != nil && moment.CommentID != nil {
			return &commentTarget{AreaID: *moment.CommentID}, nil, nil
		}
	}
	if s.thinkingRepo != nil {
		if item, err := s.thinkingRepo.FindByActivityPubObjectID(ctx, inReplyTo); err == nil && item != nil {
			return &commentTarget{AreaID: item.CommentID}, nil, nil
		}
	}

	if target := s.resolveContentByLocalURL(ctx, baseURL, inReplyTo); target != nil {
		return target, nil, nil
	}
	return nil, nil, nil
}

func (s *Service) resolveContentByLocalURL(ctx context.Context, baseURL, raw string) *commentTarget {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return nil
	}
	local, err := url.Parse(strings.TrimRight(baseURL, "/"))
	if err != nil {
		return nil
	}
	if !strings.EqualFold(u.Hostname(), local.Hostname()) {
		return nil
	}
	path := strings.TrimSpace(u.Path)
	if strings.HasPrefix(path, "/posts/") {
		slug := strings.TrimPrefix(path, "/posts/")
		slug = strings.Trim(slug, "/")
		if slug == "" {
			return nil
		}
		item, err := s.contentRepo.GetArticleByShortURL(ctx, slug)
		if err != nil {
			return nil
		}
		if item.CommentID == nil {
			return nil
		}
		return &commentTarget{AreaID: *item.CommentID}
	}
	if strings.HasPrefix(path, "/moments/") {
		slug := strings.TrimPrefix(path, "/moments/")
		parts := strings.Split(strings.Trim(slug, "/"), "/")
		if len(parts) == 0 {
			return nil
		}
		shortURL := strings.TrimSpace(parts[len(parts)-1])
		if shortURL == "" {
			return nil
		}
		item, err := s.contentRepo.GetMomentByShortURL(ctx, shortURL)
		if err != nil || item == nil || item.CommentID == nil {
			return nil
		}
		return &commentTarget{AreaID: *item.CommentID}
	}
	if strings.HasPrefix(path, "/thinkings") {
		fragment := strings.TrimSpace(u.Fragment)
		if strings.HasPrefix(fragment, "thinking-") {
			rawID := strings.TrimPrefix(fragment, "thinking-")
			id, err := strconv.ParseInt(rawID, 10, 64)
			if err != nil {
				return nil
			}
			if s.thinkingRepo == nil {
				return nil
			}
			item, err := s.thinkingRepo.FindByID(ctx, id)
			if err != nil || item == nil {
				return nil
			}
			return &commentTarget{AreaID: item.CommentID}
		}
	}
	if strings.HasPrefix(path, "/ap/objects/article-") {
		rawID := strings.TrimPrefix(path, "/ap/objects/article-")
		id, err := strconv.ParseInt(rawID, 10, 64)
		if err != nil {
			return nil
		}
		item, err := s.contentRepo.GetArticleByID(ctx, id)
		if err != nil {
			return nil
		}
		if item.CommentID == nil {
			return nil
		}
		return &commentTarget{AreaID: *item.CommentID}
	}
	if strings.HasPrefix(path, "/ap/objects/moment-") {
		rawID := strings.TrimPrefix(path, "/ap/objects/moment-")
		id, err := strconv.ParseInt(rawID, 10, 64)
		if err != nil {
			return nil
		}
		item, err := s.contentRepo.GetMomentByID(ctx, id)
		if err != nil || item == nil || item.CommentID == nil {
			return nil
		}
		return &commentTarget{AreaID: *item.CommentID}
	}
	if strings.HasPrefix(path, "/ap/objects/thinking-") {
		rawID := strings.TrimPrefix(path, "/ap/objects/thinking-")
		id, err := strconv.ParseInt(rawID, 10, 64)
		if err != nil || s.thinkingRepo == nil {
			return nil
		}
		item, err := s.thinkingRepo.FindByID(ctx, id)
		if err != nil || item == nil {
			return nil
		}
		return &commentTarget{AreaID: item.CommentID}
	}
	return nil
}

func (s *Service) resolveLocalActorProfile(ctx context.Context, baseURL string) localActorProfile {
	profile := localActorProfile{
		URL: strings.TrimRight(baseURL, "/") + "/",
	}
	if s.cfgSvc == nil {
		return profile
	}

	siteInfo, err := s.cfgSvc.WebsiteInfo(ctx)
	if err == nil {
		profile.URL = firstNonEmpty(
			resolveRelativeURL(baseURL, strings.TrimSpace(siteInfo["public_url"])),
			profile.URL,
		)
		profile.Summary = firstNonEmpty(
			strings.TrimSpace(siteInfo["description"]),
			strings.TrimSpace(siteInfo["og_description"]),
		)
		profile.IconURL = firstNonEmpty(
			resolveRelativeURL(baseURL, strings.TrimSpace(siteInfo["og_image"])),
			resolveRelativeURL(baseURL, strings.TrimSpace(siteInfo["favicon"])),
		)
		profile.ImageURL = resolveRelativeURL(baseURL, strings.TrimSpace(siteInfo["og_image"]))
	}

	themeRaw, err := s.cfgSvc.ThemeExtendInfo(ctx)
	if err == nil {
		avatarURL, description := parseThemeHeroProfile(themeRaw)
		if description != "" {
			profile.Summary = description
		}
		if avatarURL != "" {
			profile.IconURL = resolveRelativeURL(baseURL, avatarURL)
		}
	}
	if profile.ImageURL == "" {
		profile.ImageURL = profile.IconURL
	}
	return profile
}

func parseThemeHeroProfile(raw json.RawMessage) (avatarURL, description string) {
	if len(raw) == 0 {
		return "", ""
	}
	var root map[string]any
	if err := json.Unmarshal(raw, &root); err != nil {
		return "", ""
	}
	container := root
	if home, ok := asMap(root["home"]); ok {
		container = home
	}
	hero, ok := asMap(container["hero"])
	if !ok {
		return "", ""
	}
	return firstNonEmpty(valueAsString(hero["avatarUrl"])), firstNonEmpty(valueAsString(hero["description"]))
}

func resolveRelativeURL(baseURL, raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if parsed, err := url.Parse(raw); err == nil {
		if parsed.Scheme != "" {
			return parsed.String()
		}
		base, baseErr := url.Parse(strings.TrimRight(baseURL, "/") + "/")
		if baseErr != nil {
			return raw
		}
		return base.ResolveReference(parsed).String()
	}
	return raw
}

func buildActorImageRef(rawURL string) *actorImageRef {
	url := strings.TrimSpace(rawURL)
	if url == "" {
		return nil
	}
	return &actorImageRef{
		Type:      "Image",
		MediaType: detectImageMIMEType(url),
		URL:       url,
	}
}

func detectImageMIMEType(url string) string {
	lower := strings.ToLower(strings.TrimSpace(url))
	if lower == "" {
		return "image/jpeg"
	}
	if idx := strings.Index(lower, "?"); idx >= 0 {
		lower = lower[:idx]
	}
	switch {
	case strings.HasSuffix(lower, ".jpg"), strings.HasSuffix(lower, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(lower, ".png"):
		return "image/png"
	case strings.HasSuffix(lower, ".gif"):
		return "image/gif"
	case strings.HasSuffix(lower, ".webp"):
		return "image/webp"
	case strings.HasSuffix(lower, ".avif"):
		return "image/avif"
	case strings.HasSuffix(lower, ".svg"):
		return "image/svg+xml"
	case strings.HasSuffix(lower, ".bmp"):
		return "image/bmp"
	case strings.HasSuffix(lower, ".ico"):
		return "image/x-icon"
	default:
		return "image/jpeg"
	}
}

func (a *remoteActor) UnmarshalJSON(raw []byte) error {
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return err
	}
	a.ID = parseStringLike(payload["id"])
	a.PreferredUsername = parseStringLike(payload["preferredUsername"])
	a.Name = parseStringLike(payload["name"])
	a.Inbox = parseURLLike(payload["inbox"])
	a.Icon = parseRemoteMediaRef(payload["icon"])
	a.Image = parseRemoteMediaRef(payload["image"])
	if endpoints, ok := asMap(payload["endpoints"]); ok {
		a.Endpoints.SharedInbox = parseURLLike(endpoints["sharedInbox"])
	}
	if pubKey, ok := asMap(payload["publicKey"]); ok {
		a.PublicKey.ID = parseStringLike(pubKey["id"])
		a.PublicKey.PublicKeyPEM = parseStringLike(pubKey["publicKeyPem"])
	}
	return nil
}

func parseRemoteMediaRef(raw any) *remoteMediaRef {
	switch t := raw.(type) {
	case nil:
		return nil
	case string:
		url := parseURLLike(t)
		if url == "" {
			return nil
		}
		return &remoteMediaRef{Type: "Image", URL: url}
	case []any:
		for _, item := range t {
			if parsed := parseRemoteMediaRef(item); parsed != nil && parsed.URL != "" {
				return parsed
			}
		}
		return nil
	case map[string]any:
		ref := &remoteMediaRef{
			Type:      parseStringLike(t["type"]),
			MediaType: parseStringLike(t["mediaType"]),
			URL:       parseURLLike(t["url"]),
		}
		if ref.URL == "" {
			ref.URL = parseURLLike(t["href"])
		}
		if ref.URL == "" {
			ref.URL = parseURLLike(t["id"])
		}
		if ref.Type == "" {
			ref.Type = "Image"
		}
		if ref.URL == "" {
			return nil
		}
		return ref
	default:
		return nil
	}
}

func parseURLLike(raw any) string {
	switch t := raw.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(t)
	case []any:
		for _, item := range t {
			if out := parseURLLike(item); out != "" {
				return out
			}
		}
		return ""
	case map[string]any:
		if out := parseURLLike(t["url"]); out != "" {
			return out
		}
		if out := parseURLLike(t["href"]); out != "" {
			return out
		}
		if out := parseURLLike(t["id"]); out != "" {
			return out
		}
		if out := parseURLLike(t["@id"]); out != "" {
			return out
		}
		return ""
	default:
		return ""
	}
}

func parseStringLike(raw any) string {
	switch t := raw.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(t)
	case []any:
		for _, item := range t {
			if out := parseStringLike(item); out != "" {
				return out
			}
		}
		return ""
	case map[string]any:
		for _, key := range []string{"@value", "value", "name", "summary", "preferredUsername", "id"} {
			if out := parseStringLike(t[key]); out != "" {
				return out
			}
		}
		for _, value := range t {
			if out := parseStringLike(value); out != "" {
				return out
			}
		}
		return ""
	default:
		return ""
	}
}

func asMap(raw any) (map[string]any, bool) {
	out, ok := raw.(map[string]any)
	return out, ok
}

func (a *remoteActor) avatarURL() string {
	if a == nil {
		return ""
	}
	if a.Icon != nil {
		if url := strings.TrimSpace(a.Icon.URL); url != "" {
			return url
		}
	}
	if a.Image != nil {
		if url := strings.TrimSpace(a.Image.URL); url != "" {
			return url
		}
	}
	return ""
}

func (s *Service) notifyAdminsMention(ctx context.Context, actorID string, note noteObject) error {
	if s.identityRepo == nil || s.notifSvc == nil {
		return nil
	}
	admins, err := s.identityRepo.ListAdmins(ctx)
	if err != nil || len(admins) == 0 {
		return err
	}
	snippet := strings.TrimSpace(stripHTML(note.Content))
	if len([]rune(snippet)) > 120 {
		snippet = string([]rune(snippet)[:120]) + "..."
	}
	title := "收到 ActivityPub 提及"
	contentText := "收到来自联邦网络的提及"
	if actorID != "" {
		contentText += "：" + actorID
	}
	if snippet != "" {
		contentText += "，内容：" + snippet
	}
	payload := map[string]any{
		"actor":       actorID,
		"note_id":     strings.TrimSpace(note.ID),
		"note_url":    strings.TrimSpace(note.URL),
		"in_reply_to": strings.TrimSpace(note.InReplyTo),
	}
	for _, admin := range admins {
		if _, err := s.notifSvc.Create(ctx, admin.ID, "activitypub.mention.received", title, contentText, payload); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) verifyRequestSignature(ctx context.Context, req *http.Request, body []byte) error {
	signatureHeader := strings.TrimSpace(req.Header.Get("Signature"))
	if signatureHeader == "" {
		return errors.New("missing signature")
	}
	if len(body) > 0 {
		if err := verifyDigest(req.Header.Get("Digest"), body); err != nil {
			return err
		}
	}
	reqTime, err := time.Parse(http.TimeFormat, strings.TrimSpace(req.Header.Get("Date")))
	if err != nil {
		return err
	}
	skew := time.Since(reqTime)
	if skew > 10*time.Minute || skew < -10*time.Minute {
		return errors.New("signature date out of range")
	}
	verifier, err := httpsig.NewVerifier(req)
	if err != nil {
		return err
	}
	keyID := strings.TrimSpace(verifier.KeyId())
	if keyID == "" {
		return errors.New("missing keyId")
	}
	pubKey, alg, err := s.resolvePublicKeyForKeyID(ctx, keyID)
	if err != nil {
		return err
	}
	if err := verifier.Verify(pubKey, alg); err != nil {
		if alg != httpsig.ED25519 {
			if edErr := verifier.Verify(pubKey, httpsig.ED25519); edErr == nil {
				return nil
			}
		}
		return err
	}
	return nil
}

func (s *Service) resolvePublicKeyForKeyID(ctx context.Context, keyID string) (crypto.PublicKey, httpsig.Algorithm, error) {
	u, err := url.Parse(keyID)
	if err != nil {
		return nil, "", err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, "", errors.New("invalid key id scheme")
	}
	if err := fedinfra.ValidateRemoteURL(ctx, keyID); err != nil {
		return nil, "", err
	}
	actorURL := strings.TrimRight(keyID, "#")
	if u.Fragment != "" {
		u.Fragment = ""
		actorURL = u.String()
	}
	remote, err := s.fetchRemoteActor(ctx, actorURL)
	if err != nil {
		return nil, "", err
	}
	pemData := strings.TrimSpace(remote.PublicKey.PublicKeyPEM)
	if pemData == "" {
		return nil, "", errors.New("actor public key is empty")
	}
	pubKey, err := parsePublicKey(pemData)
	if err != nil {
		return nil, "", err
	}
	switch pubKey.(type) {
	case *rsa.PublicKey:
		return pubKey, httpsig.RSA_SHA256, nil
	case ed25519.PublicKey:
		return pubKey, httpsig.ED25519, nil
	default:
		return nil, "", errors.New("unsupported actor public key type")
	}
}

func (s *Service) fetchRemoteActor(ctx context.Context, actorID string) (*remoteActor, error) {
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("actor id is empty")
	}
	if err := fedinfra.ValidateRemoteURL(ctx, actorID); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, actorID, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", `application/activity+json, application/ld+json; profile="https://www.w3.org/ns/activitystreams", application/ld+json`)
	req.Header.Set("User-Agent", "grtblog-activitypub/2")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetch actor failed: %s", resp.Status)
	}
	var actor remoteActor
	if err := json.NewDecoder(resp.Body).Decode(&actor); err != nil {
		return nil, err
	}
	if strings.TrimSpace(actor.ID) == "" {
		actor.ID = actorID
	}
	return &actor, nil
}

func (s *Service) sendActivity(ctx context.Context, settings sysconfig.ActivityPubSettings, baseURL string, targetURL string, payload []byte) error {
	result := s.sendActivityWithResult(ctx, settings, baseURL, targetURL, payload)
	return result.Error
}

func (s *Service) sendActivityWithResult(ctx context.Context, settings sysconfig.ActivityPubSettings, baseURL string, targetURL string, payload []byte) sendResult {
	if err := fedinfra.ValidateRemoteURL(ctx, targetURL); err != nil {
		return sendResult{Error: err}
	}
	if strings.TrimSpace(settings.PrivateKey) == "" {
		return sendResult{Error: errors.New("private key not configured")}
	}
	privateKey, err := parsePrivateKey(settings.PrivateKey)
	if err != nil {
		return sendResult{Error: err}
	}
	algorithm := strings.TrimSpace(settings.SignatureAlg)
	if algorithm == "" {
		algorithm = "rsa-sha256"
	}
	signer, err := fedinfra.NewSigner(algorithm)
	if err != nil {
		return sendResult{Error: err}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimSpace(targetURL), bytes.NewReader(payload))
	if err != nil {
		return sendResult{Error: err}
	}
	req.Header.Set("Accept", "application/activity+json, application/ld+json")
	req.Header.Set("Content-Type", "application/activity+json")
	keyID := actorKeyID(baseURL)
	if err := signer.SignRequest(req, payload, keyID, privateKey); err != nil {
		return sendResult{Error: err}
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return sendResult{Error: err}
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return sendResult{HTTPStatus: resp.StatusCode, Error: fmt.Errorf("deliver activity failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(raw)))}
	}
	return sendResult{HTTPStatus: resp.StatusCode}
}

func (s *Service) buildObjectForSource(ctx context.Context, baseURL string, sourceType string, sourceID int64, summaryOverride string, publishTemplate string, persist bool) (map[string]any, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	switch sourceType {
	case "article":
		article, err := s.contentRepo.GetArticleByID(ctx, sourceID)
		if err != nil {
			return nil, err
		}
		if article == nil || !article.IsPublished {
			return nil, errors.New("article is not published")
		}
		sourceURL := strings.TrimRight(baseURL, "/") + "/posts/" + article.ShortURL
		objectID := ""
		if article.ActivityPubObjectID != nil && strings.TrimSpace(*article.ActivityPubObjectID) != "" {
			objectID = strings.TrimSpace(*article.ActivityPubObjectID)
		} else {
			objectID = buildObjectID(baseURL, sourceType, article.ID)
		}
		if persist {
			article.ActivityPubObjectID = &objectID
			publishedAt := time.Now().UTC()
			article.ActivityPubLastPublishedAt = &publishedAt
			if err := s.contentRepo.UpdateArticle(ctx, article); err != nil {
				return nil, err
			}
		}
		summary := strings.TrimSpace(firstNonEmpty(summaryOverride, article.Summary))
		return map[string]any{
			"id":        objectID,
			"type":      "Note",
			"url":       sourceURL,
			"content":   renderFederatedHTML(publishTemplate, article.Title, summary, sourceURL, sourceType),
			"published": now,
		}, nil
	case "moment":
		moment, err := s.contentRepo.GetMomentByID(ctx, sourceID)
		if err != nil {
			return nil, err
		}
		if moment == nil || !moment.IsPublished {
			return nil, errors.New("moment is not published")
		}
		sourceURL := strings.TrimRight(baseURL, "/") + "/moments/" +
			moment.CreatedAt.UTC().Format("2006/01/02") + "/" + moment.ShortURL
		objectID := ""
		if moment.ActivityPubObjectID != nil && strings.TrimSpace(*moment.ActivityPubObjectID) != "" {
			objectID = strings.TrimSpace(*moment.ActivityPubObjectID)
		} else {
			objectID = buildObjectID(baseURL, sourceType, moment.ID)
		}
		if persist {
			moment.ActivityPubObjectID = &objectID
			publishedAt := time.Now().UTC()
			moment.ActivityPubLastPublishedAt = &publishedAt
			if err := s.contentRepo.UpdateMoment(ctx, moment); err != nil {
				return nil, err
			}
		}
		summary := strings.TrimSpace(firstNonEmpty(summaryOverride, moment.Summary))
		return map[string]any{
			"id":        objectID,
			"type":      "Note",
			"url":       sourceURL,
			"content":   renderFederatedHTML(publishTemplate, moment.Title, summary, sourceURL, sourceType),
			"published": now,
		}, nil
	case "thinking":
		if s.thinkingRepo == nil {
			return nil, errors.New("thinking repository not configured")
		}
		item, err := s.thinkingRepo.FindByID(ctx, sourceID)
		if err != nil {
			return nil, err
		}
		sourceURL := strings.TrimRight(baseURL, "/") + "/thinkings#thinking-" + strconv.FormatInt(item.ID, 10)
		objectID := ""
		if item.ActivityPubObjectID != nil && strings.TrimSpace(*item.ActivityPubObjectID) != "" {
			objectID = strings.TrimSpace(*item.ActivityPubObjectID)
		} else {
			objectID = buildObjectID(baseURL, sourceType, item.ID)
		}
		if persist {
			item.ActivityPubObjectID = &objectID
			publishedAt := time.Now().UTC()
			item.ActivityPubLastPublishedAt = &publishedAt
			if err := s.thinkingRepo.Update(ctx, item); err != nil {
				return nil, err
			}
		}
		summary := strings.TrimSpace(firstNonEmpty(summaryOverride, stripHTML(item.Content)))
		return map[string]any{
			"id":        objectID,
			"type":      "Note",
			"url":       sourceURL,
			"content":   renderFederatedHTML(publishTemplate, "思考", summary, sourceURL, sourceType),
			"published": now,
		}, nil
	default:
		return nil, errors.New("unsupported source type")
	}
}

func verifyDigest(digestHeader string, body []byte) error {
	digestHeader = strings.TrimSpace(digestHeader)
	if digestHeader == "" {
		return errors.New("missing digest")
	}
	parts := strings.SplitN(digestHeader, "=", 2)
	if len(parts) != 2 {
		return errors.New("invalid digest format")
	}
	if !strings.EqualFold(strings.TrimSpace(parts[0]), "SHA-256") {
		return errors.New("unsupported digest algorithm")
	}
	expected := strings.TrimSpace(parts[1])
	sum := sha256.Sum256(body)
	actual := base64.StdEncoding.EncodeToString(sum[:])
	if expected != actual {
		return errors.New("digest mismatch")
	}
	return nil
}

func parsePublicKey(pemData string) (crypto.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, errors.New("invalid public key pem")
	}
	if key, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		return key, nil
	}
	if key, err := x509.ParsePKCS1PublicKey(block.Bytes); err == nil {
		return key, nil
	}
	return nil, errors.New("unsupported public key format")
}

func parsePrivateKey(pemData string) (crypto.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, errors.New("invalid private key")
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		if rsaKey, ok := key.(*rsa.PrivateKey); ok {
			return rsaKey, nil
		}
		if edKey, ok := key.(ed25519.PrivateKey); ok {
			return edKey, nil
		}
		return nil, errors.New("unsupported private key type")
	}
	return nil, errors.New("unsupported private key format")
}

func renderFederatedHTML(rawTemplate, title, summary, sourceURL, sourceType string) string {
	tplText := strings.TrimSpace(rawTemplate)
	if tplText == "" {
		tplText = `<p><strong>{{ .Title }}</strong></p>{{ if .Summary }}<p>{{ .Summary }}</p>{{ end }}{{ if .URL }}<p><a href="{{ .URL }}" rel="nofollow noopener noreferrer">阅读全文</a></p>{{ end }}`
	}
	tpl, err := htmltemplate.New("activitypub.publish").Option("missingkey=error").Parse(tplText)
	if err != nil {
		return ""
	}
	var buf bytes.Buffer
	err = tpl.Execute(&buf, map[string]string{
		"Title":       strings.TrimSpace(title),
		"Summary":     strings.TrimSpace(summary),
		"URL":         strings.TrimSpace(sourceURL),
		"ContentType": activityPubContentTypeLabel(sourceType),
	})
	if err != nil {
		return ""
	}
	return strings.TrimSpace(buf.String())
}

func activityPubContentTypeLabel(sourceType string) string {
	switch strings.ToLower(strings.TrimSpace(sourceType)) {
	case "article":
		return "文章"
	case "moment":
		return "手记"
	case "thinking":
		return "思考"
	default:
		return ""
	}
}

func stripHTML(raw string) string {
	clean := html.UnescapeString(strings.TrimSpace(raw))
	re := regexp.MustCompile(`<[^>]+>`)
	clean = re.ReplaceAllString(clean, " ")
	clean = strings.Join(strings.Fields(clean), " ")
	return clean
}

func buildObjectID(baseURL string, sourceType string, sourceID int64) string {
	return strings.TrimRight(baseURL, "/") + "/ap/objects/" + sourceType + "-" + strconv.FormatInt(sourceID, 10)
}

func actorURL(baseURL string) string {
	return strings.TrimRight(baseURL, "/") + "/ap/actor"
}

func actorKeyID(baseURL string) string {
	return actorURL(baseURL) + "#main-key"
}

func inboxURL(baseURL string) string {
	return strings.TrimRight(baseURL, "/") + "/ap/inbox"
}

func outboxURL(baseURL string) string {
	return strings.TrimRight(baseURL, "/") + "/ap/outbox"
}

func followersURL(baseURL string) string {
	return strings.TrimRight(baseURL, "/") + "/ap/followers"
}

func acctURI(baseURL string, username string) string {
	u, err := url.Parse(strings.TrimRight(baseURL, "/"))
	if err != nil {
		return ""
	}
	return "acct:" + username + "@" + u.Host
}

func parseActorID(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var val string
	if err := json.Unmarshal(raw, &val); err == nil {
		return strings.TrimSpace(val)
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err == nil {
		if id, ok := obj["id"].(string); ok {
			return strings.TrimSpace(id)
		}
	}
	return ""
}

func parseStringValue(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var val string
	if err := json.Unmarshal(raw, &val); err == nil {
		return strings.TrimSpace(val)
	}
	return ""
}

func parseObjectID(raw json.RawMessage) string {
	if val := parseStringValue(raw); val != "" {
		return val
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err == nil {
		if id, ok := obj["id"].(string); ok {
			return strings.TrimSpace(id)
		}
	}
	return ""
}

func extractDisplayNameFromActor(actorID string) string {
	u, err := url.Parse(strings.TrimSpace(actorID))
	if err != nil {
		return ""
	}
	path := strings.Trim(strings.TrimSpace(u.Path), "/")
	if path == "" {
		return ""
	}
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return ""
	}
	return strings.TrimSpace(parts[len(parts)-1])
}

func isMentionToLocal(baseURL string, localActor string, note noteObject, username string) bool {
	for _, target := range append([]string{}, append(note.To, note.CC...)...) {
		if sameURL(target, localActor) {
			return true
		}
	}
	host := ""
	if parsed, err := url.Parse(strings.TrimRight(baseURL, "/")); err == nil {
		host = parsed.Host
	}
	expectedName := "@" + username
	if host != "" {
		expectedName = "@" + username + "@" + host
	}
	for _, tag := range note.Tag {
		if !strings.EqualFold(strings.TrimSpace(tag.Type), "Mention") {
			continue
		}
		if sameURL(tag.Href, localActor) {
			return true
		}
		if strings.EqualFold(strings.TrimSpace(tag.Name), expectedName) {
			return true
		}
	}
	return false
}

func ptrValue(raw *string) string {
	if raw == nil {
		return ""
	}
	return strings.TrimSpace(*raw)
}

func strPtr(raw string) *string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	return &raw
}

func valueAsString(raw any) string {
	if raw == nil {
		return ""
	}
	switch t := raw.(type) {
	case string:
		return strings.TrimSpace(t)
	case fmt.Stringer:
		return strings.TrimSpace(t.String())
	default:
		return ""
	}
}

func preferredUsername(settings sysconfig.ActivityPubSettings) string {
	username := strings.TrimSpace(settings.ActorUsername)
	if username == "" {
		return "blog"
	}
	return username
}

func allowPublishType(raw json.RawMessage, sourceType string) bool {
	sourceType = strings.ToLower(strings.TrimSpace(sourceType))
	if sourceType == "" {
		return false
	}
	if len(raw) == 0 {
		return sourceType == "article" || sourceType == "moment" || sourceType == "thinking"
	}
	var values []string
	if err := json.Unmarshal(raw, &values); err != nil {
		return sourceType == "article" || sourceType == "moment" || sourceType == "thinking"
	}
	for _, item := range values {
		if strings.EqualFold(strings.TrimSpace(item), sourceType) {
			return true
		}
	}
	return false
}

func firstNonEmpty(values ...string) string {
	for _, item := range values {
		if strings.TrimSpace(item) != "" {
			return strings.TrimSpace(item)
		}
	}
	return ""
}

func sameURL(a, b string) bool {
	a = strings.TrimRight(strings.TrimSpace(a), "/")
	b = strings.TrimRight(strings.TrimSpace(b), "/")
	if a == "" || b == "" {
		return false
	}
	return strings.EqualFold(a, b)
}
