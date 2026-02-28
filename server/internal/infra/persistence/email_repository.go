package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	domainemail "github.com/grtsinry43/grtblog-v2/server/internal/domain/email"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence/model"
)

type EmailRepository struct {
	db *gorm.DB
}

func NewEmailRepository(db *gorm.DB) *EmailRepository {
	return &EmailRepository{db: db}
}

func (r *EmailRepository) CreateTemplate(ctx context.Context, tpl *domainemail.Template) error {
	rec, err := mapTemplateToModel(tpl)
	if err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Create(&rec).Error; err != nil {
		if isTemplateCodeConstraint(err) {
			return domainemail.ErrEmailTemplateCodeExists
		}
		return err
	}
	tpl.ID = rec.ID
	tpl.CreatedAt = rec.CreatedAt
	tpl.UpdatedAt = rec.UpdatedAt
	return nil
}

func (r *EmailRepository) UpdateTemplate(ctx context.Context, tpl *domainemail.Template) error {
	rec, err := mapTemplateToModel(tpl)
	if err != nil {
		return err
	}
	result := r.db.WithContext(ctx).
		Model(&model.EmailTemplate{}).
		Where("code = ?", tpl.Code).
		Updates(map[string]any{
			"name":             rec.Name,
			"event_name":       rec.EventName,
			"subject_template": rec.SubjectTemplate,
			"html_template":    rec.HTMLTemplate,
			"text_template":    rec.TextTemplate,
			"to_emails":        rec.ToEmails,
			"is_enabled":       rec.IsEnabled,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domainemail.ErrEmailTemplateNotFound
	}
	return nil
}

func (r *EmailRepository) DeleteTemplateByCode(ctx context.Context, code string) error {
	result := r.db.WithContext(ctx).Where("code = ?", code).Delete(&model.EmailTemplate{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domainemail.ErrEmailTemplateNotFound
	}
	return nil
}

func (r *EmailRepository) GetTemplateByCode(ctx context.Context, code string) (*domainemail.Template, error) {
	var rec model.EmailTemplate
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&rec).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainemail.ErrEmailTemplateNotFound
		}
		return nil, err
	}
	return mapTemplateToDomain(rec)
}

func (r *EmailRepository) ListTemplates(ctx context.Context) ([]*domainemail.Template, error) {
	var records []model.EmailTemplate
	if err := r.db.WithContext(ctx).Order("updated_at DESC").Find(&records).Error; err != nil {
		return nil, err
	}
	items := make([]*domainemail.Template, len(records))
	for i, rec := range records {
		item, err := mapTemplateToDomain(rec)
		if err != nil {
			return nil, err
		}
		items[i] = item
	}
	return items, nil
}

func (r *EmailRepository) ListEnabledTemplatesByEvent(ctx context.Context, eventName string) ([]*domainemail.Template, error) {
	var records []model.EmailTemplate
	if err := r.db.WithContext(ctx).
		Where("event_name = ?", eventName).
		Where("is_enabled = ?", true).
		Order("updated_at DESC").
		Find(&records).Error; err != nil {
		return nil, err
	}
	items := make([]*domainemail.Template, len(records))
	for i, rec := range records {
		item, err := mapTemplateToDomain(rec)
		if err != nil {
			return nil, err
		}
		items[i] = item
	}
	return items, nil
}

func (r *EmailRepository) CreateOutbox(ctx context.Context, item *domainemail.Outbox) error {
	rec, err := mapOutboxToModel(item)
	if err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Create(&rec).Error; err != nil {
		return err
	}
	item.ID = rec.ID
	item.CreatedAt = rec.CreatedAt
	item.UpdatedAt = rec.UpdatedAt
	return nil
}

