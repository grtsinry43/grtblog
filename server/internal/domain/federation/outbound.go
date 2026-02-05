package federation

import (
	"encoding/json"
	"time"
)

const (
	DeliveryTypeFriendLink = "friendlink"
	DeliveryTypeCitation   = "citation"
	DeliveryTypeMention    = "mention"
)

const (
	DeliveryStatusQueued   = "queued"
	DeliveryStatusSending  = "sending"
	DeliveryStatusAccepted = "accepted"
	DeliveryStatusApproved = "approved"
	DeliveryStatusRejected = "rejected"
	DeliveryStatusFailed   = "failed"
	DeliveryStatusTimeout  = "timeout"
	DeliveryStatusDead     = "dead"
)

type OutboundDelivery struct {
	ID                int64
	RequestID         string
	DeliveryType      string
	SourceArticleID   *int64
	TargetInstanceURL string
	TargetEndpoint    string
	Payload           json.RawMessage
	Status            string
	AttemptCount      int
	MaxAttempts       int
	NextRetryAt       *time.Time
	HTTPStatus        *int
	ResponseBody      *string
	ErrorMessage      *string
	RemoteTicketID    *string
	TraceID           *string
	LastCallbackAt    *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type OutboundDeliveryListOptions struct {
	RequestID string
	Type      string
	Status    string
	Target    string
	Page      int
	PageSize  int
}
