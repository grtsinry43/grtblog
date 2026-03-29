package persistence

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	domainlike "github.com/grtsinry43/grtblog-v2/server/internal/domain/like"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence/model"
)

type LikeRepository struct {
	db *gorm.DB
}

func NewLikeRepository(db *gorm.DB) *LikeRepository {
	return &LikeRepository{db: db}
}

func (r *LikeRepository) ExistsTarget(ctx context.Context, targetType domainlike.TargetType, targetID int64) (bool, error) {
	var count int64
	q := r.db.WithContext(ctx)
	switch targetType {
	case domainlike.TargetArticle:
		if err := q.Model(&model.Article{}).Where("id = ? AND is_published = ?", targetID, true).Count(&count).Error; err != nil {
			return false, err
		}
	case domainlike.TargetMoment:
		if err := q.Model(&model.Moment{}).Where("id = ? AND is_published = ?", targetID, true).Count(&count).Error; err != nil {
			return false, err
		}
	case domainlike.TargetPage:
		if err := q.Model(&model.Page{}).Where("id = ? AND is_enabled = ?", targetID, true).Count(&count).Error; err != nil {
			return false, err
		}
	case domainlike.TargetThinking:
		if err := q.Model(&model.Thinking{}).Where("id = ?", targetID).Count(&count).Error; err != nil {
			return false, err
		}
	case domainlike.TargetAlbum:
		if err := q.Model(&model.Album{}).Where("id = ? AND is_published = ?", targetID, true).Count(&count).Error; err != nil {
			return false, err
		}
	default:
		return false, domainlike.ErrInvalidTargetType
	}
	return count > 0, nil
}

func (r *LikeRepository) CreateIfAbsent(ctx context.Context, entity *domainlike.ContentLike) (bool, error) {
	rec := model.ContentLike{
		TargetType: string(entity.TargetType),
		TargetID:   entity.TargetID,
		UserID:     entity.UserID,
	}
	if entity.VisitorID != nil {
		rec.VisitorID = *entity.VisitorID
	}

	tx := r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&rec)
	if tx.Error != nil {
		return false, tx.Error
	}
	if tx.RowsAffected == 0 {
		return false, nil
	}

	entity.ID = rec.ID
	entity.CreatedAt = rec.CreatedAt
	return true, nil
}

func (r *LikeRepository) CreateBatchIfAbsent(ctx context.Context, entities []*domainlike.ContentLike) (int64, error) {
	if len(entities) == 0 {
		return 0, nil
	}

	recs := make([]model.ContentLike, 0, len(entities))
	for _, entity := range entities {
		if entity == nil || entity.VisitorID == nil {
			continue
		}
		rec := model.ContentLike{
			TargetType: string(entity.TargetType),
			TargetID:   entity.TargetID,
			UserID:     entity.UserID,
			VisitorID:  *entity.VisitorID,
		}
		if !entity.CreatedAt.IsZero() {
			rec.CreatedAt = entity.CreatedAt
		}
		recs = append(recs, rec)
	}
	if len(recs) == 0 {
		return 0, nil
	}

	tx := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		CreateInBatches(recs, 500)
	if tx.Error != nil {
		return 0, tx.Error
	}
	return tx.RowsAffected, nil
}
