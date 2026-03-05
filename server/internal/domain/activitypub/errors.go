package activitypub

import "errors"

var (
	ErrFollowerNotFound       = errors.New("activitypub follower not found")
	ErrOutboxItemNotFound     = errors.New("activitypub outbox item not found")
	ErrOutboxItemNotRetryable = errors.New("activitypub outbox item not retryable")
)
