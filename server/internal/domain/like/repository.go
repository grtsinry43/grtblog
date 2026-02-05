package like

import "context"

type Repository interface {
	ExistsTarget(ctx context.Context, targetType TargetType, targetID int64) (bool, error)
	CreateIfAbsent(ctx context.Context, like *ContentLike) (bool, error)
}
