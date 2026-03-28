package persistence

import (
	"context"
	"strings"
	"time"

	"github.com/grtsinry43/grtblog-v2/server/internal/domain/navigation"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence/model"
	"gorm.io/gorm"
)

type NavMenuRepository struct {
	repo *GormRepository[model.NavMenu]
	db   *gorm.DB
}

func NewNavMenuRepository(db *gorm.DB) *NavMenuRepository {
	return &NavMenuRepository{
		repo: NewGormRepository[model.NavMenu](db),
		db:   db,
	}
}

func (r *NavMenuRepository) List(ctx context.Context) ([]*navigation.NavMenu, error) {
	items, err := r.repo.List(ctx, func(db *gorm.DB) *gorm.DB {
		return db.Order("parent_id").Order("sort").Order("id")
	})
	if err != nil {
		return nil, err
	}
	result := make([]*navigation.NavMenu, 0, len(items))
	for i := range items {
		item := items[i]
		result = append(result, r.modelToNavMenu(&item))
	}
	return result, nil
}

func (r *NavMenuRepository) GetByID(ctx context.Context, id int64) (*navigation.NavMenu, error) {
	item, err := r.repo.FirstByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return r.modelToNavMenu(item), nil
}

func (r *NavMenuRepository) Create(ctx context.Context, menu *navigation.NavMenu) error {
	modelItem := r.navMenuToModel(menu)
	if err := r.repo.Create(ctx, modelItem); err != nil {
		return err
	}
	menu.ID = modelItem.ID
	return nil
}

func (r *NavMenuRepository) Update(ctx context.Context, menu *navigation.NavMenu) error {
	modelItem := r.navMenuToModel(menu)
	modelItem.ID = menu.ID
	return r.repo.Save(ctx, modelItem)
}

func (r *NavMenuRepository) Delete(ctx context.Context, id int64) error {
	return r.repo.DeleteByID(ctx, id)
}

func (r *NavMenuRepository) NextSort(ctx context.Context, parentID *int64) (int, error) {
	query := r.db.WithContext(ctx).Model(&model.NavMenu{})
	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}
	var maxSort int
	if err := query.Select("COALESCE(MAX(sort), 0)").Scan(&maxSort).Error; err != nil {
		return 0, err
	}
	return maxSort + 1, nil
}

func (r *NavMenuRepository) UpdateOrder(ctx context.Context, updates []navigation.NavMenuOrderUpdate) error {
	if len(updates) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 延迟唯一约束检查到事务提交时，避免行级中间状态冲突。
		if err := tx.Exec("SET CONSTRAINTS uq_nav_menu_parent_sort DEFERRED").Error; err != nil {
			return err
		}
		var sb strings.Builder
		args := make([]any, 0, len(updates)*3)
		sb.WriteString("UPDATE nav_menu AS m SET parent_id = v.parent_id, sort = v.sort, updated_at = NOW() FROM (VALUES ")
		for i, u := range updates {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString("(?::bigint, ?::bigint, ?::int)")
			args = append(args, u.ID, u.ParentID, u.Sort)
		}
		sb.WriteString(") AS v(id, parent_id, sort) WHERE m.id = v.id AND m.deleted_at IS NULL")
		return tx.Exec(sb.String(), args...).Error
	})
}

func (r *NavMenuRepository) modelToNavMenu(item *model.NavMenu) *navigation.NavMenu {
	if item == nil {
		return nil
	}
	var deletedAt *time.Time
	if item.DeletedAt.Valid {
		value := item.DeletedAt.Time
		deletedAt = &value
	}
	return &navigation.NavMenu{
		ID:        item.ID,
		Name:      item.Name,
		URL:       item.URL,
		Icon:      item.Icon,
		Sort:      item.Sort,
		ParentID:  item.ParentID,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
		DeletedAt: deletedAt,
	}
}

func (r *NavMenuRepository) navMenuToModel(item *navigation.NavMenu) *model.NavMenu {
	if item == nil {
		return nil
	}
	return &model.NavMenu{
		ID:        item.ID,
		Name:      item.Name,
		URL:       item.URL,
		Icon:      item.Icon,
		Sort:      item.Sort,
		ParentID:  item.ParentID,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}
