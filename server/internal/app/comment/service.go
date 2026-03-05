package comment

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/sysconfig"
	domaincomment "github.com/grtsinry43/grtblog-v2/server/internal/domain/comment"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/identity"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/social"
)

const (
	defaultMaxDepth        = 3
	commentContentMaxRunes = 500
)

type RequestMeta struct {
	IP        string
	UserAgent string
}

type ClientInfo struct {
	Platform string
	Browser  string
}

type ClientInfoResolver interface {
	Resolve(userAgent string) ClientInfo
}

type GeoIPResolver interface {
	Resolve(ip string) string
}

type Service struct {
	repo           domaincomment.CommentRepository
	userRepo       identity.Repository
	friendLinkRepo social.FriendLinkRepository
	sysCfg         *sysconfig.Service
	clientInfo     ClientInfoResolver
	geoIP          GeoIPResolver
	maxDepthLimit  int
	events         appEvent.Bus
}

func NewService(
	repo domaincomment.CommentRepository,
	userRepo identity.Repository,
	friendLinkRepo social.FriendLinkRepository,
	sysCfg *sysconfig.Service,
	clientInfo ClientInfoResolver,
	geoIP GeoIPResolver,
	events appEvent.Bus,
) *Service {
	if events == nil {
		events = appEvent.NopBus{}
	}
	return &Service{
		repo:           repo,
		userRepo:       userRepo,
		friendLinkRepo: friendLinkRepo,
		sysCfg:         sysCfg,
		clientInfo:     clientInfo,
		geoIP:          geoIP,
		maxDepthLimit:  defaultMaxDepth,
		events:         events,
	}
}

type CommentNode struct {
	Comment  *domaincomment.Comment
	Children []*CommentNode
	Floor    string
}

type PublicCommentPage struct {
	Items             []*CommentNode
	Total             int64
	Page              int
	Size              int
	IsClosed          bool
	RequireModeration bool
}

