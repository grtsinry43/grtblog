package config

import "context"

// SysConfigRepository 抽象系统配置存储。
type SysConfigRepository interface {
	GetByKey(ctx context.Context, key string) (*SysConfig, error)
	List(ctx context.Context, keys []string) ([]SysConfig, error)
	Upsert(ctx context.Context, configs []SysConfig) error
}
