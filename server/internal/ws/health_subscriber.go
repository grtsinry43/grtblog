package ws

import (
	"context"
	"encoding/json"
	"time"

	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/health"
)

// RegisterHealthSubscriber listens for health state changes and broadcasts
// them to all clients in the realtime lobby room.
func RegisterHealthSubscriber(bus appEvent.Bus, manager *Manager) {
	if bus == nil || manager == nil {
		return
	}
	bus.Subscribe(health.EventNameStateChanged, handlerFunc(func(ctx context.Context, event appEvent.Event) error {
		changed, ok := event.(health.StateChanged)
		if !ok {
			return nil
		}
		snap := changed.Snapshot
		payload := map[string]any{
			"type":        "system.health.state",
			"healthBits":  snap.HealthBits,
			"maintenance": snap.Maintenance,
			"mode":        snap.Mode,
			"components":  snap.Components,
			"isDev":       snap.IsDev,
			"timestamp":   snap.Timestamp.UTC().Format(time.RFC3339),
		}
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		manager.Broadcast(RealtimeRoomKey(), data)
		return nil
	}))
}
