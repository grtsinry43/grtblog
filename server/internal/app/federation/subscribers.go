package federation

import (
	"context"
	"errors"
	"log"

	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	domainfed "github.com/grtsinry43/grtblog-v2/server/internal/domain/federation"
)

type handlerFunc func(ctx context.Context, event appEvent.Event) error

func (h handlerFunc) Handle(ctx context.Context, event appEvent.Event) error {
	return h(ctx, event)
}

// dispatchResult normalizes dispatch outcomes for the article signal registry:
//   - delivery record created (even if the first send failed): success — the
//     retry worker owns it from here, so the signal must be marked delivered;
//   - self-target: swallow with a log so the marker isn't re-fired on every save;
//   - record creation failed: propagate so the signal stays pending.
func dispatchResult(delivery *domainfed.OutboundDelivery, err error, kind string) error {
	if delivery != nil {
		return nil
	}
	if errors.Is(err, ErrSelfTarget) {
		log.Printf("[federation] 跳过指向本实例的%s信号", kind)
		return nil
	}
	return err
}

// RegisterSubscribers wires federation outbound handlers to the event bus.
func RegisterSubscribers(bus appEvent.Bus, svc *DeliveryService) {
	if bus == nil || svc == nil {
		return
	}
	bus.Subscribe(MentionDetected{}.Name(), handlerFunc(func(ctx context.Context, event appEvent.Event) error {
		payload, ok := event.(MentionDetected)
		if !ok {
			return nil
		}
		delivery, err := svc.DispatchMention(ctx, payload, nil)
		return dispatchResult(delivery, err, "提及")
	}))
	bus.Subscribe(CitationDetected{}.Name(), handlerFunc(func(ctx context.Context, event appEvent.Event) error {
		payload, ok := event.(CitationDetected)
		if !ok {
			return nil
		}
		delivery, err := svc.DispatchCitation(ctx, payload, nil)
		return dispatchResult(delivery, err, "引用")
	}))
}
