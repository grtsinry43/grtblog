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
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"code.superseriousbusiness.org/httpsig"
	"github.com/google/uuid"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/activitypubconfig"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/adminnotification"
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
	cfgSvc       *activitypubconfig.Service
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
	SourceType string
	SourceID   int64
	Summary    string
}

type PublishResult struct {
	Item          domainap.OutboxItem
	Deliveries    int
	SuccessCount  int
	FailureCount  int
	FailedTargets []string
}

type ActorDocument struct {
	Context           []string       `json:"@context"`
	ID                string         `json:"id"`
	Type              string         `json:"type"`
	PreferredUsername string         `json:"preferredUsername"`
	Name              string         `json:"name,omitempty"`
	Inbox             string         `json:"inbox"`
	Outbox            string         `json:"outbox"`
	Followers         string         `json:"followers"`
	PublicKey         actorPublicKey `json:"publicKey"`
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
	ID                string `json:"id"`
	PreferredUsername string `json:"preferredUsername"`
	Name              string `json:"name"`
	Inbox             string `json:"inbox"`
	Endpoints         struct {
		SharedInbox string `json:"sharedInbox"`
	} `json:"endpoints"`
	PublicKey struct {
		ID           string `json:"id"`
		PublicKeyPEM string `json:"publicKeyPem"`
	} `json:"publicKey"`
}

