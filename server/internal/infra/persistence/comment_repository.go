package persistence

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sort"
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

type visitorProfileAggRow struct {
	VisitorID        string    `gorm:"column:visitor_id"`
	TotalComments    int64     `gorm:"column:total_comments"`
	ApprovedComments int64     `gorm:"column:approved_comments"`
	PendingComments  int64     `gorm:"column:pending_comments"`
	RejectedComments int64     `gorm:"column:rejected_comments"`
	BlockedComments  int64     `gorm:"column:blocked_comments"`
	DeletedComments  int64     `gorm:"column:deleted_comments"`
	TopComments      int64     `gorm:"column:top_comments"`
	ActiveDays       int64     `gorm:"column:active_days"`
	FirstSeenAt      time.Time `gorm:"column:first_seen_at"`
	LastSeenAt       time.Time `gorm:"column:last_seen_at"`
}

type visitorLatestRow struct {
	VisitorID  string    `gorm:"column:visitor_id"`
	NickName   string    `gorm:"column:nick_name"`
	Email      string    `gorm:"column:email"`
	Website    string    `gorm:"column:website"`
	IP         string    `gorm:"column:ip"`
	Location   string    `gorm:"column:location"`
	Platform   string    `gorm:"column:platform"`
	Browser    string    `gorm:"column:browser"`
	ObservedAt time.Time `gorm:"column:observed_at"`
	RN         int64     `gorm:"column:rn"`
}

type visitorPoolRow struct {
	VisitorID    string    `gorm:"column:visitor_id"`
	LastActiveAt time.Time `gorm:"column:last_active_at"`
}

type visitorLikeAggRow struct {
	VisitorID        string     `gorm:"column:visitor_id"`
	TotalLikes       int64      `gorm:"column:total_likes"`
	UniqueLikedItems int64      `gorm:"column:unique_liked_items"`
	LastLikedAt      *time.Time `gorm:"column:last_liked_at"`
}

type visitorViewAggRow struct {
	VisitorID       string     `gorm:"column:visitor_id"`
	TotalViews      int64      `gorm:"column:total_views"`
	UniqueViewItems int64      `gorm:"column:unique_view_items"`
	LastViewedAt    *time.Time `gorm:"column:last_viewed_at"`
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

func (r *CommentRepository) GetContentTitleByTypeAndID(ctx context.Context, areaType string, contentID int64) (string, error) {
	normalizedType := strings.ToLower(strings.TrimSpace(areaType))
	if normalizedType == "" || contentID <= 0 {
		return "", nil
	}
	switch normalizedType {
	case "article":
		var row struct {
			Title string `gorm:"column:title"`
		}
		if err := r.db.WithContext(ctx).
			Model(&model.Article{}).
			Select("title").
			Where("id = ? AND deleted_at IS NULL", contentID).
			Take(&row).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "", nil
			}
			return "", err
		}
		return strings.TrimSpace(row.Title), nil
	case "moment":
		var row struct {
			Title string `gorm:"column:title"`
		}
		if err := r.db.WithContext(ctx).
			Model(&model.Moment{}).
			Select("title").
			Where("id = ? AND deleted_at IS NULL", contentID).
			Take(&row).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "", nil
			}
			return "", err
		}
		return strings.TrimSpace(row.Title), nil
	case "page":
		var row struct {
			Title string `gorm:"column:title"`
		}
		if err := r.db.WithContext(ctx).
			Model(&model.Page{}).
			Select("title").
			Where("id = ? AND deleted_at IS NULL", contentID).
			Take(&row).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "", nil
			}
			return "", err
		}
		return strings.TrimSpace(row.Title), nil
	case "thinking":
		var row struct {
			Content string `gorm:"column:content"`
		}
		if err := r.db.WithContext(ctx).
			Model(&model.Thinking{}).
			Select("content").
			Where("id = ?", contentID).
			Take(&row).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "", nil
			}
			return "", err
		}
		return summarizeThinkingContent(row.Content), nil
	default:
		return "", nil
	}
}

func (r *CommentRepository) GetContentViewPathByTypeAndID(ctx context.Context, areaType string, contentID int64) (string, error) {
	normalizedType := strings.ToLower(strings.TrimSpace(areaType))
	if normalizedType == "" || contentID <= 0 {
		return "", nil
	}
	switch normalizedType {
	case "article":
		var row struct {
			ShortURL string `gorm:"column:short_url"`
		}
		if err := r.db.WithContext(ctx).
			Model(&model.Article{}).
			Select("short_url").
			Where("id = ? AND deleted_at IS NULL", contentID).
			Take(&row).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "", nil
			}
			return "", err
		}
		slug := strings.TrimSpace(row.ShortURL)
		if slug == "" {
			return "", nil
		}
		return "/posts/" + url.PathEscape(slug), nil
	case "moment":
		var row struct {
			ShortURL  string    `gorm:"column:short_url"`
			CreatedAt time.Time `gorm:"column:created_at"`
		}
		if err := r.db.WithContext(ctx).
			Model(&model.Moment{}).
			Select("short_url", "created_at").
			Where("id = ? AND deleted_at IS NULL", contentID).
			Take(&row).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "", nil
			}
			return "", err
		}
		slug := strings.TrimSpace(row.ShortURL)
		if slug == "" {
			return "", nil
		}
		created := row.CreatedAt.UTC()
		return fmt.Sprintf(
			"/moments/%04d/%02d/%02d/%s",
			created.Year(),
			int(created.Month()),
			created.Day(),
			url.PathEscape(slug),
		), nil
	case "page":
		var row struct {
			ShortURL string `gorm:"column:short_url"`
		}
		if err := r.db.WithContext(ctx).
			Model(&model.Page{}).
			Select("short_url").
			Where("id = ? AND deleted_at IS NULL", contentID).
			Take(&row).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "", nil
			}
			return "", err
		}
		slug := strings.TrimSpace(row.ShortURL)
		if slug == "" {
			return "", nil
		}
		return "/" + url.PathEscape(slug), nil
	case "thinking":
		var row struct {
			ID int64 `gorm:"column:id"`
		}
		if err := r.db.WithContext(ctx).
			Model(&model.Thinking{}).
			Select("id").
			Where("id = ?", contentID).
			Take(&row).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "", nil
			}
			return "", err
		}
		return fmt.Sprintf("/thinkings#thinking-%d", row.ID), nil
	default:
		return "", nil
	}
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