func (s *Service) CreateCommentLogin(ctx context.Context, userID int64, cmd CreateCommentLoginCmd, meta RequestMeta) (*domaincomment.Comment, error) {
	if err := s.ensureCommentAllowed(ctx); err != nil {
		return nil, err
	}
	if err := s.ensureContentValid(cmd.Content); err != nil {
		return nil, err
	}
	if err := s.ensureAreaCommentable(ctx, cmd.AreaID); err != nil {
		return nil, err
	}
	if cmd.ParentID != nil {
		if err := s.ensureParentValid(ctx, cmd.AreaID, *cmd.ParentID); err != nil {
			return nil, err
		}
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	nickname := strings.TrimSpace(user.Nickname)
	if nickname == "" {
		nickname = strings.TrimSpace(user.Username)
	}
	nicknamePtr := toPtr(nickname)
	emailPtr := toPtr(strings.TrimSpace(user.Email))
	visitorID := strings.TrimSpace(cmd.VisitorID)

	isFriend := false
	if !user.IsAdmin && s.friendLinkRepo != nil {
		active, err := s.friendLinkRepo.ExistsActiveByUserID(ctx, user.ID)
		if err != nil {
			return nil, err
		}
		isFriend = active
	}

	status, isViewed, err := s.resolveCreateStatus(ctx, user.IsAdmin, &user.ID, emailPtr)
	if err != nil {
		return nil, err
	}

	commentEntity := &domaincomment.Comment{
		AreaID:    cmd.AreaID,
		Content:   strings.TrimSpace(cmd.Content),
		AuthorID:  &user.ID,
		VisitorID: toPtr(visitorID),
		NickName:  nicknamePtr,
		Email:     emailPtr,
		Website:   nil,
		IsOwner:   user.IsAdmin,
		// "本文作者" 只允许人工标记，避免将站长/登录用户自动等同为内容作者。
		IsAuthor: false,
		IsFriend: isFriend,
		IsViewed: isViewed,
		IsTop:    false,
		IsMy:     true,
		CanReply: true,
		Status:   status,
		ParentID: cmd.ParentID,
	}
	s.applyRequestMeta(commentEntity, meta)
	commentEntity.Avatar = s.resolveCommentAvatar(ctx, commentEntity, nil)

	if err := s.repo.Create(ctx, commentEntity); err != nil {
		return nil, err
	}
	_ = s.events.Publish(ctx, CommentCreated{
		ID:       commentEntity.ID,
		AreaID:   commentEntity.AreaID,
		ParentID: commentEntity.ParentID,
		AuthorID: commentEntity.AuthorID,
		NickName: toValue(commentEntity.NickName),
		Email:    toValue(commentEntity.Email),
		Content:  commentEntity.Content,
		Status:   string(commentEntity.Status),
		At:       time.Now(),
	})
	s.publishReplyEventIfNeeded(ctx, commentEntity)
	return commentEntity, nil
}

func (s *Service) CreateCommentVisitor(ctx context.Context, cmd CreateCommentVisitorCmd, meta RequestMeta) (*domaincomment.Comment, error) {
	if err := s.ensureCommentAllowed(ctx); err != nil {
		return nil, err
	}
	if err := s.ensureContentValid(cmd.Content); err != nil {
		return nil, err
	}
	if err := s.ensureAreaCommentable(ctx, cmd.AreaID); err != nil {
		return nil, err
	}
	if cmd.ParentID != nil {
		if err := s.ensureParentValid(ctx, cmd.AreaID, *cmd.ParentID); err != nil {
			return nil, err
		}
	}

	nickname := strings.TrimSpace(cmd.NickName)
	email := strings.TrimSpace(cmd.Email)
	website := strings.TrimSpace(toValue(cmd.Website))
	emailPtr := toPtr(email)
	visitorID := strings.TrimSpace(cmd.VisitorID)

	status, isViewed, err := s.resolveCreateStatus(ctx, false, nil, emailPtr)
	if err != nil {
		return nil, err
	}

	commentEntity := &domaincomment.Comment{
		AreaID:    cmd.AreaID,
		Content:   strings.TrimSpace(cmd.Content),
		AuthorID:  nil,
		VisitorID: toPtr(visitorID),
		NickName:  toPtr(nickname),
		Email:     emailPtr,
		Website:   toPtr(website),
		IsOwner:   false,
		IsAuthor:  false,
		IsFriend:  false,
		IsViewed:  isViewed,
		IsTop:     false,
		IsMy:      true,
		CanReply:  true,
		Status:    status,
		ParentID:  cmd.ParentID,
	}
	s.applyRequestMeta(commentEntity, meta)
	commentEntity.Avatar = s.resolveCommentAvatar(ctx, commentEntity, nil)

	if err := s.repo.Create(ctx, commentEntity); err != nil {
		return nil, err
	}
	_ = s.events.Publish(ctx, CommentCreated{
		ID:       commentEntity.ID,
		AreaID:   commentEntity.AreaID,
		ParentID: commentEntity.ParentID,
		AuthorID: commentEntity.AuthorID,
		NickName: toValue(commentEntity.NickName),
		Email:    toValue(commentEntity.Email),
		Content:  commentEntity.Content,
		Status:   string(commentEntity.Status),
		At:       time.Now(),
	})
	s.publishReplyEventIfNeeded(ctx, commentEntity)
	return commentEntity, nil
}

func (s *Service) ImportComment(ctx context.Context, cmd ImportCommentCmd) (*domaincomment.Comment, error) {
	if err := s.ensureAreaExists(ctx, cmd.AreaID); err != nil {
		return nil, err
	}

	content := strings.TrimSpace(cmd.Content)
	if content == "" && cmd.DeletedAt == nil {
		return nil, domaincomment.ErrCommentContentEmpty
	}

	if cmd.ParentID != nil {
		parent, err := s.repo.FindByID(ctx, *cmd.ParentID)
		if err != nil {
			if errors.Is(err, domaincomment.ErrCommentNotFound) {
				return nil, domaincomment.ErrCommentParentNotFound
			}
			return nil, err
		}
		if parent.AreaID != cmd.AreaID {
			return nil, domaincomment.ErrCommentParentNotFound
		}
	}

	status := domaincomment.CommentStatusApproved
	if cmd.Status != nil {
		status = normalizeCommentStatus(*cmd.Status)
		if status == "" {
			return nil, domaincomment.ErrCommentStatusInvalid
		}
	}

	entity := &domaincomment.Comment{
		AreaID:            cmd.AreaID,
		Content:           content,
		AuthorID:          normalizeInt64Ptr(cmd.AuthorID),
		VisitorID:         normalizePtr(cmd.VisitorID),
		NickName:          normalizePtr(cmd.NickName),
		IP:                normalizePtr(cmd.IP),
		Location:          normalizePtr(cmd.Location),
		Platform:          normalizePtr(cmd.Platform),
		Browser:           normalizePtr(cmd.Browser),
		Email:             normalizePtr(cmd.Email),
		Website:           normalizePtr(cmd.Website),
		IsOwner:           boolOrDefault(cmd.IsOwner, false),
		IsFriend:          boolOrDefault(cmd.IsFriend, false),
		IsAuthor:          boolOrDefault(cmd.IsAuthor, false),
		IsViewed:          boolOrDefault(cmd.IsViewed, false),
		IsTop:             boolOrDefault(cmd.IsTop, false),
		IsMy:              false,
		IsFederated:       boolOrDefault(cmd.IsFederated, false),
		FederatedProtocol: normalizePtr(cmd.FederatedProtocol),
		FederatedActor:    normalizePtr(cmd.FederatedActor),
		FederatedObjectID: normalizePtr(cmd.FederatedObjectID),
		CanReply:          boolOrDefault(cmd.CanReply, true),
		Status:            status,
		ParentID:          normalizeInt64Ptr(cmd.ParentID),
		CreatedAt:         timeOrZero(cmd.CreatedAt),
		UpdatedAt:         timeOrZero(cmd.UpdatedAt),
		DeletedAt:         cmd.DeletedAt,
	}

	if cmd.ID != nil && *cmd.ID > 0 {
		entity.ID = *cmd.ID
	}

	if err := s.repo.Create(ctx, entity); err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *Service) ListPublicComments(ctx context.Context, cmd ListPublicCommentsCmd) (*PublicCommentPage, error) {
	area, err := s.repo.GetAreaByID(ctx, cmd.AreaID)
	if err != nil {
		return nil, err
	}
	requireModeration := false
	if s.sysCfg != nil {
		requireModeration = s.sysCfg.CommentSettings(ctx).RequireModeration
	}
	page, size := normalizePage(cmd.Page, cmd.PageSize)

	items, err := s.repo.ListPublicByAreaID(ctx, domaincomment.PublicListOptions{
		AreaID:          cmd.AreaID,
		ViewerAuthorID:  cmd.ViewerAuthorID,
		ViewerVisitorID: strings.TrimSpace(cmd.ViewerVisitorID),
	})
	if err != nil {
		return nil, err
	}

	s.populateCommentOwnership(items, cmd.ViewerAuthorID, cmd.ViewerVisitorID)
	s.populateCommentAvatars(ctx, items)
	tree := buildCommentTree(items)
	assignFloors(tree)
	total := len(tree)
	start := (page - 1) * size
	if start >= total {
		return &PublicCommentPage{
			Items:             []*CommentNode{},
			Total:             int64(total),
			Page:              page,
			Size:              size,
			IsClosed:          area.IsClosed,
			RequireModeration: requireModeration,
		}, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return &PublicCommentPage{
		Items:             tree[start:end],
		Total:             int64(total),
		Page:              page,
		Size:              size,
		IsClosed:          area.IsClosed,
		RequireModeration: requireModeration,
	}, nil
}

func (s *Service) SetAreaClosed(ctx context.Context, areaID int64, isClosed bool) error {
	if areaID <= 0 {
		return domaincomment.ErrCommentAreaNotFound
	}
	if _, err := s.repo.GetAreaByID(ctx, areaID); err != nil {
		return err
	}
	return s.repo.SetAreaClosed(ctx, areaID, isClosed)
}

func (s *Service) ListAdminComments(ctx context.Context, cmd ListAdminCommentsCmd) ([]*domaincomment.Comment, int64, error) {
	page, size := normalizePage(cmd.Page, cmd.PageSize)
	items, total, err := s.repo.ListForAdmin(ctx, domaincomment.AdminListOptions{
		AreaID:       cmd.AreaID,
		Status:       strings.TrimSpace(cmd.Status),
		OnlyUnviewed: cmd.OnlyUnviewed,
		Page:         page,
		PageSize:     size,
	})
	if err != nil {
		return nil, 0, err
	}
	s.populateCommentAvatars(ctx, items)
	return items, total, nil
}

func (s *Service) ListAdminVisitors(ctx context.Context, cmd ListAdminVisitorsCmd) ([]domaincomment.VisitorProfile, int64, error) {
	page, size := normalizePage(cmd.Page, cmd.PageSize)
	return s.repo.ListVisitors(ctx, domaincomment.AdminVisitorListOptions{
		Keyword:  strings.TrimSpace(cmd.Keyword),
		Page:     page,
		PageSize: size,
	})
}

func (s *Service) GetVisitorProfile(ctx context.Context, cmd GetVisitorProfileCmd) (*domaincomment.VisitorProfile, []domaincomment.VisitorRecentComment, error) {
	visitorID := strings.TrimSpace(cmd.VisitorID)
	if visitorID == "" {
		return nil, nil, domaincomment.ErrVisitorNotFound
	}

	recentLimit := cmd.RecentLimit
	if recentLimit <= 0 {
		recentLimit = 20
	}
	if recentLimit > 100 {
		recentLimit = 100
	}

	return s.repo.GetVisitorProfile(ctx, visitorID, recentLimit)
}

func (s *Service) GetVisitorInsights(ctx context.Context, cmd GetVisitorInsightsCmd) (*domaincomment.VisitorInsights, error) {
	days := cmd.Days
	if days <= 0 {
		days = 30
	}
	if days > 180 {
		days = 180
	}
	return s.repo.GetVisitorInsights(ctx, days)
}

func (s *Service) MarkCommentsViewed(ctx context.Context, cmd MarkCommentsViewedCmd) error {
	if len(cmd.IDs) == 0 {
		return nil
	}
	ids := make([]int64, 0, len(cmd.IDs))
	seen := make(map[int64]struct{}, len(cmd.IDs))
	for _, id := range cmd.IDs {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil
	}
	return s.repo.SetViewedStatus(ctx, ids, cmd.IsViewed)
}

func (s *Service) ReplyComment(ctx context.Context, cmd ReplyCommentCmd) (*domaincomment.Comment, error) {
	if err := s.ensureContentValid(cmd.Content); err != nil {
		return nil, err
	}
	parent, err := s.repo.FindByID(ctx, cmd.ParentID)
	if err != nil {
		return nil, err
	}
	if parent.IsFederated {
		return nil, domaincomment.ErrCommentReplyDisabled
	}
	if err := s.ensureParentValid(ctx, parent.AreaID, parent.ID); err != nil {
		return nil, err
	}
	adminUser, err := s.userRepo.FindByID(ctx, cmd.AdminID)
	if err != nil {
		return nil, err
	}
	nickname := strings.TrimSpace(adminUser.Nickname)
	if nickname == "" {
		nickname = strings.TrimSpace(adminUser.Username)
	}

	reply := &domaincomment.Comment{
		AreaID:   parent.AreaID,
		Content:  strings.TrimSpace(cmd.Content),
		AuthorID: &adminUser.ID,
		NickName: toPtr(nickname),
		Email:    toPtr(strings.TrimSpace(adminUser.Email)),
		IsOwner:  true,
		IsFriend: false,
		// "本文作者" 只允许人工标记，管理员回复默认不自动标记为本文作者。
		IsAuthor: false,
		IsViewed: true,
		IsTop:    false,
		IsMy:     true,
		CanReply: true,
		Status:   domaincomment.CommentStatusApproved,
		ParentID: &parent.ID,
	}
	reply.Avatar = s.resolveCommentAvatar(ctx, reply, nil)
	if err := s.repo.Create(ctx, reply); err != nil {
		return nil, err
	}
	if shouldSkipReplyNotification(parent, adminUser) {
		return reply, nil
	}
	payload := s.buildCommentReplyPayload(ctx, parent, reply)
	_ = s.events.Publish(ctx, appEvent.Generic{
		EventName: "comment.reply",
		At:        time.Now(),
		Payload:   payload,
	})
	return reply, nil
}

func shouldSkipReplyNotification(parent *domaincomment.Comment, replier *identity.User) bool {
	if parent == nil || replier == nil {
		return false
	}
	replierAuthorID := &replier.ID
	replierEmail := toPtr(strings.TrimSpace(replier.Email))
	return shouldSkipReplyNotificationByIdentity(parent, replierAuthorID, replierEmail)
}

func shouldSkipReplyNotificationByIdentity(parent *domaincomment.Comment, replierAuthorID *int64, replierEmail *string) bool {
	if parent == nil {
		return false
	}
	if parent.AuthorID != nil && replierAuthorID != nil && *parent.AuthorID == *replierAuthorID {
		return true
	}
	parentEmail := strings.TrimSpace(toValue(parent.Email))
	replyEmail := strings.TrimSpace(toValue(replierEmail))
	if parentEmail != "" && replyEmail != "" && strings.EqualFold(parentEmail, replyEmail) {
		return true
	}
	return false
}

// publishReplyEventIfNeeded publishes a comment.reply event when a newly created
// comment is a reply (has ParentID) and is immediately approved. Comments that
// are pending moderation will have the event published later via UpdateCommentStatus.
func (s *Service) publishReplyEventIfNeeded(ctx context.Context, comment *domaincomment.Comment) {
	if comment.ParentID == nil {
		return
	}
	if comment.Status != domaincomment.CommentStatusApproved {
		return
	}
	parent, err := s.repo.FindByID(ctx, *comment.ParentID)
	if err != nil {
		return
	}
	if shouldSkipReplyNotificationByIdentity(parent, comment.AuthorID, comment.Email) {
		return
	}
	payload := s.buildCommentReplyPayload(ctx, parent, comment)
	_ = s.events.Publish(ctx, appEvent.Generic{
		EventName: "comment.reply",
		At:        time.Now(),
		Payload:   payload,
	})
}

func (s *Service) buildCommentReplyPayload(ctx context.Context, parent *domaincomment.Comment, reply *domaincomment.Comment) map[string]any {
	contentType, contentTitle, viewURL := s.loadReplyContentMeta(ctx, reply.AreaID)
	parentID := int64(0)
	if parent != nil {
		parentID = parent.ID
	}
	return map[string]any{
		"ID":             reply.ID,
		"ParentID":       parentID,
		"AreaID":         reply.AreaID,
		"ContentType":    contentType,
		"ContentTitle":   contentTitle,
		"viewUrl":        viewURL,
		"ParentContent":  toCommentContent(parent),
		"ReplyContent":   strings.TrimSpace(reply.Content),
		"ParentNickName": toCommentNickName(parent),
		"ReplyNickName":  toValue(reply.NickName),
		"recipientEmail": toCommentEmail(parent),
		"Status":         strings.TrimSpace(reply.Status),
	}
}

func (s *Service) loadReplyContentMeta(ctx context.Context, areaID int64) (string, string, string) {
	if areaID <= 0 {
		return "", "", ""
	}
	area, err := s.repo.GetAreaByID(ctx, areaID)
	if err != nil || area == nil {
		return "", "", ""
	}

	rawContentType := strings.TrimSpace(area.Type)
	contentType := commentAreaTypeLabel(rawContentType)
	contentTitle := ""
	viewURL := ""
	if area.ContentID != nil && *area.ContentID > 0 && rawContentType != "" {
		if title, titleErr := s.repo.GetContentTitleByTypeAndID(ctx, rawContentType, *area.ContentID); titleErr == nil {
			contentTitle = title
		}
		if path, pathErr := s.repo.GetContentViewPathByTypeAndID(ctx, rawContentType, *area.ContentID); pathErr == nil {
			viewURL = path
		}
	}
	return contentType, contentTitle, viewURL
}

func toCommentContent(item *domaincomment.Comment) string {
	if item == nil {
		return ""
	}
	return strings.TrimSpace(item.Content)
}

func toCommentNickName(item *domaincomment.Comment) string {
	if item == nil {
		return ""
	}
	return toValue(item.NickName)
}

func toCommentEmail(item *domaincomment.Comment) string {
	if item == nil {
		return ""
	}
	return toValue(item.Email)
}

func commentAreaTypeLabel(areaType string) string {
	switch strings.ToLower(strings.TrimSpace(areaType)) {
	case "article":
		return "文章"
	case "moment":
		return "手记"
	case "page":
		return "页面"
	case "thinking":
		return "思考"
	default:
		return strings.TrimSpace(areaType)
	}
}

func (s *Service) UpdateCommentStatus(ctx context.Context, cmd UpdateCommentStatusCmd) error {
	commentEntity, err := s.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}
	status := normalizeCommentStatus(cmd.Status)
	if status == "" {
		return domaincomment.ErrCommentStatusInvalid
	}
	if err := s.repo.UpdateStatus(ctx, cmd.ID, status); err != nil {
		return err
	}
	eventName := "comment.updated"
	if status == domaincomment.CommentStatusBlocked {
		eventName = "comment.blocked"
	}
	_ = s.events.Publish(ctx, appEvent.Generic{
		EventName: eventName,
		At:        time.Now(),
		Payload: map[string]any{
			"ID":     cmd.ID,
			"AreaID": commentEntity.AreaID,
			"Status": status,
		},
	})
	if status == domaincomment.CommentStatusApproved &&
		commentEntity.Status != domaincomment.CommentStatusApproved &&
		commentEntity.ParentID != nil {
		if parent, parentErr := s.repo.FindByID(ctx, *commentEntity.ParentID); parentErr == nil &&
			!shouldSkipReplyNotificationByIdentity(parent, commentEntity.AuthorID, commentEntity.Email) {
			payload := s.buildCommentReplyPayload(ctx, parent, commentEntity)
			_ = s.events.Publish(ctx, appEvent.Generic{
				EventName: "comment.reply",
				At:        time.Now(),
				Payload:   payload,
			})
		}
	}
	return nil
}

func (s *Service) SetCommentAuthor(ctx context.Context, cmd SetCommentAuthorCmd) error {
	if _, err := s.repo.FindByID(ctx, cmd.ID); err != nil {
		return err
	}
	return s.repo.SetAuthorStatus(ctx, cmd.ID, cmd.IsAuthor)
}

func (s *Service) SetCommentTop(ctx context.Context, cmd SetCommentTopCmd) error {
	if _, err := s.repo.FindByID(ctx, cmd.ID); err != nil {
		return err
	}
	return s.repo.SetTopStatus(ctx, cmd.ID, cmd.IsTop)
}

func (s *Service) ensureOwnership(comment *domaincomment.Comment, viewerAuthorID *int64, viewerVisitorID string) error {
	if viewerAuthorID != nil && comment.AuthorID != nil && *viewerAuthorID == *comment.AuthorID {
		return nil
	}
	visitorID := strings.TrimSpace(viewerVisitorID)
	if visitorID != "" {
		if itemVisitorID := strings.TrimSpace(toValue(comment.VisitorID)); itemVisitorID != "" && itemVisitorID == visitorID {
			return nil
		}
	}
	return domaincomment.ErrCommentNotOwner
}

func (s *Service) EditComment(ctx context.Context, cmd EditCommentCmd) (*domaincomment.Comment, error) {
	if err := s.ensureContentValid(cmd.Content); err != nil {
		return nil, err
	}
	commentEntity, err := s.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}
	if commentEntity.DeletedAt != nil {
		return nil, domaincomment.ErrCommentAlreadyDeleted
	}
	if err := s.ensureOwnership(commentEntity, cmd.ViewerAuthorID, cmd.ViewerVisitorID); err != nil {
		return nil, err
	}

	newContent := strings.TrimSpace(cmd.Content)
	if newContent != commentEntity.Content {
		commentEntity.IsEdited = true
	}
	commentEntity.Content = newContent

	if s.sysCfg != nil && s.sysCfg.CommentSettings(ctx).RequireModeration &&
		commentEntity.Status == domaincomment.CommentStatusApproved {
		commentEntity.Status = domaincomment.CommentStatusPending
	}

	if err := s.repo.Update(ctx, commentEntity); err != nil {
		return nil, err
	}
	_ = s.events.Publish(ctx, appEvent.Generic{
		EventName: "comment.edited",
		At:        time.Now(),
		Payload: map[string]any{
			"ID":     commentEntity.ID,
			"AreaID": commentEntity.AreaID,
			"Status": string(commentEntity.Status),
		},
	})
	return commentEntity, nil
}

