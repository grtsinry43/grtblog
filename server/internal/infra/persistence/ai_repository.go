package persistence

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	domainai "github.com/grtsinry43/grtblog-v2/server/internal/domain/ai"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence/model"
)

type AIRepository struct {
	db           *gorm.DB
	providerRepo *GormRepository[model.AIProvider]
	modelRepo    *GormRepository[model.AIModel]
	taskLogRepo  *GormRepository[model.AITaskLog]
}

func NewAIRepository(db *gorm.DB) *AIRepository {
	return &AIRepository{
		db:           db,
		providerRepo: NewGormRepository[model.AIProvider](db),
		modelRepo:    NewGormRepository[model.AIModel](db),
		taskLogRepo:  NewGormRepository[model.AITaskLog](db),
	}
}

// ── Provider CRUD ──

func (r *AIRepository) CreateProvider(ctx context.Context, p *domainai.Provider) error {
	rec := toProviderModel(p)
	if err := r.providerRepo.Create(ctx, &rec); err != nil {
		return err
	}
	p.ID = rec.ID
	p.CreatedAt = rec.CreatedAt
	p.UpdatedAt = rec.UpdatedAt
	return nil
}

func (r *AIRepository) GetProviderByID(ctx context.Context, id int64) (*domainai.Provider, error) {
	rec, err := r.providerRepo.FirstByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainai.ErrProviderNotFound
		}
		return nil, err
	}
	return toProviderDomain(rec), nil
}

func (r *AIRepository) ListProviders(ctx context.Context) ([]*domainai.Provider, error) {
	records, err := r.providerRepo.List(ctx, func(db *gorm.DB) *gorm.DB {
		return db.Order("id ASC")
	})
	if err != nil {
		return nil, err
	}
	result := make([]*domainai.Provider, len(records))
	for i := range records {
		result[i] = toProviderDomain(&records[i])
	}
	return result, nil
}

func (r *AIRepository) UpdateProvider(ctx context.Context, p *domainai.Provider) error {
	rec := toProviderModel(p)
	_, err := r.providerRepo.UpdateByID(ctx, p.ID, &rec)
	return err
}

func (r *AIRepository) DeleteProvider(ctx context.Context, id int64) error {
	return r.providerRepo.DeleteByID(ctx, id)
}

// ── Model CRUD ──

func (r *AIRepository) CreateModel(ctx context.Context, m *domainai.Model) error {
	rec := toModelModel(m)
	if err := r.modelRepo.Create(ctx, &rec); err != nil {
		return err
	}
	m.ID = rec.ID
	m.CreatedAt = rec.CreatedAt
	m.UpdatedAt = rec.UpdatedAt
	return nil
}

func (r *AIRepository) GetModelByID(ctx context.Context, id int64) (*domainai.Model, error) {
	rec, err := r.modelRepo.FirstByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainai.ErrModelNotFound
		}
		return nil, err
	}
	return toModelDomain(rec), nil
}

func (r *AIRepository) GetModelWithProvider(ctx context.Context, modelID int64) (*domainai.Model, *domainai.Provider, error) {
	m, err := r.GetModelByID(ctx, modelID)
	if err != nil {
		return nil, nil, err
	}
	p, err := r.GetProviderByID(ctx, m.ProviderID)
	if err != nil {
		return nil, nil, err
	}
	return m, p, nil
}

func (r *AIRepository) ListModels(ctx context.Context) ([]*domainai.Model, error) {
	records, err := r.modelRepo.List(ctx, func(db *gorm.DB) *gorm.DB {
		return db.Order("provider_id ASC, id ASC")
	})
	if err != nil {
		return nil, err
	}
	result := make([]*domainai.Model, len(records))
	for i := range records {
		result[i] = toModelDomain(&records[i])
	}
	return result, nil
}

func (r *AIRepository) ListModelsByProvider(ctx context.Context, providerID int64) ([]*domainai.Model, error) {
	records, err := r.modelRepo.List(ctx, func(db *gorm.DB) *gorm.DB {
		return db.Where("provider_id = ?", providerID).Order("id ASC")
	})
	if err != nil {
		return nil, err
	}
	result := make([]*domainai.Model, len(records))
	for i := range records {
		result[i] = toModelDomain(&records[i])
	}
	return result, nil
}

func (r *AIRepository) UpdateModel(ctx context.Context, m *domainai.Model) error {
	rec := toModelModel(m)
	_, err := r.modelRepo.UpdateByID(ctx, m.ID, &rec)
	return err
}

func (r *AIRepository) DeleteModel(ctx context.Context, id int64) error {
	return r.modelRepo.DeleteByID(ctx, id)
}