func (r *CommentRepository) FindByFederatedObjectID(ctx context.Context, objectID string) (*comment.Comment, error) {
	objectID = strings.TrimSpace(objectID)
	if objectID == "" {
		return nil, comment.ErrCommentNotFound
	}
	var rec model.Comment
	if err := r.db.WithContext(ctx).
		Where("federated_object_id = ?", objectID).
		Limit(1).
		Find(&rec).Error; err != nil {
		return nil, err
	}
	if rec.ID == 0 {
		return nil, comment.ErrCommentNotFound
	}
	entity := mapCommentToDomain(rec)
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
		Order("is_top DESC, created_at DESC, id DESC").
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
			"is_edited":  rec.IsEdited,
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
	result := r.db.WithContext(ctx).Unscoped().
		Model(&model.Comment{}).
		Where("id IN ?", ids).
		Update("is_viewed", isViewed)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		return nil
	}

	// If no rows were updated, treat it as success only when records already exist
	// and are already in the target viewed state; otherwise return not found.
	var matched int64
	if err := r.db.WithContext(ctx).Unscoped().
		Model(&model.Comment{}).
		Where("id IN ?", ids).
		Count(&matched).Error; err != nil {
		return err
	}
	if matched == 0 {
		return comment.ErrCommentNotFound
	}
	return nil
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

func (r *CommentRepository) ListVisitors(ctx context.Context, options comment.AdminVisitorListOptions) ([]comment.VisitorProfile, int64, error) {
	keyword := strings.TrimSpace(options.Keyword)
	poolRows, total, err := r.listVisitorPool(ctx, keyword, options.Page, options.PageSize)
	if err != nil {
		return nil, 0, err
	}
	if len(poolRows) == 0 {
		return []comment.VisitorProfile{}, total, nil
	}

	visitorIDs := make([]string, 0, len(poolRows))
	lastActiveMap := make(map[string]time.Time, len(poolRows))
	for _, row := range poolRows {
		visitorIDs = append(visitorIDs, row.VisitorID)
		lastActiveMap[row.VisitorID] = row.LastActiveAt
	}

	commentAggMap, err := r.loadVisitorCommentAggMap(ctx, visitorIDs)
	if err != nil {
		return nil, 0, err
	}
	likeAggMap, err := r.loadVisitorLikeAggMap(ctx, visitorIDs)
	if err != nil {
		return nil, 0, err
	}
	viewAggMap, err := r.loadVisitorViewAggMap(ctx, visitorIDs)
	if err != nil {
		return nil, 0, err
	}
	latestByVisitor, err := r.loadLatestVisitorRows(ctx, visitorIDs)
	if err != nil {
		return nil, 0, err
	}

	items := make([]comment.VisitorProfile, 0, len(visitorIDs))
	for _, visitorID := range visitorIDs {
		item := comment.VisitorProfile{
			VisitorID: visitorID,
		}
		if commentAgg, ok := commentAggMap[visitorID]; ok {
			item.TotalComments = commentAgg.TotalComments
			item.ApprovedComments = commentAgg.ApprovedComments
			item.PendingComments = commentAgg.PendingComments
			item.RejectedComments = commentAgg.RejectedComments
			item.BlockedComments = commentAgg.BlockedComments
			item.DeletedComments = commentAgg.DeletedComments
			item.TopComments = commentAgg.TopComments
			item.ActiveDays = commentAgg.ActiveDays
			item.FirstSeenAt = commentAgg.FirstSeenAt
			item.LastSeenAt = commentAgg.LastSeenAt
		}
		if likeAgg, ok := likeAggMap[visitorID]; ok {
			item.TotalLikes = likeAgg.TotalLikes
			item.UniqueLikedItems = likeAgg.UniqueLikedItems
			item.LastLikedAt = likeAgg.LastLikedAt
		}
		if viewAgg, ok := viewAggMap[visitorID]; ok {
			item.TotalViews = viewAgg.TotalViews
			item.UniqueViewItems = viewAgg.UniqueViewItems
			item.LastViewedAt = viewAgg.LastViewedAt
		}
		if latest, ok := latestByVisitor[visitorID]; ok {
			item.NickName = toPtr(latest.NickName)
			item.Email = toPtr(latest.Email)
			item.Website = toPtr(latest.Website)
			item.IP = toPtr(latest.IP)
			item.Location = toPtr(latest.Location)
			item.Platform = toPtr(latest.Platform)
			item.Browser = toPtr(latest.Browser)
		}

		if item.FirstSeenAt.IsZero() {
			item.FirstSeenAt = lastActiveMap[visitorID]
		}
		if item.LastSeenAt.IsZero() {
			item.LastSeenAt = lastActiveMap[visitorID]
		} else if active := lastActiveMap[visitorID]; active.After(item.LastSeenAt) {
			item.LastSeenAt = active
		}
		items = append(items, item)
	}

	return items, total, nil
}

