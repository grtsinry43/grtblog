package email

import (
	"context"
	"encoding/json"
	"math"
	"time"

	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	domainconfig "github.com/grtsinry43/grtblog-v2/server/internal/domain/config"
	domainemail "github.com/grtsinry43/grtblog-v2/server/internal/domain/email"
)

type Dispatcher struct {
	repo        domainemail.Repository
	sender      *Sender
	websiteInfo domainconfig.WebsiteInfoRepository
	queue       chan *domainemail.Outbox
	maxRetries  int
}

func NewDispatcher(repo domainemail.Repository, sender *Sender, websiteInfo domainconfig.WebsiteInfoRepository, workers int, queueSize int, maxRetries int, pollInterval time.Duration) *Dispatcher {
	if workers <= 0 {
		workers = 1
	}
	if queueSize <= 0 {
		queueSize = 1
	}
	if maxRetries <= 0 {
		maxRetries = 1
	}
	if pollInterval <= 0 {
		pollInterval = 2 * time.Second
	}
	d := &Dispatcher{
		repo:        repo,
		sender:      sender,
		websiteInfo: websiteInfo,
		queue:       make(chan *domainemail.Outbox, queueSize),
		maxRetries:  maxRetries,
	}
	for i := 0; i < workers; i++ {
		go d.worker()
	}
	go d.pollLoop(pollInterval, workers*2)
	return d
}

func (d *Dispatcher) Handle(ctx context.Context, event appEvent.Event) error {
	if event == nil {
		return nil
	}
	templates, err := d.repo.ListEnabledTemplatesByEvent(ctx, event.Name())
	if err != nil {
		return err
	}
	if len(templates) == 0 {
		return nil
	}
	variables := appEvent.BuildGlobalTemplateVariables(ctx, d.websiteInfo)
	for key, value := range mapFromEvent(event) {
		variables[key] = value
	}
	subscribers, err := d.repo.ListActiveSubscriberEmailsByEvent(ctx, event.Name())
	if err != nil {
		return err
	}
	now := time.Now()
	for _, tpl := range templates {
		rendered, renderErr := RenderTemplate(tpl, variables)
		recipients := normalizeRecipients(append(append([]string{}, tpl.ToEmails...), subscribers...))
		item := &domainemail.Outbox{
			TemplateID:   &tpl.ID,
			TemplateCode: tpl.Code,
			EventName:    event.Name(),
			ToEmails:     recipients,
			NextRetryAt:  now,
		}
		if renderErr != nil {
			item.Status = domainemail.OutboxStatusFailed
			item.RetryCount = d.maxRetries
			item.LastError = renderErr.Error()
		} else {
			item.Status = domainemail.OutboxStatusPending
			item.Subject = rendered.Subject
			item.HTMLBody = rendered.HTMLBody
			item.TextBody = rendered.TextBody
		}
		_ = d.repo.CreateOutbox(context.Background(), item)
	}
	return nil
}

func (d *Dispatcher) pollLoop(interval time.Duration, batchSize int) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		items, err := d.repo.ClaimDueOutbox(context.Background(), batchSize, time.Now(), d.maxRetries)
		if err != nil || len(items) == 0 {
			continue
		}
		for _, item := range items {
			d.queue <- item
		}
	}
}

func (d *Dispatcher) worker() {
	for item := range d.queue {
		err := d.sender.Send(context.Background(), Message{
			To:       item.ToEmails,
			Subject:  item.Subject,
			HTMLBody: item.HTMLBody,
			TextBody: item.TextBody,
		})
		if err == nil {
			_ = d.repo.MarkOutboxSent(context.Background(), item.ID, time.Now())
			continue
		}
		nextRetryCount := item.RetryCount + 1
		backoffMinutes := 1 << min(nextRetryCount, 6)
		delayMinutes := int(math.Min(float64(backoffMinutes), 60))
		nextRetryAt := time.Now().Add(time.Duration(delayMinutes) * time.Minute)
		_ = d.repo.MarkOutboxFailed(context.Background(), item.ID, nextRetryCount, nextRetryAt, err.Error())
	}
}

func RegisterSubscribers(bus appEvent.Bus, handler appEvent.Handler) {
	if bus == nil || handler == nil {
		return
	}
	for _, name := range AvailableEventNames {
		bus.Subscribe(name, handler)
	}
}

func mapFromEvent(event appEvent.Event) map[string]any {
	result := map[string]any{
		"eventName":  event.Name(),
		"occurredAt": event.OccurredAt().Format(time.RFC3339),
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return result
	}
	var object map[string]any
	if err := json.Unmarshal(payload, &object); err != nil {
		return result
	}
	for key, value := range object {
		result[key] = value
	}
	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