func (r *EmailRepository) ClaimDueOutbox(ctx context.Context, limit int, dueAt time.Time, maxRetries int) ([]*domainemail.Outbox, error) {
	if limit <= 0 {
		limit = 1
	}
	if maxRetries <= 0 {
		maxRetries = 1
	}

	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var records []model.EmailOutbox
	if err := tx.
		Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
		Where("status IN ?", []string{domainemail.OutboxStatusPending, domainemail.OutboxStatusFailed}).
		Where("next_retry_at <= ?", dueAt).
		Where("retry_count < ?", maxRetries).
		Order("next_retry_at ASC").
		Limit(limit).
		Find(&records).Error; err != nil {
		return nil, err
	}
	if len(records) == 0 {
		if err := tx.Commit().Error; err != nil {
			return nil, err
		}
		return []*domainemail.Outbox{}, nil
	}

	ids := make([]int64, len(records))
	for i, rec := range records {
		ids[i] = rec.ID
		rec.Status = domainemail.OutboxStatusSending
		records[i] = rec
	}
	now := time.Now()
	if err := tx.Model(&model.EmailOutbox{}).
		Where("id IN ?", ids).
		Updates(map[string]any{
			"status":     domainemail.OutboxStatusSending,
			"updated_at": now,
		}).Error; err != nil {
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	items := make([]*domainemail.Outbox, len(records))
	for i, rec := range records {
		item, err := mapOutboxToDomain(rec)
		if err != nil {
			return nil, err
		}
		item.Status = domainemail.OutboxStatusSending
		items[i] = item
	}
	return items, nil
}

func (r *EmailRepository) MarkOutboxSent(ctx context.Context, id int64, sentAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&model.EmailOutbox{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":     domainemail.OutboxStatusSent,
			"sent_at":    sentAt,
			"updated_at": sentAt,
			"last_error": "",
		}).Error
}

func (r *EmailRepository) MarkOutboxFailed(ctx context.Context, id int64, retryCount int, nextRetryAt time.Time, lastError string) error {
	return r.db.WithContext(ctx).
		Model(&model.EmailOutbox{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":        domainemail.OutboxStatusFailed,
			"retry_count":   retryCount,
			"next_retry_at": nextRetryAt,
			"last_error":    lastError,
			"updated_at":    time.Now(),
		}).Error
}

func (r *EmailRepository) ListOutbox(ctx context.Context, options domainemail.OutboxListOptions) ([]*domainemail.Outbox, int64, error) {
	page := options.Page
	if page <= 0 {
		page = 1
	}
	pageSize := options.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	query := r.db.WithContext(ctx).Model(&model.EmailOutbox{})
	if options.Status != nil && strings.TrimSpace(*options.Status) != "" {
		query = query.Where("status = ?", strings.TrimSpace(*options.Status))
	}
	if options.EventName != nil && strings.TrimSpace(*options.EventName) != "" {
		query = query.Where("event_name = ?", strings.TrimSpace(*options.EventName))
	}
	if options.Search != nil && strings.TrimSpace(*options.Search) != "" {
		kw := "%" + strings.TrimSpace(*options.Search) + "%"
		query = query.Where("subject ILIKE ? OR template_code ILIKE ?", kw, kw)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var records []model.EmailOutbox
	if err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&records).Error; err != nil {
		return nil, 0, err
	}
	items := make([]*domainemail.Outbox, len(records))
	for i, rec := range records {
		item, err := mapOutboxToDomain(rec)
		if err != nil {
			return nil, 0, err
		}
		items[i] = item
	}
	return items, total, nil
}

func (r *EmailRepository) GetOutboxByID(ctx context.Context, id int64) (*domainemail.Outbox, error) {
	var rec model.EmailOutbox
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&rec).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainemail.ErrEmailOutboxNotFound
		}
		return nil, err
	}
	return mapOutboxToDomain(rec)
}

func (r *EmailRepository) CreateOrUpdateSubscription(ctx context.Context, sub *domainemail.Subscription) error {
	rec := mapSubscriptionToModel(sub)
	now := time.Now()
	updateMap := map[string]any{
		"status":          rec.Status,
		"token":           rec.Token,
		"source_ip":       rec.SourceIP,
		"updated_at":      now,
		"unsubscribed_at": nil,
	}
	if rec.Status == domainemail.SubscriptionStatusUnsubscribed {
		updateMap["unsubscribed_at"] = now
	}
	if rec.Status == domainemail.SubscriptionStatusBlocked {
		updateMap["unsubscribed_at"] = nil
	}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "email"}, {Name: "event_name"}},
			DoUpdates: clause.Assignments(updateMap),
		}).
		Create(&rec).Error
}

