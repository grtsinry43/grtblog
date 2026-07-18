package media

import (
	"sync"
	"testing"
	"time"
)

func TestMutationGateWaitsForMutationBeforeSnapshot(t *testing.T) {
	t.Parallel()
	gate := NewMutationGate()
	mutationStarted := make(chan struct{})
	releaseMutation := make(chan struct{})
	mutationDone := make(chan struct{})
	go func() {
		_ = gate.WithMutation(func() error {
			close(mutationStarted)
			<-releaseMutation
			return nil
		})
		close(mutationDone)
	}()
	<-mutationStarted

	var mu sync.Mutex
	snapshotRan := false
	snapshotDone := make(chan struct{})
	go func() {
		_ = gate.WithSnapshot(func() error {
			mu.Lock()
			snapshotRan = true
			mu.Unlock()
			return nil
		})
		close(snapshotDone)
	}()

	select {
	case <-snapshotDone:
		t.Fatal("snapshot ran before the active mutation finished")
	case <-time.After(20 * time.Millisecond):
	}
	close(releaseMutation)
	<-mutationDone
	<-snapshotDone
	mu.Lock()
	defer mu.Unlock()
	if !snapshotRan {
		t.Fatal("snapshot did not run")
	}
}
