package health

import (
	"sync"
	"time"
)

const (
	BitNginx       = 5 // MSB
	BitBackend     = 4
	BitDatabase    = 3
	BitRedis       = 2
	BitRenderer    = 1
	BitMaintenance = 0 // LSB: 1 = normal, 0 = maintenance active

	HealthyValue = 0b111111 // 63
)

type SystemMode string

const (
	ModeHealthy     SystemMode = "healthy"     // 63
	ModeMaintenance SystemMode = "maintenance" // bit 0 = 0
	ModeDegraded    SystemMode = "degraded"    // minor components down (Redis/Renderer)
	ModeCritical    SystemMode = "critical"    // Database down
	ModeOutage      SystemMode = "outage"      // Backend or Nginx down
)

// HealthSnapshot is a serialisable snapshot of the current state.
type HealthSnapshot struct {
	HealthBits  uint8           `json:"healthBits"`
	Maintenance bool            `json:"maintenance"`
	Mode        SystemMode      `json:"mode"`
	Components  map[string]bool `json:"components"`
	IsDev       bool            `json:"isDev"`
	Timestamp   time.Time       `json:"timestamp"`
}

// State is the central, concurrency-safe health state machine.
type State struct {
	mu        sync.RWMutex
	value     uint8
	isDev     bool
	updatedAt time.Time
}

// NewState creates a new State.
// All bits start healthy (111111); the checker clears bits when failures are
// detected or when manual maintenance is enabled (bit 0 → 0).
func NewState(isDev bool) *State {
	return &State{
		value:     HealthyValue,
		isDev:     isDev,
		updatedAt: time.Now(),
	}
}

// DeriveMode maps a 6-bit value to a SystemMode.
//
//	Bit 0 = 0         → maintenance  (admin enabled maintenance)
//	63     (111111)   → healthy      all components up, maintenance off
//	top-5 >= 11100    → degraded     Redis / Renderer down
//	top-5 >= 11000    → critical     Database down
//	else              → outage       Backend or Nginx down
func DeriveMode(value uint8) SystemMode {
	// Bit 0 clear = maintenance mode active.
	if value&1 == 0 {
		return ModeMaintenance
	}
	if value == HealthyValue {
		return ModeHealthy
	}
	top5 := value >> 1 // bits 5-1
	switch {
	case top5 >= 0b11100: // 28: nginx + backend + DB all up
		return ModeDegraded
	case top5 >= 0b11000: // 24: nginx + backend up, DB down
		return ModeCritical
	default:
		return ModeOutage
	}
}

// SetBit sets or clears a single component bit. Returns (prev, next) values.
func (s *State) SetBit(bit int, healthy bool) (prev, next uint8) {
	if bit < 0 || bit > 5 {
		s.mu.RLock()
		v := s.value
		s.mu.RUnlock()
		return v, v
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	prev = s.value
	if healthy {
		s.value |= 1 << uint(bit)
	} else {
		s.value &^= 1 << uint(bit)
	}
	next = s.value
	if prev != next {
		s.updatedAt = time.Now()
	}
	return prev, next
}

// SetMaintenance sets bit 0: on=true clears it (maintenance active),
// on=false sets it (normal). Returns whether the value changed.
func (s *State) SetMaintenance(on bool) bool {
	prev, next := s.SetBit(BitMaintenance, !on)
	return prev != next
}

// Snapshot returns a read-consistent copy of the current state.
func (s *State) Snapshot() HealthSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return HealthSnapshot{
		HealthBits:  s.value,
		Maintenance: s.value&1 == 0,
		Mode:        DeriveMode(s.value),
		Components:  componentsFromBits(s.value),
		IsDev:       s.isDev,
		Timestamp:   s.updatedAt,
	}
}

// Value returns the current 6-bit value.
func (s *State) Value() uint8 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.value
}

// Mode returns the derived mode.
func (s *State) Mode() SystemMode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return DeriveMode(s.value)
}

func componentsFromBits(v uint8) map[string]bool {
	return map[string]bool{
		"nginx":    v&(1<<BitNginx) != 0,
		"backend":  v&(1<<BitBackend) != 0,
		"database": v&(1<<BitDatabase) != 0,
		"redis":    v&(1<<BitRedis) != 0,
		"renderer": v&(1<<BitRenderer) != 0,
	}
}
