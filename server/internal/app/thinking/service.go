package thinking

import (
	"context"
	"time"

	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/comment"
	domainthinking "github.com/grtsinry43/grtblog-v2/server/internal/domain/thinking"
)

type Service struct {
	repo        domainthinking.ThinkingRepository
	commentRepo comment.CommentRepository
	events      appEvent.Bus
}

func NewService(repo domainthinking.ThinkingRepository, commentRepo comment.CommentRepository, events appEvent.Bus) *Service {
	if events == nil {
		events = appEvent.NopBus{}
	}
	return &Service{
		repo:        repo,
		commentRepo: commentRepo,
		events:      events,
	}
}

func (s *Service) Create(ctx context.Context, cmd CreateThinkingCmd) (*domainthinking.Thinking, error) {
	if cmd.Content == "" {
		return nil, domainthinking.ErrThinkingContentEmpty
	}

	t := &domainthinking.Thinking{
		Content:  cmd.Content,
		AuthorID: cmd.AuthorID,
	}
	if err := s.repo.Create(ctx, t); err != nil {
		return nil, err
	}
	if cmd.AllowComment != nil {
		if err := s.commentRepo.SetAreaClosed(ctx, t.CommentID, !*cmd.AllowComment); err != nil {
			return nil, err
		}
	}
	_ = s.events.Publish(ctx, ThinkingCreated{
		ID:       t.ID,
		AuthorID: t.AuthorID,
		Content:  t.Content,
		At:       time.Now(),
	})

	return t, nil
}

func (s *Service) Update(ctx context.Context, cmd UpdateThinkingCmd) (*domainthinking.Thinking, error) {
	if cmd.Content == "" {
		return nil, domainthinking.ErrThinkingContentEmpty
	}
	t, err := s.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}
	t.Content = cmd.Content
	if err := s.repo.Update(ctx, t); err != nil {
		return nil, err
	}
	if cmd.AllowComment != nil {
		if err := s.commentRepo.SetAreaClosed(ctx, t.CommentID, !*cmd.AllowComment); err != nil {
			return nil, err
		}
	}
	_ = s.events.Publish(ctx, appEvent.Generic{
		EventName: "thinking.updated",
		At:        time.Now(),
		Payload: map[string]any{
			"ID":       t.ID,
			"AuthorID": t.AuthorID,
			"Content":  t.Content,
		},
	})
	return t, nil
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]*domainthinking.Thinking, int64, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	t, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, t.ID); err != nil {
		return err
	}
	_ = s.events.Publish(ctx, appEvent.Generic{
		EventName: "thinking.deleted",
		At:        time.Now(),
		Payload: map[string]any{
			"ID":       t.ID,
			"AuthorID": t.AuthorID,
		},
	})
	return nil
}

// BatchDelete 批量删除思考。
func (s *Service) BatchDelete(ctx context.Context, cmd BatchDeleteCmd) error {
	ids := normalizeIDs(cmd.IDs)
	for _, id := range ids {
		if err := s.Delete(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) FindByID(ctx context.Context, id int64) (*domainthinking.Thinking, error) {
	_ = s.repo.IncView(ctx, id)
	return s.repo.FindByID(ctx, id)
}

func normalizeIDs(ids []int64) []int64 {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
