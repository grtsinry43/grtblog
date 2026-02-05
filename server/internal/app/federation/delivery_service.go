package federation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	domainfed "github.com/grtsinry43/grtblog-v2/server/internal/domain/federation"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
)

type CallbackResultCmd struct {
	RequestID      string
	Type           string
	Status         string
	RemoteTicketID string
	Reason         string
	ProcessedAt    *time.Time
}

type DeliveryService struct {
	repo     domainfed.OutboundDeliveryRepository
	outbound *OutboundService
	events   appEvent.Bus
}

func NewDeliveryService(repo domainfed.OutboundDeliveryRepository, outbound *OutboundService, events appEvent.Bus) *DeliveryService {
	if events == nil {
		events = appEvent.NopBus{}
	}
	return &DeliveryService{repo: repo, outbound: outbound, events: events}
}

func (s *DeliveryService) DispatchFriendLink(ctx context.Context, target, message, rssURL string, traceID *string) (*domainfed.OutboundDelivery, error) {
	payload, _ := json.Marshal(contract.FederationFriendLinkRequestReq{
		RequesterURL: "",
		Message:      message,
		RSSURL:       rssURL,
	})
	delivery := s.newDelivery(domainfed.DeliveryTypeFriendLink, nil, target, payload, traceID)
	if err := s.repo.Create(ctx, delivery); err != nil {
		return nil, err
	}
	return s.sendFriendLink(ctx, delivery, message, rssURL)
}

func (s *DeliveryService) DispatchCitation(ctx context.Context, ev CitationDetected, traceID *string) (*domainfed.OutboundDelivery, error) {
	payload, _ := json.Marshal(ev)
	articleID := ev.ArticleID
	delivery := s.newDelivery(domainfed.DeliveryTypeCitation, &articleID, ev.TargetInstance, payload, traceID)
	if err := s.repo.Create(ctx, delivery); err != nil {
		return nil, err
	}
	return s.sendCitation(ctx, delivery, ev)
}

func (s *DeliveryService) DispatchMention(ctx context.Context, ev MentionDetected, traceID *string) (*domainfed.OutboundDelivery, error) {
	payload, _ := json.Marshal(ev)
	articleID := ev.ArticleID
	delivery := s.newDelivery(domainfed.DeliveryTypeMention, &articleID, ev.TargetInstance, payload, traceID)
	if err := s.repo.Create(ctx, delivery); err != nil {
		return nil, err
	}
	return s.sendMention(ctx, delivery, ev)
}

func (s *DeliveryService) Retry(ctx context.Context, id int64) (*domainfed.OutboundDelivery, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	switch item.DeliveryType {
	case domainfed.DeliveryTypeFriendLink:
		var payload contract.FederationFriendLinkRequestReq
		_ = json.Unmarshal(item.Payload, &payload)
		return s.sendFriendLink(ctx, item, payload.Message, payload.RSSURL)
	case domainfed.DeliveryTypeCitation:
		var ev CitationDetected
		if err := json.Unmarshal(item.Payload, &ev); err != nil {
			return nil, err
		}
		return s.sendCitation(ctx, item, ev)
	case domainfed.DeliveryTypeMention:
		var ev MentionDetected
		if err := json.Unmarshal(item.Payload, &ev); err != nil {
			return nil, err
		}
		return s.sendMention(ctx, item, ev)
	default:
		return nil, errors.New("unsupported delivery type")
	}
}

