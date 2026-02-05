package federation

import (
	"context"

	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
)

type handlerFunc func(ctx context.Context, event appEvent.Event) error

func (h handlerFunc) Handle(ctx context.Context, event appEvent.Event) error {
	return h(ctx, event)
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
		_, err := svc.DispatchMention(ctx, payload, nil)
		return err
	}))
	bus.Subscribe(CitationDetected{}.Name(), handlerFunc(func(ctx context.Context, event appEvent.Event) error {
		payload, ok := event.(CitationDetected)
		if !ok {
			return nil
		}
		_, err := svc.DispatchCitation(ctx, payload, nil)
		return err
	}))
}