func (r *CommentRepository) GetVisitorProfile(ctx context.Context, visitorID string, recentLimit int) (*comment.VisitorProfile, []comment.VisitorRecentComment, error) {
	visitorID = strings.TrimSpace(visitorID)
	if visitorID == "" {
		return nil, nil, comment.ErrVisitorNotFound
	}

	ids := []string{visitorID}
	commentAggMap, err := r.loadVisitorCommentAggMap(ctx, ids)
	if err != nil {
		return nil, nil, err
	}
	likeAggMap, err := r.loadVisitorLikeAggMap(ctx, ids)
	if err != nil {
		return nil, nil, err
	}
	viewAggMap, err := r.loadVisitorViewAggMap(ctx, ids)
	if err != nil {
		return nil, nil, err
	}
	latestByVisitor, err := r.loadLatestVisitorRows(ctx, ids)
	if err != nil {
		return nil, nil, err
	}

	_, hasCommentAgg := commentAggMap[visitorID]
	_, hasLikeAgg := likeAggMap[visitorID]
	_, hasViewAgg := viewAggMap[visitorID]
	_, hasLatest := latestByVisitor[visitorID]
	if !hasCommentAgg && !hasLikeAgg && !hasViewAgg && !hasLatest {
		return nil, nil, comment.ErrVisitorNotFound
	}

	profile := &comment.VisitorProfile{
		VisitorID: visitorID,
	}
	if agg, ok := commentAggMap[visitorID]; ok {
		profile.TotalComments = agg.TotalComments
		profile.ApprovedComments = agg.ApprovedComments
		profile.PendingComments = agg.PendingComments
		profile.RejectedComments = agg.RejectedComments
		profile.BlockedComments = agg.BlockedComments
		profile.DeletedComments = agg.DeletedComments
		profile.TopComments = agg.TopComments
		profile.ActiveDays = agg.ActiveDays
		profile.FirstSeenAt = agg.FirstSeenAt
		profile.LastSeenAt = agg.LastSeenAt
	}
	if agg, ok := likeAggMap[visitorID]; ok {
		profile.TotalLikes = agg.TotalLikes
		profile.UniqueLikedItems = agg.UniqueLikedItems
		profile.LastLikedAt = agg.LastLikedAt
	}
	if agg, ok := viewAggMap[visitorID]; ok {
		profile.TotalViews = agg.TotalViews
		profile.UniqueViewItems = agg.UniqueViewItems
		profile.LastViewedAt = agg.LastViewedAt
	}
	if latest, ok := latestByVisitor[visitorID]; ok {
		profile.NickName = toPtr(latest.NickName)
		profile.Email = toPtr(latest.Email)
		profile.Website = toPtr(latest.Website)
		profile.IP = toPtr(latest.IP)
		profile.Location = toPtr(latest.Location)
		profile.Platform = toPtr(latest.Platform)
		profile.Browser = toPtr(latest.Browser)
	}

	lastActivityAt, err := r.loadVisitorLastActivityAt(ctx, visitorID)
	if err != nil {
		return nil, nil, err
	}
	if profile.FirstSeenAt.IsZero() {
		profile.FirstSeenAt = lastActivityAt
	}
	if profile.LastSeenAt.IsZero() || lastActivityAt.After(profile.LastSeenAt) {
		profile.LastSeenAt = lastActivityAt
	}

	if !hasCommentAgg {
		return profile, []comment.VisitorRecentComment{}, nil
	}

	if recentLimit <= 0 {
		recentLimit = 20
	}
	if recentLimit > 100 {
		recentLimit = 100
	}
	type recentCommentRow struct {
		ID        int64      `gorm:"column:id"`
		AreaID    int64      `gorm:"column:area_id"`
		Content   string     `gorm:"column:content"`
		Status    string     `gorm:"column:status"`
		CreatedAt time.Time  `gorm:"column:created_at"`
		DeletedAt *time.Time `gorm:"column:deleted_at"`
	}
	var recRows []recentCommentRow
	if err := r.db.WithContext(ctx).Unscoped().Model(&model.Comment{}).
		Select("id", "area_id", "content", "status", "created_at", "deleted_at").
		Where("visitor_id = ?", visitorID).
		Order("created_at DESC").
		Order("id DESC").
		Limit(recentLimit).
		Scan(&recRows).Error; err != nil {
		return nil, nil, err
	}

	recent := make([]comment.VisitorRecentComment, 0, len(recRows))
	for _, row := range recRows {
		recent = append(recent, comment.VisitorRecentComment{
			ID:        row.ID,
			AreaID:    row.AreaID,
			Content:   row.Content,
			Status:    row.Status,
			CreatedAt: row.CreatedAt,
			IsDeleted: row.DeletedAt != nil,
		})
	}

	return profile, recent, nil
}