func (s *Service) DeleteOwnComment(ctx context.Context, cmd DeleteOwnCommentCmd) error {
	commentEntity, err := s.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}
	if commentEntity.DeletedAt != nil {
		return domaincomment.ErrCommentAlreadyDeleted
	}
	if err := s.ensureOwnership(commentEntity, cmd.ViewerAuthorID, cmd.ViewerVisitorID); err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, cmd.ID); err != nil {
		return err
	}
	_ = s.events.Publish(ctx, appEvent.Generic{
		EventName: "comment.deleted",
		At:        time.Now(),
		Payload: map[string]any{
			"ID":     commentEntity.ID,
			"AreaID": commentEntity.AreaID,
			"Status": string(commentEntity.Status),
		},
	})
	return nil
}

func (s *Service) DeleteComment(ctx context.Context, id int64) error {
	commentEntity, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	_ = s.events.Publish(ctx, appEvent.Generic{
		EventName: "comment.deleted",
		At:        time.Now(),
		Payload: map[string]any{
			"ID":     commentEntity.ID,
			"AreaID": commentEntity.AreaID,
			"Status": string(commentEntity.Status),
		},
	})
	return nil
}

func (s *Service) applyRequestMeta(commentEntity *domaincomment.Comment, meta RequestMeta) {
	ip := strings.TrimSpace(meta.IP)
	if ip != "" {
		commentEntity.IP = &ip
	}
	if s.clientInfo != nil {
		info := s.clientInfo.Resolve(meta.UserAgent)
		if strings.TrimSpace(info.Platform) != "" {
			commentEntity.Platform = toPtr(info.Platform)
		}
		if strings.TrimSpace(info.Browser) != "" {
			commentEntity.Browser = toPtr(info.Browser)
		}
	}
	if s.geoIP != nil {
		location := strings.TrimSpace(s.geoIP.Resolve(ip))
		if location != "" {
			commentEntity.Location = &location
		}
	}
}

