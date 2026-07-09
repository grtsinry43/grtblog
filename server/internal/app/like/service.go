package like

import (
	"context"
	"fmt"
	"strings"
	"time"

	domainlike "github.com/grtsinry43/grtblog-v2/server/internal/domain/like"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/clientinfo"
)

type RequestMeta struct {
	IP        string
	UserAgent string
}

type TrackLikeResult struct {
	VisitorID string
	Affected  bool
}

type ImportLikeBatchResult struct {
	Inserted int64
}

type Service struct {
	repo domainlike.Repository
	now  func() time.Time
}

func NewService(repo domainlike.Repository) *Service {
	return &Service{
		repo: repo,
		now:  time.Now,
	}
}

func (s *Service) TrackLike(ctx context.Context, cmd TrackLikeCmd, meta RequestMeta) (*TrackLikeResult, error) {
	targetType, err := normalizeTargetType(cmd.ContentType)
	if err != nil {
		return nil, err
	}
	if cmd.ContentID <= 0 {
		return nil, domainlike.ErrInvalidTargetID
	}

	exists, err := s.repo.ExistsTarget(ctx, targetType, cmd.ContentID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, domainlike.ErrTargetNotFound
	}

	visitorID := strings.TrimSpace(cmd.VisitorID)
	if visitorID == "" {
		visitorID = fallbackVisitorID(meta.IP, meta.UserAgent, s.now().UnixNano())
	}
	clientFP := clientinfo.ClientFingerprint(meta.IP, meta.UserAgent)

	likeEntity := &domainlike.ContentLike{
		TargetType: targetType,
		TargetID:   cmd.ContentID,
		VisitorID:  &visitorID,
		ClientFP:   clientFP,
	}
	liked, err := s.repo.CreateIfAbsent(ctx, likeEntity)
	if err != nil {
		return nil, err
	}

	return &TrackLikeResult{
		VisitorID: visitorID,
		Affected:  liked,
	}, nil
}

func (s *Service) ImportLikeBatch(ctx context.Context, cmd ImportLikeBatchCmd) (*ImportLikeBatchResult, error) {
	targetType, err := normalizeTargetType(cmd.ContentType)
	if err != nil {
		return nil, err
	}
	if cmd.ContentID <= 0 {
		return nil, domainlike.ErrInvalidTargetID
	}

	exists, err := s.repo.ExistsTarget(ctx, targetType, cmd.ContentID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, domainlike.ErrTargetNotFound
	}

	entities := make([]*domainlike.ContentLike, 0, len(cmd.VisitorIDs))
	for _, raw := range cmd.VisitorIDs {
		visitorID := strings.TrimSpace(raw)
		if visitorID == "" {
			continue
		}
		entities = append(entities, &domainlike.ContentLike{
			TargetType: targetType,
			TargetID:   cmd.ContentID,
			VisitorID:  &visitorID,
		})
	}

	inserted, err := s.repo.CreateBatchIfAbsent(ctx, entities)
	if err != nil {
		return nil, err
	}
	return &ImportLikeBatchResult{
		Inserted: inserted,
	}, nil
}

func normalizeTargetType(raw string) (domainlike.TargetType, error) {
	targetType := domainlike.TargetType(strings.ToLower(strings.TrimSpace(raw)))
	switch targetType {
	case domainlike.TargetArticle, domainlike.TargetMoment, domainlike.TargetPage, domainlike.TargetThinking, domainlike.TargetAlbum:
		return targetType, nil
	default:
		return "", domainlike.ErrInvalidTargetType
	}
}

func fallbackVisitorID(ip, ua string, seed int64) string {
	fp := clientinfo.ClientFingerprint(ip, ua)
	if strings.TrimSpace(ip) == "" && strings.TrimSpace(ua) == "" {
		return fmt.Sprintf("anonymous-%d", seed)
	}
	return fp
}
