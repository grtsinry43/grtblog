package friendlink

import (
	"context"

	"github.com/grtsinry43/grtblog-v2/server/internal/domain/social"
)

type LinkService struct {
	repo social.FriendLinkRepository
}

func NewLinkService(repo social.FriendLinkRepository) *LinkService {
	return &LinkService{repo: repo}
}

func (s *LinkService) List(ctx context.Context, options FriendLinkListOptions) ([]social.FriendLink, int64, error) {
	return s.repo.List(ctx, social.FriendLinkListOptions{
		IsActive: options.IsActive,
		Kind:     options.Kind,
		SyncMode: options.SyncMode,
		Keyword:  options.Keyword,
		Page:     options.Page,
		PageSize: options.PageSize,
	})
}