func (r *CommentRepository) listVisitorPool(ctx context.Context, keyword string, page, pageSize int) ([]visitorPoolRow, int64, error) {
	baseSQL := `
WITH union_activities AS (
    SELECT visitor_id, MAX(created_at) AS last_at
    FROM comment
    WHERE visitor_id <> ''
    GROUP BY visitor_id
    UNION ALL
    SELECT visitor_id, MAX(created_at) AS last_at
    FROM content_like
    WHERE visitor_id <> ''
    GROUP BY visitor_id
    UNION ALL
    SELECT visitor_id, MAX(last_view_at) AS last_at
    FROM analytics_visitor_view
    WHERE visitor_id <> ''
    GROUP BY visitor_id
),
visitor_pool AS (
    SELECT visitor_id, MAX(last_at) AS last_active_at
    FROM union_activities
    GROUP BY visitor_id
)
`
	keyword = strings.TrimSpace(keyword)
	filterSQL := "SELECT visitor_id, last_active_at FROM visitor_pool"
	countSQL := "SELECT COUNT(*) FROM visitor_pool"
	args := []any{}
	if keyword != "" {
		like := "%" + keyword + "%"
		filterSQL += `
 WHERE visitor_id LIKE ?
    OR EXISTS (
        SELECT 1
        FROM comment c
        WHERE c.visitor_id = visitor_pool.visitor_id
          AND (c.nick_name LIKE ? OR c.email LIKE ? OR c.ip LIKE ? OR c.location LIKE ? OR c.platform LIKE ? OR c.browser LIKE ?)
    )
    OR EXISTS (
        SELECT 1
        FROM analytics_visitor_view v
        WHERE v.visitor_id = visitor_pool.visitor_id
          AND (v.last_ip LIKE ? OR v.location LIKE ? OR v.platform LIKE ? OR v.browser LIKE ?)
    )`
		countSQL += `
 WHERE visitor_id LIKE ?
    OR EXISTS (
        SELECT 1
        FROM comment c
        WHERE c.visitor_id = visitor_pool.visitor_id
          AND (c.nick_name LIKE ? OR c.email LIKE ? OR c.ip LIKE ? OR c.location LIKE ? OR c.platform LIKE ? OR c.browser LIKE ?)
    )
    OR EXISTS (
        SELECT 1
        FROM analytics_visitor_view v
        WHERE v.visitor_id = visitor_pool.visitor_id
          AND (v.last_ip LIKE ? OR v.location LIKE ? OR v.platform LIKE ? OR v.browser LIKE ?)
    )`
		args = append(args, like, like, like, like, like, like, like, like, like, like, like)
	}

	var total int64
	if err := r.db.WithContext(ctx).Raw(baseSQL+countSQL, args...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []visitorPoolRow{}, 0, nil
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	pagedSQL := filterSQL + " ORDER BY last_active_at DESC OFFSET ? LIMIT ?"
	queryArgs := append(args, offset, pageSize)
	var rows []visitorPoolRow
	if err := r.db.WithContext(ctx).Raw(baseSQL+pagedSQL, queryArgs...).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *CommentRepository) loadVisitorCommentAggMap(ctx context.Context, visitorIDs []string) (map[string]visitorProfileAggRow, error) {
	if len(visitorIDs) == 0 {
		return map[string]visitorProfileAggRow{}, nil
	}
	var rows []visitorProfileAggRow
	if err := r.db.WithContext(ctx).Unscoped().Model(&model.Comment{}).
		Select(
			"visitor_id",
			"COUNT(*) AS total_comments",
			"SUM(CASE WHEN status = 'approved' THEN 1 ELSE 0 END) AS approved_comments",
			"SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) AS pending_comments",
			"SUM(CASE WHEN status = 'rejected' THEN 1 ELSE 0 END) AS rejected_comments",
			"SUM(CASE WHEN status = 'blocked' THEN 1 ELSE 0 END) AS blocked_comments",
			"SUM(CASE WHEN deleted_at IS NOT NULL THEN 1 ELSE 0 END) AS deleted_comments",
			"SUM(CASE WHEN is_top = TRUE THEN 1 ELSE 0 END) AS top_comments",
			"COUNT(DISTINCT DATE(created_at)) AS active_days",
			"MIN(created_at) AS first_seen_at",
			"MAX(created_at) AS last_seen_at",
		).
		Where("visitor_id IN ?", visitorIDs).
		Group("visitor_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make(map[string]visitorProfileAggRow, len(rows))
	for _, row := range rows {
		result[row.VisitorID] = row
	}
	return result, nil
}

func (r *CommentRepository) loadVisitorLikeAggMap(ctx context.Context, visitorIDs []string) (map[string]visitorLikeAggRow, error) {
	if len(visitorIDs) == 0 {
		return map[string]visitorLikeAggRow{}, nil
	}
	var rows []visitorLikeAggRow
	if err := r.db.WithContext(ctx).Model(&model.ContentLike{}).
		Select(
			"visitor_id",
			"COUNT(*) AS total_likes",
			"COUNT(DISTINCT CONCAT(target_type, ':', target_id)) AS unique_liked_items",
			"MAX(created_at) AS last_liked_at",
		).
		Where("visitor_id IN ?", visitorIDs).
		Group("visitor_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make(map[string]visitorLikeAggRow, len(rows))
	for _, row := range rows {
		result[row.VisitorID] = row
	}
	return result, nil
}

func (r *CommentRepository) loadVisitorViewAggMap(ctx context.Context, visitorIDs []string) (map[string]visitorViewAggRow, error) {
	if len(visitorIDs) == 0 {
		return map[string]visitorViewAggRow{}, nil
	}
	var rows []visitorViewAggRow
	if err := r.db.WithContext(ctx).Model(&model.AnalyticsVisitorView{}).
		Select(
			"visitor_id",
			"COALESCE(SUM(view_count), 0) AS total_views",
			"COUNT(DISTINCT CONCAT(content_type, ':', content_id)) AS unique_view_items",
			"MAX(last_view_at) AS last_viewed_at",
		).
		Where("visitor_id IN ?", visitorIDs).
		Group("visitor_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make(map[string]visitorViewAggRow, len(rows))
	for _, row := range rows {
		result[row.VisitorID] = row
	}
	return result, nil
}

func (r *CommentRepository) loadVisitorLastActivityAt(ctx context.Context, visitorID string) (time.Time, error) {
	type row struct {
		LastActiveAt *time.Time `gorm:"column:last_active_at"`
	}
	var result row
	err := r.db.WithContext(ctx).Raw(`
WITH activities AS (
    SELECT MAX(created_at) AS at FROM comment WHERE visitor_id = ?
    UNION ALL
    SELECT MAX(created_at) AS at FROM content_like WHERE visitor_id = ?
    UNION ALL
    SELECT MAX(last_view_at) AS at FROM analytics_visitor_view WHERE visitor_id = ?
)
SELECT MAX(at) AS last_active_at FROM activities
`, visitorID, visitorID, visitorID).Scan(&result).Error
	if err != nil {
		return time.Time{}, err
	}
	if result.LastActiveAt == nil {
		return time.Time{}, comment.ErrVisitorNotFound
	}
	return *result.LastActiveAt, nil
}

func (r *CommentRepository) GetVisitorInsights(ctx context.Context, days int) (*comment.VisitorInsights, error) {
	if days <= 0 {
		days = 30
	}
	if days > 180 {
		days = 180
	}
	start := time.Now().UTC().AddDate(0, 0, -(days - 1)).Truncate(24 * time.Hour)

	platformTop, browserTop, locationTop, err := r.queryVisitorSourceTop(ctx, start, 8)
	if err != nil {
		return nil, err
	}
	trend, err := r.queryVisitorTrend(ctx, start, days)
	if err != nil {
		return nil, err
	}
	funnel, err := r.queryVisitorFunnel(ctx, start)
	if err != nil {
		return nil, err
	}
	segments, err := r.queryVisitorSegments(ctx, start)
	if err != nil {
		return nil, err
	}

	return &comment.VisitorInsights{
		Days:        days,
		GeneratedAt: time.Now().UTC(),
		DataSource:  "api",
		PlatformTop: platformTop,
		BrowserTop:  browserTop,
		LocationTop: locationTop,
		Trend:       trend,
		Funnel:      funnel,
		Segments:    segments,
	}, nil
}

func (r *CommentRepository) queryVisitorSourceTop(ctx context.Context, start time.Time, topN int) ([]comment.VisitorDistributionItem, []comment.VisitorDistributionItem, []comment.VisitorDistributionItem, error) {
	type row struct {
		VisitorID string `gorm:"column:visitor_id"`
		Platform  string `gorm:"column:platform"`
		Browser   string `gorm:"column:browser"`
		Location  string `gorm:"column:location"`
	}
	var rows []row
	err := r.db.WithContext(ctx).Raw(`
WITH view_latest AS (
    SELECT
        visitor_id,
        platform,
        browser,
        location,
        ROW_NUMBER() OVER (PARTITION BY visitor_id ORDER BY last_view_at DESC, updated_at DESC) AS rn
    FROM analytics_visitor_view
    WHERE visitor_id <> '' AND last_view_at >= ?
),
comment_latest AS (
    SELECT
        visitor_id,
        platform,
        browser,
        location,
        ROW_NUMBER() OVER (PARTITION BY visitor_id ORDER BY created_at DESC, id DESC) AS rn
    FROM comment
    WHERE visitor_id <> '' AND created_at >= ?
),
merged AS (
    SELECT visitor_id, platform, browser, location
    FROM view_latest
    WHERE rn = 1
    UNION ALL
    SELECT c.visitor_id, c.platform, c.browser, c.location
    FROM comment_latest c
    WHERE c.rn = 1
      AND NOT EXISTS (
          SELECT 1 FROM view_latest v
          WHERE v.rn = 1 AND v.visitor_id = c.visitor_id
      )
)
SELECT visitor_id, platform, browser, location FROM merged
`, start, start).Scan(&rows).Error
	if err != nil {
		return nil, nil, nil, err
	}

	platformMap := map[string]int64{}
	browserMap := map[string]int64{}
	locationMap := map[string]int64{}
	for _, row := range rows {
		platform := strings.TrimSpace(row.Platform)
		if platform == "" {
			platform = "Unknown"
		}
		browser := strings.TrimSpace(row.Browser)
		if browser == "" {
			browser = "Unknown"
		}
		location := strings.TrimSpace(row.Location)
		if location == "" {
			location = "Unknown"
		}
		platformMap[platform]++
		browserMap[browser]++
		locationMap[location]++
	}

	return topDistribution(platformMap, topN), topDistribution(browserMap, topN), topDistribution(locationMap, topN), nil
}

func (r *CommentRepository) queryVisitorTrend(ctx context.Context, start time.Time, days int) ([]comment.VisitorTrendPoint, error) {
	activeMap, err := r.queryDayCountMapRaw(ctx, `
SELECT TO_CHAR(day, 'YYYY-MM-DD') AS day, COUNT(DISTINCT visitor_id) AS count
FROM (
    SELECT DATE(last_view_at) AS day, visitor_id
    FROM analytics_visitor_view
    WHERE visitor_id <> '' AND last_view_at >= ?
    UNION ALL
    SELECT DATE(created_at) AS day, visitor_id
    FROM content_like
    WHERE visitor_id <> '' AND created_at >= ?
    UNION ALL
    SELECT DATE(created_at) AS day, visitor_id
    FROM comment
    WHERE visitor_id <> '' AND created_at >= ?
) t
GROUP BY day
`, start, start, start)
	if err != nil {
		return nil, err
	}

	newMap, err := r.queryDayCountMapRaw(ctx, `
WITH first_touch AS (
    SELECT visitor_id, MIN(first_at) AS first_at
    FROM (
        SELECT visitor_id, MIN(first_view_at) AS first_at
        FROM analytics_visitor_view
        WHERE visitor_id <> ''
        GROUP BY visitor_id
        UNION ALL
        SELECT visitor_id, MIN(created_at) AS first_at
        FROM content_like
        WHERE visitor_id <> ''
        GROUP BY visitor_id
        UNION ALL
        SELECT visitor_id, MIN(created_at) AS first_at
        FROM comment
        WHERE visitor_id <> ''
        GROUP BY visitor_id
    ) x
    GROUP BY visitor_id
)
SELECT TO_CHAR(DATE(first_at), 'YYYY-MM-DD') AS day, COUNT(*) AS count
FROM first_touch
WHERE first_at >= ?
GROUP BY DATE(first_at)
`, start)
	if err != nil {
		return nil, err
	}

	viewMap, err := r.queryDayCountMapRaw(ctx, `
SELECT TO_CHAR(DATE(last_view_at), 'YYYY-MM-DD') AS day, COALESCE(SUM(view_count), 0) AS count
FROM analytics_visitor_view
WHERE visitor_id <> '' AND last_view_at >= ?
GROUP BY DATE(last_view_at)
`, start)
	if err != nil {
		return nil, err
	}

	likeMap, err := r.queryDayCountMapRaw(ctx, `
SELECT TO_CHAR(DATE(created_at), 'YYYY-MM-DD') AS day, COUNT(*) AS count
FROM content_like
WHERE visitor_id <> '' AND created_at >= ?
GROUP BY DATE(created_at)
`, start)
	if err != nil {
		return nil, err
	}

	commentMap, err := r.queryDayCountMapRaw(ctx, `
SELECT TO_CHAR(DATE(created_at), 'YYYY-MM-DD') AS day, COUNT(*) AS count
FROM comment
WHERE visitor_id <> '' AND created_at >= ?
GROUP BY DATE(created_at)
`, start)
	if err != nil {
		return nil, err
	}

	out := make([]comment.VisitorTrendPoint, 0, days)
	for i := 0; i < days; i++ {
		day := start.AddDate(0, 0, i).Format("2006-01-02")
		active := activeMap[day]
		newCount := newMap[day]
		returning := active - newCount
		if returning < 0 {
			returning = 0
		}
		out = append(out, comment.VisitorTrendPoint{
			Date:              day,
			ActiveVisitors:    active,
			NewVisitors:       newCount,
			ReturningVisitors: returning,
			Views:             viewMap[day],
			Likes:             likeMap[day],
			Comments:          commentMap[day],
		})
	}
	return out, nil
}

func (r *CommentRepository) queryVisitorFunnel(ctx context.Context, start time.Time) (comment.VisitorFunnel, error) {
	type row struct {
		ViewVisitors    int64 `gorm:"column:view_visitors"`
		LikeVisitors    int64 `gorm:"column:like_visitors"`
		CommentVisitors int64 `gorm:"column:comment_visitors"`
	}
	var result row
	err := r.db.WithContext(ctx).Raw(`
SELECT
    (SELECT COUNT(DISTINCT visitor_id) FROM analytics_visitor_view WHERE visitor_id <> '' AND last_view_at >= ?) AS view_visitors,
    (SELECT COUNT(DISTINCT visitor_id) FROM content_like WHERE visitor_id <> '' AND created_at >= ?) AS like_visitors,
    (SELECT COUNT(DISTINCT visitor_id) FROM comment WHERE visitor_id <> '' AND created_at >= ?) AS comment_visitors
`, start, start, start).Scan(&result).Error
	if err != nil {
		return comment.VisitorFunnel{}, err
	}
	funnel := comment.VisitorFunnel{
		ViewVisitors:    result.ViewVisitors,
		LikeVisitors:    result.LikeVisitors,
		CommentVisitors: result.CommentVisitors,
	}
	if funnel.ViewVisitors > 0 {
		funnel.LikeRate = float64(funnel.LikeVisitors) / float64(funnel.ViewVisitors)
		funnel.CommentRateByView = float64(funnel.CommentVisitors) / float64(funnel.ViewVisitors)
	}
	if funnel.LikeVisitors > 0 {
		funnel.CommentRateByLike = float64(funnel.CommentVisitors) / float64(funnel.LikeVisitors)
	}
	return funnel, nil
}

func (r *CommentRepository) queryVisitorSegments(ctx context.Context, start time.Time) (comment.VisitorSegments, error) {
	type row struct {
		Active1D      int64 `gorm:"column:active_1d"`
		Active3D      int64 `gorm:"column:active_3d"`
		Active7D      int64 `gorm:"column:active_7d"`
		Active30D     int64 `gorm:"column:active_30d"`
		HighlyEngaged int64 `gorm:"column:highly_engaged"`
	}
	now := time.Now().UTC()
	d1 := now.Add(-24 * time.Hour)
	d3 := now.Add(-72 * time.Hour)
	d7 := now.Add(-7 * 24 * time.Hour)
	d30 := now.Add(-30 * 24 * time.Hour)

	var result row
	err := r.db.WithContext(ctx).Raw(`
WITH last_activity AS (
    SELECT visitor_id, MAX(at) AS last_at
    FROM (
        SELECT visitor_id, MAX(last_view_at) AS at
        FROM analytics_visitor_view
        WHERE visitor_id <> ''
        GROUP BY visitor_id
        UNION ALL
        SELECT visitor_id, MAX(created_at) AS at
        FROM content_like
        WHERE visitor_id <> ''
        GROUP BY visitor_id
        UNION ALL
        SELECT visitor_id, MAX(created_at) AS at
        FROM comment
        WHERE visitor_id <> ''
        GROUP BY visitor_id
    ) t
    GROUP BY visitor_id
),
period_score AS (
    SELECT visitor_id, SUM(score) AS total_score
    FROM (
        SELECT visitor_id, COALESCE(SUM(view_count),0)::bigint AS score
        FROM analytics_visitor_view
        WHERE visitor_id <> '' AND last_view_at >= ?
        GROUP BY visitor_id
        UNION ALL
        SELECT visitor_id, COUNT(*)::bigint AS score
        FROM content_like
        WHERE visitor_id <> '' AND created_at >= ?
        GROUP BY visitor_id
        UNION ALL
        SELECT visitor_id, COUNT(*)::bigint AS score
        FROM comment
        WHERE visitor_id <> '' AND created_at >= ?
        GROUP BY visitor_id
    ) s
    GROUP BY visitor_id
)
SELECT
    (SELECT COUNT(*) FROM last_activity WHERE last_at >= ?) AS active_1d,
    (SELECT COUNT(*) FROM last_activity WHERE last_at >= ?) AS active_3d,
    (SELECT COUNT(*) FROM last_activity WHERE last_at >= ?) AS active_7d,
    (SELECT COUNT(*) FROM last_activity WHERE last_at >= ?) AS active_30d,
    (SELECT COUNT(*) FROM period_score WHERE total_score >= 20) AS highly_engaged
`, start, start, start, d1, d3, d7, d30).Scan(&result).Error
	if err != nil {
		return comment.VisitorSegments{}, err
	}

	return comment.VisitorSegments{
		Active1D:      result.Active1D,
		Active3D:      result.Active3D,
		Active7D:      result.Active7D,
		Active30D:     result.Active30D,
		HighlyEngaged: result.HighlyEngaged,
	}, nil
}

func (r *CommentRepository) queryDayCountMapRaw(ctx context.Context, sql string, args ...any) (map[string]int64, error) {
	type row struct {
		Day   string `gorm:"column:day"`
		Count int64  `gorm:"column:count"`
	}
	var rows []row
	if err := r.db.WithContext(ctx).Raw(sql, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[string]int64, len(rows))
	for _, row := range rows {
		day := strings.TrimSpace(row.Day)
		if day == "" {
			continue
		}
		out[day] = row.Count
	}
	return out, nil
}

func topDistribution(values map[string]int64, topN int) []comment.VisitorDistributionItem {
	if topN <= 0 {
		topN = 8
	}
	items := make([]comment.VisitorDistributionItem, 0, len(values))
	for name, count := range values {
		items = append(items, comment.VisitorDistributionItem{Name: name, Count: count})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Count == items[j].Count {
			return items[i].Name < items[j].Name
		}
		return items[i].Count > items[j].Count
	})
	if len(items) > topN {
		items = items[:topN]
	}
	return items
}

func (r *CommentRepository) loadLatestVisitorRows(ctx context.Context, visitorIDs []string) (map[string]visitorLatestRow, error) {
	if len(visitorIDs) == 0 {
		return map[string]visitorLatestRow{}, nil
	}

	commentLatest, err := r.loadLatestCommentVisitorRows(ctx, visitorIDs)
	if err != nil {
		return nil, err
	}
	viewLatest, err := r.loadLatestViewVisitorRows(ctx, visitorIDs)
	if err != nil {
		return nil, err
	}

	result := make(map[string]visitorLatestRow, len(visitorIDs))
	for _, visitorID := range visitorIDs {
		commentRow, hasComment := commentLatest[visitorID]
		viewRow, hasView := viewLatest[visitorID]
		switch {
		case hasComment && hasView:
			primary := commentRow
			secondary := viewRow
			if viewRow.ObservedAt.After(commentRow.ObservedAt) {
				primary = viewRow
				secondary = commentRow
			}
			merged := primary
			merged.NickName = preferNonEmpty(merged.NickName, secondary.NickName)
			merged.Email = preferNonEmpty(merged.Email, secondary.Email)
			merged.Website = preferNonEmpty(merged.Website, secondary.Website)
			merged.IP = preferNonEmpty(merged.IP, secondary.IP)
			merged.Location = preferNonEmpty(merged.Location, secondary.Location)
			merged.Platform = preferNonEmpty(merged.Platform, secondary.Platform)
			merged.Browser = preferNonEmpty(merged.Browser, secondary.Browser)
			if merged.ObservedAt.IsZero() {
				merged.ObservedAt = secondary.ObservedAt
			}
			result[visitorID] = merged
		case hasComment:
			result[visitorID] = commentRow
		case hasView:
			result[visitorID] = viewRow
		}
	}
	return result, nil
}

func (r *CommentRepository) loadLatestCommentVisitorRows(ctx context.Context, visitorIDs []string) (map[string]visitorLatestRow, error) {
	if len(visitorIDs) == 0 {
		return map[string]visitorLatestRow{}, nil
	}
	sub := r.db.WithContext(ctx).Unscoped().Table("comment").
		Select(
			"visitor_id",
			"nick_name",
			"email",
			"website",
			"ip",
			"location",
			"platform",
			"browser",
			"created_at AS observed_at",
			"ROW_NUMBER() OVER (PARTITION BY visitor_id ORDER BY created_at DESC, id DESC) AS rn",
		).
		Where("visitor_id IN ?", visitorIDs)

	var rows []visitorLatestRow
	if err := r.db.WithContext(ctx).Table("(?) AS visitor_latest", sub).Where("rn = 1").Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make(map[string]visitorLatestRow, len(rows))
	for _, row := range rows {
		result[row.VisitorID] = row
	}
	return result, nil
}

func (r *CommentRepository) loadLatestViewVisitorRows(ctx context.Context, visitorIDs []string) (map[string]visitorLatestRow, error) {
	if len(visitorIDs) == 0 {
		return map[string]visitorLatestRow{}, nil
	}
	sub := r.db.WithContext(ctx).Table("analytics_visitor_view").
		Select(
			"visitor_id",
			"'' AS nick_name",
			"'' AS email",
			"'' AS website",
			"last_ip AS ip",
			"location",
			"platform",
			"browser",
			"last_view_at AS observed_at",
			"ROW_NUMBER() OVER (PARTITION BY visitor_id ORDER BY last_view_at DESC, updated_at DESC) AS rn",
		).
		Where("visitor_id IN ?", visitorIDs)

	var rows []visitorLatestRow
	if err := r.db.WithContext(ctx).Table("(?) AS visitor_latest", sub).Where("rn = 1").Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make(map[string]visitorLatestRow, len(rows))
	for _, row := range rows {
		result[row.VisitorID] = row
	}
	return result, nil
}

func preferNonEmpty(primary, fallback string) string {
	if strings.TrimSpace(primary) != "" {
		return primary
	}
	return fallback
}

func mapCommentToDomain(rec model.Comment) comment.Comment {
	status := strings.TrimSpace(rec.Status)
	if status == "" {
		status = comment.CommentStatusApproved
	}
	canReply := rec.AllowLocalReply
	if !rec.IsFederated && !canReply {
		canReply = true
	}
	return comment.Comment{
		ID:                rec.ID,
		AreaID:            rec.AreaID,
		Content:           rec.Content,
		AuthorID:          rec.AuthorID,
		VisitorID:         toPtr(rec.VisitorID),
		NickName:          toPtr(rec.NickName),
		IP:                toPtr(rec.IP),
		Location:          toPtr(rec.Location),
		Platform:          toPtr(rec.Platform),
		Browser:           toPtr(rec.Browser),
		Email:             toPtr(rec.Email),
		Website:           toPtr(rec.Website),
		Avatar:            toPtr(rec.Avatar),
		IsOwner:           rec.IsOwner,
		IsFriend:          rec.IsFriend,
		IsAuthor:          rec.IsAuthor,
		IsViewed:          rec.IsViewed,
		IsTop:             rec.IsTop,
		IsMy:              false,
		IsFederated:       rec.IsFederated,
		FederatedProtocol: toPtr(rec.FederatedProtocol),
		FederatedActor:    toPtr(rec.FederatedActor),
		FederatedObjectID: toPtr(rec.FederatedObjectID),
		CanReply:          canReply,
		Status:            status,
		IsEdited:          rec.IsEdited,
		ParentID:          rec.ParentID,
		CreatedAt:         rec.CreatedAt,
		UpdatedAt:         rec.UpdatedAt,
		DeletedAt:         timeToPtr(rec.DeletedAt),
	}
}

func mapCommentToModel(entity *comment.Comment) model.Comment {
	status := strings.TrimSpace(entity.Status)
	if status == "" {
		status = comment.CommentStatusPending
	}
	allowLocalReply := entity.CanReply
	if !entity.IsFederated && !allowLocalReply {
		allowLocalReply = true
	}
	return model.Comment{
		ID:                entity.ID,
		AreaID:            entity.AreaID,
		Content:           strings.TrimSpace(entity.Content),
		AuthorID:          entity.AuthorID,
		VisitorID:         toValue(entity.VisitorID),
		NickName:          toValue(entity.NickName),
		IP:                toValue(entity.IP),
		Location:          toValue(entity.Location),
		Platform:          toValue(entity.Platform),
		Browser:           toValue(entity.Browser),
		Email:             toValue(entity.Email),
		Website:           toValue(entity.Website),
		Avatar:            toValue(entity.Avatar),
		IsOwner:           entity.IsOwner,
		IsFriend:          entity.IsFriend,
		IsAuthor:          entity.IsAuthor,
		IsViewed:          entity.IsViewed,
		IsTop:             entity.IsTop,
		IsFederated:       entity.IsFederated,
		FederatedProtocol: toValue(entity.FederatedProtocol),
		FederatedActor:    toValue(entity.FederatedActor),
		FederatedObjectID: toValue(entity.FederatedObjectID),
		AllowLocalReply:   allowLocalReply,
		Status:            status,
		IsEdited:          entity.IsEdited,
		ParentID:          entity.ParentID,
		CreatedAt:         entity.CreatedAt,
		UpdatedAt:         entity.UpdatedAt,
		DeletedAt:         gorm.DeletedAt{Time: timeToValue(entity.DeletedAt), Valid: entity.DeletedAt != nil},
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

func summarizeThinkingContent(content string) string {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return ""
	}
	const maxRunes = 40
	runes := []rune(trimmed)
	if len(runes) <= maxRunes {
		return trimmed
	}
	return strings.TrimSpace(string(runes[:maxRunes])) + "..."
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