func (s *Service) ensureCommentAllowed(ctx context.Context) error {
	if s.sysCfg == nil {
		return nil
	}
	settings := s.sysCfg.CommentSettings(ctx)
	if settings.Disabled {
		return domaincomment.ErrCommentDisabled
	}
	return nil
}

func (s *Service) resolveCreateStatus(ctx context.Context, isAdmin bool, authorID *int64, email *string) (status string, isViewed bool, err error) {
	if !isAdmin {
		blocked, err := s.repo.ExistsBlockedIdentity(ctx, authorID, email)
		if err != nil {
			return "", false, err
		}
		if blocked {
			return "", false, domaincomment.ErrCommentBlocked
		}
	}

	if isAdmin {
		return domaincomment.CommentStatusApproved, true, nil
	}

	if s.sysCfg != nil && s.sysCfg.CommentSettings(ctx).RequireModeration {
		return domaincomment.CommentStatusPending, false, nil
	}
	return domaincomment.CommentStatusApproved, false, nil
}

func (s *Service) ensureContentValid(content string) error {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return domaincomment.ErrCommentContentEmpty
	}
	if utf8.RuneCountInString(trimmed) > commentContentMaxRunes {
		return domaincomment.ErrCommentContentTooLong
	}
	return nil
}

func (s *Service) ensureAreaExists(ctx context.Context, areaID int64) error {
	if areaID <= 0 {
		return domaincomment.ErrCommentAreaNotFound
	}
	_, err := s.repo.GetAreaByID(ctx, areaID)
	if err != nil {
		if errors.Is(err, domaincomment.ErrCommentAreaNotFound) {
			return err
		}
		return err
	}
	return nil
}

