package event

import "time"

// Generic is a lightweight event for cases where introducing a dedicated event type is unnecessary.
type Generic struct {
	EventName string
	At        time.Time
	Payload   map[string]any
}

func (e Generic) Name() string {
	return e.EventName
}

func (e Generic) OccurredAt() time.Time {
	if e.At.IsZero() {
		return time.Now()
	}
	return e.At
}