func (s *DeliveryService) HandleCallback(ctx context.Context, cmd CallbackResultCmd) (*domainfed.OutboundDelivery, error) {
	item, err := s.repo.GetByRequestID(ctx, strings.TrimSpace(cmd.RequestID))
	if err != nil {
		return nil, err
	}
	if typ := strings.TrimSpace(cmd.Type); typ != "" && typ != item.DeliveryType {
		return nil, errors.New("callback type mismatch")
	}
	status := normalizeCallbackStatus(cmd.Status)
	if status == "" {
		return nil, errors.New("invalid callback status")
	}
	transitioned := canTransition(item.Status, status)
	if !transitioned {
		return item, nil
	}
	item.Status = status
	if strings.TrimSpace(cmd.RemoteTicketID) != "" {
		val := strings.TrimSpace(cmd.RemoteTicketID)
		item.RemoteTicketID = &val
	}
	if strings.TrimSpace(cmd.Reason) != "" {
		reason := strings.TrimSpace(cmd.Reason)
		item.ErrorMessage = &reason
	}
	now := time.Now().UTC()
	if cmd.ProcessedAt != nil {
		now = cmd.ProcessedAt.UTC()
	}
	item.LastCallbackAt = &now
	item.NextRetryAt = nil
	if err := s.repo.Update(ctx, item); err != nil {
		return nil, err
	}
	s.publishStatus(ctx, item)
	return item, nil
}

func (s *DeliveryService) Get(ctx context.Context, id int64) (*domainfed.OutboundDelivery, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *DeliveryService) List(ctx context.Context, options domainfed.OutboundDeliveryListOptions) ([]domainfed.OutboundDelivery, int64, error) {
	return s.repo.List(ctx, options)
}

func (s *DeliveryService) ProcessRetryQueue(ctx context.Context, limit int) error {
	items, err := s.repo.ListRetryable(ctx, time.Now().UTC(), limit)
	if err != nil {
		return err
	}
	for i := range items {
		_, _ = s.Retry(ctx, items[i].ID)
	}
	return nil
}

func (s *DeliveryService) newDelivery(deliveryType string, sourceArticleID *int64, target string, payload json.RawMessage, traceID *string) *domainfed.OutboundDelivery {
	return &domainfed.OutboundDelivery{
		RequestID:         strings.ReplaceAll(uuid.NewString(), "-", ""),
		DeliveryType:      deliveryType,
		SourceArticleID:   sourceArticleID,
		TargetInstanceURL: strings.TrimSpace(target),
		TargetEndpoint:    "",
		Payload:           payload,
		Status:            domainfed.DeliveryStatusQueued,
		AttemptCount:      0,
		MaxAttempts:       6,
		TraceID:           traceID,
	}
}

func (s *DeliveryService) sendFriendLink(ctx context.Context, item *domainfed.OutboundDelivery, message, rssURL string) (*domainfed.OutboundDelivery, error) {
	if !s.beginSend(ctx, item) {
		return item, nil
	}
	resp, raw, endpoint, err := s.outbound.SendFriendLinkRequest(ctx, item.TargetInstanceURL, message, rssURL)
	return s.finishSend(ctx, item, resp, raw, endpoint, err)
}

func (s *DeliveryService) sendCitation(ctx context.Context, item *domainfed.OutboundDelivery, ev CitationDetected) (*domainfed.OutboundDelivery, error) {
	if !s.beginSend(ctx, item) {
		return item, nil
	}
	ev.RequestID = item.RequestID
	resp, raw, endpoint, err := s.outbound.SendCitation(ctx, ev)
	return s.finishSend(ctx, item, resp, raw, endpoint, err)
}

func (s *DeliveryService) sendMention(ctx context.Context, item *domainfed.OutboundDelivery, ev MentionDetected) (*domainfed.OutboundDelivery, error) {
	if !s.beginSend(ctx, item) {
		return item, nil
	}
	ev.RequestID = item.RequestID
	resp, raw, endpoint, err := s.outbound.SendMention(ctx, ev)
	return s.finishSend(ctx, item, resp, raw, endpoint, err)
}

func (s *DeliveryService) beginSend(ctx context.Context, item *domainfed.OutboundDelivery) bool {
	if item.Status == domainfed.DeliveryStatusDead {
		return false
	}
	item.Status = domainfed.DeliveryStatusSending
	item.AttemptCount++
	if err := s.repo.Update(ctx, item); err != nil {
		return false
	}
	s.publishStatus(ctx, item)
	return true
}