func (s *Service) ensureAreaCommentable(ctx context.Context, areaID int64) error {
	if areaID <= 0 {
		return domaincomment.ErrCommentAreaNotFound
	}
	area, err := s.repo.GetAreaByID(ctx, areaID)
	if err != nil {
		if errors.Is(err, domaincomment.ErrCommentAreaNotFound) {
			return err
		}
		return err
	}
	if area.IsClosed {
		return domaincomment.ErrCommentAreaClosed
	}
	return nil
}

func (s *Service) ensureParentValid(ctx context.Context, areaID int64, parentID int64) error {
	parent, err := s.repo.FindByID(ctx, parentID)
	if err != nil {
		if errors.Is(err, domaincomment.ErrCommentNotFound) {
			return domaincomment.ErrCommentParentNotFound
		}
		return err
	}
	if parent.AreaID != areaID {
		return domaincomment.ErrCommentParentNotFound
	}
	if !parent.CanReply {
		return domaincomment.ErrCommentReplyDisabled
	}

	chainLength := 1
	current := parent
	for current.ParentID != nil {
		if chainLength+1 >= s.maxDepthLimit {
			return domaincomment.ErrCommentTooDeep
		}
		next, err := s.repo.FindByID(ctx, *current.ParentID)
		if err != nil {
			if errors.Is(err, domaincomment.ErrCommentNotFound) {
				return domaincomment.ErrCommentParentNotFound
			}
			return err
		}
		chainLength++
		current = next
	}
	if chainLength+1 > s.maxDepthLimit {
		return domaincomment.ErrCommentTooDeep
	}
	return nil
}

