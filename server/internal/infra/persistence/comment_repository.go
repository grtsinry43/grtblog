package persistence

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/grtsinry43/grtblog-v2/server/internal/domain/comment"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence/model"
)

type CommentRepository struct {
	db        *gorm.DB
	commentDB *GormRepository[model.Comment]
	areaDB    *GormRepository[model.CommentArea]
}

func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{
		db:        db,
		commentDB: NewGormRepository[model.Comment](db),
		areaDB:    NewGormRepository[model.CommentArea](db),
	}
}

func (r *CommentRepository) GetAreaByID(ctx context.Context, id int64) (*comment.CommentArea, error) {
	rec, err := r.areaDB.FirstByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, comment.ErrCommentAreaNotFound
		}
		return nil, err
	}
	return mapCommentAreaToDomain(*rec), nil
}

func (r *CommentRepository) SetAreaClosed(ctx context.Context, areaID int64, isClosed bool) error {
	return r.db.WithContext(ctx).
		Model(&model.CommentArea{}).
		Where("id = ?", areaID).
		Update("is_closed", isClosed).Error
}

func (r *CommentRepository) FindByID(ctx context.Context, id int64) (*comment.Comment, error) {
	rec, err := r.commentDB.FirstByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, comment.ErrCommentNotFound
		}
		return nil, err
	}
	entity := mapCommentToDomain(*rec)
	return &entity, nil
}

func (r *CommentRepository) ListPublicByAreaID(ctx context.Context, options comment.PublicListOptions) ([]*comment.Comment, error) {
	var recs []model.Comment
	query := r.db.WithContext(ctx).Unscoped().Where("area_id = ?", options.AreaID)

	approvedCond := "status = ?"
	args := []any{comment.CommentStatusApproved}

	if options.ViewerAuthorID != nil && *options.ViewerAuthorID > 0 {
		approvedCond += " OR author_id = ?"
		args = append(args, *options.ViewerAuthorID)
	}

	if visitorID := strings.TrimSpace(options.ViewerVisitorID); visitorID != "" {
		approvedCond += " OR visitor_id = ?"
		args = append(args, visitorID)
	}

	if err := query.
		Where("("+approvedCond+")", args...).
		Order("is_top DESC, created_at ASC").
		Find(&recs).Error; err != nil {
		return nil, err
	}

	out := make([]*comment.Comment, len(recs))
	for i, rec := range recs {
		entity := mapCommentToDomain(rec)
		out[i] = &entity
	}
	return out, nil
}

func (r *CommentRepository) ListForAdmin(ctx context.Context, options comment.AdminListOptions) ([]*comment.Comment, int64, error) {
	query := r.db.WithContext(ctx).Unscoped().Model(&model.Comment{}).
		Joins("JOIN comment_area ON comment_area.id = comment.area_id AND comment_area.deleted_at IS NULL").
		Joins("LEFT JOIN article ON article.comment_id = comment_area.id AND comment_area.area_type = ? AND article.deleted_at IS NULL", "article").
		Joins("LEFT JOIN moment ON moment.comment_id = comment_area.id AND comment_area.area_type = ? AND moment.deleted_at IS NULL", "moment").
		Joins("LEFT JOIN page ON page.comment_id = comment_area.id AND comment_area.area_type = ? AND page.deleted_at IS NULL", "page").
		Joins("LEFT JOIN thinking ON thinking.comment_id = comment_area.id AND comment_area.area_type = ?", "thinking").
		Where(
			"(comment_area.area_type = ? AND article.id IS NOT NULL) OR "+
				"(comment_area.area_type = ? AND moment.id IS NOT NULL) OR "+
				"(comment_area.area_type = ? AND page.id IS NOT NULL) OR "+
				"(comment_area.area_type = ? AND thinking.id IS NOT NULL)",
			"article", "moment", "page", "thinking",
		)
	if options.AreaID != nil {
		query = query.Where("comment.area_id = ?", *options.AreaID)
	}
	if strings.TrimSpace(options.Status) != "" {
		query = query.Where("comment.status = ?", options.Status)
	}
	if options.OnlyUnviewed {
		query = query.Where("comment.is_viewed = ?", false)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (options.Page - 1) * options.PageSize
	type adminCommentRow struct {
		model.Comment
		AreaType   string `gorm:"column:area_type"`
		AreaName   string `gorm:"column:area_name"`
		AreaRefID  *int64 `gorm:"column:area_ref_id"`
		AreaClosed bool   `gorm:"column:area_closed"`
		AreaTitle  string `gorm:"column:area_title"`
	}
	var recs []adminCommentRow
	if err := query.
		Select(
			"comment.*",
			"comment_area.area_type",
			"comment_area.area_name",
			"comment_area.content_id AS area_ref_id",
			"comment_area.is_closed AS area_closed",
			"COALESCE(article.title, moment.title, page.title, comment_area.area_name) AS area_title",
		).
		Order("comment.created_at DESC").
		Offset(offset).
		Limit(options.PageSize).
		Find(&recs).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*comment.Comment, len(recs))
	for i := range recs {
		entity := mapCommentToDomain(recs[i].Comment)
		entity.AreaType = toPtr(recs[i].AreaType)
		entity.AreaName = toPtr(recs[i].AreaName)
		entity.AreaRefID = recs[i].AreaRefID
		entity.AreaTitle = toPtr(recs[i].AreaTitle)
		entity.AreaClosed = &recs[i].AreaClosed
		items[i] = &entity
	}
	return items, total, nil
}