// ── Mapper helpers ──

func toProviderModel(p *domainai.Provider) model.AIProvider {
	return model.AIProvider{
		ID:        p.ID,
		Name:      p.Name,
		Type:      p.Type,
		APIURL:    p.APIURL,
		APIKey:    p.APIKey,
		IsActive:  p.IsActive,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

func toProviderDomain(rec *model.AIProvider) *domainai.Provider {
	return &domainai.Provider{
		ID:        rec.ID,
		Name:      rec.Name,
		Type:      rec.Type,
		APIURL:    rec.APIURL,
		APIKey:    rec.APIKey,
		IsActive:  rec.IsActive,
		CreatedAt: rec.CreatedAt,
		UpdatedAt: rec.UpdatedAt,
	}
}

func toModelModel(m *domainai.Model) model.AIModel {
	return model.AIModel{
		ID:         m.ID,
		ProviderID: m.ProviderID,
		Name:       m.Name,
		ModelID:    m.ModelID,
		IsActive:   m.IsActive,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
}

func toModelDomain(rec *model.AIModel) *domainai.Model {
	return &domainai.Model{
		ID:         rec.ID,
		ProviderID: rec.ProviderID,
		Name:       rec.Name,
		ModelID:    rec.ModelID,
		IsActive:   rec.IsActive,
		CreatedAt:  rec.CreatedAt,
		UpdatedAt:  rec.UpdatedAt,
	}
}

// ── TaskLog CRUD ──

func (r *AIRepository) CreateTaskLog(ctx context.Context, l *domainai.TaskLog) error {
	rec := toTaskLogModel(l)
	if err := r.taskLogRepo.Create(ctx, &rec); err != nil {
		return err
	}
	l.ID = rec.ID
	l.CreatedAt = rec.CreatedAt
	l.UpdatedAt = rec.UpdatedAt
	return nil
}

func (r *AIRepository) UpdateTaskLog(ctx context.Context, l *domainai.TaskLog) error {
	rec := toTaskLogModel(l)
	return r.db.WithContext(ctx).Save(&rec).Error
}

func (r *AIRepository) GetTaskLogByID(ctx context.Context, id int64) (*domainai.TaskLog, error) {
	rec, err := r.taskLogRepo.FirstByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("task log not found")
		}
		return nil, err
	}
	return toTaskLogDomain(rec), nil
}

func (r *AIRepository) ListTaskLogs(ctx context.Context, opts domainai.TaskLogListOptions) ([]*domainai.TaskLog, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.AITaskLog{})

	if opts.TaskType != nil && *opts.TaskType != "" {
		query = query.Where("task_type = ?", *opts.TaskType)
	}
	if opts.Status != nil && *opts.Status != "" {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Search != nil && *opts.Search != "" {
		like := "%" + *opts.Search + "%"
		query = query.Where("input_text ILIKE ? OR output_text ILIKE ? OR model_name ILIKE ?", like, like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := opts.Page
	if page < 1 {
		page = 1
	}
	pageSize := opts.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	var records []model.AITaskLog
	if err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&records).Error; err != nil {
		return nil, 0, err
	}

	result := make([]*domainai.TaskLog, len(records))
	for i := range records {
		result[i] = toTaskLogDomain(&records[i])
	}
	return result, total, nil
}

func toTaskLogModel(l *domainai.TaskLog) model.AITaskLog {
	rec := model.AITaskLog{
		ID:            l.ID,
		TaskType:      l.TaskType,
		ModelName:     l.ModelName,
		ProviderName:  l.ProviderName,
		Status:        l.Status,
		InputText:     l.InputText,
		OutputText:    l.OutputText,
		DurationMs:    l.DurationMs,
		TriggerSource: l.TriggerSource,
		CreatedAt:     l.CreatedAt,
		UpdatedAt:     l.UpdatedAt,
	}
	if l.ErrorMessage != "" {
		rec.ErrorMessage = &l.ErrorMessage
	}
	return rec
}

func toTaskLogDomain(rec *model.AITaskLog) *domainai.TaskLog {
	l := &domainai.TaskLog{
		ID:            rec.ID,
		TaskType:      rec.TaskType,
		ModelName:     rec.ModelName,
		ProviderName:  rec.ProviderName,
		Status:        rec.Status,
		InputText:     rec.InputText,
		OutputText:    rec.OutputText,
		DurationMs:    rec.DurationMs,
		TriggerSource: rec.TriggerSource,
		CreatedAt:     rec.CreatedAt,
		UpdatedAt:     rec.UpdatedAt,
	}
	if rec.ErrorMessage != nil {
		l.ErrorMessage = *rec.ErrorMessage
	}
	return l
}
