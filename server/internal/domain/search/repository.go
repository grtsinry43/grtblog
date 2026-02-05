package search

import "context"

type Repository interface {
	SearchSite(ctx context.Context, keyword string, limitPerGroup int) ([]Hit, error)
}
