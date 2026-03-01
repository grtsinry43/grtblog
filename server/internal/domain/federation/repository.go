package federation

import (
	"context"
	"time"
)

// FederationInstanceRepository manages remote instance records.
type FederationInstanceRepository interface {
	GetByID(ctx context.Context, id int64) (*FederationInstance, error)
	GetByBaseURL(ctx context.Context, baseURL string) (*FederationInstance, error)
	Create(ctx context.Context, instance *FederationInstance) error
	Update(ctx context.Context, instance *FederationInstance) error
	ListActive(ctx context.Context) ([]FederationInstance, error)
	List(ctx context.Context, status string, keyword string, page int, pageSize int) ([]FederationInstance, int64, error)
}

// AuthorInfo represents a de-duplicated author extracted from cached posts.
type AuthorInfo struct {
	Name         string
	InstanceURL  string
	InstanceName string
}

// FederatedPostCacheRepository stores cached timeline posts.
type FederatedPostCacheRepository interface {
	UpsertBatch(ctx context.Context, posts []FederatedPostCache) error
	ListByInstance(ctx context.Context, instanceID int64, since *time.Time, limit int) ([]FederatedPostCache, error)
	ListRecent(ctx context.Context, limit int) ([]FederatedPostCache, error)
	ListTimeline(ctx context.Context, page, pageSize int) ([]FederatedPostCache, int64, error)
	SearchPostsByInstance(ctx context.Context, instanceID int64, keyword string, limit int) ([]FederatedPostCache, error)
	SearchAuthors(ctx context.Context, keyword string, limit int) ([]AuthorInfo, error)
}

// FederatedCitationRepository stores citation workflows.
type FederatedCitationRepository interface {
	Create(ctx context.Context, citation *FederatedCitation) error
	GetByID(ctx context.Context, id int64) (*FederatedCitation, error)
	UpdateStatus(ctx context.Context, id int64, status string, reason *string) error
	ListByTarget(ctx context.Context, articleID int64, status string) ([]FederatedCitation, error)
	List(ctx context.Context, status string, limit int) ([]FederatedCitation, error)
}

// FederatedMentionRepository stores mentions delivered to local users.
type FederatedMentionRepository interface {
	Create(ctx context.Context, mention *FederatedMention) error
	GetByID(ctx context.Context, id int64) (*FederatedMention, error)
	UpdateStatus(ctx context.Context, id int64, status string, reason *string) error
	MarkRead(ctx context.Context, id int64) error
	ListByUser(ctx context.Context, userID int64, unreadOnly bool) ([]FederatedMention, error)
	List(ctx context.Context, status string, limit int) ([]FederatedMention, error)
}

// OutboundDeliveryRepository stores local outbound federation deliveries.
type OutboundDeliveryRepository interface {
	Create(ctx context.Context, delivery *OutboundDelivery) error
	GetByID(ctx context.Context, id int64) (*OutboundDelivery, error)
	GetByRequestID(ctx context.Context, requestID string) (*OutboundDelivery, error)
	Update(ctx context.Context, delivery *OutboundDelivery) error
	List(ctx context.Context, options OutboundDeliveryListOptions) ([]OutboundDelivery, int64, error)
	ListRetryable(ctx context.Context, now time.Time, limit int) ([]OutboundDelivery, error)
	ListBySourceArticle(ctx context.Context, articleID int64, limit int) ([]OutboundDelivery, error)
}
