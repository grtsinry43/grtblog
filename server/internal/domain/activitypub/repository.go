package activitypub

import "context"

type FollowerRepository interface {
	Upsert(ctx context.Context, follower *Follower) error
	GetByActorID(ctx context.Context, actorID string) (*Follower, error)
	List(ctx context.Context, status string, page, pageSize int) ([]Follower, int64, error)
	ListActive(ctx context.Context) ([]Follower, error)
}

type OutboxRepository interface {
	Create(ctx context.Context, item *OutboxItem) error
	List(ctx context.Context, page, pageSize int) ([]OutboxItem, int64, error)
}