func (r *EmailRepository) GetSubscriptionByEmailEvent(ctx context.Context, email string, eventName string) (*domainemail.Subscription, error) {
	var rec model.EmailSubscription
	if err := r.db.WithContext(ctx).
		Where("email = ? AND event_name = ?", email, eventName).
		First(&rec).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainemail.ErrEmailSubscriptionNotFound
		}
		return nil, err
	}
	return mapSubscriptionToDomain(rec), nil
}

func (r *EmailRepository) UnsubscribeByToken(ctx context.Context, token string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&model.EmailSubscription{}).
		Where("token = ?", token).
		Updates(map[string]any{
			"status":          domainemail.SubscriptionStatusUnsubscribed,
			"unsubscribed_at": now,
			"updated_at":      now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domainemail.ErrEmailSubscriptionNotFound
	}
	return nil
}

func (r *EmailRepository) UnsubscribeByEmailEvent(ctx context.Context, email string, eventName string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&model.EmailSubscription{}).
		Where("email = ? AND event_name = ?", email, eventName).
		Updates(map[string]any{
			"status":          domainemail.SubscriptionStatusUnsubscribed,
			"unsubscribed_at": now,
			"updated_at":      now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domainemail.ErrEmailSubscriptionNotFound
	}
	return nil
}

func (r *EmailRepository) ListSubscriptions(ctx context.Context, options domainemail.SubscriptionListOptions) ([]*domainemail.Subscription, int64, error) {
	page := options.Page
	if page <= 0 {
		page = 1
	}
	pageSize := options.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	query := r.db.WithContext(ctx).Model(&model.EmailSubscription{})
	if options.EventName != nil && strings.TrimSpace(*options.EventName) != "" {
		query = query.Where("event_name = ?", strings.TrimSpace(*options.EventName))
	}
	if options.Status != nil && strings.TrimSpace(*options.Status) != "" {
		query = query.Where("status = ?", strings.TrimSpace(*options.Status))
	}
	if options.Search != nil && strings.TrimSpace(*options.Search) != "" {
		kw := "%" + strings.TrimSpace(*options.Search) + "%"
		query = query.Where("email ILIKE ?", kw)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var records []model.EmailSubscription
	if err := query.Order("updated_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&records).Error; err != nil {
		return nil, 0, err
	}
	items := make([]*domainemail.Subscription, len(records))
	for i, rec := range records {
		items[i] = mapSubscriptionToDomain(rec)
	}
	return items, total, nil
}

func (r *EmailRepository) BatchUpdateSubscriptionStatus(ctx context.Context, ids []int64, status string) error {
	if len(ids) == 0 {
		return nil
	}
	now := time.Now()
	updates := map[string]any{
		"status":     status,
		"updated_at": now,
	}
	switch status {
	case domainemail.SubscriptionStatusUnsubscribed:
		updates["unsubscribed_at"] = now
	default:
		updates["unsubscribed_at"] = nil
	}
	result := r.db.WithContext(ctx).
		Model(&model.EmailSubscription{}).
		Where("id IN ?", ids).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domainemail.ErrEmailSubscriptionNotFound
	}
	return nil
}

func (r *EmailRepository) ListActiveSubscriberEmailsByEvent(ctx context.Context, eventName string) ([]string, error) {
	var emails []string
	if err := r.db.WithContext(ctx).
		Model(&model.EmailSubscription{}).
		Where("event_name = ?", eventName).
		Where("status = ?", domainemail.SubscriptionStatusActive).
		Distinct("email").
		Pluck("email", &emails).Error; err != nil {
		return nil, err
	}
	return emails, nil
}

func mapTemplateToModel(tpl *domainemail.Template) (model.EmailTemplate, error) {
	if tpl.ToEmails == nil {
		tpl.ToEmails = []string{}
	}
	toBytes, err := json.Marshal(tpl.ToEmails)
	if err != nil {
		return model.EmailTemplate{}, err
	}
	return model.EmailTemplate{
		ID:              tpl.ID,
		Code:            tpl.Code,
		Name:            tpl.Name,
		EventName:       tpl.EventName,
		SubjectTemplate: tpl.SubjectTemplate,
		HTMLTemplate:    tpl.HTMLTemplate,
		TextTemplate:    tpl.TextTemplate,
		ToEmails:        datatypes.JSON(toBytes),
		IsEnabled:       tpl.IsEnabled,
		IsInternal:      tpl.IsInternal,
	}, nil
}