func (s *Service) populateCommentOwnership(items []*domaincomment.Comment, viewerAuthorID *int64, viewerVisitorID string) {
	visitorID := strings.TrimSpace(viewerVisitorID)
	for _, item := range items {
		if item == nil {
			continue
		}

		item.IsMy = false
		if viewerAuthorID != nil && item.AuthorID != nil && *viewerAuthorID == *item.AuthorID {
			item.IsMy = true
			continue
		}

		if visitorID != "" {
			if itemVisitorID := strings.TrimSpace(toValue(item.VisitorID)); itemVisitorID != "" && itemVisitorID == visitorID {
				item.IsMy = true
			}
		}
	}
}

func buildCommentTree(items []*domaincomment.Comment) []*CommentNode {
	nodes := make(map[int64]*CommentNode, len(items))
	for _, item := range items {
		nodes[item.ID] = &CommentNode{Comment: item}
	}

	var roots []*CommentNode
	for _, item := range items {
		node := nodes[item.ID]
		if item.ParentID != nil {
			if parent, ok := nodes[*item.ParentID]; ok {
				parent.Children = append(parent.Children, node)
				continue
			}
		}
		roots = append(roots, node)
	}
	return roots
}

// assignFloors assigns chronological floor numbers to the comment tree.
// Root comments get sequential floors (1, 2, 3…) based on creation time.
// Children are reordered chronologically and get sub-floors (e.g. 1-1, 1-2).
// After assignment, roots are sorted back to display order (pinned first, newest first).
func assignFloors(roots []*CommentNode) {
	if len(roots) == 0 {
		return
	}

	// Sort roots chronologically (oldest first) to assign floor numbers.
	sort.SliceStable(roots, func(i, j int) bool {
		ci, cj := roots[i].Comment, roots[j].Comment
		if !ci.CreatedAt.Equal(cj.CreatedAt) {
			return ci.CreatedAt.Before(cj.CreatedAt)
		}
		return ci.ID < cj.ID
	})

	for i, root := range roots {
		root.Floor = fmt.Sprintf("%d", i+1)
		assignChildFloors(root)
	}

	// Sort back to display order: pinned first, then newest first.
	sort.SliceStable(roots, func(i, j int) bool {
		ci, cj := roots[i].Comment, roots[j].Comment
		if ci.IsTop != cj.IsTop {
			return ci.IsTop
		}
		if !ci.CreatedAt.Equal(cj.CreatedAt) {
			return ci.CreatedAt.After(cj.CreatedAt)
		}
		return ci.ID > cj.ID
	})
}