func (s *DeliveryService) finishSend(ctx context.Context, item *domainfed.OutboundDelivery, resp *http.Response, raw []byte, endpoint string, sendErr error) (*domainfed.OutboundDelivery, error) {
	if endpoint != "" {
		item.TargetEndpoint = endpoint
	}
	if sendErr != nil {
		msg := sendErr.Error()
		item.ErrorMessage = &msg
		if isTimeoutErr(sendErr) {
			item.Status = domainfed.DeliveryStatusTimeout
		} else {
			item.Status = domainfed.DeliveryStatusFailed
		}
		scheduleRetry(item)
		_ = s.repo.Update(ctx, item)
		s.publishStatus(ctx, item)
		return item, sendErr
	}

	body := string(raw)
	item.ResponseBody = &body
	if resp != nil {
		item.HTTPStatus = &resp.StatusCode
	}
	if resp != nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
		item.Status = domainfed.DeliveryStatusAccepted
		item.NextRetryAt = nil
		if ticket := parseRemoteTicketID(raw); ticket != "" {
			item.RemoteTicketID = &ticket
		}
	} else {
		item.Status = domainfed.DeliveryStatusFailed
		scheduleRetry(item)
	}
	_ = s.repo.Update(ctx, item)
	s.publishStatus(ctx, item)
	return item, nil
}

func (s *DeliveryService) publishStatus(ctx context.Context, item *domainfed.OutboundDelivery) {
	if item == nil {
		return
	}
	_ = s.events.Publish(ctx, DeliveryStatusChanged{
		DeliveryID:      item.ID,
		RequestID:       item.RequestID,
		DeliveryType:    item.DeliveryType,
		SourceArticleID: item.SourceArticleID,
		Status:          item.Status,
		HTTPStatus:      item.HTTPStatus,
		ErrorMessage:    item.ErrorMessage,
		RemoteTicketID:  item.RemoteTicketID,
		At:              time.Now().UTC(),
	})
}

func scheduleRetry(item *domainfed.OutboundDelivery) {
	if item.AttemptCount >= item.MaxAttempts {
		item.Status = domainfed.DeliveryStatusDead
		item.NextRetryAt = nil
		return
	}
	backoff := retryBackoff(item.AttemptCount)
	at := time.Now().UTC().Add(backoff)
	item.NextRetryAt = &at
}

func retryBackoff(attempt int) time.Duration {
	switch {
	case attempt <= 1:
		return 30 * time.Second
	case attempt == 2:
		return 2 * time.Minute
	case attempt == 3:
		return 10 * time.Minute
	case attempt == 4:
		return 30 * time.Minute
	default:
		return time.Hour
	}
}

func isTimeoutErr(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

func parseRemoteTicketID(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ""
	}
	keys := []string{"applicationId", "citation_id", "mention_id", "id"}
	for _, key := range keys {
		val, ok := payload[key]
		if !ok || val == nil {
			continue
		}
		switch t := val.(type) {
		case string:
			if strings.TrimSpace(t) != "" {
				return strings.TrimSpace(t)
			}
		case float64:
			return fmt.Sprintf("%.0f", t)
		case json.Number:
			return t.String()
		}
	}
	return ""
}

func normalizeCallbackStatus(status string) string {
	switch strings.TrimSpace(strings.ToLower(status)) {
	case domainfed.DeliveryStatusAccepted:
		return domainfed.DeliveryStatusAccepted
	case domainfed.DeliveryStatusApproved:
		return domainfed.DeliveryStatusApproved
	case domainfed.DeliveryStatusRejected:
		return domainfed.DeliveryStatusRejected
	case domainfed.DeliveryStatusFailed:
		return domainfed.DeliveryStatusFailed
	default:
		return ""
	}
}

func canTransition(from, to string) bool {
	if from == to {
		return false
	}
	order := map[string]int{
		domainfed.DeliveryStatusQueued:   0,
		domainfed.DeliveryStatusSending:  1,
		domainfed.DeliveryStatusAccepted: 2,
		domainfed.DeliveryStatusApproved: 3,
		domainfed.DeliveryStatusRejected: 3,
		domainfed.DeliveryStatusFailed:   2,
		domainfed.DeliveryStatusTimeout:  2,
		domainfed.DeliveryStatusDead:     4,
	}
	src, okFrom := order[from]
	dst, okTo := order[to]
	if !okFrom || !okTo {
		return false
	}
	return dst >= src
}
