package like

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	domainlike "github.com/grtsinry43/grtblog-v2/server/internal/domain/like"
)

type RequestMeta struct {
	IP        string
	UserAgent string
}

type TrackLikeResult struct {
	VisitorID string
	Affected  bool
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

	likeEntity := &domainlike.ContentLike{
		TargetType: targetType,
		TargetID:   cmd.ContentID,
		VisitorID:  &visitorID,
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

func normalizeTargetType(raw string) (domainlike.TargetType, error) {
	targetType := domainlike.TargetType(strings.ToLower(strings.TrimSpace(raw)))
	switch targetType {
	case domainlike.TargetArticle, domainlike.TargetMoment, domainlike.TargetPage, domainlike.TargetThinking:
		return targetType, nil
	default:
		return "", domainlike.ErrInvalidTargetType
	}
}

func fallbackVisitorID(ip, ua string, seed int64) string {
	raw := strings.TrimSpace(ip) + "|" + strings.TrimSpace(ua)
	if raw == "|" {
		raw = fmt.Sprintf("anonymous-%d", seed)
	}
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:16])
}