func assignChildFloors(node *CommentNode) {
	if len(node.Children) == 0 {
		return
	}
	// Sort children chronologically (oldest first).
	sort.SliceStable(node.Children, func(i, j int) bool {
		ci, cj := node.Children[i].Comment, node.Children[j].Comment
		if !ci.CreatedAt.Equal(cj.CreatedAt) {
			return ci.CreatedAt.Before(cj.CreatedAt)
		}
		return ci.ID < cj.ID
	})
	for i, child := range node.Children {
		child.Floor = fmt.Sprintf("%s-%d", node.Floor, i+1)
		assignChildFloors(child)
	}
}

type commentAuthorSnapshot struct {
	avatar string
	email  string
	found  bool
}

func (s *Service) populateCommentAvatars(ctx context.Context, items []*domaincomment.Comment) {
	if len(items) == 0 {
		return
	}
	cache := make(map[int64]commentAuthorSnapshot)
	for _, item := range items {
		if item == nil {
			continue
		}
		item.Avatar = s.resolveCommentAvatar(ctx, item, cache)
	}
}

func (s *Service) resolveCommentAvatar(ctx context.Context, item *domaincomment.Comment, cache map[int64]commentAuthorSnapshot) *string {
	if item == nil {
		return nil
	}

	email := strings.TrimSpace(toValue(item.Email))
	if item.AuthorID != nil && s.userRepo != nil {
		uid := *item.AuthorID
		var (
			info commentAuthorSnapshot
			ok   bool
		)
		if cache != nil {
			info, ok = cache[uid]
		}
		if !ok {
			user, err := s.userRepo.FindByID(ctx, uid)
			if err == nil && user != nil {
				info = commentAuthorSnapshot{
					avatar: strings.TrimSpace(user.Avatar),
					email:  strings.TrimSpace(user.Email),
					found:  true,
				}
			} else {
				info = commentAuthorSnapshot{found: false}
			}
			if cache != nil {
				cache[uid] = info
			}
		}
		if info.found {
			if info.avatar != "" {
				return toPtr(info.avatar)
			}
			if email == "" {
				email = info.email
			}
		}
	}

	return buildCavatarURL(email)
}

