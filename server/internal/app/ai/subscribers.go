package ai

import (
	"context"
	"log"

	appcomment "github.com/grtsinry43/grtblog-v2/server/internal/app/comment"
	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
)

// CommentStatusUpdater is a minimal interface for updating comment moderation status.
// Implemented by comment.Service or a simple adapter wrapping the comment repository.
type CommentStatusUpdater interface {
	UpdateCommentStatus(ctx context.Context, cmd appcomment.UpdateCommentStatusCmd) error
}

type subscriberHandlerFunc func(ctx context.Context, event appEvent.Event) error

func (h subscriberHandlerFunc) Handle(ctx context.Context, event appEvent.Event) error {
	return h(ctx, event)
}

// RegisterSubscribers wires event-driven AI features.
// Currently subscribes to comment.created for auto-moderation.
func RegisterSubscribers(bus appEvent.Bus, aiSvc *Service, commentUpdater CommentStatusUpdater) {
	if bus == nil || aiSvc == nil || commentUpdater == nil {
		return
	}

	bus.Subscribe(appcomment.CommentCreated{}.Name(), subscriberHandlerFunc(func(ctx context.Context, event appEvent.Event) error {
		payload, ok := event.(appcomment.CommentCreated)
		if !ok {
			return nil
		}

		// Only moderate comments that are pending
		if payload.Status != "pending" {
			return nil
		}

		// Check if AI is enabled
		enabled, _ := aiSvc.cfgGet.GetConfigValue(ctx, "ai.enabled")
		if enabled != "true" {
			return nil
		}

		// Check if a moderation model is assigned
		modelID, _ := aiSvc.cfgGet.GetConfigValue(ctx, taskKeyCommentModeration)
		if modelID == "" || modelID == "0" {
			return nil
		}

		// Call AI moderation
		result, err := aiSvc.ModerateComment(ctx, payload.Content, "auto")
		if err != nil {
			log.Printf("[AI] moderate comment #%d failed: %v", payload.ID, err)
			return nil // Don't block the event chain; comment stays pending
		}

		newStatus := "approved"
		if !result.Approved {
			newStatus = "rejected"
		}

		log.Printf("[AI] comment #%d moderated: %s (score=%.2f, reason=%s)", payload.ID, newStatus, result.Score, result.Reason)

		if err := commentUpdater.UpdateCommentStatus(ctx, appcomment.UpdateCommentStatusCmd{
			ID:     payload.ID,
			Status: newStatus,
		}); err != nil {
			log.Printf("[AI] update comment #%d status failed: %v", payload.ID, err)
		}
		return nil
	}))
}
