package navigation

import (
	"context"
	"strings"
	"time"

	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	domain "github.com/grtsinry43/grtblog-v2/server/internal/domain/navigation"
)

type Service struct {
	repo   domain.Repository
	events appEvent.Bus
}

func NewService(repo domain.Repository, events appEvent.Bus) *Service {
	if events == nil {
		events = appEvent.NopBus{}
	}
	return &Service{
		repo:   repo,
		events: events,
	}
}

func (s *Service) List(ctx context.Context) ([]*domain.NavMenu, error) {
	return s.repo.List(ctx)
}

func (s *Service) Create(ctx context.Context, cmd CreateNavMenuCmd) (*domain.NavMenu, error) {
	if cmd.ParentID != nil {
		if _, err := s.repo.GetByID(ctx, *cmd.ParentID); err != nil {
			return nil, err
		}
	}

	sort, err := s.repo.NextSort(ctx, cmd.ParentID)
	if err != nil {
		return nil, err
	}

	icon, err := normalizeIcon(cmd.Icon)
	if err != nil {
		return nil, err
	}

	menu := &domain.NavMenu{
		Name:     strings.TrimSpace(cmd.Name),
		URL:      strings.TrimSpace(cmd.URL),
		Icon:     icon,
		Sort:     sort,
		ParentID: cmd.ParentID,
	}

	if err := s.repo.Create(ctx, menu); err != nil {
		return nil, err
	}
	s.publishMenuUpdated(ctx, "create", menu.ID)

	return menu, nil
}

func (s *Service) Update(ctx context.Context, cmd UpdateNavMenuCmd) (*domain.NavMenu, error) {
	menu, err := s.repo.GetByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	menu.Name = strings.TrimSpace(cmd.Name)
	menu.URL = strings.TrimSpace(cmd.URL)

	if cmd.Icon != nil {
		icon, err := normalizeIcon(cmd.Icon)
		if err != nil {
			return nil, err
		}
		menu.Icon = icon
	}

	if cmd.ParentID != nil {
		if menu.ParentID == nil || *menu.ParentID != *cmd.ParentID {
			if _, err := s.repo.GetByID(ctx, *cmd.ParentID); err != nil {
				return nil, err
			}
			menu.ParentID = cmd.ParentID
			if cmd.Sort != nil {
				menu.Sort = *cmd.Sort
			} else {
				nextSort, err := s.repo.NextSort(ctx, cmd.ParentID)
				if err != nil {
					return nil, err
				}
				menu.Sort = nextSort
			}
		}
	} else if cmd.Sort != nil {
		menu.Sort = *cmd.Sort
	}

	if err := s.repo.Update(ctx, menu); err != nil {
		return nil, err
	}
	s.publishMenuUpdated(ctx, "update", menu.ID)

	return menu, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	s.publishMenuUpdated(ctx, "delete", id)
	return nil
}

func (s *Service) UpdateOrder(ctx context.Context, items []NavMenuOrderItem) error {
	updates := make([]domain.NavMenuOrderUpdate, 0, len(items))
	for _, item := range items {
		updates = append(updates, domain.NavMenuOrderUpdate{
			ID:       item.ID,
			ParentID: item.ParentID,
			Sort:     item.Sort,
		})
	}
	if err := s.repo.UpdateOrder(ctx, updates); err != nil {
		return err
	}
	s.publishMenuUpdated(ctx, "reorder", 0)
	return nil
}

func (s *Service) publishMenuUpdated(ctx context.Context, action string, id int64) {
	_ = s.events.Publish(ctx, appEvent.Generic{
		EventName: "navmenu.updated",
		At:        time.Now(),
		Payload: map[string]any{
			"action": strings.TrimSpace(action),
			"id":     id,
		},
	})
}

func normalizeIcon(icon *string) (*string, error) {
	if icon == nil {
		return nil, nil
	}
	value := strings.TrimSpace(*icon)
	if value == "" {
		return nil, nil
	}

	if _, ok := navMenuIconWhitelist[value]; !ok {
		return nil, domain.ErrInvalidNavMenuIcon
	}
	return &value, nil
}

var navMenuIconWhitelist = map[string]struct{}{
	"house":     {},
	"book-open": {},
	"pen-tool":  {},
	"archive":   {},
	"image":     {},
	"user":      {},
	"terminal":  {},
	"coffee":    {},
	"sparkles":  {},
	"code":      {},
	"list":      {},
}