func mapTemplateToDomain(rec model.EmailTemplate) (*domainemail.Template, error) {
	toEmails := []string{}
	if len(rec.ToEmails) > 0 {
		if err := json.Unmarshal(rec.ToEmails, &toEmails); err != nil {
			return nil, err
		}
	}
	return &domainemail.Template{
		ID:              rec.ID,
		Code:            rec.Code,
		Name:            rec.Name,
		EventName:       rec.EventName,
		SubjectTemplate: rec.SubjectTemplate,
		HTMLTemplate:    rec.HTMLTemplate,
		TextTemplate:    rec.TextTemplate,
		ToEmails:        toEmails,
		IsEnabled:       rec.IsEnabled,
		IsInternal:      rec.IsInternal,
		CreatedAt:       rec.CreatedAt,
		UpdatedAt:       rec.UpdatedAt,
		DeletedAt:       deletedAtToPtr(rec.DeletedAt),
	}, nil
}

func mapOutboxToModel(item *domainemail.Outbox) (model.EmailOutbox, error) {
	if item.ToEmails == nil {
		item.ToEmails = []string{}
	}
	toBytes, err := json.Marshal(item.ToEmails)
	if err != nil {
		return model.EmailOutbox{}, err
	}
	status := item.Status
	if status == "" {
		status = domainemail.OutboxStatusPending
	}
	nextRetryAt := item.NextRetryAt
	if nextRetryAt.IsZero() {
		nextRetryAt = time.Now()
	}
	return model.EmailOutbox{
		ID:           item.ID,
		TemplateID:   item.TemplateID,
		TemplateCode: item.TemplateCode,
		EventName:    item.EventName,
		ToEmails:     datatypes.JSON(toBytes),
		Subject:      item.Subject,
		HTMLBody:     item.HTMLBody,
		TextBody:     item.TextBody,
		Status:       status,
		RetryCount:   item.RetryCount,
		NextRetryAt:  nextRetryAt,
		LastError:    item.LastError,
		SentAt:       item.SentAt,
	}, nil
}

func mapOutboxToDomain(rec model.EmailOutbox) (*domainemail.Outbox, error) {
	toEmails := []string{}
	if len(rec.ToEmails) > 0 {
		if err := json.Unmarshal(rec.ToEmails, &toEmails); err != nil {
			return nil, err
		}
	}
	return &domainemail.Outbox{
		ID:           rec.ID,
		TemplateID:   rec.TemplateID,
		TemplateCode: rec.TemplateCode,
		EventName:    rec.EventName,
		ToEmails:     toEmails,
		Subject:      rec.Subject,
		HTMLBody:     rec.HTMLBody,
		TextBody:     rec.TextBody,
		Status:       rec.Status,
		RetryCount:   rec.RetryCount,
		NextRetryAt:  rec.NextRetryAt,
		LastError:    rec.LastError,
		SentAt:       rec.SentAt,
		CreatedAt:    rec.CreatedAt,
		UpdatedAt:    rec.UpdatedAt,
	}, nil
}

func mapSubscriptionToModel(sub *domainemail.Subscription) model.EmailSubscription {
	return model.EmailSubscription{
		ID:             sub.ID,
		Email:          sub.Email,
		EventName:      sub.EventName,
		Status:         sub.Status,
		Token:          sub.Token,
		SourceIP:       sub.SourceIP,
		UnsubscribedAt: sub.UnsubscribedAt,
	}
}

func mapSubscriptionToDomain(rec model.EmailSubscription) *domainemail.Subscription {
	return &domainemail.Subscription{
		ID:             rec.ID,
		Email:          rec.Email,
		EventName:      rec.EventName,
		Status:         rec.Status,
		Token:          rec.Token,
		SourceIP:       rec.SourceIP,
		UnsubscribedAt: rec.UnsubscribedAt,
		CreatedAt:      rec.CreatedAt,
		UpdatedAt:      rec.UpdatedAt,
	}
}

func isTemplateCodeConstraint(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "uq_email_template_code")
}