func (r *CommentRepository) Create(ctx context.Context, commentEntity *comment.Comment) error {
	rec := mapCommentToModel(commentEntity)
	if err := r.db.WithContext(ctx).Create(&rec).Error; err != nil {
		return err
	}
	commentEntity.ID = rec.ID
	commentEntity.CreatedAt = rec.CreatedAt
	commentEntity.UpdatedAt = rec.UpdatedAt
	return nil
}

func (r *CommentRepository) Update(ctx context.Context, commentEntity *comment.Comment) error {
	rec := mapCommentToModel(commentEntity)
	return r.db.WithContext(ctx).
		Model(&model.Comment{}).
		Where("id = ?", commentEntity.ID).
		Updates(map[string]any{
			"content":    rec.Content,
			"nick_name":  rec.NickName,
			"email":      rec.Email,
			"website":    rec.Website,
			"is_owner":   rec.IsOwner,
			"is_friend":  rec.IsFriend,
			"is_author":  rec.IsAuthor,
			"is_viewed":  rec.IsViewed,
			"is_top":     rec.IsTop,
			"status":     rec.Status,
			"updated_at": time.Now(),
		}).Error
}

func (r *CommentRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Comment{}, id).Error
}

func (r *CommentRepository) SetViewedStatus(ctx context.Context, ids []int64, isViewed bool) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).
		Model(&model.Comment{}).
		Where("id IN ?", ids).
		Update("is_viewed", isViewed).Error
}

func (r *CommentRepository) SetAuthorStatus(ctx context.Context, id int64, isAuthor bool) error {
	return r.db.WithContext(ctx).Model(&model.Comment{}).
		Where("id = ?", id).
		Update("is_author", isAuthor).Error
}

func (r *CommentRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	return r.db.WithContext(ctx).Model(&model.Comment{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *CommentRepository) SetTopStatus(ctx context.Context, id int64, isTop bool) error {
	return r.db.WithContext(ctx).Model(&model.Comment{}).
		Where("id = ?", id).
		Update("is_top", isTop).Error
}

func (r *CommentRepository) ExistsBlockedIdentity(ctx context.Context, authorID *int64, email *string) (bool, error) {
	query := r.db.WithContext(ctx).Unscoped().Model(&model.Comment{}).Where("status = ?", comment.CommentStatusBlocked)
	switch {
	case authorID != nil && *authorID > 0:
		query = query.Where("author_id = ?", *authorID)
	case email != nil && strings.TrimSpace(*email) != "":
		query = query.Where("LOWER(email) = LOWER(?)", strings.TrimSpace(*email))
	default:
		return false, nil
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func mapCommentToDomain(rec model.Comment) comment.Comment {
	status := strings.TrimSpace(rec.Status)
	if status == "" {
		status = comment.CommentStatusApproved
	}
	return comment.Comment{
		ID:        rec.ID,
		AreaID:    rec.AreaID,
		Content:   rec.Content,
		AuthorID:  rec.AuthorID,
		VisitorID: toPtr(rec.VisitorID),
		NickName:  toPtr(rec.NickName),
		IP:        toPtr(rec.IP),
		Location:  toPtr(rec.Location),
		Platform:  toPtr(rec.Platform),
		Browser:   toPtr(rec.Browser),
		Email:     toPtr(rec.Email),
		Website:   toPtr(rec.Website),
		IsOwner:   rec.IsOwner,
		IsFriend:  rec.IsFriend,
		IsAuthor:  rec.IsAuthor,
		IsViewed:  rec.IsViewed,
		IsTop:     rec.IsTop,
		IsMy:      false,
		Status:    status,
		ParentID:  rec.ParentID,
		CreatedAt: rec.CreatedAt,
		UpdatedAt: rec.UpdatedAt,
		DeletedAt: timeToPtr(rec.DeletedAt),
	}
}

func mapCommentToModel(entity *comment.Comment) model.Comment {
	status := strings.TrimSpace(entity.Status)
	if status == "" {
		status = comment.CommentStatusPending
	}
	return model.Comment{
		ID:        entity.ID,
		AreaID:    entity.AreaID,
		Content:   strings.TrimSpace(entity.Content),
		AuthorID:  entity.AuthorID,
		VisitorID: toValue(entity.VisitorID),
		NickName:  toValue(entity.NickName),
		IP:        toValue(entity.IP),
		Location:  toValue(entity.Location),
		Platform:  toValue(entity.Platform),
		Browser:   toValue(entity.Browser),
		Email:     toValue(entity.Email),
		Website:   toValue(entity.Website),
		IsOwner:   entity.IsOwner,
		IsFriend:  entity.IsFriend,
		IsAuthor:  entity.IsAuthor,
		IsViewed:  entity.IsViewed,
		IsTop:     entity.IsTop,
		Status:    status,
		ParentID:  entity.ParentID,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
		DeletedAt: gorm.DeletedAt{Time: timeToValue(entity.DeletedAt), Valid: entity.DeletedAt != nil},
	}
}

func mapCommentAreaToDomain(rec model.CommentArea) *comment.CommentArea {
	return &comment.CommentArea{
		ID:        rec.ID,
		Name:      rec.AreaName,
		Type:      rec.AreaType,
		ContentID: rec.ContentID,
		IsClosed:  rec.IsClosed,
		CreatedAt: rec.CreatedAt,
		UpdatedAt: rec.UpdatedAt,
	}
}

func toPtr(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func toValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func timeToPtr(val gorm.DeletedAt) *time.Time {
	if !val.Valid {
		return nil
	}
	t := val.Time
	return &t
}

func timeToValue(val *time.Time) time.Time {
	if val == nil {
		return time.Time{}
	}
	return *val
}
