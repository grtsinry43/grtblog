package health

import "time"

const EventNameStateChanged = "system.health.changed"

// StateChanged is published whenever the aggregate health state changes.
type StateChanged struct {
	Prev      uint8
	Next      uint8
	Snapshot  HealthSnapshot
	At        time.Time
}

func (e StateChanged) Name() string        { return EventNameStateChanged }
func (e StateChanged) OccurredAt() time.Time { return e.At }
