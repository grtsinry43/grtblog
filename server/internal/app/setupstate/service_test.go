package setupstate

import (
	"context"
	"errors"
	"testing"
)

func TestCompleteUpgradeGuideRejectsUnknownVersion(t *testing.T) {
	svc := &Service{}
	err := svc.CompleteUpgradeGuide(context.Background(), "../../arbitrary")
	if !errors.Is(err, ErrUnknownUpgradeGuide) {
		t.Fatalf("expected ErrUnknownUpgradeGuide, got %v", err)
	}
}
