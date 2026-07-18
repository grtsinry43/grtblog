package media

import "sync"

// MutationGate keeps filesystem snapshots consistent with upload mutations.
// Ordinary upload writes take a shared lock; a backup briefly takes the
// exclusive lock while it snapshots the upload tree.
type MutationGate struct{ mu sync.RWMutex }

func NewMutationGate() *MutationGate { return &MutationGate{} }

func (s *Service) beginMutation() func() {
	if s.gate == nil {
		return func() {}
	}
	s.gate.mu.RLock()
	return s.gate.mu.RUnlock
}

func (g *MutationGate) WithMutation(fn func() error) error {
	if g == nil {
		return fn()
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	return fn()
}

func (g *MutationGate) WithSnapshot(fn func() error) error {
	if g == nil {
		return fn()
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	return fn()
}
