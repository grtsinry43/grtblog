package ai

import "context"

type Repository interface {
	// Provider CRUD
	CreateProvider(ctx context.Context, p *Provider) error
	GetProviderByID(ctx context.Context, id int64) (*Provider, error)
	ListProviders(ctx context.Context) ([]*Provider, error)
	UpdateProvider(ctx context.Context, p *Provider) error
	DeleteProvider(ctx context.Context, id int64) error

	// Model CRUD
	CreateModel(ctx context.Context, m *Model) error
	GetModelByID(ctx context.Context, id int64) (*Model, error)
	GetModelWithProvider(ctx context.Context, modelID int64) (*Model, *Provider, error)
	ListModels(ctx context.Context) ([]*Model, error)
	ListModelsByProvider(ctx context.Context, providerID int64) ([]*Model, error)
	UpdateModel(ctx context.Context, m *Model) error
	DeleteModel(ctx context.Context, id int64) error

	// TaskLog
	CreateTaskLog(ctx context.Context, log *TaskLog) error
	UpdateTaskLog(ctx context.Context, log *TaskLog) error
	GetTaskLogByID(ctx context.Context, id int64) (*TaskLog, error)
	ListTaskLogs(ctx context.Context, opts TaskLogListOptions) ([]*TaskLog, int64, error)
}
