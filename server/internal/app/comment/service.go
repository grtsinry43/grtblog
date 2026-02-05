package comment

import (
	"context"
	"errors"
	"strings"
	"time"

	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/sysconfig"
	domaincomment "github.com/grtsinry43/grtblog-v2/server/internal/domain/comment"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/identity"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/social"
)

const defaultMaxDepth = 3

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
}

type PublicCommentPage struct {
	Items    []*CommentNode
	Total    int64
	Page     int
	Size     int
	IsClosed bool
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
		AreaID:   cmd.AreaID,
		Content:  strings.TrimSpace(cmd.Content),
		AuthorID: &user.ID,
		NickName: nicknamePtr,
		Email:    emailPtr,
		Website:  nil,
		IsOwner:  user.IsAdmin,
		IsAuthor: user.IsAdmin,
		IsFriend: isFriend,
		IsViewed: isViewed,
		IsTop:    false,
		Status:   status,
		ParentID: cmd.ParentID,
	}
	s.applyRequestMeta(commentEntity, meta)

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

	status, isViewed, err := s.resolveCreateStatus(ctx, false, nil, emailPtr)
	if err != nil {
		return nil, err
	}

	commentEntity := &domaincomment.Comment{
		AreaID:   cmd.AreaID,
		Content:  strings.TrimSpace(cmd.Content),
		AuthorID: nil,
		NickName: toPtr(nickname),
		Email:    emailPtr,
		Website:  toPtr(website),
		IsOwner:  false,
		IsAuthor: false,
		IsFriend: false,
		IsViewed: isViewed,
		IsTop:    false,
		Status:   status,
		ParentID: cmd.ParentID,
	}
	s.applyRequestMeta(commentEntity, meta)

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
	return commentEntity, nil
}

func (s *Service) ListPublicComments(ctx context.Context, cmd ListPublicCommentsCmd) (*PublicCommentPage, error) {
	area, err := s.repo.GetAreaByID(ctx, cmd.AreaID)
	if err != nil {
		return nil, err
	}
	page, size := normalizePage(cmd.Page, cmd.PageSize)

	items, err := s.repo.ListPublicByAreaID(ctx, domaincomment.PublicListOptions{AreaID: cmd.AreaID})
	if err != nil {
		return nil, err
	}
	tree := buildCommentTree(items)
	total := len(tree)
	start := (page - 1) * size
	if start >= total {
		return &PublicCommentPage{
			Items:    []*CommentNode{},
			Total:    int64(total),
			Page:     page,
			Size:     size,
			IsClosed: area.IsClosed,
		}, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return &PublicCommentPage{
		Items:    tree[start:end],
		Total:    int64(total),
		Page:     page,
		Size:     size,
		IsClosed: area.IsClosed,
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
	return items, total, nil
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
		IsAuthor: true,
		IsViewed: true,
		IsTop:    false,
		Status:   domaincomment.CommentStatusApproved,
		ParentID: &parent.ID,
	}
	if err := s.repo.Create(ctx, reply); err != nil {
		return nil, err
	}
	if shouldSkipReplyNotification(parent, adminUser) {
		return reply, nil
	}
	_ = s.events.Publish(ctx, appEvent.Generic{
		EventName: "comment.reply",
		At:        time.Now(),
		Payload: map[string]any{
			"ID":             reply.ID,
			"ParentID":       parent.ID,
			"AreaID":         reply.AreaID,
			"ParentContent":  parent.Content,
			"ReplyContent":   reply.Content,
			"ParentNickName": toValue(parent.NickName),
			"ReplyNickName":  toValue(reply.NickName),
			"recipientEmail": toValue(parent.Email),
			"Status":         string(reply.Status),
		},
	})
	return reply, nil
}

func shouldSkipReplyNotification(parent *domaincomment.Comment, replier *identity.User) bool {
	if parent == nil || replier == nil {
		return false
	}
	if parent.IsOwner {
		return true
	}
	if parent.AuthorID != nil && *parent.AuthorID == replier.ID {
		return true
	}
	parentEmail := ""
	if parent.Email != nil {
		parentEmail = strings.TrimSpace(*parent.Email)
	}
	replierEmail := strings.TrimSpace(replier.Email)
	if parentEmail != "" && replierEmail != "" && strings.EqualFold(parentEmail, replierEmail) {
		return true
	}
	return false
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
			return domaincomment.CommentStatusRejected, false, nil
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
	if strings.TrimSpace(content) == "" {
		return domaincomment.ErrCommentContentEmpty
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