func buildCavatarURL(email string) *string {
	normalized := strings.ToLower(strings.TrimSpace(email))
	if normalized == "" {
		return nil
	}
	hash := md5.Sum([]byte(normalized))
	return toPtr(fmt.Sprintf("https://cravatar.cn/avatar/%s?d=mp&s=240", hex.EncodeToString(hash[:])))
}

func normalizePage(page, size int) (int, int) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	if size > 50 {
		size = 50
	}
	return page, size
}

func normalizeCommentStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case domaincomment.CommentStatusPending:
		return domaincomment.CommentStatusPending
	case domaincomment.CommentStatusApproved:
		return domaincomment.CommentStatusApproved
	case domaincomment.CommentStatusRejected:
		return domaincomment.CommentStatusRejected
	case domaincomment.CommentStatusBlocked:
		return domaincomment.CommentStatusBlocked
	default:
		return ""
	}
}

func toPtr(val string) *string {
	trimmed := strings.TrimSpace(val)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func toValue(val *string) string {
	if val == nil {
		return ""
	}
	return *val
}

func normalizePtr(val *string) *string {
	if val == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*val)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeInt64Ptr(val *int64) *int64 {
	if val == nil {
		return nil
	}
	if *val <= 0 {
		return nil
	}
	v := *val
	return &v
}

func boolOrDefault(val *bool, fallback bool) bool {
	if val == nil {
		return fallback
	}
	return *val
}

func timeOrZero(val *time.Time) time.Time {
	if val == nil {
		return time.Time{}
	}
	return *val
}