func NewService(
	cfgSvc *activitypubconfig.Service,
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

func (s *Service) ResolveBaseURL(ctx context.Context, fallbackBaseURL string) (string, activitypubconfig.Settings, error) {
	if s.cfgSvc == nil {
		return "", activitypubconfig.Settings{}, errors.New("activitypub config service not configured")
	}
	settings, err := s.cfgSvc.Settings(ctx)
	if err != nil {
		return "", activitypubconfig.Settings{}, err
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
			"localPosts": 0,
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
	baseURL, _, err := s.ResolveBaseURL(ctx, baseURL)
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
	object, err := s.buildObjectForSource(ctx, baseURL, sourceType, sourceID, "", false)
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
	object, err := s.buildObjectForSource(ctx, baseURL, sourceType, cmd.SourceID, cmd.Summary, true)
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

	outboxItem := &domainap.OutboxItem{
		ActivityID:  activityID,
		ObjectID:    objectID,
		SourceType:  sourceType,
		SourceID:    cmd.SourceID,
		SourceURL:   sourceURL,
		Summary:     strings.TrimSpace(stripHTML(contentHTML)),
		Activity:    raw,
		PublishedAt: now,
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
		return result, nil
	}

	for _, follower := range followers {
		target := strings.TrimSpace(firstNonEmpty(ptrValue(follower.SharedInboxURL), follower.InboxURL))
		if target == "" {
			result.FailureCount++
			result.FailedTargets = append(result.FailedTargets, follower.ActorID)
			continue
		}
		if err := s.sendActivity(ctx, settings, baseURL, target, raw); err != nil {
			result.FailureCount++
			result.FailedTargets = append(result.FailedTargets, target)
			continue
		}
		result.SuccessCount++
	}
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
		// Minimal compatibility: ignore silently.
		return nil
	default:
		return nil
	}
}

func (s *Service) handleFollow(ctx context.Context, baseURL string, settings activitypubconfig.Settings, activity activityEnvelope, raw []byte) error {
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

func (s *Service) handleCreate(ctx context.Context, baseURL string, settings activitypubconfig.Settings, activity activityEnvelope) error {
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

	article, parent, err := s.resolveCommentTarget(ctx, baseURL, inReplyTo)
	if err != nil || article == nil || article.CommentID == nil {
		return nil
	}

	nick := extractDisplayNameFromActor(actorID)
	if remote, err := s.fetchRemoteActor(ctx, actorID); err == nil && remote != nil {
		if strings.TrimSpace(remote.PreferredUsername) != "" {
			nick = strings.TrimSpace(remote.PreferredUsername)
		}
		if strings.TrimSpace(remote.Name) != "" && nick == "" {
			nick = strings.TrimSpace(remote.Name)
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
		AreaID:            *article.CommentID,
		Content:           contentText,
		AuthorID:          nil,
		VisitorID:         strPtr(actorID),
		NickName:          strPtr(nick),
		Email:             nil,
		Website:           strPtr(strings.TrimSpace(actorID)),
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

func (s *Service) resolveCommentTarget(ctx context.Context, baseURL string, inReplyTo string) (*content.Article, *domaincomment.Comment, error) {
	if s.commentRepo != nil {
		if parent, err := s.commentRepo.FindByFederatedObjectID(ctx, inReplyTo); err == nil && parent != nil {
			area, err := s.commentRepo.GetAreaByID(ctx, parent.AreaID)
			if err != nil || area == nil || area.Type != "article" || area.ContentID == nil {
				return nil, nil, err
			}
			article, err := s.contentRepo.GetArticleByID(ctx, *area.ContentID)
			return article, parent, err
		}
	}

	if article, err := s.contentRepo.GetArticleByActivityPubObjectID(ctx, inReplyTo); err == nil {
		return article, nil, nil
	}

	if article := s.resolveArticleByLocalURL(ctx, baseURL, inReplyTo); article != nil {
		return article, nil, nil
	}
	return nil, nil, nil
}

func (s *Service) resolveArticleByLocalURL(ctx context.Context, baseURL, raw string) *content.Article {
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
		return item
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
		return item
	}
	return nil
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
	if err := validateRemoteURL(ctx, keyID); err != nil {
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
	if err := validateRemoteURL(ctx, actorID); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, actorID, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/activity+json, application/ld+json")
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

func (s *Service) sendActivity(ctx context.Context, settings activitypubconfig.Settings, baseURL string, targetURL string, payload []byte) error {
	if err := validateRemoteURL(ctx, targetURL); err != nil {
		return err
	}
	if strings.TrimSpace(settings.PrivateKey) == "" {
		return errors.New("private key not configured")
	}
	privateKey, err := parsePrivateKey(settings.PrivateKey)
	if err != nil {
		return err
	}
	algorithm := strings.TrimSpace(settings.SignatureAlg)
	if algorithm == "" {
		algorithm = "rsa-sha256"
	}
	signer, err := fedinfra.NewSigner(algorithm)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimSpace(targetURL), bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/activity+json, application/ld+json")
	req.Header.Set("Content-Type", "application/activity+json")
	keyID := actorKeyID(baseURL)
	if err := signer.SignRequest(req, payload, keyID, privateKey); err != nil {
		return err
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("deliver activity failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	return nil
}

func (s *Service) buildObjectForSource(ctx context.Context, baseURL string, sourceType string, sourceID int64, summaryOverride string, persist bool) (map[string]any, error) {
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
			"content":   renderFederatedHTML(article.Title, summary, sourceURL),
			"name":      strings.TrimSpace(article.Title),
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
		sourceURL := strings.TrimRight(baseURL, "/") + "/timeline?moment=" + moment.ShortURL
		summary := strings.TrimSpace(firstNonEmpty(summaryOverride, moment.Summary))
		return map[string]any{
			"id":        buildObjectID(baseURL, sourceType, moment.ID),
			"type":      "Note",
			"url":       sourceURL,
			"content":   renderFederatedHTML(moment.Title, summary, sourceURL),
			"name":      strings.TrimSpace(moment.Title),
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
		sourceURL := strings.TrimRight(baseURL, "/") + "/timeline?thinking=" + strconv.FormatInt(item.ID, 10)
		summary := strings.TrimSpace(firstNonEmpty(summaryOverride, stripHTML(item.Content)))
		return map[string]any{
			"id":        buildObjectID(baseURL, sourceType, item.ID),
			"type":      "Note",
			"url":       sourceURL,
			"content":   renderFederatedHTML("思考", summary, sourceURL),
			"name":      "思考",
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

func renderFederatedHTML(title, summary, sourceURL string) string {
	title = html.EscapeString(strings.TrimSpace(title))
	summary = html.EscapeString(strings.TrimSpace(summary))
	link := html.EscapeString(strings.TrimSpace(sourceURL))
	parts := make([]string, 0, 3)
	if title != "" {
		parts = append(parts, "<p><strong>"+title+"</strong></p>")
	}
	if summary != "" {
		parts = append(parts, "<p>"+summary+"</p>")
	}
	if link != "" {
		parts = append(parts, `<p><a href="`+link+`" rel="nofollow noopener noreferrer">阅读全文</a></p>`)
	}
	return strings.Join(parts, "")
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

func preferredUsername(settings activitypubconfig.Settings) string {
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

func validateRemoteURL(ctx context.Context, raw string) error {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return err
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("unsupported remote scheme")
	}
	host := strings.TrimSpace(parsed.Hostname())
	if host == "" {
		return errors.New("missing remote host")
	}
	if isBlockedHost(host) {
		return errors.New("blocked remote host")
	}
	ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return err
	}
	if len(ips) == 0 {
		return errors.New("remote host resolves to no ip")
	}
	for _, ip := range ips {
		addr, ok := netip.AddrFromSlice(ip)
		if !ok {
			return errors.New("invalid remote ip")
		}
		if isPrivateAddr(addr) {
			//return errors.New("private ip blocked")
		}
	}
	return nil
}

func isBlockedHost(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" {
		return true
	}
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return true
	}
	if strings.HasSuffix(host, ".local") {
		return true
	}
	return false
}

func isPrivateAddr(ip netip.Addr) bool {
	return ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() || ip.IsMulticast() || ip.IsUnspecified()
}
